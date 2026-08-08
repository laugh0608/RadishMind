import { useEffect, useMemo, useRef, useState } from "react";

import {
  createGatewayPlaygroundRequestId,
  initialModelGatewayPlaygroundResult,
  modelGatewayPlaygroundConfigForApplication,
  modelGatewayPlaygroundConfigForAPIKey,
  readModelGatewayPlaygroundConfig,
  submitModelGatewayPlaygroundRequest,
  type ModelGatewayPlaygroundConfig,
  type ModelGatewayPlaygroundProtocol,
} from "./modelGatewayPlaygroundConsumer.ts";
import {
  applicationApiIntegrationConfigFromGateway,
  initialApplicationModelCatalogState,
  loadApplicationModelCatalog,
} from "./applicationApiIntegrationConsumer.ts";
import { requestApplicationModelCatalogReady } from "./applicationApiIntegrationEvents.ts";
import {
  MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
  requestGatewayRequestHistoryReview,
  type ModelGatewayPlaygroundHandoffEventDetail,
} from "./modelGatewayPlaygroundEvents.ts";

const baseConfig = readModelGatewayPlaygroundConfig();
const DEFAULT_INPUT = "请用简洁的中文说明 RadishMind Gateway 当前请求的处理结果。";

export default function ModelGatewayPlaygroundPanel({
  selectedApplicationId,
  workspaceId,
  applicationActive,
  active,
}: {
  selectedApplicationId: string;
  workspaceId: string;
  applicationActive: boolean;
  active: boolean;
}) {
  const [applicationId, setApplicationId] = useState(baseConfig.applicationId);
  const [apiKeyCredential, setAPIKeyCredential] = useState<{ apiKeyId: string; token: string } | null>(null);
  const [protocol, setProtocol] = useState<ModelGatewayPlaygroundProtocol>("chat_completions");
  const [model, setModel] = useState(baseConfig.defaultModel);
  const [inputText, setInputText] = useState(DEFAULT_INPUT);
  const [stream, setStream] = useState(false);
  const [result, setResult] = useState(() => initialModelGatewayPlaygroundResult(baseConfig));
  const [catalog, setCatalog] = useState(() => initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(baseConfig), baseConfig.applicationId));
  const activeController = useRef<AbortController | null>(null);
  const activeCatalogController = useRef<AbortController | null>(null);
  const config = useMemo(
    () => apiKeyCredential
      ? modelGatewayPlaygroundConfigForAPIKey(baseConfig, applicationId, apiKeyCredential.apiKeyId, apiKeyCredential.token)
      : modelGatewayPlaygroundConfigForApplication(baseConfig, applicationId),
    [apiKeyCredential, applicationId],
  );
  const workspaceScopeMatches = Boolean(workspaceId.trim()) && baseConfig.workspaceId === workspaceId.trim();
  const selectedCatalogModel = catalog.models.find((item) => item.id === catalog.selectedModel) ?? null;
  const supportedProtocols = selectedCatalogModel?.protocols ?? [];

  useEffect(() => {
    function receiveApplicationHandoff(event: Event) {
      const detail = (event as CustomEvent<ModelGatewayPlaygroundHandoffEventDetail>).detail;
      if (!detail?.applicationId || !detail.model || !applicationActive || !workspaceScopeMatches ||
        detail.applicationId !== selectedApplicationId.trim()) return;
      activeController.current?.abort();
      activeController.current = null;
      activeCatalogController.current?.abort();
      activeCatalogController.current = null;
      const nextCredential = detail.apiKeyCredential ?? null;
      const nextConfig = nextCredential
        ? modelGatewayPlaygroundConfigForAPIKey(baseConfig, detail.applicationId, nextCredential.apiKeyId, nextCredential.token)
        : modelGatewayPlaygroundConfigForApplication(baseConfig, detail.applicationId);
      setApplicationId(detail.applicationId);
      setAPIKeyCredential(nextCredential);
      setProtocol(detail.protocol);
      setModel(detail.model);
      setInputText(DEFAULT_INPUT);
      setStream(false);
      setResult(initialModelGatewayPlaygroundResult(nextConfig));
      setCatalog(initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(nextConfig), detail.applicationId));
    }
    window.addEventListener(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, receiveApplicationHandoff);
    return () => window.removeEventListener(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, receiveApplicationHandoff);
  }, [applicationActive, selectedApplicationId, workspaceScopeMatches]);

  useEffect(() => {
    const normalizedApplicationId = selectedApplicationId.trim();
    if (normalizedApplicationId === applicationId) return;
    activeController.current?.abort();
    activeController.current = null;
    activeCatalogController.current?.abort();
    activeCatalogController.current = null;
    setApplicationId(normalizedApplicationId);
    setAPIKeyCredential(null);
    setProtocol("chat_completions");
    setModel(baseConfig.defaultModel);
    setInputText(DEFAULT_INPUT);
    setStream(false);
    const cleared = modelGatewayPlaygroundConfigForApplication(baseConfig, normalizedApplicationId);
    setResult(initialModelGatewayPlaygroundResult(cleared));
    setCatalog(initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(cleared), normalizedApplicationId));
  }, [applicationId, selectedApplicationId]);

  useEffect(() => {
    if (!applicationActive || !workspaceScopeMatches) {
      activeController.current?.abort();
      activeController.current = null;
      activeCatalogController.current?.abort();
      activeCatalogController.current = null;
      setAPIKeyCredential(null);
      const cleared = modelGatewayPlaygroundConfigForApplication(baseConfig, applicationId);
      setResult(initialModelGatewayPlaygroundResult(cleared));
      setCatalog(initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(cleared), applicationId));
    }
  }, [applicationActive, applicationId, workspaceScopeMatches]);

  useEffect(() => {
    if (catalog.status !== "ready" || !selectedCatalogModel || supportedProtocols.includes(protocol)) return;
    setProtocol(supportedProtocols[0] ?? "chat_completions");
  }, [catalog.status, protocol, selectedCatalogModel, supportedProtocols]);

  useEffect(() => {
    function clearCredentialAfterRouteLeave() {
      if (window.location.hash === "#model-gateway-playground") return;
      activeController.current?.abort();
      activeController.current = null;
      activeCatalogController.current?.abort();
      activeCatalogController.current = null;
      setAPIKeyCredential(null);
      const cleared = modelGatewayPlaygroundConfigForApplication(baseConfig, applicationId);
      setResult(initialModelGatewayPlaygroundResult(cleared));
      setCatalog(initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(cleared), applicationId));
    }
    window.addEventListener("hashchange", clearCredentialAfterRouteLeave);
    return () => window.removeEventListener("hashchange", clearCredentialAfterRouteLeave);
  }, [applicationId]);

  useEffect(() => () => {
    activeController.current?.abort();
    activeCatalogController.current?.abort();
  }, []);

  async function loadModels(configOverride: ModelGatewayPlaygroundConfig = config) {
    if (!active || !applicationActive || !workspaceScopeMatches) return;
    const controller = new AbortController();
    activeCatalogController.current?.abort();
    activeCatalogController.current = controller;
    setCatalog((current) => ({ ...current, status: "loading", models: [], selectedModel: "", failureCode: "", summary: "Loading the scoped Gateway model catalog." }));
    try {
      const next = await loadApplicationModelCatalog(applicationApiIntegrationConfigFromGateway(configOverride), configOverride.applicationId, controller.signal);
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      setCatalog(next);
      if (next.selectedModel) setModel(next.selectedModel);
      if (next.status === "ready" && next.selectedModel) {
        requestApplicationModelCatalogReady(next.applicationId, next.models, next.selectedModel);
      }
    } catch (error) {
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      if (!(error instanceof DOMException && error.name === "AbortError")) {
        setCatalog((current) => ({ ...current, status: "failed", failureCode: "gateway_model_catalog_network_error", summary: "The Gateway model catalog could not be loaded." }));
      }
    }
  }

  function clearCredential() {
    activeController.current?.abort();
    activeController.current = null;
    activeCatalogController.current?.abort();
    activeCatalogController.current = null;
    setAPIKeyCredential(null);
    const cleared = modelGatewayPlaygroundConfigForApplication(baseConfig, applicationId);
    setCatalog(initialApplicationModelCatalogState(applicationApiIntegrationConfigFromGateway(cleared), applicationId));
    setResult(initialModelGatewayPlaygroundResult(cleared));
  }

  async function submit() {
    if (!active || !applicationActive || !workspaceScopeMatches || catalog.status !== "ready" ||
      !selectedCatalogModel || !supportedProtocols.includes(protocol)) return;
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
    requestGatewayRequestHistoryReview(
      result.requestId,
      applicationId,
      apiKeyCredential ? `api_key:${apiKeyCredential.apiKeyId}` : config.consumerRef,
    );
    window.location.hash = "model-gateway-request-history";
  }

  const enabled = config.mode === "dev_gateway_playground_http";
  const credentialReady = config.authMode === "dev_headers" || Boolean(apiKeyCredential);
  const executionReady = enabled && active && applicationActive && workspaceScopeMatches && credentialReady &&
    catalog.status === "ready" && Boolean(selectedCatalogModel) && supportedProtocols.includes(protocol);
  return (
    <section className="surface-band model-gateway-overview gateway-playground" id="model-gateway-playground" aria-labelledby="model-gateway-playground-title">
      <div className="section-heading">
        <div><p className="eyebrow">Model Gateway</p><h3 id="model-gateway-playground-title">Playground and request review</h3></div>
        <span className={`status-badge ${executionReady ? "good" : "neutral"}`}>{enabled ? "dev/test controlled" : "offline"}</span>
      </div>
      {enabled && !workspaceScopeMatches ? (
        <article className="model-gateway-overview-hero gateway-playground-blocked" role="alert">
          <div><p className="eyebrow">Workspace boundary</p><h4>Gateway source scope does not match this workspace</h4><p>Configured source <code>{baseConfig.workspaceId}</code> cannot be used for Application Workspace <code>{workspaceId || "unavailable"}</code>. No model or inference request is sent.</p></div>
        </article>
      ) : enabled && !applicationActive ? (
        <article className="model-gateway-overview-hero gateway-playground-blocked" role="status">
          <div><p className="eyebrow">Archived application</p><h4>Controlled invocation is closed</h4><p>Sanitized request history and current-window evidence remain readable, but an archived Application cannot load models or send a Gateway request.</p></div>
        </article>
      ) : !enabled ? (
        <article className="model-gateway-overview-hero">
          <div><p className="eyebrow">Offline boundary</p><h4>No northbound request is sent</h4><p>Enable the explicit Gateway Playground and Request History dev/test source to run a request. Production keys, quota, billing, retry, fallback, and persistence remain disabled.</p></div>
        </article>
      ) : (
        <div className="gateway-playground-layout">
          <form className="gateway-playground-form" onSubmit={(event) => { event.preventDefault(); void submit(); }}>
            <div className="gateway-playground-scope"><p><strong>Application scope</strong><code>{applicationId || "unbound"}</code></p><p><strong>Authentication</strong><code>{config.authMode === "api_key_dev_test" ? apiKeyCredential?.apiKeyId ?? "handoff required" : "dev headers"}</code></p>{apiKeyCredential ? <button type="button" className="secondary-action" onClick={clearCredential}>Clear credential</button> : null}</div>
            <div className="gateway-playground-model-catalog">
              <div><p className="eyebrow">Scoped model catalog</p><span className={`status-badge ${catalog.status === "ready" ? "good" : catalog.status === "failed" ? "bad" : "neutral"}`}>{catalog.status}</span></div>
              <p>{catalog.summary}</p>
              <button type="button" onClick={() => void loadModels()} disabled={!applicationId || !credentialReady || catalog.status === "loading" || !active}>{catalog.status === "loading" ? "Loading models…" : "Load models"}</button>
              {catalog.models.length ? <label>Validated model<select value={catalog.selectedModel} onChange={(event) => { const selectedModel = event.target.value; const item = catalog.models.find((candidate) => candidate.id === selectedModel); setCatalog((current) => ({ ...current, selectedModel })); setModel(selectedModel); if (item && !item.protocols.includes(protocol)) setProtocol(item.protocols[0] ?? "chat_completions"); }}>{catalog.models.map((item) => <option key={item.id} value={item.id}>{item.id}</option>)}</select></label> : null}
              {catalog.failureCode ? <p className="failure-summary">{catalog.failureCode}: {catalog.summary}</p> : null}
            </div>
            <label>Protocol<select value={protocol} onChange={(event) => setProtocol(event.target.value as ModelGatewayPlaygroundProtocol)} disabled={result.status === "submitting" || catalog.status !== "ready"}><option value="chat_completions" disabled={!supportedProtocols.includes("chat_completions")}>Chat Completions</option><option value="responses" disabled={!supportedProtocols.includes("responses")}>Responses</option><option value="messages" disabled={!supportedProtocols.includes("messages")}>Messages</option></select></label>
            <label>Model<input value={model} readOnly maxLength={160} disabled={result.status === "submitting" || catalog.status !== "ready"} /></label>
            <label className="gateway-playground-input">Temporary input<textarea value={inputText} onChange={(event) => setInputText(event.target.value)} maxLength={8000} rows={7} disabled={result.status === "submitting"} /></label>
            <label className="gateway-playground-stream"><input type="checkbox" checked={stream} onChange={(event) => setStream(event.target.checked)} disabled={result.status === "submitting"} /> Stream response</label>
            <div className="gateway-playground-actions">
              <button type="submit" disabled={result.status === "submitting" || !executionReady}>Send request</button>
              <button type="button" onClick={cancel} disabled={result.status !== "submitting"}>Cancel</button>
            </div>
            <p className="boundary-note">Input, output, and any handed-off API key stay in this component and active HTTP requests. The credential is cleared when leaving this route and is never written to browser storage.</p>
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
