import "../../i18n/workflowRAGEvaluationResources.ts";
import { useTranslation } from "react-i18next";
import { ragEvaluationState, ragEvaluationMessage, ragEvaluationFailureCode, ragEvaluationMetric } from "./workflowRAGEvaluationMessages.ts";
import { formatDisplayDate, formatDisplayNumber } from "../../i18n/formatters.ts";
import { resolveUiLocale } from "../../i18n/localePreference.ts";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  archiveWorkflowRAGEvaluationDataset,
  createWorkflowRAGCandidateReview,
  createWorkflowRAGEvaluationDataset,
  listWorkflowRAGCandidateReviews,
  listWorkflowRAGEvaluationDatasets,
  readWorkflowRAGCandidateReview,
  readWorkflowRAGEvaluationConfig,
  readWorkflowRAGEvaluationDataset,
  validateWorkflowRAGEvaluationDatasetInput,
  versionWorkflowRAGEvaluationDataset,
  type WorkflowRAGCandidateReviewSummary,
  type WorkflowRAGEvaluationClassification,
  type WorkflowRAGEvaluationDatasetInput,
  type WorkflowRAGEvaluationDatasetResource,
  type WorkflowRAGEvaluationDatasetVersion,
  type WorkflowRAGEvaluationLifecycle,
  type WorkflowRAGEvaluationMetrics,
  type WorkflowRAGEvaluationOperationResult,
  type WorkflowRAGEvaluationSampleInput,
  type WorkflowRAGEvaluationSnapshotBinding,
} from "./workflowRAGEvaluationDatasetConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const config = readWorkflowRAGEvaluationConfig();
const DEFAULT_THRESHOLDS: WorkflowRAGEvaluationMetrics = {
  hitAtK: 0.8,
  expectedRecallAtK: 0.8,
  requiredOfficialRecallAtK: 1,
  meanReciprocalRank: 0.7,
  noEvidenceAccuracy: 1,
  samplePassRate: 0.8,
};

type DatasetEditor = {
  datasetKey: string;
  displayName: string;
  contentClassification: WorkflowRAGEvaluationClassification;
  baselineSnapshotJSON: string;
  thresholds: WorkflowRAGEvaluationMetrics;
  reviewSummary: string;
  samplesJSON: string;
};

type DatasetCollection = {
  active: WorkflowRAGEvaluationDatasetResource[];
  archived: WorkflowRAGEvaluationDatasetResource[];
  activeCursor: string;
  archivedCursor: string;
  failureCode: string;
};

