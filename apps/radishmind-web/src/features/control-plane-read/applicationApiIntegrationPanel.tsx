import "../../i18n/apiIntegrationResources.ts";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation("gateway");
  const [catalog, setCatalog] = useState(() => initialApplicationModelCatalogState(config, applicationId));
  const [catalogReused, setCatalogReused] = useState(false);
  const [protocol, setProtocol] = useState<ApplicationApiProtocol>("chat_completions");
  const [language, setLanguage] = useState<ApplicationApiExampleLanguage>("curl");
  const activeCatalogController = useRef<AbortController | null>(null);

  useEffect(() => {
    activeCatalogController.current?.abort();
    activeCatalogController.current = null;
    setCatalog(resetApplicationModelCatalogState(config, applicationId));
    setCatalogReused(false);
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
          summary: "",
        });
        setCatalogReused(true);
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
    if (!catalog.selectedModel || !protocolSupported) return "";
    return generateApplicationApiIntegrationExample({ protocol, language, model: catalog.selectedModel });
  }, [catalog.selectedModel, language, protocol, protocolSupported]);

  async function loadModels(preferredModel = "") {
    setCatalogReused(false);
    if (!workspaceScopeMatches) {
      setCatalog({
        status: "failed",
        applicationId,
        models: [],
        selectedModel: "",
        failureCode: "gateway_model_catalog_workspace_mismatch",
        summary: "",
      });
      return;
    }
    const controller = new AbortController();
    activeCatalogController.current?.abort();
    activeCatalogController.current = controller;
    setCatalog((current) => ({ ...current, status: "loading", models: [], selectedModel: "", failureCode: "", summary: "" }));
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
        setCatalog((current) => ({ ...current, status: "failed", failureCode: "gateway_model_catalog_network_error", summary: "" }));
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
  const catalogStatus = catalog.status === "ready" ? t($ => $.apiIntegration.statusReady)
    : catalog.status === "failed" ? t($ => $.apiIntegration.statusFailed)
    : catalog.status === "loading" ? t($ => $.apiIntegration.statusLoading)
    : catalog.status === "empty" ? t($ => $.apiIntegration.statusEmpty)
    : catalog.status === "offline" ? t($ => $.apiIntegration.offline) : t($ => $.apiIntegration.statusIdle);
  const catalogMessage = catalog.status === "ready"
    ? catalogReused ? t($ => $.apiIntegration.reusedModels, { count: catalog.models.length }) : t($ => $.apiIntegration.loadedModels, { count: catalog.models.length })
    : catalog.status === "empty" ? t($ => $.apiIntegration.validEmptyCatalog)
    : catalog.status === "loading" ? t($ => $.apiIntegration.loadingCatalog)
    : catalog.status === "idle" ? t($ => $.apiIntegration.loadCatalogPrompt)
    : catalog.status === "offline" ? t($ => $.apiIntegration.offlineCatalog)
    : catalog.failureCode === "gateway_model_catalog_workspace_mismatch" ? t($ => $.apiIntegration.workspaceMismatchSummary)
    : catalog.failureCode === "gateway_api_key_handoff_required" ? t($ => $.apiIntegration.handoffRequiredSummary)
    : catalog.failureCode === "gateway_model_catalog_scope_invalid" ? t($ => $.apiIntegration.invalidApplicationScope)
    : catalog.failureCode === "gateway_model_catalog_network_error" ? t($ => $.apiIntegration.catalogLoadFailed)
    : t($ => $.apiIntegration.catalogUnavailable, { code: catalog.failureCode || t($ => $.apiIntegration.unknownFailure) });
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
          <p className="eyebrow">{t($ => $.apiIntegration.connectApi)}</p>
          <h4 id="application-api-integration-title">{t($ => $.apiIntegration.modelProtocolExample)}</h4>
        </div>
        <span className={`status-badge ${enabled && !credentialHandoffRequired ? "good" : "neutral"}`}>
          {credentialHandoffRequired ? t($ => $.apiIntegration.keyHandoffRequired) : enabled ? t($ => $.apiIntegration.devTestScoped) : t($ => $.apiIntegration.offline)}
        </span>
      </div>

      <div className="application-api-integration-scope">
        <article><span>{t($ => $.apiIntegration.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.apiIntegration.workspace)}</span><strong>{workspaceId}</strong><p>{t($ => $.apiIntegration.gatewayScope)}<code>{config.workspaceId}</code></p></article>
      </div>

      {!workspaceScopeMatches ? (
        <article className="application-api-integration-offline" role="alert">
          <p className="eyebrow">{t($ => $.apiIntegration.scopeMismatch)}</p>
          <h5>{t($ => $.apiIntegration.gatewayRequestsBlocked)}</h5>
          <p>{t($ => $.apiIntegration.workspaceMismatch)}</p>
        </article>
      ) : !enabled ? (
        <article className="application-api-integration-offline">
          <p className="eyebrow">{t($ => $.apiIntegration.offlineBoundary)}</p>
          <h5>{t($ => $.apiIntegration.noModelOrInvocationRequest)}</h5>
          <p>{t($ => $.apiIntegration.enableDevTestSource)}</p>
        </article>
      ) : (
        <div className="application-api-integration-layout">
          <article className="application-api-models">
            <div className="application-api-card-heading">
              <div><p className="eyebrow">{t($ => $.apiIntegration.scopedModelCatalog)}</p><h5>/v1/models</h5></div>
              <span className={`status-badge ${catalog.status === "ready" ? "good" : catalog.status === "failed" ? "bad" : "neutral"}`}>{catalogStatus}</span>
            </div>
            <p>{catalogMessage}</p>
            {credentialHandoffRequired && catalog.status !== "ready" ? (
              <div className="application-api-eligibility-blocked" role="status">
                <strong>{t($ => $.apiIntegration.oneTimeCredentialRequired)}</strong>
                <p>{t($ => $.apiIntegration.issueScopedKey)}</p>
                <a href="#workspace-api-keys">{t($ => $.apiIntegration.openCredentials)}<span aria-hidden="true">→</span></a>
              </div>
            ) : (
              <button type="button" onClick={() => void loadModels()} disabled={catalog.status === "loading"}>{catalog.status === "idle" ? t($ => $.apiIntegration.loadModels) : t($ => $.apiIntegration.refreshModels)}</button>
            )}
            {catalog.status === "failed" ? <p className="failure-summary">{catalog.failureCode}: {catalogMessage}</p> : null}
            {catalog.status === "empty" ? <p className="boundary-note">{t($ => $.apiIntegration.emptyInventory)}</p> : null}
            <label>{t($ => $.apiIntegration.model)}<select value={catalog.selectedModel} onChange={(event) => setCatalog((current) => ({ ...current, selectedModel: event.target.value }))} disabled={catalog.models.length === 0 || catalog.status === "loading"}>
                {catalog.models.length === 0 ? <option value="">{t($ => $.apiIntegration.noValidatedModels)}</option> : catalog.models.map((model) => <option value={model.id} key={model.id}>{model.id} · {model.ownedBy || t($ => $.apiIntegration.unowned)}</option>)}
              </select>
            </label>
          </article>

          <article className="application-api-example">
            <div className="application-api-example-controls">
              <label>{t($ => $.apiIntegration.protocol)}<select value={protocol} onChange={(event) => setProtocol(event.target.value as ApplicationApiProtocol)} disabled={!selectedCatalogModel}><option value="chat_completions" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("chat_completions")}>{t($ => $.apiIntegration.chatCompletions)}</option><option value="responses" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("responses")}>{t($ => $.apiIntegration.responses)}</option><option value="messages" disabled={Boolean(selectedCatalogModel) && !selectedCatalogModel?.protocols.includes("messages")}>{t($ => $.apiIntegration.messages)}</option></select></label>
              <label>{t($ => $.apiIntegration.example)}<select value={language} onChange={(event) => setLanguage(event.target.value as ApplicationApiExampleLanguage)}><option value="curl">{t($ => $.apiIntegration.curl)}</option><option value="python">{t($ => $.apiIntegration.python)}</option><option value="typescript">{t($ => $.apiIntegration.typescript)}</option></select></label>
            </div>
            <pre aria-label={t($ => $.apiIntegration.generatedExample)}>{!catalog.selectedModel ? t($ => $.apiIntegration.selectValidatedModel) : !protocolSupported ? t($ => $.apiIntegration.unsupportedProtocol) : example}</pre>
            <div className="application-api-actions">
              <button type="button" onClick={openPlayground} disabled={!catalog.selectedModel || !protocolSupported}>{t($ => $.apiIntegration.openScopedPlayground)}</button>
            </div>
            <p className="boundary-note">{t($ => $.apiIntegration.eligibleProtocols)}</p>
          </article>
        </div>
      )}
      <p className="boundary-note">{t($ => $.apiIntegration.componentMemoryBoundary)}</p>
    </section>
  );
}
