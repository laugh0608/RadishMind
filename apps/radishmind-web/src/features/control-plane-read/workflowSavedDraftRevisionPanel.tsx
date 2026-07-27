import { useEffect, useMemo, useState } from "react";

import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import type {
  WorkflowSavedDraftConsumerConfig,
} from "./savedWorkflowDraftConsumer.ts";
import { compareWorkflowSavedDraftRevision } from "./workflowSavedDraftRevisionComparison.ts";
import {
  initialWorkflowSavedDraftRevisionHistoryState,
  listWorkflowSavedDraftRevisions,
  readWorkflowSavedDraftRevision,
  restoreWorkflowSavedDraftRevision,
  type WorkflowSavedDraftRevisionDetail,
  type WorkflowSavedDraftRevisionHistoryState,
  type WorkflowSavedDraftRevisionRestoreResult,
} from "./workflowSavedDraftRevisionConsumer.ts";

export function WorkflowSavedDraftRevisionPanel({
  draft,
  currentDraftVersion,
  config,
  dirty,
  disabled,
  onRestored,
}: {
  draft: WorkflowDraftDesignerDraft;
  currentDraftVersion: number;
  config: WorkflowSavedDraftConsumerConfig;
  dirty: boolean;
  disabled: boolean;
  onRestored: (
    draft: WorkflowDraftDesignerDraft,
    result: WorkflowSavedDraftRevisionRestoreResult,
  ) => void;
}) {
  const [history, setHistory] = useState<WorkflowSavedDraftRevisionHistoryState>(() =>
    initialWorkflowSavedDraftRevisionHistoryState(config),
  );
  const [selected, setSelected] = useState<WorkflowSavedDraftRevisionDetail | null>(null);
  const [detailStatus, setDetailStatus] = useState<"idle" | "loading" | "ready" | "failed">("idle");
  const [operationSummary, setOperationSummary] = useState("");
  const [confirmVersion, setConfirmVersion] = useState<number | null>(null);
  const [restoring, setRestoring] = useState(false);
  const comparison = useMemo(
    () => selected ? compareWorkflowSavedDraftRevision(selected.draft, draft) : null,
    [draft, selected],
  );

  useEffect(() => {
    setHistory(initialWorkflowSavedDraftRevisionHistoryState(config));
    setSelected(null);
    setDetailStatus("idle");
    setOperationSummary("");
    setConfirmVersion(null);
    setRestoring(false);
  }, [config.mode, draft.applicationRef, draft.draftId]);

  const loadHistory = async (cursor = "") => {
    setHistory((state) => ({
      ...state,
      status: "loading",
      failureCode: null,
      summary: cursor ? "正在读取更早的修订记录。" : "正在读取草案修订历史。",
    }));
    try {
      const result = await listWorkflowSavedDraftRevisions(draft, config, cursor);
      setHistory((state) => cursor && result.status === "ready"
        ? {
            ...result,
            revisions: [...state.revisions, ...result.revisions],
            summary: `已读取 ${state.revisions.length + result.revisions.length} 条不可变修订记录。`,
          }
        : result);
    } catch (error) {
      setHistory((state) => ({
        ...state,
        status: "failed",
        failureCode: "draft_revision_history_request_failed",
        summary: error instanceof Error ? error.message : "修订历史读取失败。",
      }));
    }
  };

  const selectRevision = async (draftVersion: number) => {
    setDetailStatus("loading");
    setOperationSummary(`正在读取版本 ${draftVersion}。`);
    setConfirmVersion(null);
    try {
      const detail = await readWorkflowSavedDraftRevision(draft, draftVersion, config);
      setSelected(detail);
      setDetailStatus("ready");
      setOperationSummary(`版本 ${draftVersion} 已加载，可与当前工作区草案比较。`);
    } catch (error) {
      setSelected(null);
      setDetailStatus("failed");
      setOperationSummary(error instanceof Error ? error.message : "修订详情读取失败。");
    }
  };

  const restoreSelectedRevision = async () => {
    if (!selected || confirmVersion !== selected.draftVersion || currentDraftVersion < 1) return;
    setRestoring(true);
    setOperationSummary(`正在从版本 ${selected.draftVersion} 创建新修订。`);
    try {
      const result = await restoreWorkflowSavedDraftRevision(
        draft,
        selected.draftVersion,
        currentDraftVersion,
        config,
      );
      if (!result.draft) {
        setOperationSummary(result.summary);
        return;
      }
      onRestored(result.draft, result);
      setOperationSummary(result.summary);
      setSelected(null);
      setConfirmVersion(null);
      await loadHistory();
    } catch (error) {
      setOperationSummary(error instanceof Error ? error.message : "修订恢复失败。");
    } finally {
      setRestoring(false);
    }
  };

  const canUseHistory = config.mode === "dev_saved_draft_http" && currentDraftVersion > 0;
  return (
    <section className="workflow-draft-revision-panel" aria-label="草案修订历史与恢复">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Revision History</p>
          <h4>草案修订历史与恢复</h4>
        </div>
        <span className="status-badge neutral">
          {currentDraftVersion > 0 ? `current v${currentDraftVersion}` : "尚未保存"}
        </span>
      </div>
      <p>{history.summary}</p>
      <div className="workflow-draft-action-row">
        <button
          type="button"
          disabled={!canUseHistory || disabled || history.status === "loading" || restoring}
          onClick={() => void loadHistory()}
        >
          刷新历史
        </button>
        {history.hasMore ? (
          <button
            type="button"
            disabled={disabled || history.status === "loading" || restoring}
            onClick={() => void loadHistory(history.nextCursor)}
          >
            读取更早版本
          </button>
        ) : null}
      </div>
      {history.revisions.length > 0 ? (
        <div className="workflow-draft-revision-grid">
          <div className="workflow-draft-revision-list" aria-label="草案修订列表">
            {history.revisions.map((revision) => (
              <button
                type="button"
                key={revision.draftVersion}
                className={selected?.draftVersion === revision.draftVersion ? "selected" : ""}
                disabled={disabled || restoring || detailStatus === "loading"}
                onClick={() => void selectRevision(revision.draftVersion)}
              >
                <strong>v{revision.draftVersion} · {revision.revisionKind}</strong>
                <span>{revision.name}</span>
                <small>{revision.updatedAt} · {revision.nodeCount} nodes · {revision.edgeCount} edges</small>
                {revision.restoredFromVersion > 0 ? <small>restored from v{revision.restoredFromVersion}</small> : null}
              </button>
            ))}
          </div>
          <article className="workflow-draft-card workflow-draft-revision-detail">
            <span>版本比较</span>
            {selected && comparison ? (
              <>
                <strong>v{selected.draftVersion} → 当前工作区</strong>
                <p>
                  元数据 {comparison.metadataChangeCount} 项、节点 {comparison.nodeChangeCount} 项、边
                  {comparison.edgeChangeCount} 项、布局 / 审查上下文
                  {comparison.reviewContextChangeCount} 项差异。
                </p>
                <ul>
                  {comparison.changes.length === 0
                    ? <li>所选修订与当前工作区草案一致。</li>
                    : comparison.changes.slice(0, 12).map((change) => (
                        <li key={`${change.kind}:${change.subject}`}>
                          <code>{change.subject}</code> {change.summary}
                        </li>
                      ))}
                </ul>
                {confirmVersion === selected.draftVersion ? (
                  <div className="workflow-draft-revision-confirm">
                    <p>
                      确认后会以版本 {selected.draftVersion} 的内容创建版本 {currentDraftVersion + 1}；
                      现有版本不会被改写{dirty ? "，当前未保存编辑将被替换" : ""}。
                    </p>
                    <button
                      type="button"
                      disabled={disabled || restoring}
                      onClick={() => void restoreSelectedRevision()}
                    >
                      {restoring ? "正在恢复…" : "确认创建新修订"}
                    </button>
                    <button type="button" disabled={restoring} onClick={() => setConfirmVersion(null)}>
                      取消
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    disabled={disabled || restoring || selected.draftVersion === currentDraftVersion}
                    onClick={() => setConfirmVersion(selected.draftVersion)}
                  >
                    准备从此版本恢复
                  </button>
                )}
              </>
            ) : (
              <p>{detailStatus === "loading" ? "正在读取修订详情。" : "选择一个历史版本以查看结构化差异。"}</p>
            )}
            {operationSummary ? <p>{operationSummary}</p> : null}
          </article>
        </div>
      ) : null}
      {history.failureCode ? <p>failure: {history.failureCode}</p> : null}
      <p className="workflow-draft-revision-stopline">
        恢复只创建新的当前修订，不会删除、覆盖或就地修改任何历史版本。
      </p>
    </section>
  );
}