export default function WorkflowRAGEvaluationDatasetPanel({
  applicationId,
  applicationName,
  applicationActive,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t, i18n } = useTranslation("workflow");
  const [collection, setCollection] = useState<DatasetCollection>(emptyCollection);
  const [filter, setFilter] = useState<WorkflowRAGEvaluationLifecycle>("active");
  const [selectedResource, setSelectedResource] = useState<WorkflowRAGEvaluationDatasetResource | null>(null);
  const [selectedVersion, setSelectedVersion] = useState<WorkflowRAGEvaluationDatasetVersion | null>(null);
  const [reviews, setReviews] = useState<WorkflowRAGCandidateReviewSummary[]>([]);
  const [selectedReview, setSelectedReview] = useState<WorkflowRAGCandidateReviewSummary | null>(null);
  const [reviewCursor, setReviewCursor] = useState("");
  const [reviewHistoryFailure, setReviewHistoryFailure] = useState("");
  const [editor, setEditor] = useState<DatasetEditor>(() => emptyEditor(applicationId));
  const [candidateJSON, setCandidateJSON] = useState(() => emptySnapshotJSON(applicationId));
  const [showCreate, setShowCreate] = useState(false);
  const [showArchiveConfirm, setShowArchiveConfirm] = useState(false);
  const [pending, setPending] = useState<"" | "listing" | "reading" | "creating" | "versioning" | "archiving" | "reviewing">("");
  const [operation, setOperation] = useState<WorkflowRAGEvaluationOperationResult | null>(null);

  const requestGeneration = useRef(0);
  const writeInFlight = useRef(false);
  const versionConflict = operation?.status === "version_conflict";
  const locale = resolveUiLocale(i18n.resolvedLanguage, []);
  const canRead = config.scopes.has("workflow_rag_evaluation_datasets:read");
  const canReadContent = canRead && config.scopes.has("workflow_rag_evaluation_datasets:read_content");
  const canWrite = applicationActive && config.scopes.has("workflow_rag_evaluation_datasets:write") && config.scopes.has("workflow_rag_snapshots:read");
  const canReview = applicationActive && canRead && config.scopes.has("workflow_rag_evaluation_datasets:review") && config.scopes.has("workflow_rag_snapshots:read");
  const canArchive = applicationActive && config.scopes.has("workflow_rag_evaluation_datasets:archive");
  const visibleResources = useMemo(() => filter === "active" ? collection.active : collection.archived, [collection, filter]);

  useEffect(() => {
    if (!onEvidenceChange) return;
    const rawFailure = collection.failureCode || reviewHistoryFailure || operation?.failureCode || "";
    const failureCode = ragEvaluationFailureCode(rawFailure) || (rawFailure ? "workflow_rag_evaluation_response_invalid" : "");
    const ownerFailed = Boolean(failureCode) && (operation?.status === "failed" || operation?.status === "scope_denied" || operation?.status === "version_conflict" || Boolean(collection.failureCode || reviewHistoryFailure));
    const compatible = selectedReview?.candidateStatus === "passed" &&
      (selectedReview.conclusion === "improved" || selectedReview.conclusion === "unchanged");
    onEvidenceChange({
      contributionId: "evaluation_review",
      status: ownerFailed ? "blocked" : compatible ? "available" : "incomplete",
      coverage: selectedReview || ownerFailed ? "complete" : collection.active.length ? "partial" : "none",
      evidenceRefs: selectedReview ? [{ kind: "evaluation", id: selectedReview.reviewId, version: selectedReview.datasetVersion }] : [],
      missingEvidence: compatible ? [] : ["Select a compatible candidate evaluation review with a passed candidate snapshot."],
      blockers: ownerFailed ? [{ code: failureCode, summary: operation?.summary || "Evaluation evidence could not be loaded from its owner." }] : [],
      failureCodes: ownerFailed ? [failureCode] : [],
    });
  }, [collection.active.length, collection.failureCode, onEvidenceChange, operation, reviewHistoryFailure, selectedReview]);

  useEffect(() => {
    const generation = ++requestGeneration.current;
    writeInFlight.current = false;
    setPending("");
    setReviewCursor("");
    setReviewHistoryFailure("");
    setCollection(emptyCollection());
    setSelectedResource(null);
    setSelectedVersion(null);
    setReviews([]);
    setSelectedReview(null);
    setEditor(emptyEditor(applicationId));
    setCandidateJSON(emptySnapshotJSON(applicationId));
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    if (config.mode === "offline" || !canRead || !applicationId.trim()) return () => { ++requestGeneration.current; };
    setPending("listing");
    Promise.all([
      listWorkflowRAGEvaluationDatasets(config, applicationId, "active"),
      listWorkflowRAGEvaluationDatasets(config, applicationId, "archived"),
    ]).then(([active, archived]) => {
      if (generation !== requestGeneration.current) return;
      setPending("");
      setCollection({ active: active.resources, archived: archived.resources, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode });
    });
    return () => { ++requestGeneration.current; writeInFlight.current = false; };
  }, [applicationId, canRead]);

  const refresh = async (generation = ++requestGeneration.current) => {
    const [active, archived] = await Promise.all([
      listWorkflowRAGEvaluationDatasets(config, applicationId, "active"),
      listWorkflowRAGEvaluationDatasets(config, applicationId, "archived"),
    ]);
    const next = { active: active.resources, archived: archived.resources, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode };
    if (generation !== requestGeneration.current) return next;
    setCollection(next);
    return next;
  };

  const selectDataset = async (resource: WorkflowRAGEvaluationDatasetResource) => {
    if (writeInFlight.current) return;
    const generation = ++requestGeneration.current;
    setSelectedResource(resource);
    setReviews([]);
    setReviewCursor("");
    setReviewHistoryFailure("");
    setSelectedVersion(null);
    setSelectedReview(null);
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setPending("reading");
    const [detail, history] = await Promise.all([
      canReadContent ? readWorkflowRAGEvaluationDataset(config, applicationId, resource.datasetId, resource.latestVersion) : Promise.resolve(null),
      listWorkflowRAGCandidateReviews(config, applicationId, resource.datasetId),
    ]);
    if (generation !== requestGeneration.current) return;
    setPending("");
    setReviews(history.reviews);
    setReviewCursor(history.nextCursor);
    setReviewHistoryFailure(history.failureCode);
    if (!detail) return;
    setOperation(detail);
    if (detail.version) {
      setSelectedVersion(detail.version);
      setEditor(editorFromVersion(detail.version));
    }
  };

  const submitCreate = async () => {
    if (writeInFlight.current || versionConflict) return;
    const input = parseEditor(editor, applicationId);
    if (typeof input === "string") return setOperation(localFailure(input));
    writeInFlight.current = true;
    const generation = ++requestGeneration.current;
    setPending("creating");
    const result = await createWorkflowRAGEvaluationDataset(config, applicationId, input);
    if (generation !== requestGeneration.current) return;
    writeInFlight.current = false;
    setPending("");
    setOperation(result);
    if (!result.resource || !result.version) return;
    setSelectedResource(result.resource);
    setSelectedVersion(result.version);
    setEditor(editorFromVersion(result.version));
    setShowCreate(false);
    setFilter("active");
    await refresh(generation);
  };

  const submitVersion = async () => {
    if (writeInFlight.current || versionConflict) return;
    if (!selectedResource) return;
    const input = parseEditor(editor, applicationId);
    if (typeof input === "string") return setOperation(localFailure(input));
    writeInFlight.current = true;
    const generation = ++requestGeneration.current;
    setPending("versioning");
    const result = await versionWorkflowRAGEvaluationDataset(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion, input);
    if (generation !== requestGeneration.current) return;
    writeInFlight.current = false;
    setPending("");
    setOperation(result);
    if (!result.resource || !result.version) return;
    setSelectedResource(result.resource);
    setSelectedVersion(result.version);
    setEditor(editorFromVersion(result.version));
    await refresh(generation);
  };

  const submitArchive = async () => {
    if (writeInFlight.current || versionConflict) return;
    if (!selectedResource) return;
    writeInFlight.current = true;
    const generation = ++requestGeneration.current;
    setPending("archiving");
    const result = await archiveWorkflowRAGEvaluationDataset(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion);
    if (generation !== requestGeneration.current) return;
    writeInFlight.current = false;
    setPending("");
    setOperation(result);
    setShowArchiveConfirm(false);
    if (!result.resource) return;
    setSelectedResource(result.resource);
    setFilter("archived");
    await refresh(generation);
  };

  const submitReview = async () => {
    if (writeInFlight.current || versionConflict) return;
    if (!selectedResource) return;
    const candidate = parseSnapshotBinding(candidateJSON, applicationId);
    if (typeof candidate === "string") return setOperation(localFailure(candidate));
    writeInFlight.current = true;
    const generation = ++requestGeneration.current;
    setPending("reviewing");
    const result = await createWorkflowRAGCandidateReview(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion, selectedResource.latestDigest, candidate);
    if (generation !== requestGeneration.current) return;
    writeInFlight.current = false;
    setPending("");
    setOperation(result);
    if (!result.review) return;
    setSelectedReview(result.review);
    const history = await listWorkflowRAGCandidateReviews(config, applicationId, selectedResource.datasetId);
    if (generation !== requestGeneration.current) return;
    setReviews(history.reviews);
    setReviewCursor(history.nextCursor);
    setReviewHistoryFailure(history.failureCode);
  };

  const selectReview = async (review: WorkflowRAGCandidateReviewSummary) => {
    if (writeInFlight.current) return;
    const generation = ++requestGeneration.current;
    if (!selectedResource) return;
    setPending("reading");
    const result = await readWorkflowRAGCandidateReview(config, applicationId, selectedResource.datasetId, review.reviewId);
    if (generation !== requestGeneration.current) return;
    setPending("");
    setOperation(result);
    setSelectedReview(result.review);
  };

  const loadMoreDatasets = async () => {
    if (writeInFlight.current) return;
    const generation = ++requestGeneration.current;
    const cursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
    if (!cursor) return;
    setPending("listing");
    const page = await listWorkflowRAGEvaluationDatasets(config, applicationId, filter, cursor);
    if (generation !== requestGeneration.current) return;
    setPending("");
    setCollection((current) => filter === "active"
      ? { ...current, active: mergeResources(current.active, page.resources), activeCursor: page.nextCursor, failureCode: page.failureCode }
      : { ...current, archived: mergeResources(current.archived, page.resources), archivedCursor: page.nextCursor, failureCode: page.failureCode });
  };

  const loadMoreReviews = async () => {
    if (writeInFlight.current) return;
    const generation = ++requestGeneration.current;
    if (!selectedResource || !reviewCursor) return;
    setPending("reading");
    const page = await listWorkflowRAGCandidateReviews(config, applicationId, selectedResource.datasetId, reviewCursor);
    if (generation !== requestGeneration.current) return;
    setPending("");
    setReviews((current) => mergeReviews(current, page.reviews));
    setReviewCursor(page.nextCursor);
    setReviewHistoryFailure(page.failureCode);
  };

  const reloadEvidence = async () => {
    if (writeInFlight.current) return;
    const generation = ++requestGeneration.current;
    setPending("listing");
    const next = await refresh(generation);
    if (generation !== requestGeneration.current) return;
    setPending("");
    if (next.failureCode) return;
    const current = [...next.active, ...next.archived].find(item => item.datasetId === selectedResource?.datasetId);
    if (current) await selectDataset(current);
    else setOperation(null);
  };

  if (config.mode === "offline") return <BoundaryPanel status="offline" summary={t($ => $.ragEvaluation.offlineBoundary)} />;
  if (!canRead || !applicationId.trim()) return <BoundaryPanel status="scope_denied" summary={t($ => $.ragEvaluation.scopeBoundary)} />;

  const datasetCursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
  return (
    <section className="workflow-rag-evaluation-panel" id="workflow-rag-evaluation-panel" aria-labelledby="workflow-rag-evaluation-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.ragEvaluation.eyebrow)}</p><h4 id="workflow-rag-evaluation-title">{t($ => $.ragEvaluation.title)}</h4></div>
        <span className={`status-badge ${collection.failureCode ? "bad" : "good"}`}>{ragEvaluationState(t, pending || (collection.failureCode ? "failed" : "ready"))}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>{t($ => $.ragEvaluation.application)}</span><strong>{applicationName || applicationId}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.ragEvaluation.detail)}</span><strong>{canReadContent ? t($ => $.ragEvaluation.authorized) : t($ => $.ragEvaluation.metadataOnly)}</strong><small>{t($ => $.ragEvaluation.exactQuery)}</small></article>
        <article><span>{t($ => $.ragEvaluation.candidateReview)}</span><strong>{canReview ? t($ => $.ragEvaluation.enabled) : t($ => $.ragEvaluation.readOnly)}</strong><small>{t($ => $.ragEvaluation.rankingBoundary)}</small></article>
        <article><span>{t($ => $.ragEvaluation.executionBoundary)}</span><strong>{t($ => $.ragEvaluation.zeroCalls)}</strong><small>{t($ => $.ragEvaluation.noExecution)}</small></article>
      </div>

      {!applicationActive ? <p className="workflow-rag-boundary-note">{t($ => $.ragEvaluation.archivedApplication)}</p> : null}
      {!canReadContent ? <p className="workflow-rag-boundary-note">{t($ => $.ragEvaluation.contentPermission)}</p> : null}
      {collection.failureCode ? <p className="workflow-rag-failure" role="alert"><code>{ragEvaluationFailureCode(collection.failureCode)}</code> {ragEvaluationMessage(t, { status: "failed", failureCode: collection.failureCode })}</p> : null}

      <div className="workflow-rag-toolbar">
        <div className="workflow-rag-filter" aria-label={t($ => $.ragEvaluation.lifecycleFilter)}>
          {LIFECYCLES.map((state) => <button key={state} type="button" className={filter === state ? "selected" : ""} onClick={() => setFilter(state)}>{ragEvaluationState(t, state)}</button>)}
        </div>
        <button type="button" disabled={pending !== "" || versionConflict || !canWrite} onClick={() => { ++requestGeneration.current; setShowArchiveConfirm(false); setReviewCursor(""); setReviewHistoryFailure(""); setShowCreate(true); setSelectedResource(null); setSelectedVersion(null); setSelectedReview(null); setReviews([]); setEditor(emptyEditor(applicationId)); setOperation(null); }}>{t($ => $.ragEvaluation.newDataset)}</button>
      </div>

      <div className="workflow-rag-evaluation-layout">
        <div className="workflow-rag-list" aria-label={t($ => $.ragEvaluation.datasetList, { value: ragEvaluationState(t, filter) })}>
          {visibleResources.map((resource) => (
            <button key={resource.datasetId} type="button" className={selectedResource?.datasetId === resource.datasetId ? "selected" : ""} disabled={writeInFlight.current} onClick={() => void selectDataset(resource)}>
              <span><strong>{resource.displayName}</strong><code>{resource.datasetKey} {t($ => $.ragEvaluation.versionSeparator)}{resource.latestVersion}</code></span>
              <span><small>{t($ => $.ragEvaluation.sampleCount, { count: resource.sampleCount })}</small><small>{ragEvaluationState(t, resource.contentClassification)}</small></span>
            </button>
          ))}
          {!visibleResources.length && pending !== "listing" ? <p>{t($ => $.ragEvaluation.emptyDatasets, { lifecycle: ragEvaluationState(t, filter) })}</p> : null}
          {datasetCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMoreDatasets()}>{t($ => $.ragEvaluation.nextPage)}</button> : null}
        </div>

        <div className="workflow-rag-evaluation-workspace">
          {showCreate || selectedVersion ? <DatasetEditorFields editor={editor} disabled={pending !== "" || (!showCreate && !canWrite)} immutableKey={Boolean(selectedVersion)} onChange={setEditor} /> : selectedResource ? <ResourceEvidence resource={selectedResource} /> : <article className="workflow-rag-empty"><strong>{t($ => $.ragEvaluation.selectDataset)}</strong><p>{t($ => $.ragEvaluation.selectionHint)}</p></article>}

          {showCreate ? <div className="workflow-rag-actions"><button type="button" disabled={pending !== "" || versionConflict || !canWrite} onClick={() => void submitCreate()}>{pending === "creating" ? t($ => $.ragEvaluation.creating) : t($ => $.ragEvaluation.create)}</button><button type="button" disabled={pending !== ""} onClick={() => setShowCreate(false)}>{t($ => $.ragEvaluation.cancel)}</button></div> : null}
          {selectedResource && selectedVersion && selectedResource.lifecycleState === "active" ? (
            <div className="workflow-rag-actions"><button type="button" disabled={pending !== "" || versionConflict || !canWrite} onClick={() => void submitVersion()}>{pending === "versioning" ? t($ => $.ragEvaluation.versioning) : t($ => $.ragEvaluation.createVersion, { value: selectedResource.latestVersion + 1 })}</button><button type="button" className="danger-action" disabled={pending !== "" || versionConflict || !canArchive} onClick={() => setShowArchiveConfirm(true)}>{t($ => $.ragEvaluation.archive)}</button></div>
          ) : null}
          {showArchiveConfirm ? <div className="workflow-rag-archive-confirm" role="alert"><p>{t($ => $.ragEvaluation.archiveConfirmation)}</p><button type="button" className="danger-action" disabled={pending !== "" || versionConflict || !canArchive} onClick={() => void submitArchive()}>{t($ => $.ragEvaluation.confirmArchive)}</button><button type="button" onClick={() => setShowArchiveConfirm(false)}>{t($ => $.ragEvaluation.cancel)}</button></div> : null}

          {selectedResource ? (
            <section className="workflow-rag-candidate-review" aria-labelledby="workflow-rag-candidate-review-title">
              <div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.ragEvaluation.exactBinding)}</p><h5 id="workflow-rag-candidate-review-title">{t($ => $.ragEvaluation.reviewTitle)}</h5></div><span className="status-badge neutral">{t($ => $.ragEvaluation.recordCount, { count: reviews.length })}</span></div>
              <label><span>{t($ => $.ragEvaluation.candidateBinding)}</span><textarea value={candidateJSON} disabled={pending !== "" || versionConflict || !canReview || selectedResource.lifecycleState !== "active"} spellCheck={false} onChange={(event) => setCandidateJSON(event.currentTarget.value)} /></label>
              <button type="button" disabled={pending !== "" || versionConflict || !canReview || selectedResource.lifecycleState !== "active"} onClick={() => void submitReview()}>{pending === "reviewing" ? t($ => $.ragEvaluation.reviewing) : t($ => $.ragEvaluation.reviewVersion, { latestVersion: selectedResource.latestVersion })}</button>
              <div className="workflow-rag-review-history">
                {reviews.map((review) => <button key={review.reviewId} type="button" className={selectedReview?.reviewId === review.reviewId ? "selected" : ""} disabled={writeInFlight.current} onClick={() => void selectReview(review)}><span><strong>{ragEvaluationState(t, review.conclusion)}</strong><code>{review.candidateSnapshot.ragRef}</code></span><small title={review.createdAt}>{formatDisplayDate(review.createdAt, locale, { timeZone: "UTC" })}</small></button>)}
                {reviewHistoryFailure ? <p role="alert" className="workflow-rag-failure">{ragEvaluationMessage(t, { status: "failed", failureCode: reviewHistoryFailure })} <code>{ragEvaluationFailureCode(reviewHistoryFailure)}</code></p> : !reviews.length ? <p>{t($ => $.ragEvaluation.noReviews)}</p> : null}
                {reviewCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMoreReviews()}>{t($ => $.ragEvaluation.moreReviews)}</button> : null}
              </div>
              {selectedReview ? <ReviewEvidence review={selectedReview} /> : null}
            </section>
          ) : null}

          {operation ? <OperationEvidence operation={operation} /> : null}
          {operation?.failureCode || collection.failureCode || reviewHistoryFailure ? <button type="button" disabled={pending !== ""} onClick={() => void reloadEvidence()}>{t($ => $.ragEvaluation.reload)}</button> : null}
        </div>
      </div>
    </section>
  );
}

