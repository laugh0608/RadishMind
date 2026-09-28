import { useTranslation } from "react-i18next";
import "../../i18n/workflowRAGSnapshotResources.ts";

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
};

type SnapshotOperation = WorkflowRAGSnapshotOperationResult & { localMessage?: "localRejected" | "fileReadFailed" };

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
  const { t } = useTranslation("workflow");
  const [collection, setCollection] = useState<SnapshotCollection>(emptyCollection);
  const [filter, setFilter] = useState<WorkflowRAGSnapshotLifecycle>("active");
  const [selectedResource, setSelectedResource] = useState<WorkflowRAGSnapshotResource | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<WorkflowRAGSnapshotRecord | null>(null);
  const [editor, setEditor] = useState<WorkflowRAGSnapshotEditor>(createEmptyWorkflowRAGSnapshotEditor);
  const [pending, setPending] = useState<PendingOperation>("");
  const [operation, setOperation] = useState<SnapshotOperation | null>(null);
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
      setOperation(localFailure("workflow_rag_material_content_invalid", "fileReadFailed"));
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
    setCollection((current) => mergePage(current, filter, result.records, result.nextCursor, result.failureCode));
  };

  if (config.mode === "offline") {
    return <BoundaryPanel status="offline" summary={t($ => $.ragSnapshot.offlineNote)} />;
  }
  if (!canRead || !applicationId.trim()) {
    return <BoundaryPanel status="scope_denied" summary={t($ => $.ragSnapshot.scopeNote)} />;
  }

  const currentCursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
  const writeDisabled = pending !== "" || !canWrite || !analysis.canSubmit;
  const editorDisabled = pending !== "" || !canWrite || Boolean(selectedRecord && selectedRecord.lifecycleState !== "active");
  const archiveDisabled = pending !== "" || !canArchive || selectedRecord?.lifecycleState !== "active";

  return (
    <section className="workflow-rag-snapshot-panel" id="workflow-rag-snapshot-panel" aria-labelledby="workflow-rag-snapshot-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.ragSnapshot.applicationKnowledge)}</p><h4 id="workflow-rag-snapshot-title">{t($ => $.ragSnapshot.heading)}</h4></div>
        <span className={`status-badge ${collection.failureCode ? "bad" : "good"}`}>{t($ => $.ragSnapshot.status[pending || (collection.failureCode ? "failed" : "ready")])}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>{t($ => $.ragSnapshot.application)}</span><strong>{applicationName || applicationId}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.ragSnapshot.repositoryScope)}</span><strong>{config.workspaceId}</strong><code>{config.tenantRef}</code></article>
        <article><span>{t($ => $.ragSnapshot.profile)}</span><strong>lexical-ngram-dev.v1</strong><small>{t($ => $.ragSnapshot.profileNote)}</small></article>
        <article><span>{t($ => $.ragSnapshot.writeBoundary)}</span><strong>{canWrite ? t($ => $.ragSnapshot.writeEnabled) : t($ => $.ragSnapshot.readOnly)}</strong><small>{t($ => $.ragSnapshot.archiveScope)}</small></article>
      </div>

      {!applicationActive ? <p className="workflow-rag-boundary-note">{t($ => $.ragSnapshot.archivedApplication)}</p> : null}
      {collection.failureCode ? <p className="workflow-rag-failure" role="alert"><code>{collection.failureCode}</code> · {t($ => $.ragSnapshot.listFailed)}</p> : null}

      <div className="workflow-rag-toolbar">
        <div className="workflow-rag-filter" aria-label={t($ => $.ragSnapshot.filterLabel)}>
          {(["active", "archived"] as const).map((state) => <button key={state} type="button" className={filter === state ? "selected" : ""} disabled={pending !== ""} onClick={() => changeFilter(state)}>{t($ => $.ragSnapshot.status[state])}</button>)}
        </div>
        <button type="button" disabled={pending !== "" || !canWrite} onClick={beginCreate}>{t($ => $.ragSnapshot.newSnapshot)}</button>
      </div>

      <div className="workflow-rag-layout">
        <div className="workflow-rag-list" aria-label={t($ => $.ragSnapshot.listLabel, { state: t($ => $.ragSnapshot.status[filter]) })}>
          {visibleResources.map((resource) => (
            <button key={resource.snapshotId} type="button" disabled={pending !== ""} className={selectedResource?.snapshotId === resource.snapshotId ? "selected" : ""} onClick={() => void selectResource(resource)}>
              <span><strong>{resource.displayName}</strong><code>{resource.latestRAGRef}</code></span>
              <span><small>{t($ => $.ragSnapshot.resourceFragments, { count: resource.fragmentCount })}</small><small>{t($ => $.ragSnapshot.resourceBytes, { count: resource.totalContentBytes })}</small></span>
            </button>
          ))}
          {!visibleResources.length && pending !== "listing" ? <p>{t($ => $.ragSnapshot.emptyList, { state: t($ => $.ragSnapshot.status[filter]) })}</p> : null}
          {currentCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMore()}>{t($ => $.ragSnapshot.loadMore)}</button> : null}
        </div>

        <div className="workflow-rag-editor">
          {showCreate || selectedRecord ? (
            <WorkflowRAGSnapshotEditorPanel editor={editor} analysis={analysis} disabled={editorDisabled} immutableSnapshotKey={Boolean(selectedRecord)} importing={pending === "importing"} onChange={changeEditor} onImportFiles={(files) => void importLocalFiles(files)} />
          ) : (
            <article className="workflow-rag-empty"><strong>{t($ => $.ragSnapshot.selectVersion)}</strong><p>{t($ => $.ragSnapshot.listNote)}</p></article>
          )}

          {showCreate ? (
            <div className="workflow-rag-actions"><button type="button" disabled={writeDisabled} onClick={() => void submitCreate()}>{pending === "creating" ? t($ => $.ragSnapshot.creating) : t($ => $.ragSnapshot.createV1)}</button><button type="button" disabled={pending !== ""} onClick={cancelCreate}>{t($ => $.ragSnapshot.cancel)}</button></div>
          ) : selectedRecord ? (
            <>
              <SnapshotRecordEvidence record={selectedRecord} />
              {selectedRecord.lifecycleState === "active" ? (
                <div className="workflow-rag-actions">
                  <button type="button" disabled={writeDisabled} onClick={() => void submitVersion()}>{pending === "versioning" ? t($ => $.ragSnapshot.versioning) : t($ => $.ragSnapshot.replaceVersion, { version: selectedRecord.snapshotVersion + 1 })}</button>
                  <button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => setShowArchiveConfirm(true)}>{t($ => $.ragSnapshot.archive)}</button>
                </div>
              ) : null}
              {showArchiveConfirm ? <div className="workflow-rag-archive-confirm" role="alert"><p>{t($ => $.ragSnapshot.archiveNote)}</p><button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => void submitArchive()}>{t($ => $.ragSnapshot.confirmArchive)}</button><button type="button" disabled={pending !== ""} onClick={() => setShowArchiveConfirm(false)}>{t($ => $.ragSnapshot.cancel)}</button></div> : null}
            </>
          ) : null}

          {operation ? <OperationEvidence operation={operation} /> : null}
        </div>
      </div>
    </section>
  );
}

