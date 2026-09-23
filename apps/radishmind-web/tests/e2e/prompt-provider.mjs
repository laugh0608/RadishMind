import { createServer } from "node:http";

// A local transport fixture, never a model or a quality evaluation.
export async function startPromptProvider() {
  const observations = [];
  const server = createServer((request, response) => {
    const send = (status, body) => {
      response.writeHead(status, { "Content-Type": "application/json", "Cache-Control": "no-store" });
      response.end(JSON.stringify(body));
    };
    async function handle() {
      const match = request.url?.match(/^\/observations\/([a-f0-9-]{36})$/);
      if (request.method === "GET" && match) {
        return send(200, observations.filter((item) => item.caseId === match[1]));
      }
      if (request.method !== "POST" || request.url !== "/v1/chat/completions") return send(404, { error: "fixture_route_not_found" });
      const chunks = [];
      let bytes = 0;
      for await (const chunk of request) {
        bytes += chunk.length;
        if (bytes > 256 * 1024) return send(413, { error: "fixture_request_too_large" });
        chunks.push(chunk);
      }
      let document;
      try { document = JSON.parse(Buffer.concat(chunks).toString("utf8")); }
      catch { return send(400, { error: "fixture_invalid_json" }); }
      const messages = document.messages;
      const marker = messages?.[1]?.content?.match(/^E2E_CASE:([a-f0-9-]{36}):(valid|invalid|missing)$/);
      const accepted = document.model === "prompt-e2e-model" && Array.isArray(messages) && messages.length === 2
        && messages[0].role === "system" && messages[0].content === "E2E_PROMPT_SYSTEM_V1"
        && messages[1].role === "user" && Boolean(marker);
      observations.push({ caseId: marker?.[1] ?? "unrecognized", accepted, mode: marker?.[2] ?? "unknown" });
      if (!accepted) return send(400, { error: "fixture_prompt_transport_mismatch" });
      const output = {
        diagnosis: "Synthetic timeout diagnosis",
        evidence: ["Synthetic upstream timeout"],
        next_checks: ["Check synthetic upstream health"],
        missing_context: ["Synthetic request trace"],
        uncertainty: "Synthetic evidence is incomplete",
      };
      if (marker[2] === "invalid") output.unexpected = "E2E_INVALID_OUTPUT";
      if (marker[2] === "missing") delete output.diagnosis;
      const content = JSON.stringify(output);
      send(200, {
        id: "prompt-e2e-response", object: "chat.completion",
        choices: [{ index: 0, message: { role: "assistant", content }, finish_reason: "stop" }],
        usage: { prompt_tokens: 7, completion_tokens: 5, total_tokens: 12 },
      });
    }
    void handle().catch(() => {
      if (!response.headersSent) send(500, { error: "fixture_request_failed" });
      else response.destroy();
    });
  });
  server.requestTimeout = 10_000;
  await new Promise((accept, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", accept);
  });
  return {
    url: `http://127.0.0.1:${server.address().port}`,
    observations,
    async close() {
      await new Promise((accept, reject) => {
        server.close((error) => error ? reject(error) : accept());
        server.closeAllConnections();
      });
    },
  };
}