const LIFECYCLES: WorkflowRAGEvaluationLifecycle[] = ["active", "archived"];

function DatasetEditorFields({ editor, disabled, immutableKey, onChange }: { editor: DatasetEditor; disabled: boolean; immutableKey: boolean; onChange: (value: DatasetEditor) => void }) {
  const { t } = useTranslation("workflow");
  return <div className="workflow-rag-evaluation-fields">
    <label><span>{t($ => $.ragEvaluation.datasetKey)}</span><input value={editor.datasetKey} disabled={disabled || immutableKey} maxLength={48} onChange={(event) => onChange({ ...editor, datasetKey: event.currentTarget.value })} /></label>
    <label><span>{t($ => $.ragEvaluation.displayName)}</span><input value={editor.displayName} disabled={disabled} maxLength={120} onChange={(event) => onChange({ ...editor, displayName: event.currentTarget.value })} /></label>
    <label><span>{t($ => $.ragEvaluation.classification)}</span><select value={editor.contentClassification} disabled={disabled} onChange={(event) => onChange({ ...editor, contentClassification: event.currentTarget.value as WorkflowRAGEvaluationClassification })}><option value="synthetic_public">{t($ => $.ragEvaluation.syntheticPublic)}</option><option value="workspace_internal">{t($ => $.ragEvaluation.workspaceInternal)}</option></select></label>
    <label className="wide"><span>{t($ => $.ragEvaluation.baselineBinding)}</span><textarea value={editor.baselineSnapshotJSON} disabled={disabled} spellCheck={false} onChange={(event) => onChange({ ...editor, baselineSnapshotJSON: event.currentTarget.value })} /></label>
    <fieldset className="workflow-rag-thresholds"><legend>{t($ => $.ragEvaluation.thresholds)}</legend>{(Object.keys(editor.thresholds) as Array<keyof WorkflowRAGEvaluationMetrics>).map((key) => <label key={key}><span title={key}>{ragEvaluationMetric(t, key)}</span><input type="number" min="0" max="1" step="0.01" value={editor.thresholds[key]} disabled={disabled} onChange={(event) => onChange({ ...editor, thresholds: { ...editor.thresholds, [key]: Number(event.currentTarget.value) } })} /></label>)}</fieldset>
    <label className="wide"><span>{t($ => $.ragEvaluation.reviewSummary)}</span><textarea value={editor.reviewSummary} disabled={disabled} maxLength={512} onChange={(event) => onChange({ ...editor, reviewSummary: event.currentTarget.value })} /></label>
    <label className="wide"><span id="rag-evaluation-samples-label">{t($ => $.ragEvaluation.samplesJSON)}</span><textarea aria-labelledby="rag-evaluation-samples-label" aria-describedby="rag-evaluation-samples-help" value={editor.samplesJSON} disabled={disabled} spellCheck={false} onChange={(event) => onChange({ ...editor, samplesJSON: event.currentTarget.value })} /><small id="rag-evaluation-samples-help">{t($ => $.ragEvaluation.replacementHint)}</small></label>
  </div>;
}