function SnapshotRecordEvidence({ record }: { record: WorkflowRAGSnapshotRecord }) {
  const { t } = useTranslation("workflow");
  return <article className="workflow-rag-record"><div><strong>{record.ragRef}</strong><span className="status-badge neutral">{t($ => $.ragSnapshot.status[record.lifecycleState])}</span></div><dl><div><dt>{t($ => $.ragSnapshot.digest)}</dt><dd>{record.snapshotDigest}</dd></div><div><dt>{t($ => $.ragSnapshot.profile)}</dt><dd>{record.profileRef}</dd></div><div><dt>{t($ => $.ragSnapshot.fragments)}</dt><dd>{record.fragmentCount}</dd></div><div><dt>{t($ => $.ragSnapshot.contentBytes)}</dt><dd>{record.totalContentBytes}</dd></div><div><dt>{t($ => $.ragSnapshot.request)}</dt><dd>{record.requestId}</dd></div><div><dt>{t($ => $.ragSnapshot.audit)}</dt><dd>{record.auditRef}</dd></div></dl></article>;
}

function OperationEvidence({ operation }: { operation: SnapshotOperation }) {
  const { t } = useTranslation("workflow");
  const summaryKey = operation.localMessage ?? ({
    version_conflict: "conflictNote", offline: "offlineNote", scope_denied: "operationScopeNote", failed: "operationFailed",
    created: "operationSucceeded", loaded: "operationSucceeded", versioned: "operationSucceeded", archived: "operationSucceeded",
  } as const)[operation.status];
  const lifecycle = operation.currentLifecycleState;
  const lifecycleLabel = lifecycle === "active" || lifecycle === "archived" ? t($ => $.ragSnapshot.status[lifecycle]) : lifecycle;
  return (
    <article className={`workflow-rag-operation ${operation.status === "failed" || operation.status === "version_conflict" ? "failed" : ""}`} aria-live="polite">
      <strong>{t($ => $.ragSnapshot.status[operation.status])}</strong>
      <p>{t($ => $.ragSnapshot[summaryKey], { status: t($ => $.ragSnapshot.status[operation.status]) })}</p>
      {operation.failureCode ? <code>{operation.failureCode}</code> : null}
      {operation.status === "version_conflict" ? <small>{t($ => $.ragSnapshot.currentVersion, { version: operation.currentLatestVersion, state: lifecycleLabel })}</small> : null}
    </article>
  );
}

