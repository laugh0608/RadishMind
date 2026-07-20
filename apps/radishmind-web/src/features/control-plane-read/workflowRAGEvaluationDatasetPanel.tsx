import { useEffect, useMemo, useState } from "react";

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
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
}) {
  const [collection, setCollection] = useState<DatasetCollection>(emptyCollection);
  const [filter, setFilter] = useState<WorkflowRAGEvaluationLifecycle>("active");
  const [selectedResource, setSelectedResource] = useState<WorkflowRAGEvaluationDatasetResource | null>(null);
  const [selectedVersion, setSelectedVersion] = useState<WorkflowRAGEvaluationDatasetVersion | null>(null);
  const [reviews, setReviews] = useState<WorkflowRAGCandidateReviewSummary[]>([]);
  const [selectedReview, setSelectedReview] = useState<WorkflowRAGCandidateReviewSummary | null>(null);
  const [reviewCursor, setReviewCursor] = useState("");
  const [editor, setEditor] = useState<DatasetEditor>(() => emptyEditor(applicationId));
  const [candidateJSON, setCandidateJSON] = useState(() => emptySnapshotJSON(applicationId));
  const [showCreate, setShowCreate] = useState(false);
  const [showArchiveConfirm, setShowArchiveConfirm] = useState(false);
  const [pending, setPending] = useState<"" | "listing" | "reading" | "creating" | "versioning" | "archiving" | "reviewing">("");
  const [operation, setOperation] = useState<WorkflowRAGEvaluationOperationResult | null>(null);

  const canRead = config.scopes.has("workflow_rag_evaluation_datasets:read");
  const canReadContent = canRead && config.scopes.has("workflow_rag_evaluation_datasets:read_content");
  const canWrite = applicationActive && config.scopes.has("workflow_rag_evaluation_datasets:write") && config.scopes.has("workflow_rag_snapshots:read");
  const canReview = applicationActive && canRead && config.scopes.has("workflow_rag_evaluation_datasets:review") && config.scopes.has("workflow_rag_snapshots:read");
  const canArchive = applicationActive && config.scopes.has("workflow_rag_evaluation_datasets:archive");
  const visibleResources = useMemo(() => filter === "active" ? collection.active : collection.archived, [collection, filter]);

  useEffect(() => {
    let cancelled = false;
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
    if (config.mode === "offline" || !canRead || !applicationId.trim()) return;
    setPending("listing");
    Promise.all([
      listWorkflowRAGEvaluationDatasets(config, applicationId, "active"),
      listWorkflowRAGEvaluationDatasets(config, applicationId, "archived"),
    ]).then(([active, archived]) => {
      if (cancelled) return;
      setPending("");
      setCollection({ active: active.resources, archived: archived.resources, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode });
    });
    return () => { cancelled = true; };
  }, [applicationId, canRead]);

  const refresh = async () => {
    const [active, archived] = await Promise.all([
      listWorkflowRAGEvaluationDatasets(config, applicationId, "active"),
      listWorkflowRAGEvaluationDatasets(config, applicationId, "archived"),
    ]);
    const next = { active: active.resources, archived: archived.resources, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode };
    setCollection(next);
    return next;
  };

  const selectDataset = async (resource: WorkflowRAGEvaluationDatasetResource) => {
    setSelectedResource(resource);
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
    setPending("");
    setReviews(history.reviews);
    setReviewCursor(history.nextCursor);
    if (!detail) return;
    setOperation(detail);
    if (detail.version) {
      setSelectedVersion(detail.version);
      setEditor(editorFromVersion(detail.version));
    }
  };

  const submitCreate = async () => {
    const input = parseEditor(editor, applicationId);
    if (typeof input === "string") return setOperation(localFailure(input));
    setPending("creating");
    const result = await createWorkflowRAGEvaluationDataset(config, applicationId, input);
    setPending("");
    setOperation(result);
    if (!result.resource || !result.version) return;
    setSelectedResource(result.resource);
    setSelectedVersion(result.version);
    setEditor(editorFromVersion(result.version));
    setShowCreate(false);
    setFilter("active");
    await refresh();
  };

  const submitVersion = async () => {
    if (!selectedResource) return;
    const input = parseEditor(editor, applicationId);
    if (typeof input === "string") return setOperation(localFailure(input));
    setPending("versioning");
    const result = await versionWorkflowRAGEvaluationDataset(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion, input);
    setPending("");
    setOperation(result);
    if (!result.resource || !result.version) return;
    setSelectedResource(result.resource);
    setSelectedVersion(result.version);
    setEditor(editorFromVersion(result.version));
    await refresh();
  };

  const submitArchive = async () => {
    if (!selectedResource) return;
    setPending("archiving");
    const result = await archiveWorkflowRAGEvaluationDataset(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion);
    setPending("");
    setOperation(result);
    setShowArchiveConfirm(false);
    if (!result.resource) return;
    setSelectedResource(result.resource);
    setFilter("archived");
    await refresh();
  };

  const submitReview = async () => {
    if (!selectedResource) return;
    const candidate = parseSnapshotBinding(candidateJSON, applicationId);
    if (typeof candidate === "string") return setOperation(localFailure(candidate));
    setPending("reviewing");
    const result = await createWorkflowRAGCandidateReview(config, applicationId, selectedResource.datasetId, selectedResource.latestVersion, selectedResource.latestDigest, candidate);
    setPending("");
    setOperation(result);
    if (!result.review) return;
    setSelectedReview(result.review);
    const history = await listWorkflowRAGCandidateReviews(config, applicationId, selectedResource.datasetId);
    setReviews(history.reviews);
    setReviewCursor(history.nextCursor);
  };

  const selectReview = async (review: WorkflowRAGCandidateReviewSummary) => {
    if (!selectedResource) return;
    setPending("reading");
    const result = await readWorkflowRAGCandidateReview(config, applicationId, selectedResource.datasetId, review.reviewId);
    setPending("");
    setOperation(result);
    setSelectedReview(result.review);
  };

  const loadMoreDatasets = async () => {
    const cursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
    if (!cursor) return;
    setPending("listing");
    const page = await listWorkflowRAGEvaluationDatasets(config, applicationId, filter, cursor);
    setPending("");
    setCollection((current) => filter === "active"
      ? { ...current, active: mergeResources(current.active, page.resources), activeCursor: page.nextCursor, failureCode: page.failureCode }
      : { ...current, archived: mergeResources(current.archived, page.resources), archivedCursor: page.nextCursor, failureCode: page.failureCode });
  };

  const loadMoreReviews = async () => {
    if (!selectedResource || !reviewCursor) return;
    setPending("reading");
    const page = await listWorkflowRAGCandidateReviews(config, applicationId, selectedResource.datasetId, reviewCursor);
    setPending("");
    setReviews((current) => mergeReviews(current, page.reviews));
    setReviewCursor(page.nextCursor);
  };

  if (config.mode === "offline") return <BoundaryPanel status="offline" summary="RAG 评测数据集保持 offline；本面板发送 0 个请求，不读取 query，也不创建候选审查。" />;
  if (!canRead || !applicationId.trim()) return <BoundaryPanel status="scope denied" summary="缺少 application scope 或 workflow_rag_evaluation_datasets:read；本面板发送 0 个请求。" />;

  const datasetCursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
  return (
    <section className="workflow-rag-evaluation-panel" id="workflow-rag-evaluation-panel" aria-labelledby="workflow-rag-evaluation-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Workflow RAG · Quality governance</p><h4 id="workflow-rag-evaluation-title">评测数据集与候选快照审查</h4></div>
        <span className={`status-badge ${collection.failureCode ? "bad" : "good"}`}>{pending || (collection.failureCode ? "failed" : "ready")}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>Application</span><strong>{applicationName || applicationId}</strong><code>{applicationId}</code></article>
        <article><span>Dataset detail</span><strong>{canReadContent ? "authorized" : "metadata-only"}</strong><small>query 只在精确版本 detail 中返回。</small></article>
        <article><span>Candidate review</span><strong>{canReview ? "enabled" : "read-only"}</strong><small>baseline 与 candidate 各执行一次 lexical ranking。</small></article>
        <article><span>Execution boundary</span><strong>0 Gateway / 0 runs</strong><small>审查不创建 run，也不触发模型调用。</small></article>
      </div>

      {!applicationActive ? <p className="workflow-rag-boundary-note">当前应用已归档；数据集与审查证据仍可读取，写入、审查和归档入口关闭。</p> : null}
      {!canReadContent ? <p className="workflow-rag-boundary-note">缺少 read_content：数据集列表与 metadata-only 审查仍可用，但 query / review note 不会请求。</p> : null}
      {collection.failureCode ? <p className="workflow-rag-failure" role="alert"><code>{collection.failureCode}</code></p> : null}

      <div className="workflow-rag-toolbar">
        <div className="workflow-rag-filter" aria-label="评测数据集生命周期筛选">
          {LIFECYCLES.map((state) => <button key={state} type="button" className={filter === state ? "selected" : ""} onClick={() => setFilter(state)}>{state}</button>)}
        </div>
        <button type="button" disabled={pending !== "" || !canWrite} onClick={() => { setShowCreate(true); setSelectedResource(null); setSelectedVersion(null); setSelectedReview(null); setReviews([]); setEditor(emptyEditor(applicationId)); setOperation(null); }}>新建评测数据集</button>
      </div>

      <div className="workflow-rag-evaluation-layout">
        <div className="workflow-rag-list" aria-label={`${filter} workflow RAG evaluation datasets`}>
          {visibleResources.map((resource) => (
            <button key={resource.datasetId} type="button" className={selectedResource?.datasetId === resource.datasetId ? "selected" : ""} onClick={() => void selectDataset(resource)}>
              <span><strong>{resource.displayName}</strong><code>{resource.datasetKey} · v{resource.latestVersion}</code></span>
              <span><small>{resource.sampleCount} samples</small><small>{resource.contentClassification}</small></span>
            </button>
          ))}
          {!visibleResources.length && pending !== "listing" ? <p>当前作用域没有 {filter} 评测数据集。</p> : null}
          {datasetCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMoreDatasets()}>加载下一页</button> : null}
        </div>

        <div className="workflow-rag-evaluation-workspace">
          {showCreate || selectedVersion ? <DatasetEditorFields editor={editor} disabled={pending !== "" || (!showCreate && !canWrite)} immutableKey={Boolean(selectedVersion)} onChange={setEditor} /> : selectedResource ? <ResourceEvidence resource={selectedResource} /> : <article className="workflow-rag-empty"><strong>选择数据集或创建 v1</strong><p>列表保持 metadata-only；query 只通过精确版本 detail 进入编辑器。</p></article>}

          {showCreate ? <div className="workflow-rag-actions"><button type="button" disabled={pending !== "" || !canWrite} onClick={() => void submitCreate()}>{pending === "creating" ? "创建中…" : "创建 dataset v1"}</button><button type="button" disabled={pending !== ""} onClick={() => setShowCreate(false)}>取消</button></div> : null}
          {selectedResource && selectedVersion && selectedResource.lifecycleState === "active" ? (
            <div className="workflow-rag-actions"><button type="button" disabled={pending !== "" || !canWrite} onClick={() => void submitVersion()}>{pending === "versioning" ? "写入中…" : `完整替换并创建 v${selectedResource.latestVersion + 1}`}</button><button type="button" className="danger-action" disabled={pending !== "" || !canArchive} onClick={() => setShowArchiveConfirm(true)}>归档数据集</button></div>
          ) : null}
          {showArchiveConfirm ? <div className="workflow-rag-archive-confirm" role="alert"><p>归档后禁止新版本与候选审查；历史版本和审查仍可读。</p><button type="button" className="danger-action" disabled={pending !== "" || !canArchive} onClick={() => void submitArchive()}>确认归档</button><button type="button" onClick={() => setShowArchiveConfirm(false)}>取消</button></div> : null}

          {selectedResource ? (
            <section className="workflow-rag-candidate-review" aria-labelledby="workflow-rag-candidate-review-title">
              <div className="section-heading compact-heading"><div><p className="eyebrow">Exact candidate binding</p><h5 id="workflow-rag-candidate-review-title">候选快照审查</h5></div><span className="status-badge neutral">{reviews.length} records</span></div>
              <label><span>Candidate snapshot binding（camelCase JSON）</span><textarea value={candidateJSON} disabled={pending !== "" || !canReview || selectedResource.lifecycleState !== "active"} spellCheck={false} onChange={(event) => setCandidateJSON(event.currentTarget.value)} /></label>
              <button type="button" disabled={pending !== "" || !canReview || selectedResource.lifecycleState !== "active"} onClick={() => void submitReview()}>{pending === "reviewing" ? "审查中…" : `对 dataset v${selectedResource.latestVersion} 执行候选审查`}</button>
              <div className="workflow-rag-review-history">
                {reviews.map((review) => <button key={review.reviewId} type="button" className={selectedReview?.reviewId === review.reviewId ? "selected" : ""} onClick={() => void selectReview(review)}><span><strong>{review.conclusion}</strong><code>{review.candidateSnapshot.ragRef}</code></span><small>{review.createdAt}</small></button>)}
                {!reviews.length ? <p>尚无候选快照审查记录。</p> : null}
                {reviewCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMoreReviews()}>加载更多审查</button> : null}
              </div>
              {selectedReview ? <ReviewEvidence review={selectedReview} /> : null}
            </section>
          ) : null}

          {operation ? <OperationEvidence operation={operation} /> : null}
        </div>
      </div>
    </section>
  );
}