function ResourceEvidence({ resource }: { resource: WorkflowRAGEvaluationDatasetResource }) {
  const { t, i18n } = useTranslation("workflow"); return <article className="workflow-rag-record"><div><strong>{resource.displayName}</strong><span className="status-badge neutral">{ragEvaluationState(t, resource.lifecycleState)}</span></div><dl><div><dt>{t($ => $.ragEvaluation.dataset)}</dt><dd>{resource.datasetId}</dd></div><div><dt>{t($ => $.ragEvaluation.version)}</dt><dd>{t($ => $.ragEvaluation.versionPrefix)}{resource.latestVersion}</dd></div><div><dt>{t($ => $.ragEvaluation.digest)}</dt><dd>{resource.latestDigest}</dd></div><div><dt>{t($ => $.ragEvaluation.baseline)}</dt><dd>{resource.baselineSnapshot.ragRef}</dd></div><div><dt>{t($ => $.ragEvaluation.samples)}</dt><dd>{formatDisplayNumber(resource.sampleCount, resolveUiLocale(i18n.resolvedLanguage, []))}</dd></div><div><dt>{t($ => $.ragEvaluation.classification)}</dt><dd>{ragEvaluationState(t, resource.contentClassification)}</dd></div></dl></article>; }

function ReviewEvidence({ review }: { review: WorkflowRAGCandidateReviewSummary }) {
  const { t, i18n } = useTranslation("workflow"); return <article className={`workflow-rag-review-evidence ${review.conclusion === "regressed" ? "failed" : ""}`}><div><strong>{ragEvaluationState(t, review.conclusion)}</strong><span>{ragEvaluationState(t, review.baselineStatus)} → {ragEvaluationState(t, review.candidateStatus)}</span></div><dl>{(Object.keys(review.metricDelta) as Array<keyof WorkflowRAGEvaluationMetrics>).map((key) => <div key={key}><dt title={key}>{ragEvaluationMetric(t, key)}</dt><dd>{formatDisplayNumber(review.metricDelta[key], resolveUiLocale(i18n.resolvedLanguage, []), { minimumFractionDigits: 4, maximumFractionDigits: 4, signDisplay: "exceptZero" })}</dd></div>)}</dl><p>{t($ => $.ragEvaluation.sampleChanges)}: {review.sampleChanges.map((sample) => `${sample.sampleId}: ${ragEvaluationState(t, sample.change)}`).join(" · ")}</p><p>{t($ => $.ragEvaluation.findings)} +[{review.addedFindingCodes.join(", ") || t($ => $.ragEvaluation.none)}] −[{review.removedFindingCodes.join(", ") || t($ => $.ragEvaluation.none)}]</p><code>{review.reviewId}</code></article>; }

