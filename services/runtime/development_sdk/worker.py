"""One pinned SDK Agent, private stdin handoff, bounded stdout metadata protocol.

No provider client, default execution tool, workspace import or model fallback.
SDK conversation state is memory-only. Run only under the trusted Go supervisor.
"""
from __future__ import annotations

import contextlib
import importlib.metadata
import ipaddress
import json
import os
import re
import signal
import stat
import sys
import threading
import urllib.parse
import urllib.request

MAX_BYTES = 262144
PROTOCOL_OUT = sys.stdout
ACTION_FIELDS = {"schema_version", "action", "path", "expected_digest", "content", "candidate_digest"}
INPUT_FIELDS = {"schema_version", "attempt_id", "executor_instance_id", "channel_url", "credential", "model", "initial_messages", "tools", "max_output_tokens", "max_model_calls"}


class WorkerFailure(RuntimeError):
    pass


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise WorkerFailure("duplicate field")
        result[key] = value
    return result


def decode(data):
    if len(data) > MAX_BYTES:
        raise WorkerFailure("message limit")
    value = json.loads(data, object_pairs_hook=unique_object,
                      parse_constant=lambda _value: (_ for _ in ()).throw(WorkerFailure("non-finite number")))
    reject_null(value)
    return value


def reject_null(value):
    if value is None:
        raise WorkerFailure("null field")
    if isinstance(value, dict):
        for item in value.values():
            reject_null(item)
    elif isinstance(value, list):
        for item in value:
            reject_null(item)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *_args, **_kwargs):
        raise WorkerFailure("channel redirects refused")


def read_line():
    data = sys.stdin.buffer.readline(MAX_BYTES + 2)
    if not data or not data.endswith(b"\n") or len(data) > MAX_BYTES:
        raise WorkerFailure("private input unavailable")
    return decode(data)


def read_input():
    value = read_line()
    if not isinstance(value, dict) or set(value) != INPUT_FIELDS or value["schema_version"] != "development_worker_input.v1":
        raise WorkerFailure("input contract")
    for key in ("attempt_id", "executor_instance_id"):
        if not isinstance(value[key], str) or not re.fullmatch(r"[a-z][a-z0-9_-]{2,63}", value[key]):
            raise WorkerFailure("input identity")
    url = urllib.parse.urlsplit(value["channel_url"])
    if url.scheme != "http" or url.username or url.password or url.path or url.query or url.fragment or not url.port or not ipaddress.ip_address(url.hostname).is_loopback:
        raise WorkerFailure("loopback channel required")
    if not re.fullmatch(r"devtask_[A-Za-z0-9_-]{43}", value["credential"]):
        raise WorkerFailure("credential contract")
    messages = value["initial_messages"]
    if not isinstance(messages, list) or len(messages) != 2:
        raise WorkerFailure("initial messages")
    for message, role in zip(messages, ("system", "user")):
        if not isinstance(message, dict) or set(message) != {"role", "content"} or message["role"] != role or not isinstance(message["content"], str) or not message["content"] or len(message["content"].encode()) > 65536:
            raise WorkerFailure("initial message contract")
    if type(value["max_model_calls"]) is not int or not 1 <= value["max_model_calls"] <= 8 or type(value["max_output_tokens"]) is not int or not 1 <= value["max_output_tokens"] <= 262144:
        raise WorkerFailure("worker limits")
    tools = value["tools"]
    if not isinstance(tools, list) or len(tools) != 1 or tools[0].get("type") != "function" or tools[0].get("function", {}).get("name") != "development_tool":
        raise WorkerFailure("tool catalog")
    return value


