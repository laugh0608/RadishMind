import { useEffect, useMemo, useRef, useState } from "react";

import {
  generateApplicationApiIntegrationExample,
  initialApplicationModelCatalogState,
  loadApplicationModelCatalog,
  readApplicationApiIntegrationConfig,
  resetApplicationModelCatalogState,
  type ApplicationApiExampleLanguage,
  type ApplicationApiProtocol,
} from "./applicationApiIntegrationConsumer.ts";
import { requestModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";
import {
  APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT,
  APPLICATION_MODEL_CATALOG_READY_EVENT,
  consumePendingApplicationApiIntegrationDraftHandoff,
  createApplicationModelCatalogReadyDetail,
  type ApplicationApiIntegrationDraftHandoffDetail,
  type ApplicationModelCatalogReadyDetail,
} from "./applicationApiIntegrationEvents.ts";

const config = readApplicationApiIntegrationConfig();

export default function ApplicationApiIntegrationPanel({
  applicationId,
  applicationName,
  workspaceId,
}: {
  applicationId: string;
  applicationName: string;
  workspaceId: string;
}) {
  const [catalog, setCatalog] = useState(() => initialApplicationModelCatalogState(config, applicationId));
  const [protocol, setProtocol] = useState<ApplicationApiProtocol>("chat_completions");
  const [language, setLanguage] = useState<ApplicationApiExampleLanguage>("curl");
  const activeCatalogController = useRef<AbortController | null>(null);

  useEffect(() => {
    activeCatalogController.current?.abort();
    activeCatalogController.current = null;
    setCatalog(resetApplicationModelCatalogState(config, applicationId));
    setProtocol("chat_completions");
    setLanguage("curl");
  }, [applicationId]);

  useEffect(() => {
    function receiveValidatedCatalog(event: Event) {
      const detail = (event as CustomEvent<ApplicationModelCatalogReadyDetail>).detail;
      try {
        const normalized = createApplicationModelCatalogReadyDetail(
          detail?.applicationId ?? "",
          detail?.models ?? [],
          detail?.selectedModel ?? "",
        );
        if (normalized.applicationId !== applicationId) return;
        activeCatalogController.current?.abort();
        activeCatalogController.current = null;
        setCatalog({
          status: "ready",
          applicationId: normalized.applicationId,
          models: normalized.models,
          selectedModel: normalized.selectedModel,
          failureCode: "",
          summary: `Reused ${normalized.models.length} models validated by the Gateway Playground.`,
        });
      } catch {
        return;
      }
    }
    window.addEventListener(APPLICATION_MODEL_CATALOG_READY_EVENT, receiveValidatedCatalog);
    return () => window.removeEventListener(APPLICATION_MODEL_CATALOG_READY_EVENT, receiveValidatedCatalog);
  }, [applicationId]);

  useEffect(() => () => activeCatalogController.current?.abort(), []);

  useEffect(() => {
    function applyDraftHandoff(detail: ApplicationApiIntegrationDraftHandoffDetail | null | undefined) {
      if (!detail || detail.applicationId !== applicationId) return;
      setProtocol(detail.protocol);
      void loadModels(detail.model);
    }
    function handleDraftHandoff(event: Event) {
      const detail = (event as CustomEvent<ApplicationApiIntegrationDraftHandoffDetail>).detail;
      if (!detail || detail.applicationId !== applicationId) return;
      consumePendingApplicationApiIntegrationDraftHandoff(applicationId);
      applyDraftHandoff(detail);
    }
    window.addEventListener(APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT, handleDraftHandoff);
    applyDraftHandoff(consumePendingApplicationApiIntegrationDraftHandoff(applicationId));
    return () => window.removeEventListener(APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT, handleDraftHandoff);
  }, [applicationId]);

  const selectedCatalogModel = useMemo(
    () => catalog.models.find((model) => model.id === catalog.selectedModel) ?? null,
    [catalog.models, catalog.selectedModel],
  );
  const protocolSupported = selectedCatalogModel?.protocols.includes(protocol) ?? false;

  useEffect(() => {
    if (!selectedCatalogModel || protocolSupported) return;
    setProtocol(selectedCatalogModel.protocols[0]);
  }, [protocolSupported, selectedCatalogModel]);

  const example = useMemo(() => {
    if (!catalog.selectedModel) return "Select a validated model to generate an integration example.";
    if (!protocolSupported) return "The selected model does not advertise this northbound protocol.";
    return generateApplicationApiIntegrationExample({ protocol, language, model: catalog.selectedModel });
  }, [catalog.selectedModel, language, protocol, protocolSupported]);

  async function loadModels(preferredModel = "") {
    if (!workspaceScopeMatches) {
      setCatalog({
        status: "failed",
        applicationId,
        models: [],
        selectedModel: "",
        failureCode: "gateway_model_catalog_workspace_mismatch",
        summary: "The selected workspace does not match the configured Gateway dev/test scope.",
      });
      return;
    }
    const controller = new AbortController();
    activeCatalogController.current?.abort();
    activeCatalogController.current = controller;
    setCatalog((current) => ({ ...current, status: "loading", models: [], selectedModel: "", failureCode: "", summary: "Loading the scoped Gateway model catalog." }));
    try {
      const next = await loadApplicationModelCatalog(config, applicationId, controller.signal);
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      setCatalog({
        ...next,
        selectedModel: preferredModel && next.models.some((model) => model.id === preferredModel)
          ? preferredModel
          : next.selectedModel,
      });
    } catch (error) {
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      if (!(error instanceof DOMException && error.name === "AbortError")) {
        setCatalog((current) => ({ ...current, status: "failed", failureCode: "gateway_model_catalog_network_error", summary: "The Gateway model catalog could not be loaded." }));
      }
    }
  }

  function openPlayground() {
    if (!catalog.selectedModel || !protocolSupported || !workspaceScopeMatches) return;
    requestModelGatewayPlaygroundHandoff(applicationId, protocol, catalog.selectedModel);
    window.location.hash = "model-gateway-playground";
  }

  const workspaceScopeMatches = config.mode !== "dev_application_api_http" || config.workspaceId === workspaceId;
  const enabled = config.mode === "dev_application_api_http" && workspaceScopeMatches;
  const credentialHandoffRequired = enabled && config.authMode === "api_key_dev_test" && !config.apiKeyToken;
  return (
    <section
      className="application-api-integration"
      id="application-api-integration"
      aria-labelledby="application-api-integration-title"
      data-auth-mode={config.authMode}
      data-workspace-scope-matches={String(workspaceScopeMatches)}
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Connect API</p>
          <h4 id="application-api-integration-title">Model, protocol, and example</h4>
        </div>
        <span className={`status-badge ${enabled && !credentialHandoffRequired ? "good" : "neutral"}`}>
          {credentialHandoffRequired ? "API key handoff required" : enabled ? "dev/test scoped" : "offline"}
        </span>
      </div>

      <div className="application-api-integration-scope">
        <article><span>Application</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>Workspace</span><strong>{workspaceId}</strong><p>Gateway scope: <code>{config.workspaceId}</code></p></article>
      </div>

      {!workspaceScopeMatches ? (
        <article className="application-api-integration-offline" role="alert">
          <p className="eyebrow">Scope mismatch</p>
          <h5>Gateway requests are blocked</h5>
          <p>The active Application Workspace and configured Gateway workspace differ. Switch to the matching workspace before loading models.</p>
        </article>
      ) : !enabled ? (
        <article className="application-api-integration-offline">
          <p className="eyebrow">Offline boundary</p>
          <h5>No model or invocation request is sent</h5>
          <p>Enable the existing Gateway Playground dev/test source to load `/v1/models` and hand this application to the existing invocation path.</p>
        </article>
      ) : (
        <div className="application-api-integration-layout">
          <article className="application-api-models">
            <div className="application-api-card-heading">
              <div><p className="eyebrow">Scoped model catalog</p><h5>/v1/models</h5></div>
              <span className={`status-badge ${catalog.status === "ready" ? "good" : catalog.status === "failed" ? "bad" : "neutral"}`}>{catalog.status}</span>
            </div>
            <p>{catalog.summary}</p>
            {credentialHandoffRequired && catalog.status !== "ready" ? (
              <div className="application-api-eligibility-blocked" role="status">
                <strong>One-time credential required</strong>
                <p>Issue a scoped Key, then hand it directly to the existing Playground. This Integration surface never stores or reconstructs the token.</p>
                <a href="#workspace-api-keys">Open credentials <span aria-hidden="true">→</span></a>
              </div>
            ) : (
              <button type="button" onClick={() => void loadModels()} disabled={catalog.status === "loading"}>{catalog.status === "idle" ? "Load models" : "Refresh models"}</button>
            )}
            {catalog.failureCode ? <p className="failure-summary">{catalog.failureCode}: {catalog.summary}</p> : null}
            {catalog.status === "empty" ? <p className="boundary-note">No selectable models were returned for the current dev/test inventory.</p> : null}
            <label>Model
              <select value={catalog.selectedModel} onChange={(event) => setCatalog((current) => ({ ...current, selectedModel: event.target.value }))} disabled={catalog.models.length === 0 || catalog.status === "loading"}>
                {catalog.models.length === 0 ? <option value="">No validated models</option> : catalog.models.map((model) => <option value={model.id} key={model.id}>{model.id} · {model.ownedBy || "unowned"}</option>)}
              </select>
            </label>
          </article>

          <article className="application-api-example">
            <div className="application-api-example-controls">
              <label>Protocol<select value={protocol} onChange={(event) => setProtocol(event.target.value as ApplicationApiProtocol)} disabled={!selectedCatalogModel}><option value="chat_completions" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("chat_completions")}>Chat Completions</option><option value="responses" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("responses")}>Responses</option><option value="messages" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("messages")}>Messages</option></select></label>
              <label>Example<select value={language} onChange={(event) => setLanguage(event.target.value as ApplicationApiExampleLanguage)}><option value="curl">cURL</option><option value="python">Python</option><option value="typescript">TypeScript</option></select></label>
            </div>
            <pre aria-label="Generated API integration example">{example}</pre>
            <div className="application-api-actions">
              <button type="button" onClick={openPlayground} disabled={!catalog.selectedModel || !protocolSupported}>Open scoped Playground</button>
            </div>
            <p className="boundary-note">Only protocols advertised by the selected model are eligible. Examples use environment placeholders; internal caller headers and credential values are never rendered.</p>
          </article>
        </div>
      )}
      <p className="boundary-note">Model selection and examples stay in component memory. Production key lifecycle, quota, billing, fallback, load balancing, and production authorization remain disabled.</p>
    </section>
  );
}
