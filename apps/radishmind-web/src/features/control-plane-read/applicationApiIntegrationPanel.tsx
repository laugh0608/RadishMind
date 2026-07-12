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

const config = readApplicationApiIntegrationConfig();

export default function ApplicationApiIntegrationPanel({
  applicationId,
  applicationName,
}: {
  applicationId: string;
  applicationName: string;
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

  useEffect(() => () => activeCatalogController.current?.abort(), []);

  const example = useMemo(() => {
    if (!catalog.selectedModel) return "Select a validated model to generate an integration example.";
    return generateApplicationApiIntegrationExample({ protocol, language, model: catalog.selectedModel });
  }, [catalog.selectedModel, language, protocol]);

  async function loadModels() {
    const controller = new AbortController();
    activeCatalogController.current?.abort();
    activeCatalogController.current = controller;
    setCatalog((current) => ({ ...current, status: "loading", models: [], selectedModel: "", failureCode: "", summary: "Loading the scoped Gateway model catalog." }));
    try {
      const next = await loadApplicationModelCatalog(config, applicationId, controller.signal);
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      setCatalog(next);
    } catch (error) {
      if (activeCatalogController.current !== controller) return;
      activeCatalogController.current = null;
      if (!(error instanceof DOMException && error.name === "AbortError")) {
        setCatalog((current) => ({ ...current, status: "failed", failureCode: "gateway_model_catalog_network_error", summary: "The Gateway model catalog could not be loaded." }));
      }
    }
  }

  function openPlayground() {
    if (!catalog.selectedModel) return;
    requestModelGatewayPlaygroundHandoff(applicationId, protocol, catalog.selectedModel);
    window.location.hash = "model-gateway-playground";
  }

  const enabled = config.mode === "dev_application_api_http";
  return (
    <section className="application-api-integration" id="application-api-integration" aria-labelledby="application-api-integration-title">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Application API Integration</p>
          <h4 id="application-api-integration-title">Models, examples, and test invocation</h4>
        </div>
        <span className={`status-badge ${enabled ? "good" : "neutral"}`}>{enabled ? "dev/test scoped" : "offline"}</span>
      </div>

      <div className="application-api-integration-scope">
        <article><span>Application</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>Workspace</span><strong>{config.workspaceId}</strong><p>Current selection is carried to models, Playground, and History.</p></article>
      </div>

      {!enabled ? (
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
            <button type="button" onClick={() => void loadModels()} disabled={catalog.status === "loading"}>{catalog.status === "idle" ? "Load models" : "Refresh models"}</button>
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
              <label>Protocol<select value={protocol} onChange={(event) => setProtocol(event.target.value as ApplicationApiProtocol)}><option value="chat_completions">Chat Completions</option><option value="responses">Responses</option><option value="messages">Messages</option></select></label>
              <label>Example<select value={language} onChange={(event) => setLanguage(event.target.value as ApplicationApiExampleLanguage)}><option value="curl">cURL</option><option value="python">Python</option><option value="typescript">TypeScript</option></select></label>
            </div>
            <pre aria-label="Generated API integration example">{example}</pre>
            <div className="application-api-actions">
              <button type="button" onClick={openPlayground} disabled={!catalog.selectedModel}>Open scoped Playground</button>
            </div>
            <p className="boundary-note">Examples use environment placeholders only. Internal dev caller headers and real credential values are never rendered.</p>
          </article>
        </div>
      )}
      <p className="boundary-note">Model selection and examples stay in component memory. Production key lifecycle, quota, billing, fallback, load balancing, and production authorization remain disabled.</p>
    </section>
  );
}