class Bridge:
    def __init__(self, config):
        self.config = config
        self.sequence = 0
        self.model_count = 0
        self.last_model = ""
        self.pending = None
        self.stage = "sdk_runtime"
        self.lock = threading.RLock()
        self.http = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())

    def emit(self, kind, *, event_type="", request_id="", call=None, error_code=None):
        with self.lock:
            self.sequence += 1
            if self.sequence > 128:
                raise WorkerFailure("event limit")
            frame = dict(schema_version="development_worker_message.v1", sequence=self.sequence,
                         kind=kind, event_type=event_type, request_id=request_id)
            if call is not None:
                frame["call"] = call
            if error_code is not None:
                frame["error_code"] = error_code
            PROTOCOL_OUT.write(json.dumps(frame, ensure_ascii=False, separators=(",", ":")) + "\n")
            PROTOCOL_OUT.flush()
            expected = {"ready": "worker_ready", "event": "sdk_event", "finished": "worker_finished"}.get(kind)
            if expected:
                reply = read_line()
                if not isinstance(reply, dict) or set(reply) != {"schema_version", "sequence", "request_id", "observation"} or reply != dict(schema_version="development_worker_reply.v1", sequence=self.sequence, request_id=expected, observation="") or type(reply["sequence"]) is not int:
                    raise WorkerFailure("event commit acknowledgement")
            return self.sequence

    def request(self, method, path, body=None, request_id=""):
        url = self.config["channel_url"] + "/internal/development/v1/attempts/" + self.config["attempt_id"] + path
        data = None if body is None else json.dumps(body, ensure_ascii=False, separators=(",", ":")).encode()
        if data is not None and len(data) > MAX_BYTES:
            raise WorkerFailure("request limit")
        headers = {"Authorization": "Bearer " + self.config["credential"], "Content-Type": "application/json"}
        if request_id:
            headers["X-RadishMind-Development-Request-ID"] = request_id
        request = urllib.request.Request(url, data=data, headers=headers, method=method)
        with self.http.open(request, timeout=120) as response:
            return decode(response.read(MAX_BYTES + 1))

    def observation(self, action):
        self.stage = "tool_exchange"
        with self.lock:
            call = self.pending
            if call is None or call["arguments"] != action:
                raise WorkerFailure("unmatched tool action")
            self.request("POST", "/tool-calls", call)
            sequence = self.emit("tool", request_id=call["request_id"], call=call)
            reply = read_line()
            if not isinstance(reply, dict) or set(reply) != {"schema_version", "sequence", "request_id", "observation"} or reply["schema_version"] != "development_worker_reply.v1" or type(reply["sequence"]) is not int or reply["sequence"] != sequence or reply["request_id"] != call["request_id"] or not isinstance(reply["observation"], str) or len(reply["observation"].encode()) > 65536:
                raise WorkerFailure("trusted reply contract")
            self.pending = None
            return reply["observation"]

    def event(self, event):
        name = type(event).__name__
        if name == "MessageEvent":
            role = event.llm_message.role
            if role not in ("user", "assistant"):
                raise WorkerFailure("unexpected SDK message")
            kind = role + "_message"
        else:
            kinds = {"SystemPromptEvent": "system_prompt", "ActionEvent": "action",
                     "ObservationEvent": "observation", "AgentErrorEvent": "error",
                     "ConversationErrorEvent": "error", "SecurityAnalysisEvent": "security"}
            if name not in kinds:
                raise WorkerFailure("unsupported SDK event")
            kind = kinds[name]
        code = None
        if kind == "error":
            exception = sys.exception()
            code = self.stage if isinstance(exception, WorkerFailure) else {
                "TypeError": "sdk_type_error", "AttributeError": "sdk_attribute_error",
                "ValidationError": "sdk_validation_error", "HTTPError": "model_channel",
            }.get(type(exception).__name__, "sdk_runtime")
        self.emit("event", event_type=kind, request_id=self.last_model, error_code=code)
        if kind == "error":
            raise WorkerFailure("SDK reported failure")


def validate_sdk_distribution():
    version = importlib.metadata.version("openhands-sdk")
    if version not in ("1.59.0", "1.59.0+radishmind.1"):
        raise WorkerFailure("SDK version mismatch")
    if version == "1.59.0+radishmind.1":
        for package in ("lmnr", "lmnr-claude-code-proxy"):
            try:
                importlib.metadata.distribution(package)
            except importlib.metadata.PackageNotFoundError:
                continue
            raise WorkerFailure("unexpected optional SDK distribution")