function OperationEvidence({ operation }: { operation: WorkflowRAGEvaluationOperationResult }) {
  const { t } = useTranslation("workflow"); return <article className={`workflow-rag-operation ${operation.status === "failed" || operation.status === "version_conflict" ? "failed" : ""}`} aria-live="polite"><strong>{ragEvaluationState(t, operation.status)}</strong><p>{ragEvaluationMessage(t, operation)}</p>{operation.failureCode ? <code>{ragEvaluationFailureCode(operation.failureCode)}</code> : null}{operation.status === "version_conflict" ? <small>{t($ => $.ragEvaluation.conflictCurrent, { version: operation.currentLatestVersion, lifecycle: ragEvaluationState(t, operation.currentLifecycleState) })}</small> : null}</article>; }
function BoundaryPanel({ status, summary }: { status: string; summary: string }) {
  const { t } = useTranslation("workflow"); return <section className="workflow-rag-evaluation-panel offline" aria-label={t($ => $.ragEvaluation.panelLabel)}><div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.ragEvaluation.eyebrow)}</p><h4>{t($ => $.ragEvaluation.notEnabled)}</h4></div><span className="status-badge neutral">{ragEvaluationState(t, status)}</span></div><p>{summary}</p></section>; }

function parseEditor(editor: DatasetEditor, applicationId: string): WorkflowRAGEvaluationDatasetInput | string {
  let binding: unknown;
  let samples: unknown;
  try { binding = JSON.parse(editor.baselineSnapshotJSON); samples = JSON.parse(editor.samplesJSON); } catch { return "workflow_rag_evaluation_payload_invalid"; }
  if (!isSnapshotBinding(binding) || !isSamples(samples)) return "workflow_rag_evaluation_payload_invalid";
  const input = { datasetKey: editor.datasetKey, displayName: editor.displayName, contentClassification: editor.contentClassification, baselineSnapshot: binding, thresholds: editor.thresholds, reviewSummary: editor.reviewSummary, samples };
  return validateWorkflowRAGEvaluationDatasetInput(config, applicationId, input) || input;
}

