import { workflowRAGSnapshotFindingMessage } from "./workflowRAGSnapshotMessages.ts";
import { useTranslation } from "react-i18next";
import "../../i18n/workflowRAGSnapshotResources.ts";

import {
  WORKFLOW_RAG_SNAPSHOT_LIMITS,
  WORKFLOW_RAG_SOURCE_TYPES,
  type WorkflowRAGContentClassification,
  type WorkflowRAGSourceType,
} from "./workflowRAGSnapshotConsumer.ts";
import {
  addWorkflowRAGSnapshotManualFragment,
  removeWorkflowRAGSnapshotEditorFragment,
  removeWorkflowRAGSnapshotEditorSource,
  selectWorkflowRAGSnapshotEditorFragment,
  updateWorkflowRAGSnapshotEditorFragment,
  updateWorkflowRAGSnapshotEditorSource,
  type WorkflowRAGSnapshotEditor,
  type WorkflowRAGSnapshotEditorAnalysis,
} from "./workflowRAGSnapshotEditor.ts";

export function WorkflowRAGSnapshotEditorPanel({
  editor,
  analysis,
  disabled,
  immutableSnapshotKey,
  importing,
  onChange,
  onImportFiles,
}: {
  editor: WorkflowRAGSnapshotEditor;
  analysis: WorkflowRAGSnapshotEditorAnalysis;
  disabled: boolean;
  immutableSnapshotKey: boolean;
  importing: boolean;
  onChange: (editor: WorkflowRAGSnapshotEditor) => void;
  onImportFiles: (files: File[]) => void;
}) {
  const { t } = useTranslation("workflow");
  const selectedFragment = editor.fragments.find((fragment) => fragment.fragmentId === editor.selectedFragmentId) ?? null;
  const selectedSource = selectedFragment ? editor.sources.find((source) => source.sourceId === selectedFragment.sourceId) ?? null : null;
  const findingCount = (target: "source" | "fragment", targetId: string) => analysis.findings.filter((finding) => finding.target === target && finding.targetId === targetId).length;

  return (
    <div className="workflow-rag-structured-editor">
      <div className="workflow-rag-snapshot-metadata">
        <label><span>{t($ => $.ragSnapshot.snapshotKey)}</span><input value={editor.snapshotKey} disabled={disabled || immutableSnapshotKey} maxLength={48} onChange={(event) => onChange({ ...editor, snapshotKey: event.currentTarget.value })} /></label>
        <label><span>{t($ => $.ragSnapshot.displayName)}</span><input value={editor.displayName} disabled={disabled} maxLength={120} onChange={(event) => onChange({ ...editor, displayName: event.currentTarget.value })} /></label>
        <label><span>{t($ => $.ragSnapshot.classification)}</span><select value={editor.contentClassification} disabled={disabled} onChange={(event) => onChange({ ...editor, contentClassification: event.currentTarget.value as WorkflowRAGContentClassification })}><option value="workspace_internal">{t($ => $.ragSnapshot.classifications.workspace_internal)}</option><option value="public">{t($ => $.ragSnapshot.classifications.public)}</option></select></label>
      </div>

      <section className="workflow-rag-local-import" aria-labelledby="workflow-rag-local-import-title">
        <div>
          <p className="eyebrow">{t($ => $.ragSnapshot.staging)}</p>
          <h5 id="workflow-rag-local-import-title">{t($ => $.ragSnapshot.importTitle)}</h5>
          <p>{t($ => $.ragSnapshot.importNote)}</p>
        </div>
        <label className={`workflow-rag-file-picker ${disabled ? "disabled" : ""}`}>
          <span>{importing ? t($ => $.ragSnapshot.readingFiles) : t($ => $.ragSnapshot.chooseFiles)}</span>
          <input
            type="file"
            accept=".md,.markdown,.txt,text/plain,text/markdown"
            multiple
            disabled={disabled}
            onChange={(event) => {
              const files = Array.from(event.currentTarget.files ?? []);
              event.currentTarget.value = "";
              if (files.length) onImportFiles(files);
            }}
          />
        </label>
      </section>

      <div className="workflow-rag-review-grid">
        <section className="workflow-rag-source-owner" aria-labelledby="workflow-rag-source-owner-title">
          <div className="workflow-rag-owner-heading">
            <div><p className="eyebrow">{t($ => $.ragSnapshot.sources)}</p><h5 id="workflow-rag-source-owner-title">{t($ => $.ragSnapshot.sourceReview)}</h5></div>
            <span>{editor.sources.length}</span>
          </div>
          <div className="workflow-rag-source-list">
            {editor.sources.map((source) => (
              <article key={source.sourceId} className={findingCount("source", source.sourceId) ? "blocked" : ""}>
                <div className="workflow-rag-source-heading">
                  <div><strong>{source.origin === "manual" ? t($ => $.ragSnapshot.manualSource) : source.label}</strong><small>{source.fileBytes ? formatBytes(source.fileBytes) : t($ => $.ragSnapshot.savedRecord)}{source.contentDigest ? ` · ${shortDigest(source.contentDigest)}` : ""}</small></div>
                  <button type="button" disabled={disabled} onClick={() => onChange(removeWorkflowRAGSnapshotEditorSource(editor, source.sourceId))}>{t($ => $.ragSnapshot.removeSource)}</button>
                </div>
                <div className="workflow-rag-source-fields">
                  <label><span>{t($ => $.ragSnapshot.sourceType)}</span><select value={source.sourceType} disabled={disabled} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { sourceType: event.currentTarget.value as WorkflowRAGSourceType }))}>{WORKFLOW_RAG_SOURCE_TYPES.map((sourceType) => <option key={sourceType} value={sourceType}>{t($ => $.ragSnapshot.sourceTypes[sourceType])}</option>)}</select></label>
                  <label className="wide"><span>{t($ => $.ragSnapshot.sourceRef)}</span><input value={source.sourceRef} disabled={disabled} maxLength={160} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { sourceRef: event.currentTarget.value }))} /></label>
                  <label className="workflow-rag-official-toggle"><input type="checkbox" checked={source.isOfficial} disabled={disabled} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { isOfficial: event.currentTarget.checked }))} /><span>{t($ => $.ragSnapshot.official)}</span></label>
                </div>
              </article>
            ))}
            {!editor.sources.length ? <p className="workflow-rag-owner-empty">{t($ => $.ragSnapshot.emptySources)}</p> : null}
          </div>
        </section>

        <section className="workflow-rag-fragment-owner" aria-labelledby="workflow-rag-fragment-owner-title">
          <div className="workflow-rag-owner-heading">
            <div><p className="eyebrow">{t($ => $.ragSnapshot.replacement)}</p><h5 id="workflow-rag-fragment-owner-title">{t($ => $.ragSnapshot.fragmentReview)}</h5></div>
            <button type="button" disabled={disabled || editor.fragments.length >= WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragments} onClick={() => onChange(addWorkflowRAGSnapshotManualFragment(editor))}>{t($ => $.ragSnapshot.addFragment)}</button>
          </div>
          <div className="workflow-rag-fragment-workspace">
            <div className="workflow-rag-fragment-list" aria-label={t($ => $.ragSnapshot.fragmentList)}>
              {editor.fragments.map((fragment, index) => {
                const source = editor.sources.find((candidate) => candidate.sourceId === fragment.sourceId);
                const count = findingCount("fragment", fragment.fragmentId);
                return (
                  <button key={fragment.fragmentId} type="button" disabled={disabled} className={`${editor.selectedFragmentId === fragment.fragmentId ? "selected" : ""} ${count ? "blocked" : ""}`} onClick={() => onChange(selectWorkflowRAGSnapshotEditorFragment(editor, fragment.fragmentId))}>
                    <span><strong>{String(index + 1).padStart(2, "0")} · {fragment.title || fragment.fragmentRef || t($ => $.ragSnapshot.unnamedFragment)}</strong><small>{source?.origin === "manual" ? t($ => $.ragSnapshot.manualSource) : source?.label ?? t($ => $.ragSnapshot.missingSource)}</small></span>
                    <span><small>{formatBytes(new TextEncoder().encode(fragment.content.trim()).byteLength)}</small>{count ? <em>{t($ => $.ragSnapshot.blockedCount, { count })}</em> : null}</span>
                  </button>
                );
              })}
              {!editor.fragments.length ? <p className="workflow-rag-owner-empty">{t($ => $.ragSnapshot.emptyFragments)}</p> : null}
            </div>

            {selectedFragment && selectedSource ? (
              <div className="workflow-rag-fragment-inspector">
                <div className="workflow-rag-inspector-heading"><div><p className="eyebrow">{t($ => $.ragSnapshot.selectedFragment)}</p><h6>{selectedFragment.title || selectedFragment.fragmentRef || t($ => $.ragSnapshot.unnamedFragment)}</h6></div><button type="button" disabled={disabled} onClick={() => onChange(removeWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId))}>{t($ => $.ragSnapshot.removeFragment)}</button></div>
                <div className="workflow-rag-fragment-fields">
                  <label><span>{t($ => $.ragSnapshot.fragmentRef)}</span><input value={selectedFragment.fragmentRef} disabled={disabled} maxLength={64} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { fragmentRef: event.currentTarget.value }))} /></label>
                  <label><span>{t($ => $.ragSnapshot.pageSlug)}</span><input value={selectedFragment.pageSlug} disabled={disabled} maxLength={120} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { pageSlug: event.currentTarget.value }))} /></label>
                  <label className="wide"><span>{t($ => $.ragSnapshot.title)}</span><input value={selectedFragment.title} disabled={disabled} maxLength={160} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { title: event.currentTarget.value }))} /></label>
                  <label className="wide"><span>{t($ => $.ragSnapshot.content)}</span><textarea value={selectedFragment.content} disabled={disabled} spellCheck={false} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { content: event.currentTarget.value }))} /><small>{formatBytes(new TextEncoder().encode(selectedFragment.content.trim()).byteLength)} / {formatBytes(WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragmentBytes)}</small></label>
                </div>
                <div className="workflow-rag-selected-source"><span>{t($ => $.ragSnapshot.sourceOwner)}</span><strong>{selectedSource.origin === "manual" ? t($ => $.ragSnapshot.manualSource) : selectedSource.label}</strong><code>{selectedSource.sourceRef}</code></div>
              </div>
            ) : <div className="workflow-rag-owner-empty">{t($ => $.ragSnapshot.selectFragment)}</div>}
          </div>
        </section>
      </div>

      <section className={`workflow-rag-findings ${analysis.findings.length ? "blocked" : "ready"}`} aria-live="polite">
        <div className="workflow-rag-owner-heading">
          <div><p className="eyebrow">{t($ => $.ragSnapshot.findingsBudget)}</p><h5>{analysis.findings.length ? t($ => $.ragSnapshot.blockedCount, { count: analysis.findings.length }) : t($ => $.ragSnapshot.preflightPassed)}</h5></div>
          <span>{t($ => $.ragSnapshot.budgetSummary, { sources: analysis.sourceCount, fragments: analysis.fragmentCount, bytes: formatBytes(analysis.totalContentBytes) })}</span>
        </div>
        {analysis.findings.length ? <ul>{analysis.findings.map((finding, index) => <li key={`${finding.code}:${finding.targetId}:${index}`}><code>{finding.code}</code><span>{workflowRAGSnapshotFindingMessage(t, finding.message)}</span></li>)}</ul> : <p>{t($ => $.ragSnapshot.preflightNote)}</p>}
      </section>

      <section className="workflow-rag-submit-boundary">
        <div><p className="eyebrow">{t($ => $.ragSnapshot.persistence)}</p><strong>{t($ => $.ragSnapshot.submitTitle)}</strong><p>{t($ => $.ragSnapshot.submitNote)}</p></div>
        <span className={`status-badge ${analysis.canSubmit ? "good" : "bad"}`}>{analysis.canSubmit ? t($ => $.ragSnapshot.readySubmit) : t($ => $.ragSnapshot.blockedSubmit)}</span>
      </section>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(bytes < 10 * 1024 ? 1 : 0)} KiB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`;
}

function shortDigest(digest: string): string {
  return digest.startsWith("sha256:") ? `${digest.slice(7, 15)}…${digest.slice(-4)}` : digest;
}
