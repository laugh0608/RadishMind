import { useEffect, useMemo, useRef, useState } from "react";

import {
  compareApplicationConfigurationDraft,
  createApplicationConfigurationDraft,
  initialApplicationConfigurationDraftListState,
  initialApplicationConfigurationDraftState,
  initialApplicationDraftModelCatalog,
  listApplicationConfigurationDrafts,
  loadApplicationDraftModelCatalog,
  readApplicationConfigurationDraft,
  readApplicationConfigurationDraftConfig,
  saveApplicationConfigurationDraft,
  validateApplicationConfigurationDraft,
  validateApplicationConfigurationDraftRemote,
  type ApplicationConfigurationBaseline,
  type ApplicationConfigurationDraft,
  type ApplicationConfigurationDraftOperationState,
} from "./applicationConfigurationDraftConsumer.ts";
import type { ApplicationApiProtocol } from "./applicationApiIntegrationConsumer.ts";
import { requestApplicationApiIntegrationDraftHandoff } from "./applicationApiIntegrationEvents.ts";
import { requestModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";

const config = readApplicationConfigurationDraftConfig();
const protocols: Array<{ id: ApplicationApiProtocol; label: string }> = [
  { id: "chat_completions", label: "Chat Completions" },
  { id: "responses", label: "Responses" },
  { id: "messages", label: "Messages" },
];

export default function ApplicationConfigurationDraftPanel({ baseline }: { baseline: ApplicationConfigurationBaseline }) {
  const [draft, setDraft] = useState(() => createApplicationConfigurationDraft(config, baseline));
  const [operation, setOperation] = useState(() => initialApplicationConfigurationDraftState(config));
  const [catalog, setCatalog] = useState(() => initialApplicationDraftModelCatalog(config, baseline.applicationId));
  const [list, setList] = useState(() => initialApplicationConfigurationDraftListState(config));
  const catalogController = useRef<AbortController | null>(null);

  useEffect(() => {
    catalogController.current?.abort();
    catalogController.current = null;
    setDraft(createApplicationConfigurationDraft(config, baseline));
    setOperation(initialApplicationConfigurationDraftState(config));
    setCatalog(initialApplicationDraftModelCatalog(config, baseline.applicationId));
    setList(initialApplicationConfigurationDraftListState(config));
  }, [baseline.applicationId]);

  useEffect(() => () => catalogController.current?.abort(), []);

  const differences = useMemo(() => compareApplicationConfigurationDraft(baseline, draft), [baseline, draft]);
  const currentValidation = useMemo(() => validateApplicationConfigurationDraft(draft, catalog.models), [catalog.models, draft]);
  const enabled = config.mode === "dev_application_draft_http";
  const handoffReady = operation.validation.isValid && currentValidation.isValid && catalog.status === "ready";

  function edit(patch: Partial<ApplicationConfigurationDraft>) {
    setDraft((current) => ({ ...current, ...patch }));
    setOperation((current) => ({ ...current, status: enabled ? "unsaved" : "offline", summary: "Application configuration has unsaved in-memory edits.", failureCode: "", validation: { state: "invalid", isValid: false, findings: [] } }));
  }

  async function loadModels() {
    const controller = new AbortController();
    catalogController.current?.abort();
    catalogController.current = controller;
    setCatalog((current) => ({ ...current, status: "loading", models: [], selectedModel: "", failureCode: "", summary: "Loading models for this application draft." }));
    const next = await loadApplicationDraftModelCatalog(baseline.applicationId, controller.signal);
    if (catalogController.current !== controller) return;
    catalogController.current = null;
    setCatalog(next);
    if (next.selectedModel) {
      setDraft((current) => ({ ...current, defaultModel: next.selectedModel }));
      setOperation((current) => ({ ...current, status: "unsaved", summary: "The refreshed model selection requires validation.", failureCode: "", validation: { state: "invalid", isValid: false, findings: [] } }));
    }
  }

  async function validateDraft() {
    const local = validateApplicationConfigurationDraft(draft, catalog.models);
    if (!local.isValid || !enabled) {
      setOperation((current) => ({ ...current, status: local.isValid ? "valid" : "invalid", validation: local, failureCode: local.findings[0]?.code ?? "", summary: local.isValid ? "Application configuration is valid in offline memory; saving remains disabled." : "Resolve the blocking configuration findings before saving or handoff." }));
      return;
    }
    setOperation((current) => ({ ...current, status: "validating", summary: "Validating the sanitized draft through the dev-only route.", failureCode: "", validation: local }));
    const next = await validateApplicationConfigurationDraftRemote(config, draft);
    setOperation({ ...next, currentDraftVersion: operation.currentDraftVersion });
  }

  async function saveDraft() {
    const local = validateApplicationConfigurationDraft(draft, catalog.models);
    if (!local.isValid) {
      setOperation((current) => ({ ...current, status: "invalid", validation: local, failureCode: local.findings[0]?.code ?? "application_draft_payload_invalid", summary: "Resolve the blocking configuration findings before saving." }));
      return;
    }
    setOperation((current) => ({ ...current, status: "saving", summary: "Saving the sanitized application configuration draft.", failureCode: "", validation: local }));
    const next = await saveApplicationConfigurationDraft(config, draft, operation.currentDraftVersion);
    setOperation(next);
    if (next.status === "saved") await refreshList();
  }

  async function refreshList() {
    if (!enabled) return;
    setList((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "Loading saved drafts for this application." }));
    setList(await listApplicationConfigurationDrafts(config, baseline.applicationId));
  }

  async function restoreDraft(draftId: string) {
    setOperation((current) => ({ ...current, status: "loading", summary: "Restoring the selected saved application draft.", failureCode: "" }));
    const restored = await readApplicationConfigurationDraft(config, baseline.applicationId, draftId);
    setOperation(restored.state);
    if (restored.draft) setDraft(restored.draft);
  }

  function continueAfterConflict() {
    if (operation.status !== "version_conflict") return;
    setOperation((current) => ({ ...current, status: "unsaved", failureCode: "", summary: `Local edits are preserved. The next save will compare against version ${current.currentDraftVersion}.` }));
  }

  function openIntegration() {
    if (!handoffReady) return;
    requestApplicationApiIntegrationDraftHandoff(draft.applicationId, draft.defaultProtocol, draft.defaultModel);
    window.location.hash = "application-api-integration";
  }

  function openPlayground() {
    if (!handoffReady) return;
    requestModelGatewayPlaygroundHandoff(draft.applicationId, draft.defaultProtocol, draft.defaultModel);
    window.location.hash = "model-gateway-playground";
  }

  return (
    <section className="application-configuration-draft" id="application-configuration-draft" aria-labelledby="application-configuration-draft-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Application Configuration Draft</p><h4 id="application-configuration-draft-title">Configure, validate, save, and review</h4></div>
        <span className={`status-badge ${operation.status === "saved" || operation.status === "valid" || operation.status === "restored" ? "good" : operation.status === "invalid" || operation.status === "version_conflict" || operation.status === "store_failure" ? "bad" : "neutral"}`}>{operation.status}</span>
      </div>

      <div className="application-draft-scope">
        <article><span>Application</span><strong>{baseline.displayName}</strong><code>{baseline.applicationId}</code></article>
        <article><span>Workspace</span><strong>{config.workspaceId}</strong><p>{enabled ? "dev/test repository enabled" : "offline memory only"}</p></article>
        <article><span>Version</span><strong>{operation.currentDraftVersion || "unsaved"}</strong><p>Formal application remains read-only.</p></article>
      </div>

      <div className="application-draft-layout">
        <article className="application-draft-editor">
          <div className="application-api-card-heading"><div><p className="eyebrow">Sanitized configuration</p><h5>{draft.draftId}</h5></div><span className="status-badge neutral">{draft.schemaVersion}</span></div>
          <label>Display name<input value={draft.displayName} onChange={(event) => edit({ displayName: event.target.value })} maxLength={120} /></label>
          <label>Description<textarea value={draft.description} onChange={(event) => edit({ description: event.target.value })} maxLength={1000} rows={4} placeholder="Public application purpose; never paste credentials or request content." /></label>
          <label>Application kind<select value={draft.applicationKind} onChange={(event) => edit({ applicationKind: event.target.value })}><option value="workflow_copilot">Workflow Copilot</option><option value="docs_qa">Docs QA</option><option value="agent">Agent</option><option value="prompt_application">Prompt Application</option></select></label>
          <fieldset><legend>Allowed protocols</legend>{protocols.map((protocol) => <label key={protocol.id}><input type="checkbox" checked={draft.allowedProtocols.includes(protocol.id)} onChange={(event) => edit({ allowedProtocols: event.target.checked ? [...draft.allowedProtocols, protocol.id] : draft.allowedProtocols.filter((item) => item !== protocol.id) })} />{protocol.label}</label>)}</fieldset>
          <label>Default protocol<select value={draft.defaultProtocol} onChange={(event) => edit({ defaultProtocol: event.target.value as ApplicationApiProtocol })}>{protocols.map((protocol) => <option key={protocol.id} value={protocol.id}>{protocol.label}</option>)}</select></label>
          <label>Default model<select value={draft.defaultModel} disabled={catalog.models.length === 0} onChange={(event) => edit({ defaultModel: event.target.value })}><option value="">No validated model</option>{catalog.models.map((model) => <option key={model.id} value={model.id}>{model.id} · {model.protocols.join(", ")}</option>)}</select></label>
          <div className="application-draft-actions"><button type="button" onClick={() => void loadModels()} disabled={!enabled || catalog.status === "loading"}>{catalog.status === "loading" ? "Loading models…" : "Load / refresh models"}</button><button type="button" onClick={() => void validateDraft()}>Validate configuration</button><button type="button" onClick={() => void saveDraft()} disabled={!enabled || operation.status === "saving" || operation.status === "validating" || operation.status === "version_conflict"}>Save draft</button></div>
          <p className="boundary-note">{catalog.summary}</p>
        </article>

        <article className="application-draft-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">Review state</p><h5>{operation.summary}</h5></div><span className={`status-badge ${operation.validation.isValid ? "good" : "neutral"}`}>{operation.validation.state}</span></div>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          {operation.validation.findings.length ? <ul className="application-draft-findings">{operation.validation.findings.map((finding) => <li key={`${finding.code}-${finding.field}`}><strong>{finding.field}</strong><span>{finding.code}</span><p>{finding.summary}</p></li>)}</ul> : <p className="boundary-note">No validation findings are currently available.</p>}
          {operation.status === "version_conflict" ? <div className="application-draft-conflict"><strong>Saved version {operation.currentDraftVersion} is newer.</strong><p>Your in-memory edits were not overwritten.</p><button type="button" onClick={continueAfterConflict}>Continue local edits against saved version</button>{list.summaries[0] ? <button type="button" onClick={() => void restoreDraft(list.summaries[0].draftId)}>Restore saved version</button> : null}</div> : null}
          <div className="application-draft-handoff"><button type="button" disabled={!handoffReady} onClick={openIntegration}>Open API Integration</button><button type="button" disabled={!handoffReady} onClick={openPlayground}>Test in Playground</button></div>
          <p className="boundary-note">Handoff contains only application, protocol, and validated model. It never contains form text, credentials, or request input.</p>
        </article>
      </div>

      <div className="application-draft-lower-grid">
        <article className="application-draft-diff"><div className="application-api-card-heading"><div><p className="eyebrow">Configuration comparison</p><h5>Read model → draft</h5></div><span className="status-badge neutral">{differences.filter((item) => item.changed).length} changed</span></div>{differences.map((difference) => <div className={difference.changed ? "changed" : "unchanged"} key={difference.field}><strong>{difference.field}</strong><span>{difference.before}</span><span>→</span><span>{difference.after}</span></div>)}</article>
        <article className="application-draft-saved"><div className="application-api-card-heading"><div><p className="eyebrow">Saved dev/test drafts</p><h5>{list.summary}</h5></div><button type="button" onClick={() => void refreshList()} disabled={!enabled || list.status === "loading"}>Refresh</button></div>{list.failureCode ? <p className="failure-summary">{list.failureCode}</p> : null}{list.summaries.map((summary) => <button type="button" className="application-draft-summary" key={summary.draftId} onClick={() => void restoreDraft(summary.draftId)}><strong>{summary.displayName}</strong><span>v{summary.draftVersion} · {summary.defaultProtocol} · {summary.defaultModel}</span><small>{summary.updatedAt} · {summary.updatedByActorRef}</small></button>)}</article>
      </div>

      <p className="boundary-note">Drafts do not create, publish, delete, or update formal applications. Offline edits stay in memory; production authorization, API keys, quota, billing, provider credentials, fallback, and load balancing remain disabled.</p>
    </section>
  );
}