function parseSnapshotBinding(value: string, applicationId: string): WorkflowRAGEvaluationSnapshotBinding | string {
  let parsed: unknown;
  try { parsed = JSON.parse(value); } catch { return "workflow_rag_evaluation_payload_invalid"; }
  if (!isSnapshotBinding(parsed) || parsed.tenantRef !== config.tenantRef || parsed.workspaceId !== config.workspaceId || parsed.applicationId !== applicationId) return "workflow_rag_evaluation_payload_invalid";
  return parsed;
}

function isSnapshotBinding(value: unknown): value is WorkflowRAGEvaluationSnapshotBinding { if (!isRecord(value)) return false; const keys = ["tenantRef", "workspaceId", "applicationId", "snapshotId", "snapshotVersion", "snapshotDigest", "ragRef"]; return exactKeys(value, keys) && keys.slice(0, 4).every((key) => typeof value[key] === "string") && Number.isInteger(value.snapshotVersion) && typeof value.snapshotDigest === "string" && typeof value.ragRef === "string"; }
function isSamples(value: unknown): value is WorkflowRAGEvaluationSampleInput[] { if (!Array.isArray(value)) return false; const keys = ["sampleId", "queryText", "expectation", "expectedCitationRefs", "requiredOfficialRefs", "topK", "reviewNote"]; return value.every((sample) => isRecord(sample) && exactKeys(sample, keys) && typeof sample.sampleId === "string" && typeof sample.queryText === "string" && typeof sample.expectation === "string" && Array.isArray(sample.expectedCitationRefs) && sample.expectedCitationRefs.every((ref) => typeof ref === "string") && Array.isArray(sample.requiredOfficialRefs) && sample.requiredOfficialRefs.every((ref) => typeof ref === "string") && Number.isInteger(sample.topK) && typeof sample.reviewNote === "string"); }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === "object" && value !== null && !Array.isArray(value); }
function exactKeys(value: Record<string, unknown>, keys: string[]): boolean { const expected = new Set(keys); return Object.keys(value).length === keys.length && Object.keys(value).every((key) => expected.has(key)); }