def build_agent(bridge):
    # The only dependency boundary. Host-side wire tests import no SDK packages.
    validate_sdk_distribution()
    from litellm.types.utils import ModelResponse
    from pydantic import ConfigDict
    from typing import Literal
    from openhands.sdk import Agent, LLM
    from openhands.sdk.llm import LLMResponse, Message, TextContent
    from openhands.sdk.llm.utils.metrics import MetricsSnapshot
    from openhands.sdk.tool import Action, Observation, Tool, ToolDefinition, ToolExecutor, register_tool

    class DevelopmentAction(Action):
        model_config = ConfigDict(extra="forbid", strict=True)
        schema_version: Literal["development_tool_action.v1"]
        action: Literal["read_file", "replace_file", "run_tests"]
        path: str
        expected_digest: str
        content: str
        candidate_digest: str

    class DevelopmentObservation(Observation):
        pass

    class RemoteExecutor(ToolExecutor):
        def __call__(self, action, conversation=None):
            args = action.model_dump(include=ACTION_FIELDS)
            return DevelopmentObservation(content=[TextContent(text=bridge.observation(args))])

    class DevelopmentTool(ToolDefinition):
        name = "development_tool"

        @classmethod
        def create(cls, *_args, **_kwargs):
            return [cls(description="Request a bounded development action through the task controller.",
                        action_type=DevelopmentAction, observation_type=DevelopmentObservation,
                        executor=RemoteExecutor())]

    class ControlledLLM(LLM):
        # Explicit transport extension: no base completion/responses client is used.
        def resolve_runtime_metadata(self, *, force=False):
            return None

        def generate(self, messages, tools=None, include=None, store=None,
                     add_security_risk_prediction=False, on_token=None, call_context=None, **kwargs):
            bridge.stage = "generate_options"
            if kwargs or include or store or on_token is not None or tools is None or [t.name for t in tools] != ["development_tool"]:
                raise WorkerFailure("unsupported SDK request")
            bridge.model_count += 1
            if bridge.model_count > bridge.config["max_model_calls"]:
                raise WorkerFailure("model call limit")
            bridge.stage = "message_serialization"
            native = []
            for message in messages:
                if message.reasoning_content or message.thinking_blocks or message.responses_reasoning_item or any(type(c).__name__ != "TextContent" for c in message.content):
                    raise WorkerFailure("unsupported message content")
                native.append(message.to_chat_dict(cache_enabled=False, vision_enabled=False, function_calling_enabled=True, force_string_serializer=True, send_reasoning_content=False))
            bridge.last_model = "model_" + str(bridge.model_count)
            bridge.stage = "model_channel"
            response = bridge.request("POST", "/chat/completions", {
                "model": bridge.config["model"], "messages": native,
                "max_completion_tokens": bridge.config["max_output_tokens"], "tools": bridge.config["tools"],
            }, bridge.last_model)
            bridge.stage = "response_conversion"
            raw = ModelResponse(**response)
            result = Message.from_llm_chat_message(raw.choices[0].message)
            if result.tool_calls:
                if len(result.tool_calls) != 1 or result.tool_calls[0].name != "development_tool":
                    raise WorkerFailure("unexpected tool call")
                call = result.tool_calls[0]
                args = decode(call.arguments.encode())
                if not isinstance(args, dict) or set(args) != ACTION_FIELDS:
                    raise WorkerFailure("tool argument contract")
                offer = bridge.request("GET", "/tool-offers/" + urllib.parse.quote(call.id, safe=""))
                bridge.pending = dict(schema_version="development_tool_call.v1", attempt_id=bridge.config["attempt_id"],
                                      request_id=offer["request_id"], parent_request_id=bridge.last_model,
                                      tool_call_id=call.id, arguments=args)
            # SDK-local metrics never settle Go's authoritative budget/usage.
            return LLMResponse(message=result, metrics=MetricsSnapshot(model_name=bridge.config["model"]), raw_response=raw)

        def completion(self, *_args, **_kwargs):
            raise WorkerFailure("completion bypass refused")

        def responses(self, *_args, **_kwargs):
            raise WorkerFailure("responses bypass refused")

        async def agenerate(self, *_args, **_kwargs):
            raise WorkerFailure("async bypass refused")

        async def acompletion(self, *_args, **_kwargs):
            raise WorkerFailure("async completion bypass refused")

        async def aresponses(self, *_args, **_kwargs):
            raise WorkerFailure("async responses bypass refused")

    class ControlledAgent(Agent):
        def get_dynamic_context(self, state):
            # All prompt context must be part of the host-confirmed first messages.
            return None

    register_tool("development_tool", DevelopmentTool)
    llm = ControlledLLM(model="openai/gpt-4o", num_retries=0, timeout=120, stream=False,
                        max_output_tokens=bridge.config["max_output_tokens"], fallback_strategy=None,
                        native_tool_calling=True, disable_vision=True, caching_prompt=False,
                        log_completions=False, reasoning_effort=None)
    agent = ControlledAgent(llm=llm, tools=[Tool(name="development_tool")], include_default_tools=[],
                            mcp_config={}, condenser=None, critic=None, agent_context=None,
                            system_prompt=bridge.config["initial_messages"][0]["content"], tool_concurrency_limit=1)
    return agent


def watch_parent():
    # FD 3 is a dedicated Go-owned liveness pipe, never SDK input or a secret.
    # Each worker has its own OS process group; parent death stops that group.
    if os.getpgrp() != os.getpid() or not stat.S_ISFIFO(os.fstat(3).st_mode):
        raise WorkerFailure("owned liveness pipe required")
    os.set_inheritable(3, False)
    def wait():
        try:
            os.read(3, 1)
        finally:
            os.killpg(os.getpgrp(), signal.SIGKILL)
    threading.Thread(target=wait, daemon=True, name="parent-liveness").start()


def main():
    bridge = None
    stage = "input"
    try:
        watch_parent()
        config = read_input()
        bridge = Bridge(config)
        # Keep all SDK banners, logs and progress off the metadata pipe.
        with contextlib.redirect_stdout(sys.stderr):
            stage = "agent_setup"
            agent = build_agent(bridge)
            from openhands.sdk.conversation import LocalConversation
            from openhands.sdk.io import InMemoryFileStore
            stage = "conversation_setup"
            conversation = LocalConversation(agent=agent, workspace=os.getcwd(), plugins=[],
                file_store=InMemoryFileStore(), callbacks=[bridge.event], visualizer=None,
                max_iteration_per_run=config["max_model_calls"], stuck_detection=False)
            try:
                bridge.emit("ready")
                stage = "run"
                conversation.send_message(config["initial_messages"][1]["content"])
                conversation.run()
                if conversation.state.execution_status.value != "finished":
                    raise WorkerFailure("SDK did not finish")
                stage = "close"
            finally:
                conversation.close()
        bridge.emit("finished")
        return 0
    except BaseException:
        if bridge is not None:
            try:
                bridge.emit("failed", event_type=stage)
            except BaseException:
                pass
        return 1


if __name__ == "__main__":
    sys.exit(main())
