import "../../i18n/configurationResources.ts";
import { useTranslation } from "react-i18next";
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
import {
  APPLICATION_MODEL_CATALOG_READY_EVENT,
  readLatestApplicationModelCatalogReady,
  createApplicationModelCatalogReadyDetail,
  requestApplicationApiIntegrationDraftHandoff,
  type ApplicationModelCatalogReadyDetail,
} from "./applicationApiIntegrationEvents.ts";
import { requestModelGatewayPlaygroundHandoff } from "./modelGatewayPlaygroundEvents.ts";
import {
  findExactEligibleWorkflowRAGPromotionBinding,
  initialWorkflowRAGPromotionListResult,
  listWorkflowRAGPromotionCandidates,
  readWorkflowRAGPromotionConfig,
  type WorkflowRAGPromotionListResult,
  type WorkflowRAGPromotionSummary,
} from "./workflowRAGPromotionConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const config = readApplicationConfigurationDraftConfig();
const promotionConfig = readWorkflowRAGPromotionConfig();
const protocols: Array<{ id: ApplicationApiProtocol; label: string }> = [
  { id: "chat_completions", label: "Chat Completions" },
  { id: "responses", label: "Responses" },
  { id: "messages", label: "Messages" },
];

export default function ApplicationConfigurationDraftPanel({
  baseline,
  readOnly = false,
  handoffBindingCandidateId = "",
  handoffId = "",
  onHandoffConsumed,
  onEvidenceChange,
  onOpenPublishReview,
}: {
  baseline: ApplicationConfigurationBaseline;
  readOnly?: boolean;
  handoffBindingCandidateId?: string;
  handoffId?: string;
  onHandoffConsumed?: (handoffId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
  onOpenPublishReview?: (draftId: string) => void;
}) {
  const { t } = useTranslation("applications");
  const [draft, setDraft] = useState(() => createApplicationConfigurationDraft(config, baseline));
  const [operation, setOperation] = useState(() => initialApplicationConfigurationDraftState(config));
  const [catalog, setCatalog] = useState(() => initialApplicationDraftModelCatalog(config, baseline.applicationId));
  const [list, setList] = useState(() => initialApplicationConfigurationDraftListState(config));
  const [bindings, setBindings] = useState<WorkflowRAGPromotionListResult>(() => initialWorkflowRAGPromotionListResult(promotionConfig));
  const [selectedBindingCandidateId, setSelectedBindingCandidateId] = useState("");
  const [bindingHandoffState, setBindingHandoffState] = useState<{ kind: "readOnly" | "loading" | "reloaded" | "unavailable" | "failed"; candidateId: string; bindingId?: string } | null>(null);
  const catalogController = useRef<AbortController | null>(null);
  const handledHandoffIdRef = useRef("");

  const operationStatusLabel = () => {
    switch (operation.status) {
      case "offline": return t($ => $.configuration.offlineStatus);
      case "unsaved": return t($ => $.configuration.unsaved);
      case "validating": return t($ => $.configuration.validatingStatus);
      case "invalid": return t($ => $.configuration.invalidStatus);
      case "valid": return t($ => $.configuration.validStatus);
      case "saving": return t($ => $.configuration.savingStatus);
      case "saved": return t($ => $.configuration.savedStatus);
      case "loading": return t($ => $.configuration.loadingStatus);
      case "restored": return t($ => $.configuration.restoredStatus);
      case "version_conflict": return t($ => $.configuration.versionConflictStatus);
      case "scope_denied": return t($ => $.configuration.scopeDeniedStatus);
      case "store_failure": return t($ => $.configuration.storeFailureStatus);
    }
  };
  const operationMessage = () => {
    if (operation.status === "version_conflict") return t($ => $.configuration.versionConflictNotice);
    if (operation.status === "scope_denied") return t($ => $.configuration.scopeDeniedNotice);
    if (operation.status === "store_failure") return t($ => $.configuration.storeFailureNotice);
    if (operation.failureCode) return t($ => $.configuration.configurationOperationFailed);
    switch (operation.status) {
      case "offline": return t($ => $.configuration.offlineDraftNotice);
      case "unsaved": return t($ => $.configuration.unsavedMemoryEdits);
      case "validating": return t($ => $.configuration.validatingSanitizedDraft);
      case "invalid": return t($ => $.configuration.resolveFindingsBeforeSaveOrHandoff);
      case "valid": return t($ => $.configuration.configurationValidForReview);
      case "saving": return t($ => $.configuration.savingSanitizedDraft);
      case "saved": return t($ => $.configuration.savedDraftVersion, { version: operation.currentDraftVersion });
      case "loading": return t($ => $.configuration.loadingDraftState);
      case "restored": return t($ => $.configuration.restoredDraftVersion, { version: operation.currentDraftVersion });
    }
  };
  const findingMessage = (code: string, field: string) => {
    if (code === "application_draft_secret_material_forbidden") return t($ => $.configuration.secretMaterialForbidden);
    if (code === "application_draft_model_unavailable") return t($ => $.configuration.modelUnavailableFinding);
    if (code === "application_draft_protocol_incompatible") return t($ => $.configuration.protocolIncompatibleFinding);
    if (code !== "application_draft_payload_invalid") return t($ => $.configuration.configurationFinding, { field });
    switch (field) {
      case "display_name": return t($ => $.configuration.displayNameFinding);
      case "description": return t($ => $.configuration.descriptionFinding);
      case "application_kind": return t($ => $.configuration.applicationKindFinding);
      case "allowed_protocols": return t($ => $.configuration.allowedProtocolsFinding);
      case "default_model": return t($ => $.configuration.selectValidatedModelFinding);
      default: return t($ => $.configuration.configurationFinding, { field });
    }
  };
  const listMessage = list.status === "offline" ? t($ => $.configuration.offlineSavedDraftsNotice)
    : list.status === "idle" ? t($ => $.configuration.loadSavedDraftsNotice)
      : list.status === "loading" ? t($ => $.configuration.loadingSavedDrafts)
        : list.status === "failed" ? t($ => $.configuration.savedDraftsLoadFailed)
          : list.status === "empty" ? t($ => $.configuration.noSavedDrafts)
            : t($ => $.configuration.loadedSavedDraftCount, { count: list.summaries.length });
  const catalogMessage = catalog.status === "loading" ? t($ => $.configuration.loadingDraftModels)
    : catalog.status === "ready" ? t($ => $.configuration.loadedValidatedModelCount, { count: catalog.models.length })
      : catalog.status === "failed" ? t($ => $.configuration.modelCatalogLoadFailed)
        : t($ => $.configuration.modelCatalogNotLoaded);

  useEffect(() => {
    catalogController.current?.abort();
    catalogController.current = null;
    setDraft(createApplicationConfigurationDraft(config, baseline));
    setOperation(initialApplicationConfigurationDraftState(config));
    setCatalog(initialApplicationDraftModelCatalog(config, baseline.applicationId));
    setList(initialApplicationConfigurationDraftListState(config));
    setBindings(initialWorkflowRAGPromotionListResult(promotionConfig));
    setSelectedBindingCandidateId("");
    setBindingHandoffState(null);
    handledHandoffIdRef.current = "";
  }, [baseline.applicationId]);

  useEffect(() => () => catalogController.current?.abort(), []);

  useEffect(() => {
    function applyValidatedCatalog(detail: ApplicationModelCatalogReadyDetail) {
      if (detail.applicationId !== baseline.applicationId) return;
      catalogController.current?.abort();
      catalogController.current = null;
      setCatalog({
        status: "ready",
        applicationId: detail.applicationId,
        models: detail.models,
        selectedModel: detail.selectedModel,
        failureCode: "",
        summary: `Reused ${detail.models.length} models validated by the Gateway Playground.`,
      });
      setDraft((current) => ({ ...current, defaultModel: detail.selectedModel }));
      setOperation((current) => ({
        ...current,
        status: "unsaved",
        summary: "The Playground model catalog is ready for configuration validation.",
        failureCode: "",
        validation: { state: "invalid", isValid: false, findings: [] },
      }));
    }
    function receiveValidatedCatalog(event: Event) {
      const detail = (event as CustomEvent<ApplicationModelCatalogReadyDetail>).detail;
      try {
        const normalized = createApplicationModelCatalogReadyDetail(
          detail?.applicationId ?? "",
          detail?.models ?? [],
          detail?.selectedModel ?? "",
        );
        applyValidatedCatalog(normalized);
      } catch {
        return;
      }
    }
    const latest = readLatestApplicationModelCatalogReady(baseline.applicationId);
    if (latest) applyValidatedCatalog(latest);
    window.addEventListener(APPLICATION_MODEL_CATALOG_READY_EVENT, receiveValidatedCatalog);
    return () => window.removeEventListener(APPLICATION_MODEL_CATALOG_READY_EVENT, receiveValidatedCatalog);
  }, [baseline.applicationId]);

  const differences = useMemo(() => compareApplicationConfigurationDraft(baseline, draft), [baseline, draft]);
  const currentValidation = useMemo(() => validateApplicationConfigurationDraft(draft, catalog.models), [catalog.models, draft]);
  const enabled = config.mode === "dev_application_draft_http";
  const mutationEnabled = enabled && !readOnly;
  const bindingEnabled = mutationEnabled && promotionConfig.mode === "dev_workflow_rag_promotion_http";
  const selectedBinding = bindings.summaries.find((item) => item.candidateId === selectedBindingCandidateId && item.bindingRef && item.eligibilityStatus === "eligible") ?? null;
  const bindingSourceReady = Boolean(selectedBinding && operation.currentDraftVersion === selectedBinding.sourceDraft.draftVersion && draft.draftId === selectedBinding.sourceDraft.draftId && draft.draftDigest === selectedBinding.sourceDraft.draftDigest);
  const handoffReady = operation.validation.isValid && currentValidation.isValid && catalog.status === "ready";

  useEffect(() => {
    if (!onEvidenceChange) return;
    const failed = operation.status === "version_conflict" || operation.status === "store_failure" || operation.status === "scope_denied";
    const saved = (operation.status === "saved" || operation.status === "restored") && operation.validation.isValid;
    onEvidenceChange({
      contributionId: "configuration_draft",
      status: failed ? "blocked" : saved ? "available" : "incomplete",
      coverage: saved || failed ? "complete" : draft.draftId ? "partial" : "none",
      evidenceRefs: draft.draftId && operation.currentDraftVersion > 0
        ? [{ kind: "draft", id: draft.draftId, version: operation.currentDraftVersion }]
        : [],
      missingEvidence: saved ? [] : [failed ? "Resolve the current configuration owner failure." : "Save and validate the current Application configuration draft."],
      blockers: failed ? [{
        code: operation.failureCode || "application_configuration_blocked",
        summary: operation.summary,
      }] : [],
      failureCodes: failed && operation.failureCode ? [operation.failureCode] : [],
    });
  }, [draft.draftId, onEvidenceChange, operation]);

  useEffect(() => {
    if (readOnly && enabled) void refreshList();
  }, [baseline.applicationId, readOnly]);

  useEffect(() => {
    if (!handoffId || !handoffBindingCandidateId || handledHandoffIdRef.current === handoffId) return;
    handledHandoffIdRef.current = handoffId;
    if (!bindingEnabled) {
      setBindingHandoffState({ kind: "readOnly", candidateId: handoffBindingCandidateId });
      onHandoffConsumed?.(handoffId);
      return;
    }
    setBindingHandoffState({ kind: "loading", candidateId: handoffBindingCandidateId });
    void loadApprovedBindings(handoffBindingCandidateId)
      .then((selected) => setBindingHandoffState(selected
        ? { kind: "reloaded", candidateId: handoffBindingCandidateId, bindingId: selected.bindingRef?.bindingId }
        : { kind: "unavailable", candidateId: handoffBindingCandidateId }))
      .catch(() => setBindingHandoffState({ kind: "failed", candidateId: handoffBindingCandidateId }))
      .finally(() => onHandoffConsumed?.(handoffId));
  }, [baseline.applicationId, bindingEnabled, handoffBindingCandidateId, handoffId, onHandoffConsumed]);

  function edit(patch: Partial<ApplicationConfigurationDraft>) {
    setDraft((current) => ({ ...current, ...patch }));
    setOperation((current) => ({ ...current, status: enabled ? "unsaved" : "offline", summary: "Application configuration has unsaved in-memory edits.", failureCode: "", validation: { state: "invalid", isValid: false, findings: [] } }));
  }

  async function loadModels() {
    if (readOnly) return;
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
    if (readOnly) return;
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
    if (readOnly) return;
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

  async function loadApprovedBindings(preferredCandidateId = ""): Promise<WorkflowRAGPromotionSummary | null> {
    if (!bindingEnabled) return null;
    const next = await listWorkflowRAGPromotionCandidates(promotionConfig, baseline.applicationId);
    setBindings(next);
    const selected = preferredCandidateId
      ? findExactEligibleWorkflowRAGPromotionBinding(next.summaries, preferredCandidateId)
      : next.summaries.find((item) => item.candidateState === "approved" && item.eligibilityStatus === "eligible" && item.bindingRef) ?? null;
    setSelectedBindingCandidateId(selected?.candidateId ?? "");
    return selected;
  }

  async function restoreBindingSource() {
    if (!selectedBinding) return;
    setOperation((current) => ({ ...current, status: "loading", summary: "Restoring the binding's exact source draft before attach or replace.", failureCode: "" }));
    const restored = await readApplicationConfigurationDraft(config, baseline.applicationId, selectedBinding.sourceDraft.draftId);
    if (!restored.draft || restored.state.currentDraftVersion !== selectedBinding.sourceDraft.draftVersion || restored.draft.draftDigest !== selectedBinding.sourceDraft.draftDigest) {
      setOperation({ ...restored.state, status: "store_failure", failureCode: "workflow_rag_promotion_draft_changed", summary: "The exact source draft is no longer current; binding attach failed closed." });
      return;
    }
    setDraft(restored.draft);
    setOperation(restored.state);
  }

  async function attachBinding() {
    if (!selectedBinding?.bindingRef || !bindingSourceReady) return;
    const bindingDraft: ApplicationConfigurationDraft = {
      ...draft,
      schemaVersion: "application_configuration_draft.v2",
      workflowRAGBindingRef: { ...selectedBinding.bindingRef },
    };
    setOperation((current) => ({ ...current, status: "saving", summary: "Attaching the approved immutable RAG binding as the only draft change.", failureCode: "" }));
    const saved = await saveApplicationConfigurationDraft(config, bindingDraft, selectedBinding.sourceDraft.draftVersion);
    setOperation(saved);
    if (saved.status !== "saved") return;
    const restored = await readApplicationConfigurationDraft(config, baseline.applicationId, bindingDraft.draftId);
    setOperation(restored.state);
    if (restored.draft) setDraft(restored.draft);
    await refreshList();
  }

  function continueAfterConflict() {
    if (operation.status !== "version_conflict") return;
    setOperation((current) => ({ ...current, status: "unsaved", failureCode: "", summary: `Local edits are preserved. The next save will compare against version ${current.currentDraftVersion}.` }));
  }

  function openIntegration() {
    if (readOnly || !handoffReady) return;
    requestApplicationApiIntegrationDraftHandoff(draft.applicationId, draft.defaultProtocol, draft.defaultModel);
    window.location.hash = "application-api-integration";
  }

  function openPlayground() {
    if (readOnly || !handoffReady) return;
    requestModelGatewayPlaygroundHandoff(draft.applicationId, draft.defaultProtocol, draft.defaultModel);
    window.location.hash = "model-gateway-playground";
  }

  function openPublishReview() {
    if (readOnly || operation.status !== "saved" && operation.status !== "restored") return;
    if (onOpenPublishReview) onOpenPublishReview(draft.draftId);
    else window.location.hash = "application-publish-review";
  }

  return (
    <section className="application-configuration-draft" id="application-configuration-draft" aria-labelledby="application-configuration-draft-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.configuration.configurationDraftTitle)}</p><h4 id="application-configuration-draft-title">{t($ => $.configuration.configurationDraftSubtitle)}</h4></div>
        <span className={`status-badge ${readOnly || operation.status === "saved" || operation.status === "valid" || operation.status === "restored" ? "good" : operation.status === "invalid" || operation.status === "version_conflict" || operation.status === "store_failure" ? "bad" : "neutral"}`}>{readOnly ? t($ => $.configuration.archivedReadOnly) : operationStatusLabel()}</span>
      </div>

      <div className="application-draft-scope">
        <article><span>{t($ => $.configuration.application)}</span><strong>{baseline.displayName}</strong><code>{baseline.applicationId}</code></article>
        <article><span>{t($ => $.configuration.workspace)}</span><strong>{config.workspaceId}</strong><p>{enabled ? t($ => $.configuration.developmentRepositoryEnabled) : t($ => $.configuration.offlineMemoryOnly)}</p></article>
        <article><span>{t($ => $.configuration.version)}</span><strong>{operation.currentDraftVersion || t($ => $.configuration.unsaved)}</strong><p>{t($ => $.configuration.formalApplicationReadOnly)}</p></article>
      </div>
      {bindingHandoffState ? <p className="boundary-note" role="status">{bindingHandoffState.kind === "readOnly" ? t($ => $.configuration.bindingHandoffReadOnly, { candidateId: bindingHandoffState.candidateId }) : bindingHandoffState.kind === "loading" ? t($ => $.configuration.loadingExactBindingCandidate, { candidateId: bindingHandoffState.candidateId }) : bindingHandoffState.kind === "reloaded" ? t($ => $.configuration.exactBindingReloaded, { bindingId: bindingHandoffState.bindingId ?? "" }) : bindingHandoffState.kind === "unavailable" ? t($ => $.configuration.bindingUnavailableNoFallback, { candidateId: bindingHandoffState.candidateId }) : t($ => $.configuration.bindingReloadFailedNoFallback, { candidateId: bindingHandoffState.candidateId })}</p> : null}

      <div className="application-draft-layout">
        <article className="application-draft-editor">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configuration.sanitizedConfiguration)}</p><h5>{draft.draftId}</h5></div><span className="status-badge neutral">{draft.schemaVersion}</span></div>
          <label>{t($ => $.configuration.displayName)}<input value={draft.displayName} onChange={(event) => edit({ displayName: event.target.value })} maxLength={120} disabled={readOnly} /></label>
          <label>{t($ => $.configuration.description)}<textarea value={draft.description} onChange={(event) => edit({ description: event.target.value })} maxLength={1000} rows={4} placeholder={t($ => $.configuration.publicPurposePlaceholder)} disabled={readOnly} /></label>
          <label>{t($ => $.configuration.applicationKind)}<select value={draft.applicationKind} onChange={(event) => edit({ applicationKind: event.target.value })} disabled={readOnly}><option value="workflow_copilot">{t($ => $.configuration.workflowCopilot)}</option><option value="docs_qa">{t($ => $.configuration.docsQa)}</option><option value="agent">{t($ => $.configuration.agent)}</option><option value="prompt_application">{t($ => $.configuration.promptApplication)}</option></select></label>
          <fieldset disabled={readOnly}><legend>{t($ => $.configuration.allowedProtocols)}</legend>{protocols.map((protocol) => <label key={protocol.id}><input type="checkbox" checked={draft.allowedProtocols.includes(protocol.id)} onChange={(event) => edit({ allowedProtocols: event.target.checked ? [...draft.allowedProtocols, protocol.id] : draft.allowedProtocols.filter((item) => item !== protocol.id) })} />{protocol.label}</label>)}</fieldset>
          <label>{t($ => $.configuration.defaultProtocol)}<select value={draft.defaultProtocol} onChange={(event) => edit({ defaultProtocol: event.target.value as ApplicationApiProtocol })} disabled={readOnly}>{protocols.map((protocol) => <option key={protocol.id} value={protocol.id}>{protocol.label}</option>)}</select></label>
          <label>{t($ => $.configuration.defaultModel)}<select value={draft.defaultModel} disabled={readOnly || catalog.models.length === 0} onChange={(event) => edit({ defaultModel: event.target.value })}><option value="">{t($ => $.configuration.noValidatedModel)}</option>{readOnly && draft.defaultModel && !catalog.models.some((model) => model.id === draft.defaultModel) ? <option value={draft.defaultModel}>{draft.defaultModel} {t($ => $.configuration.savedDraftSuffix)}</option> : null}{catalog.models.map((model) => <option key={model.id} value={model.id}>{model.id} · {model.protocols.join(", ")}</option>)}</select></label>
          {readOnly ? <p className="boundary-note">{t($ => $.configuration.archivedDraftReadOnlyNotice)}</p> : <div className="application-draft-actions"><button type="button" onClick={() => void loadModels()} disabled={!mutationEnabled || catalog.status === "loading"}>{catalog.status === "loading" ? t($ => $.configuration.loadingModels) : t($ => $.configuration.loadOrRefreshModels)}</button><button type="button" onClick={() => void validateDraft()}>{t($ => $.configuration.validateConfiguration)}</button><button type="button" onClick={() => void saveDraft()} disabled={!mutationEnabled || operation.status === "saving" || operation.status === "validating" || operation.status === "version_conflict"}>{t($ => $.configuration.saveDraft)}</button></div>}
          <p className="boundary-note">{catalogMessage}</p>
        </article>

        <article className="application-draft-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configuration.reviewState)}</p><h5>{operationMessage()}</h5></div><span className={`status-badge ${operation.validation.isValid ? "good" : "neutral"}`}>{operation.validation.isValid ? t($ => $.configuration.validStatus) : t($ => $.configuration.invalidStatus)}</span></div>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          {operation.validation.findings.length ? <ul className="application-draft-findings">{operation.validation.findings.map((finding) => <li key={`${finding.code}-${finding.field}`}><strong>{finding.field}</strong><span>{finding.code}</span><p>{findingMessage(finding.code, finding.field)}</p></li>)}</ul> : <p className="boundary-note">{t($ => $.configuration.noValidationFindings)}</p>}
          {operation.status === "version_conflict" ? <div className="application-draft-conflict"><strong>{t($ => $.configuration.savedVersionIsNewer, { version: operation.currentDraftVersion })}</strong><p>{t($ => $.configuration.memoryEditsPreserved)}</p><button type="button" onClick={continueAfterConflict}>{t($ => $.configuration.continueLocalEdits)}</button>{list.summaries[0] ? <button type="button" onClick={() => void restoreDraft(list.summaries[0].draftId)}>{t($ => $.configuration.restoreSavedVersion)}</button> : null}</div> : null}
          {!readOnly ? <div className="application-draft-handoff"><button type="button" disabled={!handoffReady} onClick={openIntegration}>{t($ => $.configuration.openApiIntegration)}</button><button type="button" disabled={!handoffReady} onClick={openPlayground}>{t($ => $.configuration.testInPlayground)}</button><button type="button" disabled={operation.status !== "saved" && operation.status !== "restored"} onClick={openPublishReview}>{t($ => $.configuration.openPublishReview)}</button></div> : null}
          <p className="boundary-note">{t($ => $.configuration.handoffScopeNotice)}</p>
        </article>
      </div>

      <div className="application-draft-lower-grid">
        <article className="application-draft-diff"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configuration.configurationComparison)}</p><h5>{t($ => $.configuration.readModelToDraft)}</h5></div><span className="status-badge neutral">{differences.filter((item) => item.changed).length} {t($ => $.configuration.changed)}</span></div>{differences.map((difference) => <div className={difference.changed ? "changed" : "unchanged"} key={difference.field}><strong>{difference.field}</strong><span>{difference.before}</span><span>→</span><span>{difference.after}</span></div>)}</article>
        <article className="application-draft-saved"><div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configuration.savedDevelopmentDrafts)}</p><h5>{listMessage}</h5></div><button type="button" onClick={() => void refreshList()} disabled={!enabled || list.status === "loading"}>{t($ => $.configuration.refresh)}</button></div>{list.failureCode ? <p className="failure-summary">{list.failureCode}</p> : null}{list.summaries.map((summary) => <button type="button" className="application-draft-summary" key={summary.draftId} onClick={() => void restoreDraft(summary.draftId)}><strong>{summary.displayName}</strong><span>v{summary.draftVersion} · {summary.defaultProtocol} · {summary.defaultModel}</span><small>{summary.updatedAt} · {summary.updatedByActorRef}</small></button>)}</article>
      </div>

      {baseline.applicationKind === "workflow_copilot" || baseline.applicationKind === "docs_qa" ? <article className="application-draft-rag-binding">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configuration.workflowRagBinding)}</p><h5>{t($ => $.configuration.explicitAttachOrReplace)}</h5></div><button type="button" onClick={() => void loadApprovedBindings()} disabled={!bindingEnabled}>{t($ => $.configuration.loadApprovedBindings)}</button></div>
        <label>{t($ => $.configuration.eligibleImmutableBinding)}<select value={selectedBindingCandidateId} onChange={(event) => setSelectedBindingCandidateId(event.target.value)} disabled={!bindingEnabled || bindings.summaries.length === 0}><option value="">{t($ => $.configuration.noApprovedBindingSelected)}</option>{bindings.summaries.map((item) => <option key={item.candidateId} value={item.candidateId} disabled={item.candidateState !== "approved" || item.eligibilityStatus !== "eligible" || !item.bindingRef}>{item.bindingRef?.bindingId ?? item.candidateId} · {item.candidateState} · {item.eligibilityStatus}</option>)}</select></label>
        {selectedBinding ? <div className="application-draft-binding-evidence"><strong>{t($ => $.configuration.sourceDraft)}{selectedBinding.sourceDraft.draftId} · v{selectedBinding.sourceDraft.draftVersion}</strong><code>{selectedBinding.sourceDraft.draftDigest}</code><code>{selectedBinding.bindingRef?.bindingDigest}</code></div> : <p className="boundary-note">{t($ => $.configuration.approvePromotionCandidateFirst)}</p>}
        {bindings.failureCode ? <p className="failure-summary">{bindings.failureCode}</p> : null}
        <div className="application-draft-actions"><button type="button" onClick={() => void restoreBindingSource()} disabled={!selectedBinding}>{t($ => $.configuration.restoreExactSourceDraft)}</button><button type="button" onClick={() => void attachBinding()} disabled={!bindingSourceReady || !selectedBinding?.bindingRef}>{t($ => $.configuration.attachImmutableBinding)}</button></div>
        {draft.workflowRAGBindingRef ? <p className="binding-status"><strong>{t($ => $.configuration.currentDraftBinding)}</strong><code>{draft.workflowRAGBindingRef.bindingId} · v{draft.workflowRAGBindingRef.bindingVersion}</code><code>{draft.workflowRAGBindingRef.bindingDigest}</code></p> : null}
        <p className="boundary-note">{t($ => $.configuration.attachBindingCasNotice)}</p>
      </article> : null}

      <p className="boundary-note">{t($ => $.configuration.draftBoundaryNotice)}</p>
    </section>
  );
}