function editorFromVersion(version: WorkflowRAGEvaluationDatasetVersion): DatasetEditor { return { datasetKey: version.datasetKey, displayName: version.displayName, contentClassification: version.contentClassification, baselineSnapshotJSON: JSON.stringify(version.snapshot, null, 2), thresholds: version.thresholds, reviewSummary: version.reviewSummary, samplesJSON: JSON.stringify(version.samples, null, 2) }; }
function emptyEditor(applicationId: string): DatasetEditor { return { datasetKey: "", displayName: "", contentClassification: "workspace_internal", baselineSnapshotJSON: emptySnapshotJSON(applicationId), thresholds: { ...DEFAULT_THRESHOLDS }, reviewSummary: "", samplesJSON: JSON.stringify([{ sampleId: "evidence_sample", queryText: "", expectation: "evidence_required", expectedCitationRefs: ["fragment_ref"], requiredOfficialRefs: [], topK: 5, reviewNote: "" }], null, 2) }; }
function emptySnapshotJSON(applicationId: string): string { return JSON.stringify({ tenantRef: config.tenantRef, workspaceId: config.workspaceId, applicationId, snapshotId: "", snapshotVersion: 1, snapshotDigest: "", ragRef: "" }, null, 2); }
function emptyCollection(): DatasetCollection { return { active: [], archived: [], activeCursor: "", archivedCursor: "", failureCode: "" }; }
function mergeResources(current: WorkflowRAGEvaluationDatasetResource[], incoming: WorkflowRAGEvaluationDatasetResource[]): WorkflowRAGEvaluationDatasetResource[] { return [...new Map([...current, ...incoming].map((value) => [value.datasetId, value])).values()].sort((left, right) => left.datasetKey.localeCompare(right.datasetKey)); }
function mergeReviews(current: WorkflowRAGCandidateReviewSummary[], incoming: WorkflowRAGCandidateReviewSummary[]): WorkflowRAGCandidateReviewSummary[] { return [...new Map([...current, ...incoming].map((value) => [value.reviewId, value])).values()].sort((left, right) => right.createdAt.localeCompare(left.createdAt) || right.reviewId.localeCompare(left.reviewId)); }
function localFailure(failureCode: string): WorkflowRAGEvaluationOperationResult { return { status: "failed", resource: null, version: null, review: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary: "评测数据集输入在请求发送前被拒绝。" }; }