function BoundaryPanel({ status, summary }: { status: "offline" | "scope_denied"; summary: string }) {
  const { t } = useTranslation("workflow");
  return <section className="workflow-rag-snapshot-panel offline" aria-label={t($ => $.ragSnapshot.boundaryLabel)}><div className="section-heading compact-heading"><div><p className="eyebrow">{t($ => $.ragSnapshot.applicationKnowledge)}</p><h4>{t($ => $.ragSnapshot.disabledTitle)}</h4></div><span className="status-badge neutral">{t($ => $.ragSnapshot.status[status])}</span></div><p>{summary}</p></section>;
}

function emptyCollection(): SnapshotCollection {
  return { active: [], archived: [], activeCursor: "", archivedCursor: "", failureCode: "" };
}

function collectionFromResults(active: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>, archived: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>): SnapshotCollection {
  return { active: active.records, archived: archived.records, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode };
}

function mergePage(current: SnapshotCollection, lifecycle: WorkflowRAGSnapshotLifecycle, records: WorkflowRAGSnapshotResource[], cursor: string, failureCode: string): SnapshotCollection {
  const merged = mergeResources(lifecycle === "active" ? current.active : current.archived, records);
  return lifecycle === "active" ? { ...current, active: merged, activeCursor: cursor, failureCode } : { ...current, archived: merged, archivedCursor: cursor, failureCode };
}

function mergeResources(current: WorkflowRAGSnapshotResource[], incoming: WorkflowRAGSnapshotResource[]): WorkflowRAGSnapshotResource[] {
  return [...new Map([...current, ...incoming].map((resource) => [resource.snapshotId, resource])).values()].sort((left, right) => left.snapshotKey.localeCompare(right.snapshotKey));
}

function localFailure(failureCode: string, localMessage: SnapshotOperation["localMessage"] = "localRejected"): SnapshotOperation {
  return { status: "failed", record: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary: "", localMessage };
}
