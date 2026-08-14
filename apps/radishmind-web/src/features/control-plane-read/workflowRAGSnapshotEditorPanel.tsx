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
  const selectedFragment = editor.fragments.find((fragment) => fragment.fragmentId === editor.selectedFragmentId) ?? null;
  const selectedSource = selectedFragment ? editor.sources.find((source) => source.sourceId === selectedFragment.sourceId) ?? null : null;
  const findingCount = (target: "source" | "fragment", targetId: string) => analysis.findings.filter((finding) => finding.target === target && finding.targetId === targetId).length;

  return (
    <div className="workflow-rag-structured-editor">
      <div className="workflow-rag-snapshot-metadata">
        <label><span>Snapshot key</span><input value={editor.snapshotKey} disabled={disabled || immutableSnapshotKey} maxLength={48} onChange={(event) => onChange({ ...editor, snapshotKey: event.currentTarget.value })} /></label>
        <label><span>Display name</span><input value={editor.displayName} disabled={disabled} maxLength={120} onChange={(event) => onChange({ ...editor, displayName: event.currentTarget.value })} /></label>
        <label><span>Classification</span><select value={editor.contentClassification} disabled={disabled} onChange={(event) => onChange({ ...editor, contentClassification: event.currentTarget.value as WorkflowRAGContentClassification })}><option value="workspace_internal">workspace_internal</option><option value="public">public</option></select></label>
      </div>

      <section className="workflow-rag-local-import" aria-labelledby="workflow-rag-local-import-title">
        <div>
          <p className="eyebrow">Browser-local staging</p>
          <h5 id="workflow-rag-local-import-title">导入 Markdown / Text 材料</h5>
          <p>严格 UTF-8、本地确定性切分；新选择会完整替换当前未提交 fragment staging。</p>
        </div>
        <label className={`workflow-rag-file-picker ${disabled ? "disabled" : ""}`}>
          <span>{importing ? "读取中…" : "选择本地文件"}</span>
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
            <div><p className="eyebrow">Sources</p><h5 id="workflow-rag-source-owner-title">来源审查</h5></div>
            <span>{editor.sources.length}</span>
          </div>
          <div className="workflow-rag-source-list">
            {editor.sources.map((source) => (
              <article key={source.sourceId} className={findingCount("source", source.sourceId) ? "blocked" : ""}>
                <div className="workflow-rag-source-heading">
                  <div><strong>{source.label}</strong><small>{source.fileBytes ? formatBytes(source.fileBytes) : "saved record"}{source.contentDigest ? ` · ${shortDigest(source.contentDigest)}` : ""}</small></div>
                  <button type="button" disabled={disabled} onClick={() => onChange(removeWorkflowRAGSnapshotEditorSource(editor, source.sourceId))}>移除来源</button>
                </div>
                <div className="workflow-rag-source-fields">
                  <label><span>Source type</span><select value={source.sourceType} disabled={disabled} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { sourceType: event.currentTarget.value as WorkflowRAGSourceType }))}>{WORKFLOW_RAG_SOURCE_TYPES.map((sourceType) => <option key={sourceType} value={sourceType}>{sourceType}</option>)}</select></label>
                  <label className="wide"><span>Stable source ref</span><input value={source.sourceRef} disabled={disabled} maxLength={160} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { sourceRef: event.currentTarget.value }))} /></label>
                  <label className="workflow-rag-official-toggle"><input type="checkbox" checked={source.isOfficial} disabled={disabled} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorSource(editor, source.sourceId, { isOfficial: event.currentTarget.checked }))} /><span>Official source</span></label>
                </div>
              </article>
            ))}
            {!editor.sources.length ? <p className="workflow-rag-owner-empty">选择本地文件或新增手工 fragment 以建立来源。</p> : null}
          </div>
        </section>

        <section className="workflow-rag-fragment-owner" aria-labelledby="workflow-rag-fragment-owner-title">
          <div className="workflow-rag-owner-heading">
            <div><p className="eyebrow">Full replacement</p><h5 id="workflow-rag-fragment-owner-title">Fragment 审查</h5></div>
            <button type="button" disabled={disabled || editor.fragments.length >= WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragments} onClick={() => onChange(addWorkflowRAGSnapshotManualFragment(editor))}>新增手工 fragment</button>
          </div>
          <div className="workflow-rag-fragment-workspace">
            <div className="workflow-rag-fragment-list" aria-label="知识片段列表">
              {editor.fragments.map((fragment, index) => {
                const source = editor.sources.find((candidate) => candidate.sourceId === fragment.sourceId);
                const count = findingCount("fragment", fragment.fragmentId);
                return (
                  <button key={fragment.fragmentId} type="button" disabled={disabled} className={`${editor.selectedFragmentId === fragment.fragmentId ? "selected" : ""} ${count ? "blocked" : ""}`} onClick={() => onChange(selectWorkflowRAGSnapshotEditorFragment(editor, fragment.fragmentId))}>
                    <span><strong>{String(index + 1).padStart(2, "0")} · {fragment.title || fragment.fragmentRef || "未命名 fragment"}</strong><small>{source?.label ?? "missing source"}</small></span>
                    <span><small>{formatBytes(new TextEncoder().encode(fragment.content.trim()).byteLength)}</small>{count ? <em>{count} blocked</em> : null}</span>
                  </button>
                );
              })}
              {!editor.fragments.length ? <p className="workflow-rag-owner-empty">当前 replacement 没有 fragment，提交保持阻塞。</p> : null}
            </div>

            {selectedFragment && selectedSource ? (
              <div className="workflow-rag-fragment-inspector">
                <div className="workflow-rag-inspector-heading"><div><p className="eyebrow">Selected fragment</p><h6>{selectedFragment.title || selectedFragment.fragmentRef || "未命名 fragment"}</h6></div><button type="button" disabled={disabled} onClick={() => onChange(removeWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId))}>删除 fragment</button></div>
                <div className="workflow-rag-fragment-fields">
                  <label><span>Fragment ref</span><input value={selectedFragment.fragmentRef} disabled={disabled} maxLength={64} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { fragmentRef: event.currentTarget.value }))} /></label>
                  <label><span>Page slug</span><input value={selectedFragment.pageSlug} disabled={disabled} maxLength={120} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { pageSlug: event.currentTarget.value }))} /></label>
                  <label className="wide"><span>Title</span><input value={selectedFragment.title} disabled={disabled} maxLength={160} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { title: event.currentTarget.value }))} /></label>
                  <label className="wide"><span>Content</span><textarea value={selectedFragment.content} disabled={disabled} spellCheck={false} onChange={(event) => onChange(updateWorkflowRAGSnapshotEditorFragment(editor, selectedFragment.fragmentId, { content: event.currentTarget.value }))} /><small>{formatBytes(new TextEncoder().encode(selectedFragment.content.trim()).byteLength)} / {formatBytes(WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragmentBytes)}</small></label>
                </div>
                <div className="workflow-rag-selected-source"><span>Source owner</span><strong>{selectedSource.label}</strong><code>{selectedSource.sourceRef}</code></div>
              </div>
            ) : <div className="workflow-rag-owner-empty">选择一个 fragment 进行结构化审查。</div>}
          </div>
        </section>
      </div>

      <section className={`workflow-rag-findings ${analysis.findings.length ? "blocked" : "ready"}`} aria-live="polite">
        <div className="workflow-rag-owner-heading">
          <div><p className="eyebrow">Findings / budget</p><h5>{analysis.findings.length ? `${analysis.findings.length} 个阻塞项` : "本地预检通过"}</h5></div>
          <span>{analysis.sourceCount} sources · {analysis.fragmentCount} fragments · {formatBytes(analysis.totalContentBytes)}</span>
        </div>
        {analysis.findings.length ? <ul>{analysis.findings.map((finding, index) => <li key={`${finding.code}:${finding.targetId}:${index}`}><code>{finding.code}</code><span>{finding.summary}</span></li>)}</ul> : <p>最终 replacement 已通过本地完整校验；服务端仍会执行权威校验。</p>}
      </section>

      <section className="workflow-rag-submit-boundary">
        <div><p className="eyebrow">Persistence boundary</p><strong>只提交已审查的七字段 fragment replacement</strong><p>原始文件、basename、解析中间态、被删除片段和未提交 staging 不上传，也不进入浏览器持久化介质。</p></div>
        <span className={`status-badge ${analysis.canSubmit ? "good" : "bad"}`}>{analysis.canSubmit ? "ready to submit" : "submission blocked"}</span>
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
