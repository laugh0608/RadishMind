import { formatDisplayDate } from "../../i18n/formatters.ts";
import "../../i18n/workflowRevisionResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import { workflowRevisionHistoryMessage, workflowRevisionOperationMessage, workflowRevisionChangeMessage, type WorkflowRevisionOperation } from "./workflowRevisionMessages.ts";
import { useTranslation } from "react-i18next";
import "../../i18n/workflowDraftResources.ts";
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
  currentLifecycleVersion,
  lifecycleState,
  config,
  dirty,
  disabled,
  onRestored,
}: {
  draft: WorkflowDraftDesignerDraft;
  currentDraftVersion: number;
  currentLifecycleVersion: number;
  lifecycleState: "active" | "archived" | "unknown";
  config: WorkflowSavedDraftConsumerConfig;
  dirty: boolean;
  disabled: boolean;
  onRestored: (
    draft: WorkflowDraftDesignerDraft,
    result: WorkflowSavedDraftRevisionRestoreResult,
  ) => void;
}) {
  const { t, i18n } = useTranslation("workflow");
  const [history, setHistory] = useState<WorkflowSavedDraftRevisionHistoryState>(() =>
    initialWorkflowSavedDraftRevisionHistoryState(config),
  );
  const [selected, setSelected] = useState<WorkflowSavedDraftRevisionDetail | null>(null);
  const [detailStatus, setDetailStatus] = useState<"idle" | "loading" | "ready" | "failed">("idle");
  const [operation, setOperation] = useState<WorkflowRevisionOperation | null>(null);
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
    setOperation(null);
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
    } catch {
      setHistory((state) => ({
        ...state,
        status: "failed",
        failureCode: "draft_revision_history_request_failed",
        summary: "修订历史读取失败。",
      }));
    }
  };

  const selectRevision = async (draftVersion: number) => {
    setDetailStatus("loading");
    setOperation({ code: "reading", version: draftVersion });
    setConfirmVersion(null);
    try {
      const detail = await readWorkflowSavedDraftRevision(draft, draftVersion, config);
      setSelected(detail);
      setDetailStatus("ready");
      setOperation({ code: "ready", version: draftVersion });
    } catch {
      setSelected(null);
      setDetailStatus("failed");
      setOperation({ code: "read_failed" });
    }
  };

  const restoreSelectedRevision = async () => {
    if (
      !selected ||
      confirmVersion !== selected.draftVersion ||
      currentDraftVersion < 1 ||
      currentLifecycleVersion < 1 ||
      lifecycleState !== "active"
    ) return;
    setRestoring(true);
    setOperation({ code: "restoring", version: selected.draftVersion });
    try {
      const result = await restoreWorkflowSavedDraftRevision(
        draft,
        selected.draftVersion,
        currentDraftVersion,
        currentLifecycleVersion,
        config,
      );
      if (!result.draft) {
        setOperation({ code: "restore_failed", failureCode: result.failureCode ?? "draft_revision_restore_failed" });
        return;
      }
      onRestored(result.draft, result);
      setOperation({ code: "restored", version: result.currentDraftVersion });
      setSelected(null);
      setConfirmVersion(null);
      await loadHistory();
    } catch {
      setOperation({ code: "restore_failed", failureCode: "draft_revision_restore_request_failed" });
    } finally {
      setRestoring(false);
    }
  };

  const canUseHistory = config.mode === "dev_saved_draft_http" && currentDraftVersion > 0;
  const restoreBlocked = lifecycleState !== "active";
  return (
    <section className="workflow-draft-revision-panel" aria-label={t($ => $.revision.revisionPanel)}>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.revision.revisionHistory)}</p>
          <h4>{t($ => $.revision.revisionPanel)}</h4>
        </div>
        <span className="status-badge neutral">
          {currentDraftVersion > 0
            ? t($ => $.revision.revisionCurrent, { content: currentDraftVersion, lifecycle: currentLifecycleVersion, state: workflowDraftStatusLabel(t, lifecycleState) })
            : t($ => $.revision.neverSaved)}
        </span>
      </div>
      {lifecycleState === "archived" ? (
        <p className="workflow-draft-revision-stopline">
          {t($ => $.revision.archivedRestore)}
        </p>
      ) : null}
      <p>{workflowRevisionHistoryMessage(t, history)}</p>
      <div className="workflow-draft-action-row">
        <button
          type="button"
          disabled={!canUseHistory || disabled || history.status === "loading" || restoring}
          onClick={() => void loadHistory()}
        >{t($ => $.revision.refreshHistory)}</button>
        {history.hasMore ? (
          <button
            type="button"
            disabled={disabled || history.status === "loading" || restoring}
            onClick={() => void loadHistory(history.nextCursor)}
          >{t($ => $.revision.olderRevisions)}</button>
        ) : null}
      </div>
      {history.revisions.length > 0 ? (
        <div className="workflow-draft-revision-grid">
          <div className="workflow-draft-revision-list" aria-label={t($ => $.revision.revisionList)}>
            {history.revisions.map((revision) => (
              <button
                type="button"
                key={revision.draftVersion}
                className={selected?.draftVersion === revision.draftVersion ? "selected" : ""}
                disabled={disabled || restoring || detailStatus === "loading"}
                onClick={() => void selectRevision(revision.draftVersion)}
              >
                <strong>v{revision.draftVersion} · {workflowDraftStatusLabel(t, revision.revisionKind)}</strong>
                <span>{revision.name}</span>
                <small title={revision.updatedAt}>{t($ => $.revision.revisionFacts, { time: formatDisplayDate(revision.updatedAt, i18n.language === "en-US" ? "en-US" : "zh-CN") ?? t($ => $.draft.statusUnknown), nodes: revision.nodeCount, edges: revision.edgeCount })}</small>
                {revision.restoredFromVersion > 0 ? <small>{t($ => $.revision.restoredFrom, { version: revision.restoredFromVersion })}</small> : null}
              </button>
            ))}
          </div>
          <article className="workflow-draft-card workflow-draft-revision-detail">
            <span>{t($ => $.revision.comparison)}</span>
            {selected && comparison ? (
              <>
                <strong>{t($ => $.revision.compareCurrent, { version: selected.draftVersion })}</strong>
                <p>
                  {t($ => $.revision.differenceCounts, { metadata: comparison.metadataChangeCount, nodes: comparison.nodeChangeCount, edges: comparison.edgeChangeCount, context: comparison.reviewContextChangeCount })}
                </p>
                <ul>
                  {comparison.changes.length === 0
                    ? <li>{t($ => $.revision.sameRevision)}</li>
                    : comparison.changes.slice(0, 12).map((change) => (
                        <li key={`${change.kind}:${change.subject}`}>
                          <code>{change.subject}</code> {workflowRevisionChangeMessage(t, change)}
                        </li>
                      ))}
                </ul>
                {confirmVersion === selected.draftVersion ? (
                  <div className="workflow-draft-revision-confirm">
                    <p>
                      {t($ => $.revision.restoreConfirm, { source: selected.draftVersion, next: currentDraftVersion + 1 })}
                      {dirty ? <strong>{t($ => $.revision.replaceUnsaved)}</strong> : null}
                    </p>
                    <button
                      type="button"
                      disabled={disabled || restoring || restoreBlocked}
                      onClick={() => void restoreSelectedRevision()}
                    >
                      {restoring ? t($ => $.revision.restoring) : t($ => $.revision.confirmRestore)}
                    </button>
                    <button type="button" disabled={restoring} onClick={() => setConfirmVersion(null)}>{t($ => $.draft.cancel)}</button>
                  </div>
                ) : (
                  <button
                    type="button"
                    disabled={
                      disabled ||
                      restoring ||
                      restoreBlocked ||
                      selected.draftVersion === currentDraftVersion
                    }
                    onClick={() => setConfirmVersion(selected.draftVersion)}
                  >{t($ => $.revision.prepareRestore)}</button>
                )}
                {restoreBlocked ? (
                  <p>{t($ => $.revision.restoreUnavailable, { state: workflowDraftStatusLabel(t, lifecycleState) })}</p>
                ) : null}
              </>
            ) : (
              <p>{detailStatus === "loading" ? t($ => $.revision.detailLoading) : t($ => $.revision.selectRevision)}</p>
            )}
            {operation ? <p role="status">{workflowRevisionOperationMessage(t, operation)}</p> : null}
          </article>
        </div>
      ) : null}
      {history.failureCode ? <p>{t($ => $.draft.failure)}: <code>{history.failureCode}</code></p> : null}
      <p className="workflow-draft-revision-stopline">{t($ => $.revision.restoreBoundary)}</p>
    </section>
  );
}
