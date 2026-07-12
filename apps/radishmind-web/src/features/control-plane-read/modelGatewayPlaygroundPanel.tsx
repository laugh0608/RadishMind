import { useEffect, useMemo, useRef, useState } from "react";

import {
  createGatewayPlaygroundRequestId,
  initialModelGatewayPlaygroundResult,
  modelGatewayPlaygroundConfigForApplication,
  readModelGatewayPlaygroundConfig,
  submitModelGatewayPlaygroundRequest,
  type ModelGatewayPlaygroundProtocol,
} from "./modelGatewayPlaygroundConsumer.ts";
import {
  MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
  requestGatewayRequestHistoryReview,
  type ModelGatewayPlaygroundHandoffEventDetail,
} from "./modelGatewayPlaygroundEvents.ts";

const baseConfig = readModelGatewayPlaygroundConfig();
const DEFAULT_INPUT = "请用简洁的中文说明 RadishMind Gateway 当前请求的处理结果。";

export default function ModelGatewayPlaygroundPanel() {
  const [applicationId, setApplicationId] = useState(baseConfig.applicationId);
  const [protocol, setProtocol] = useState<ModelGatewayPlaygroundProtocol>("chat_completions");
  const [model, setModel] = useState(baseConfig.defaultModel);
  const [inputText, setInputText] = useState(DEFAULT_INPUT);
  const [stream, setStream] = useState(false);
  const [result, setResult] = useState(() => initialModelGatewayPlaygroundResult(baseConfig));
  const activeController = useRef<AbortController | null>(null);
  const config = useMemo(
    () => modelGatewayPlaygroundConfigForApplication(baseConfig, applicationId),
    [applicationId],
  );

  useEffect(() => {
    function receiveApplicationHandoff(event: Event) {
      const detail = (event as CustomEvent<ModelGatewayPlaygroundHandoffEventDetail>).detail;
      if (!detail?.applicationId || !detail.model) return;
      activeController.current?.abort();
      activeController.current = null;
      const nextConfig = modelGatewayPlaygroundConfigForApplication(baseConfig, detail.applicationId);
      setApplicationId(detail.applicationId);
      setProtocol(detail.protocol);
      setModel(detail.model);
      setInputText(DEFAULT_INPUT);
      setStream(false);
      setResult(initialModelGatewayPlaygroundResult(nextConfig));
    }
    window.addEventListener(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, receiveApplicationHandoff);
    return () => window.removeEventListener(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, receiveApplicationHandoff);
  }, []);

  async function submit() {
    const controller = new AbortController();
    activeController.current = controller;
    const requestId = createGatewayPlaygroundRequestId();
    setResult({
      status: "submitting", requestId, route: "", protocol, stream, outputText: "", httpStatus: 0,
      failureCode: "", failureBoundary: "", summary: stream ? "Gateway stream is in progress." : "Gateway request is in progress.",
      historyReviewAvailable: false,
    });
    const next = await submitModelGatewayPlaygroundRequest(
      config,
      { protocol, model, inputText, stream, requestId },
      controller.signal,
      (outputText) => setResult((current) => ({ ...current, outputText })),
    );
    if (activeController.current !== controller) return;
    activeController.current = null;
    setResult(next);
  }

  function cancel() {
    activeController.current?.abort();
  }

  function reviewHistory() {
    requestGatewayRequestHistoryReview(result.requestId, applicationId);
    window.location.hash = "model-gateway-request-history";
  }

  const enabled = config.mode === "dev_gateway_playground_http";
  return (
    <section className="surface-band model-gateway-overview gateway-playground" id="model-gateway-playground" aria-labelledby="model-gateway-playground-title">
      <div className="section-heading">
        <div><p className="eyebrow">Model Gateway</p><h3 id="model-gateway-playground-title">Playground and request review</h3></div>
        <span className={`status-badge ${enabled ? "good" : "neutral"}`}>{enabled ? "dev/test interactive" : "offline"}</span>
      </div>
      {!enabled ? (
        <article className="model-gateway-overview-hero">
          <div><p className="eyebrow">Offline boundary</p><h4>No northbound request is sent</h4><p>Enable the explicit Gateway Playground and Request History dev/test source to run a request. Production keys, quota, billing, retry, fallback, and persistence remain disabled.</p></div>
        </article>
      ) : (
        <div className="gateway-playground-layout">
          <form className="gateway-playground-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
            <p className="gateway-playground-scope"><strong>Application scope</strong><code>{applicationId || "unbound"}</code></p>
            <label>Protocol<select value={protocol} onChange={(event) => setProtocol(event.target.value as ModelGatewayPlaygroundProtocol)} disabled={result.status === "submitting"}><option value="chat_completions">Chat Completions</option><option value="responses">Responses</option><option value="messages">Messages</option></select></label>
            <label>Model<input value={model} onChange={(event) => setModel(event.target.value)} maxLength={160} disabled={result.status === "submitting"} /></label>
            <label className="gateway-playground-input">Temporary input<textarea value={inputText} onChange={(event) => setInputText(event.target.value)} maxLength={8000} rows={7} disabled={result.status === "submitting"} /></label>
            <label className="gateway-playground-stream"><input type="checkbox" checked={stream} onChange={(event) => setStream(event.target.checked)} disabled={result.status === "submitting"} /> Stream response</label>
            <div className="gateway-playground-actions">
              <button type="submit" disabled={result.status === "submitting"}>Send request</button>
              <button type="button" onClick={cancel} disabled={result.status !== "submitting"}>Cancel</button>
            </div>
            <p className="boundary-note">Input and output stay in this component and the active HTTP request. They are not saved to Request History or browser storage.</p>
          </form>
          <article className="gateway-playground-result" aria-live="polite">
            <div className="model-gateway-overview-row-main">
              <div><p className="eyebrow">Current result</p><h4>{result.requestId || "No request yet"}</h4></div>
              <span className={`status-badge ${result.status === "succeeded" ? "good" : result.status === "failed" || result.status === "canceled" ? "bad" : "neutral"}`}>{result.status}</span>
            </div>
            <p>{result.summary}</p>
            {result.outputText ? <pre className="gateway-playground-output">{result.outputText}</pre> : null}
            <dl className="model-gateway-overview-meta">
              <div><dt>Route</dt><dd>{result.route || "not sent"}</dd></div>
              <div><dt>Mode</dt><dd>{result.stream ? "stream" : "unary"}</dd></div>
              <div><dt>HTTP</dt><dd>{result.httpStatus || (result.status === "idle" || result.status === "submitting" ? "pending" : "not observed")}</dd></div>
              <div><dt>Failure</dt><dd>{result.failureCode || "none"}{result.failureBoundary ? ` · ${result.failureBoundary}` : ""}</dd></div>
            </dl>
            {result.historyReviewAvailable && result.requestId ? <button type="button" onClick={reviewHistory}>Review sanitized history</button> : null}
          </article>
        </div>
      )}
    </section>
  );
}
