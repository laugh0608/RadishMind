import { useEffect, useMemo, useRef, useState } from "react";

import { importWorkflowRAGLocalMaterials, preflightWorkflowRAGLocalMaterialSelection, type WorkflowRAGLocalMaterialFile } from "./workflowRAGLocalMaterialImporter.ts";
import {
  archiveWorkflowRAGSnapshot,
  createWorkflowRAGSnapshot,
  listWorkflowRAGSnapshots,
  readWorkflowRAGSnapshot,
  readWorkflowRAGSnapshotConfig,
  versionWorkflowRAGSnapshot,
  type WorkflowRAGSnapshotLifecycle,
  type WorkflowRAGSnapshotOperationResult,
  type WorkflowRAGSnapshotRecord,
  type WorkflowRAGSnapshotResource,
} from "./workflowRAGSnapshotConsumer.ts";
import {
  analyzeWorkflowRAGSnapshotEditor,
  buildWorkflowRAGSnapshotWriteInput,
  createEmptyWorkflowRAGSnapshotEditor,
  createWorkflowRAGSnapshotEditorFromRecord,
  replaceWorkflowRAGSnapshotEditorWithImport,
  type WorkflowRAGSnapshotEditor,
} from "./workflowRAGSnapshotEditor.ts";
import { WorkflowRAGSnapshotEditorPanel } from "./workflowRAGSnapshotEditorPanel.tsx";

const config = readWorkflowRAGSnapshotConfig();

type SnapshotCollection = {
  active: WorkflowRAGSnapshotResource[];
  archived: WorkflowRAGSnapshotResource[];
  activeCursor: string;
  archivedCursor: string;
  failureCode: string;
  summary: string;
};

type PendingOperation = "" | "listing" | "reading" | "importing" | "creating" | "versioning" | "archiving";