const LIFECYCLES: WorkflowRAGEvaluationLifecycle[] = ["active", "archived"];

function DatasetEditorFields({ editor, disabled, immutableKey, onChange }: { editor: DatasetEditor; disabled: boolean; immutableKey: boolean; onChange: (value: DatasetEditor) => void }) {
  return <div className="workflow-rag-evaluation-fields">
    <label><span>Dataset key</span><input value={editor.datasetKey} disabled={disabled || immutableKey} maxLength={48} onChange={(event) => onChange({ ...editor, datasetKey: event.currentTarget.value })} /></label>
    <label><span>Display name</span><input value={editor.displayName} disabled={disabled} maxLength={120} onChange={(event) => onChange({ ...editor, displayName: event.currentTarget.value })} /></label>
    <label><span>Classification</span><select value={editor.contentClassification} disabled={disabled} onChange={(event) => onChange({ ...editor, contentClassification: event.currentTarget.value as WorkflowRAGEvaluationClassification })}><option value="synthetic_public">synthetic_public</option><option value="workspace_internal">workspace_internal</option></select></label>
    <label className="wide"><span>Baseline snapshot binding</span><textarea value={editor.baselineSnapshotJSON} disabled={disabled} spellCheck={false} onChange={(event) => onChange({ ...editor, baselineSnapshotJSON: event.currentTarget.value })} /></label>
    <fieldset className="workflow-rag-thresholds"><legend>Acceptance thresholds</legend>{(Object.keys(editor.thresholds) as Array<keyof WorkflowRAGEvaluationMetrics>).map((key) => <label key={key}><span>{key}</span><input type="number" min="0" max="1" step="0.01" value={editor.thresholds[key]} disabled={disabled} onChange={(event) => onChange({ ...editor, thresholds: { ...editor.thresholds, [key]: Number(event.currentTarget.value) } })} /></label>)}</fieldset>
    <label className="wide"><span>Human review summary</span><textarea value={editor.reviewSummary} disabled={disabled} maxLength={512} onChange={(event) => onChange({ ...editor, reviewSummary: event.currentTarget.value })} /></label>
    <label className="wide"><span>Complete samples（camelCase JSON array）</span><textarea value={editor.samplesJSON} disabled={disabled} spellCheck={false} onChange={(event) => onChange({ ...editor, samplesJSON: event.currentTarget.value })} /><small>query 最多 4096 bytes；版本化采用完整 replacement 与 expected_latest_version CAS。</small></label>
  </div>;
}

function ResourceEvidence({ resource }: { resource: WorkflowRAGEvaluationDatasetResource }) { return <article className="workflow-rag-record"><div><strong>{resource.displayName}</strong><span className="status-badge neutral">{resource.lifecycleState}</span></div><dl><div><dt>Dataset</dt><dd>{resource.datasetId}</dd></div><div><dt>Version</dt><dd>v{resource.latestVersion}</dd></div><div><dt>Digest</dt><dd>{resource.latestDigest}</dd></div><div><dt>Baseline</dt><dd>{resource.baselineSnapshot.ragRef}</dd></div><div><dt>Samples</dt><dd>{resource.sampleCount}</dd></div><div><dt>Classification</dt><dd>{resource.contentClassification}</dd></div></dl></article>; }

function ReviewEvidence({ review }: { review: WorkflowRAGCandidateReviewSummary }) { return <article className={`workflow-rag-review-evidence ${review.conclusion === "regressed" ? "failed" : ""}`}><div><strong>{review.conclusion}</strong><span>{review.baselineStatus} → {review.candidateStatus}</span></div><dl>{(Object.keys(review.metricDelta) as Array<keyof WorkflowRAGEvaluationMetrics>).map((key) => <div key={key}><dt>{key}</dt><dd>{formatDelta(review.metricDelta[key])}</dd></div>)}</dl><p>Samples: {review.sampleChanges.map((sample) => `${sample.sampleId}:${sample.change}`).join(" · ")}</p><p>Findings +[{review.addedFindingCodes.join(", ") || "none"}] −[{review.removedFindingCodes.join(", ") || "none"}]</p><code>{review.reviewId}</code></article>; }

function OperationEvidence({ operation }: { operation: WorkflowRAGEvaluationOperationResult }) { return <article className={`workflow-rag-operation ${operation.status === "failed" || operation.status === "version_conflict" ? "failed" : ""}`} aria-live="polite"><strong>{operation.status}</strong><p>{operation.summary}</p>{operation.failureCode ? <code>{operation.failureCode}</code> : null}{operation.status === "version_conflict" ? <small>Current: v{operation.currentLatestVersion} · {operation.currentLifecycleState}</small> : null}</article>; }
function BoundaryPanel({ status, summary }: { status: string; summary: string }) { return <section className="workflow-rag-evaluation-panel offline" aria-label="Workflow RAG evaluation datasets"><div className="section-heading compact-heading"><div><p className="eyebrow">Workflow RAG · Quality governance</p><h4>评测数据集未启用</h4></div><span className="status-badge neutral">{status}</span></div><p>{summary}</p></section>; }

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
function formatDelta(value: number): string { return `${value > 0 ? "+" : ""}${value.toFixed(4)}`; }