export default function WorkflowRAGSnapshotPanel({
  applicationId,
  applicationName,
  applicationActive,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
}) {
  const [collection, setCollection] = useState<SnapshotCollection>(emptyCollection);
  const [filter, setFilter] = useState<WorkflowRAGSnapshotLifecycle>("active");
  const [selectedResource, setSelectedResource] = useState<WorkflowRAGSnapshotResource | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<WorkflowRAGSnapshotRecord | null>(null);
  const [editor, setEditor] = useState<WorkflowRAGSnapshotEditor>(createEmptyWorkflowRAGSnapshotEditor);
  const [pending, setPending] = useState<PendingOperation>("");
  const [operation, setOperation] = useState<WorkflowRAGSnapshotOperationResult | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showArchiveConfirm, setShowArchiveConfirm] = useState(false);
  const requestGeneration = useRef(0);

  const visibleResources = useMemo(
    () => filter === "active" ? collection.active : collection.archived,
    [collection.active, collection.archived, filter],
  );
  const analysis = useMemo(() => analyzeWorkflowRAGSnapshotEditor(editor), [editor]);
  const canRead = config.scopes.has("workflow_rag_snapshots:read");
  const canWrite = applicationActive && config.scopes.has("workflow_rag_snapshots:write");
  const canArchive = applicationActive && config.scopes.has("workflow_rag_snapshots:archive");

  useEffect(() => {
    const generation = ++requestGeneration.current;
    setCollection(emptyCollection());
    setSelectedResource(null);
    setSelectedRecord(null);
    setEditor(createEmptyWorkflowRAGSnapshotEditor());
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setPending("");
    if (config.mode !== "offline" && canRead && applicationId.trim()) {
      setPending("listing");
      void Promise.all([
        listWorkflowRAGSnapshots(config, applicationId, "active"),
        listWorkflowRAGSnapshots(config, applicationId, "archived"),
      ]).then(([active, archived]) => {
        if (requestGeneration.current !== generation) return;
        setPending("");
        setCollection(collectionFromResults(active, archived));
      });
    }
    return () => {
      requestGeneration.current += 1;
    };
  }, [applicationId, canRead]);

  const changeEditor = (nextEditor: WorkflowRAGSnapshotEditor) => {
    requestGeneration.current += 1;
    setEditor(nextEditor);
    setOperation(null);
    setShowArchiveConfirm(false);
  };

  const changeFilter = (nextFilter: WorkflowRAGSnapshotLifecycle) => {
    requestGeneration.current += 1;
    setFilter(nextFilter);
    setSelectedResource(null);
    setSelectedRecord(null);
    setEditor(createEmptyWorkflowRAGSnapshotEditor());
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setPending("");
  };

  const beginCreate = () => {
    requestGeneration.current += 1;
    setShowCreate(true);
    setSelectedResource(null);
    setSelectedRecord(null);
    setEditor(createEmptyWorkflowRAGSnapshotEditor());
    setOperation(null);
    setShowArchiveConfirm(false);
    setPending("");
  };

  const cancelCreate = () => {
    requestGeneration.current += 1;
    setShowCreate(false);
    setEditor(createEmptyWorkflowRAGSnapshotEditor());
    setOperation(null);
    setPending("");
  };

  const selectResource = async (resource: WorkflowRAGSnapshotResource) => {
    const generation = ++requestGeneration.current;
    setSelectedResource(resource);
    setSelectedRecord(null);
    setEditor(createEmptyWorkflowRAGSnapshotEditor());
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setPending("reading");
    const result = await readWorkflowRAGSnapshot(config, applicationId, resource.snapshotId, resource.latestVersion);
    if (requestGeneration.current !== generation) return;
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(createWorkflowRAGSnapshotEditorFromRecord(result.record));
  };

  const importLocalFiles = async (files: File[]) => {
    const generation = ++requestGeneration.current;
    const baseEditor = editor;
    setPending("importing");
    setOperation(null);
    try {
      const preflight = preflightWorkflowRAGLocalMaterialSelection(files.map((file) => ({ fileName: file.name, fileBytes: file.size })));
      if (preflight) {
        setEditor(replaceWorkflowRAGSnapshotEditorWithImport(baseEditor, preflight));
        setPending("");
        return;
      }
      const localFiles: WorkflowRAGLocalMaterialFile[] = await Promise.all(files.map(async (file, selectionIndex) => ({
        fileName: file.name,
        bytes: new Uint8Array(await file.arrayBuffer()),
        selectionIndex,
      })));
      if (requestGeneration.current !== generation) return;
      const result = await importWorkflowRAGLocalMaterials(localFiles);
      if (requestGeneration.current !== generation) return;
      setEditor(replaceWorkflowRAGSnapshotEditorWithImport(baseEditor, result));
      setPending("");
    } catch {
      if (requestGeneration.current !== generation) return;
      setPending("");
      setOperation(localFailure("workflow_rag_material_content_invalid", "浏览器未能读取所选本地文件；现有 staging 保持不变。"));
    }
  };

  const refreshCollections = async (generation: number): Promise<SnapshotCollection | null> => {
    const [active, archived] = await Promise.all([
      listWorkflowRAGSnapshots(config, applicationId, "active"),
      listWorkflowRAGSnapshots(config, applicationId, "archived"),
    ]);
    if (requestGeneration.current !== generation) return null;
    const nextCollection = collectionFromResults(active, archived);
    setCollection(nextCollection);
    return nextCollection;
  };

  const submitCreate = async () => {
    const built = buildWorkflowRAGSnapshotWriteInput(editor);
    if (built.status === "blocked") {
      setOperation(localFailure(built.failureCode));
      return;
    }
    const generation = ++requestGeneration.current;
    setPending("creating");
    const result = await createWorkflowRAGSnapshot(config, applicationId, built.input);
    if (requestGeneration.current !== generation) return;
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(createWorkflowRAGSnapshotEditorFromRecord(result.record));
    setShowCreate(false);
    setFilter("active");
    const refreshed = await refreshCollections(generation);
    if (!refreshed || requestGeneration.current !== generation) return;
    setSelectedResource(refreshed.active.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const submitVersion = async () => {
    if (!selectedRecord) return;
    const built = buildWorkflowRAGSnapshotWriteInput(editor);
    if (built.status === "blocked") {
      setOperation(localFailure(built.failureCode));
      return;
    }
    const generation = ++requestGeneration.current;
    setPending("versioning");
    const result = await versionWorkflowRAGSnapshot(config, applicationId, selectedRecord.snapshotId, selectedRecord.snapshotVersion, built.input);
    if (requestGeneration.current !== generation) return;
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(createWorkflowRAGSnapshotEditorFromRecord(result.record));
    const refreshed = await refreshCollections(generation);
    if (!refreshed || requestGeneration.current !== generation) return;
    setSelectedResource(refreshed.active.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const submitArchive = async () => {
    if (!selectedRecord) return;
    const generation = ++requestGeneration.current;
    setPending("archiving");
    const result = await archiveWorkflowRAGSnapshot(config, applicationId, selectedRecord.snapshotId, selectedRecord.snapshotVersion);
    if (requestGeneration.current !== generation) return;
    setPending("");
    setOperation(result);
    setShowArchiveConfirm(false);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(createWorkflowRAGSnapshotEditorFromRecord(result.record));
    setFilter("archived");
    const refreshed = await refreshCollections(generation);
    if (!refreshed || requestGeneration.current !== generation) return;
    setSelectedResource(refreshed.archived.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const loadMore = async () => {
    const cursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
    if (!cursor) return;
    const generation = requestGeneration.current;
    setPending("listing");
    const result = await listWorkflowRAGSnapshots(config, applicationId, filter, cursor);
    if (requestGeneration.current !== generation) return;
    setPending("");
    setCollection((current) => mergePage(current, filter, result.records, result.nextCursor, result.failureCode, result.summary));
  };

  if (config.mode === "offline") {
    return <BoundaryPanel status="offline" summary="RAG 知识快照保持 offline；本面板发送 0 个请求，也不会模拟写入成功。" />;
  }
  if (!canRead || !applicationId.trim()) {
    return <BoundaryPanel status="scope denied" summary="缺少 application scope 或 workflow_rag_snapshots:read；本面板发送 0 个请求。" />;
  }

  const currentCursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
  const writeDisabled = pending !== "" || !canWrite || !analysis.canSubmit;
  const editorDisabled = pending !== "" || !canWrite || Boolean(selectedRecord && selectedRecord.lifecycleState !== "active");
  const archiveDisabled = pending !== "" || !canArchive || selectedRecord?.lifecycleState !== "active";

  return (
    <section className="workflow-rag-snapshot-panel" id="workflow-rag-snapshot-panel" aria-labelledby="workflow-rag-snapshot-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">Workflow RAG · Application knowledge</p><h4 id="workflow-rag-snapshot-title">知识快照与精确版本</h4></div>
        <span className={`status-badge ${collection.failureCode ? "bad" : "good"}`}>{pending || (collection.failureCode ? "failed" : "ready")}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>Application</span><strong>{applicationName || applicationId}</strong><code>{applicationId}</code></article>
        <article><span>Repository scope</span><strong>{config.workspaceId}</strong><code>{config.tenantRef}</code></article>
        <article><span>Profile</span><strong>lexical-ngram-dev.v1</strong><small>精确版本可绑定到独立 retrieval execution。</small></article>
        <article><span>Write boundary</span><strong>{canWrite ? "create / version enabled" : "read-only"}</strong><small>归档使用独立 archive scope。</small></article>
      </div>

      {!applicationActive ? <p className="workflow-rag-boundary-note">当前应用已归档；知识快照正文与历史仍可精确读取，但创建、版本化和归档入口均关闭。</p> : null}
      {collection.failureCode ? <p className="workflow-rag-failure" role="alert"><code>{collection.failureCode}</code> · {collection.summary}</p> : null}

      <div className="workflow-rag-toolbar">
        <div className="workflow-rag-filter" aria-label="知识快照生命周期筛选">
          {(["active", "archived"] as const).map((state) => <button key={state} type="button" className={filter === state ? "selected" : ""} disabled={pending !== ""} onClick={() => changeFilter(state)}>{state}</button>)}
        </div>
        <button type="button" disabled={pending !== "" || !canWrite} onClick={beginCreate}>新建知识快照</button>
      </div>

      <div className="workflow-rag-layout">
        <div className="workflow-rag-list" aria-label={`${filter} knowledge snapshots`}>
          {visibleResources.map((resource) => (
            <button key={resource.snapshotId} type="button" disabled={pending !== ""} className={selectedResource?.snapshotId === resource.snapshotId ? "selected" : ""} onClick={() => void selectResource(resource)}>
              <span><strong>{resource.displayName}</strong><code>{resource.latestRAGRef}</code></span>
              <span><small>{resource.fragmentCount} fragments</small><small>{resource.totalContentBytes} bytes</small></span>
            </button>
          ))}
          {!visibleResources.length && pending !== "listing" ? <p>当前作用域没有 {filter} 知识快照。</p> : null}
          {currentCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMore()}>加载下一页</button> : null}
        </div>

        <div className="workflow-rag-editor">
          {showCreate || selectedRecord ? (
            <WorkflowRAGSnapshotEditorPanel editor={editor} analysis={analysis} disabled={editorDisabled} immutableSnapshotKey={Boolean(selectedRecord)} importing={pending === "importing"} onChange={changeEditor} onImportFiles={(files) => void importLocalFiles(files)} />
          ) : (
            <article className="workflow-rag-empty"><strong>选择精确快照版本</strong><p>列表只含 metadata；选择后才以 read scope 拉取当前明确版本的正文。</p></article>
          )}

          {showCreate ? (
            <div className="workflow-rag-actions"><button type="button" disabled={writeDisabled} onClick={() => void submitCreate()}>{pending === "creating" ? "创建中…" : "创建 v1"}</button><button type="button" disabled={pending !== ""} onClick={cancelCreate}>取消</button></div>
          ) : selectedRecord ? (
            <>
              <SnapshotRecordEvidence record={selectedRecord} />
              {selectedRecord.lifecycleState === "active" ? (
                <div className="workflow-rag-actions">
                  <button type="button" disabled={writeDisabled} onClick={() => void submitVersion()}>{pending === "versioning" ? "写入中…" : `完整替换并创建 v${selectedRecord.snapshotVersion + 1}`}</button>
                  <button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => setShowArchiveConfirm(true)}>归档快照</button>
                </div>
              ) : null}
              {showArchiveConfirm ? <div className="workflow-rag-archive-confirm" role="alert"><p>归档后禁止创建新版本；现有版本仍保持精确可读。</p><button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => void submitArchive()}>确认归档</button><button type="button" disabled={pending !== ""} onClick={() => setShowArchiveConfirm(false)}>取消</button></div> : null}
            </>
          ) : null}

          {operation ? <OperationEvidence operation={operation} /> : null}
        </div>
      </div>
    </section>
  );
}

function SnapshotRecordEvidence({ record }: { record: WorkflowRAGSnapshotRecord }) {
  return <article className="workflow-rag-record"><div><strong>{record.ragRef}</strong><span className="status-badge neutral">{record.lifecycleState}</span></div><dl><div><dt>Digest</dt><dd>{record.snapshotDigest}</dd></div><div><dt>Profile</dt><dd>{record.profileRef}</dd></div><div><dt>Fragments</dt><dd>{record.fragmentCount}</dd></div><div><dt>Content bytes</dt><dd>{record.totalContentBytes}</dd></div><div><dt>Request</dt><dd>{record.requestId}</dd></div><div><dt>Audit</dt><dd>{record.auditRef}</dd></div></dl></article>;
}

function OperationEvidence({ operation }: { operation: WorkflowRAGSnapshotOperationResult }) {
  return <article className={`workflow-rag-operation ${operation.status === "failed" || operation.status === "version_conflict" ? "failed" : ""}`} aria-live="polite"><strong>{operation.status}</strong><p>{operation.summary}</p>{operation.failureCode ? <code>{operation.failureCode}</code> : null}{operation.status === "version_conflict" ? <small>Current: v{operation.currentLatestVersion} · {operation.currentLifecycleState}</small> : null}</article>;
}

function BoundaryPanel({ status, summary }: { status: string; summary: string }) {
  return <section className="workflow-rag-snapshot-panel offline" aria-label="Workflow RAG knowledge snapshot"><div className="section-heading compact-heading"><div><p className="eyebrow">Workflow RAG · Application knowledge</p><h4>知识快照未启用</h4></div><span className="status-badge neutral">{status}</span></div><p>{summary}</p></section>;
}

function emptyCollection(): SnapshotCollection {
  return { active: [], archived: [], activeCursor: "", archivedCursor: "", failureCode: "", summary: "" };
}

function collectionFromResults(active: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>, archived: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>): SnapshotCollection {
  return { active: active.records, archived: archived.records, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode, summary: active.failureCode ? active.summary : archived.failureCode ? archived.summary : `${active.records.length} active / ${archived.records.length} archived` };
}

function mergePage(current: SnapshotCollection, lifecycle: WorkflowRAGSnapshotLifecycle, records: WorkflowRAGSnapshotResource[], cursor: string, failureCode: string, summary: string): SnapshotCollection {
  const merged = mergeResources(lifecycle === "active" ? current.active : current.archived, records);
  return lifecycle === "active" ? { ...current, active: merged, activeCursor: cursor, failureCode, summary } : { ...current, archived: merged, archivedCursor: cursor, failureCode, summary };
}

function mergeResources(current: WorkflowRAGSnapshotResource[], incoming: WorkflowRAGSnapshotResource[]): WorkflowRAGSnapshotResource[] {
  return [...new Map([...current, ...incoming].map((resource) => [resource.snapshotId, resource])).values()].sort((left, right) => left.snapshotKey.localeCompare(right.snapshotKey));
}

function localFailure(failureCode: string, summary = "知识快照输入在请求发送前被拒绝。"): WorkflowRAGSnapshotOperationResult {
  return { status: "failed", record: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary };
}
