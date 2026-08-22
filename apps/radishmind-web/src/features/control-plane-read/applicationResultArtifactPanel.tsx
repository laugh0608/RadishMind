import { useEffect, useRef, useState } from "react";

import {
  applicationResultArtifactResponseMatchesScope,
  initialApplicationResultArtifactListResult,
  listApplicationResultArtifacts,
  readApplicationResultArtifact,
  transitionApplicationResultArtifactLifecycle,
  type ApplicationResultArtifactConfig,
  type ApplicationResultArtifactLifecycleState,
  type ApplicationResultArtifactReadResult,
  type ApplicationResultArtifactRequestScope,
  type ApplicationResultArtifactSummary,
} from "./applicationResultArtifactConsumer.ts";

export default function ApplicationResultArtifactPanel({
  config,
  applicationId,
  sessionId,
  saveResult,
  onSaveResultChange,
  latestArtifact,
  latestArtifactFailureCode = "",
  disabled = false,
  onOpenRun,
}: {
  config: ApplicationResultArtifactConfig;
  applicationId: string;
  sessionId: string;
  saveResult: boolean;
  onSaveResultChange: (next: boolean) => void;
  latestArtifact: ApplicationResultArtifactSummary | null;
  latestArtifactFailureCode?: string;
  disabled?: boolean;
  onOpenRun?: (runId: string) => void;
}) {
  const [lifecycleState, setLifecycleState] = useState<ApplicationResultArtifactLifecycleState>("active");
  const [listing, setListing] = useState(() => initialApplicationResultArtifactListResult(config));
  const [items, setItems] = useState<ApplicationResultArtifactSummary[]>([]);
  const [selectedSummary, setSelectedSummary] = useState<ApplicationResultArtifactSummary | null>(null);
  const [readResult, setReadResult] = useState<ApplicationResultArtifactReadResult | null>(null);
  const [pending, setPending] = useState<"" | "list" | "more" | "read" | "transition">("");
  const [operationSummary, setOperationSummary] = useState("");
  const generationRef = useRef(0);
  const requestScopeRef = useRef<ApplicationResultArtifactRequestScope>({
    generation: 0,
    applicationId: "",
    sessionId: "",
    lifecycleState: "active",
    artifactId: "",
  });
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const generation = ++generationRef.current;
    abortRef.current?.abort();
    abortRef.current = null;
    requestScopeRef.current = { generation, applicationId, sessionId, lifecycleState: "active", artifactId: "" };
    setLifecycleState("active");
    setListing(initialApplicationResultArtifactListResult(config));
    setItems([]);
    setSelectedSummary(null);
    setReadResult(null);
    setPending("");
    setOperationSummary("");
    if (config.mode !== "offline" && applicationId && sessionId) {
      void loadArtifacts("active", "", false);
    }
    return () => {
      const cleanupGeneration = ++generationRef.current;
      requestScopeRef.current = { generation: cleanupGeneration, applicationId: "", sessionId: "", lifecycleState: "active", artifactId: "" };
      abortRef.current?.abort();
      abortRef.current = null;
    };
  }, [
    applicationId,
    config.baseUrl,
    config.mode,
    config.subjectRef,
    config.tenantRef,
    config.workspaceId,
    sessionId,
  ]);

  useEffect(() => {
    if (!latestArtifact || latestArtifact.applicationId !== applicationId || latestArtifact.sessionId !== sessionId) return;
    setOperationSummary(`本次成功结果已显式保存为 ${latestArtifact.artifactId}。`);
    if (latestArtifact.lifecycleState === lifecycleState) {
      setItems((current) => mergeArtifactSummaries(current, [latestArtifact]));
      setSelectedSummary(latestArtifact);
    }
  }, [applicationId, latestArtifact?.artifactId, latestArtifact?.lifecycleVersion, lifecycleState, sessionId]);

  async function loadArtifacts(
    targetState: ApplicationResultArtifactLifecycleState,
    cursor: string,
    append: boolean,
  ) {
    if (!applicationId || !sessionId || config.mode === "offline") return;
    const expected = beginOperation(append ? "more" : "list", targetState, "");
    const controller = abortRef.current!;
    const result = await listApplicationResultArtifacts(config, {
      applicationId,
      sessionId,
      lifecycleState: targetState,
      limit: 50,
      cursor,
    }, controller.signal);
    if (!applicationResultArtifactResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setListing(result);
    setItems((current) => append ? mergeArtifactSummaries(current, result.items) : result.items);
    setOperationSummary(result.summary);
    if (!append) {
      setSelectedSummary(null);
      setReadResult(null);
    }
  }

  function changeLifecycleState(next: ApplicationResultArtifactLifecycleState) {
    if (next === lifecycleState || pending) return;
    generationRef.current += 1;
    abortRef.current?.abort();
    abortRef.current = null;
    setLifecycleState(next);
    setItems([]);
    setSelectedSummary(null);
    setReadResult(null);
    setOperationSummary("");
    void loadArtifacts(next, "", false);
  }

  async function openArtifact(summary: ApplicationResultArtifactSummary) {
    const expected = beginOperation("read", lifecycleState, summary.artifactId);
    const controller = abortRef.current!;
    setSelectedSummary(summary);
    setReadResult(null);
    const result = await readApplicationResultArtifact(config, {
      applicationId,
      sessionId,
      artifactId: summary.artifactId,
    }, controller.signal);
    if (!applicationResultArtifactResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setReadResult(result);
    setOperationSummary(result.summary);
  }

  async function transitionSelectedArtifact() {
    if (!selectedSummary || pending) return;
    const lifecycle = readResult?.lifecycle;
    const currentState = lifecycle?.lifecycleState ?? selectedSummary.lifecycleState;
    const currentVersion = lifecycle?.lifecycleVersion ?? selectedSummary.lifecycleVersion;
    const targetState: ApplicationResultArtifactLifecycleState = currentState === "active" ? "archived" : "active";
    const expected = beginOperation("transition", lifecycleState, selectedSummary.artifactId);
    const controller = abortRef.current!;
    const result = await transitionApplicationResultArtifactLifecycle(config, {
      applicationId,
      sessionId,
      artifactId: selectedSummary.artifactId,
      expectedLifecycleVersion: currentVersion,
      targetState,
    }, controller.signal);
    if (!applicationResultArtifactResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setOperationSummary(result.status === "version_conflict" || result.status === "state_conflict"
      ? `${result.failureCode}：服务端当前为 ${result.currentLifecycleState || "unknown"} v${result.currentLifecycleVersion || "?"}，请刷新当前列表后重试。`
      : result.summary);
    if (result.status !== "ready" || !result.lifecycle) {
      if (result.status === "version_conflict" || result.status === "state_conflict") {
        setSelectedSummary(null);
        setReadResult(null);
      }
      return;
    }
    const nextSummary: ApplicationResultArtifactSummary = {
      ...selectedSummary,
      lifecycleState: result.lifecycle.lifecycleState,
      lifecycleVersion: result.lifecycle.lifecycleVersion,
      archivedAt: result.lifecycle.archivedAt,
      lifecycleUpdatedAt: result.lifecycle.updatedAt,
    };
    const nextItems = targetState === lifecycleState
      ? mergeArtifactSummaries(items, [nextSummary])
      : items.filter((item) => item.artifactId !== selectedSummary.artifactId);
    setItems(nextItems);
    setListing((current) => ({
      ...current,
      items: targetState === lifecycleState
        ? mergeArtifactSummaries(current.items, [nextSummary])
        : current.items.filter((item) => item.artifactId !== selectedSummary.artifactId),
      summary: `当前已加载 ${nextItems.length} 条 ${lifecycleState} 结果资产元数据；正文仍需精确读取。`,
    }));
    setSelectedSummary(nextSummary);
    setReadResult((current) => current?.artifact ? { ...current, lifecycle: result.lifecycle } : current);
  }

  function beginOperation(
    nextPending: typeof pending,
    targetState: ApplicationResultArtifactLifecycleState,
    artifactId: string,
  ): ApplicationResultArtifactRequestScope {
    abortRef.current?.abort();
    abortRef.current = new AbortController();
    const expected = {
      generation: generationRef.current,
      applicationId,
      sessionId,
      lifecycleState: targetState,
      artifactId,
    };
    requestScopeRef.current = expected;
    setPending(nextPending);
    return expected;
  }

  const canSave = config.mode !== "offline" && Boolean(applicationId && sessionId) && !disabled;
  const currentLifecycle = readResult?.lifecycle ?? (selectedSummary ? {
    lifecycleState: selectedSummary.lifecycleState,
    lifecycleVersion: selectedSummary.lifecycleVersion,
  } : null);

  return (
    <section className="application-result-artifact-owner" aria-label="Application result artifacts">
      <div className="application-api-card-heading">
        <div><p className="eyebrow">Explicit result retention</p><h4>Saved result artifacts</h4></div>
        <span className={`status-badge ${latestArtifact ? "good" : latestArtifactFailureCode ? "bad" : "neutral"}`}>
          {latestArtifact ? "saved" : latestArtifactFailureCode ? "save failed" : "default off"}
        </span>
      </div>

      <label className={`application-result-artifact-opt-in ${saveResult ? "selected" : ""}`}>
        <input
          type="checkbox"
          checked={saveResult}
          disabled={!canSave}
          onChange={(event) => onSaveResultChange(event.target.checked)}
        />
        <span>
          <strong>显式保存下一次成功结果</strong>
          <small>默认关闭；只保存服务端 canonical result，不保存输入、prompt、provider 原始响应或完整 transcript。</small>
        </span>
      </label>

      {latestArtifactFailureCode ? <p className="failure-summary" role="alert">结果执行成功，但保存失败：{latestArtifactFailureCode}</p> : null}
      {latestArtifact ? (
        <p className="boundary-note">
          已保存 <code>{latestArtifact.artifactId}</code> · {latestArtifact.contentType} · {latestArtifact.contentBytes} bytes；正文需精确读取。
        </p>
      ) : null}

      <div className="application-result-artifact-toolbar">
        <label>Lifecycle
          <select
            value={lifecycleState}
            disabled={!sessionId || Boolean(pending)}
            onChange={(event) => changeLifecycleState(event.target.value as ApplicationResultArtifactLifecycleState)}
          >
            <option value="active">Active</option>
            <option value="archived">Archived</option>
          </select>
        </label>
        <button
          type="button"
          className="secondary-action"
          disabled={!sessionId || Boolean(pending) || config.mode === "offline"}
          onClick={() => void loadArtifacts(lifecycleState, "", false)}
        >
          {pending === "list" ? "Loading…" : "Refresh artifacts"}
        </button>
        <span>{items.length} metadata record(s)</span>
      </div>

      <div className="application-result-artifact-layout">
        <div className="application-result-artifact-list" aria-label={`${lifecycleState} result artifacts`}>
          {items.length === 0 ? <p className="empty-state">{listing.summary}</p> : items.map((item) => (
            <button
              type="button"
              key={item.artifactId}
              className={selectedSummary?.artifactId === item.artifactId ? "selected" : ""}
              disabled={Boolean(pending)}
              onClick={() => void openArtifact(item)}
            >
              <span><strong>Turn {item.turnId}</strong><code>{item.artifactId}</code></span>
              <span><small>{item.contentType} · {item.contentBytes} bytes</small><small>{item.lifecycleState} · v{item.lifecycleVersion}</small></span>
            </button>
          ))}
          {listing.nextCursor ? (
            <button
              type="button"
              className="secondary-action"
              disabled={Boolean(pending)}
              onClick={() => void loadArtifacts(lifecycleState, listing.nextCursor, true)}
            >
              {pending === "more" ? "Loading…" : "Load more"}
            </button>
          ) : null}
        </div>

        <article className="application-result-artifact-inspector">
          <div className="application-api-card-heading">
            <div><p className="eyebrow">Exact content read</p><h5>{selectedSummary?.artifactId ?? "No artifact selected"}</h5></div>
            <span className={`status-badge ${readResult?.artifact ? "good" : readResult?.failureCode ? "bad" : "neutral"}`}>
              {pending === "read" ? "reading" : currentLifecycle?.lifecycleState ?? "metadata only"}
            </span>
          </div>
          {readResult?.artifact ? (
            <>
              <dl className="tenant-meta">
                <div><dt>Source</dt><dd>{readResult.artifact.executionProfile}</dd></div>
                <div><dt>Run</dt><dd>{readResult.artifact.runRef.schemaVersion}</dd></div>
                <div><dt>Digest</dt><dd><code>{readResult.artifact.contentDigest}</code></dd></div>
                <div><dt>Lifecycle</dt><dd>{readResult.lifecycle?.lifecycleState} · v{readResult.lifecycle?.lifecycleVersion}</dd></div>
              </dl>
              <pre>{readResult.artifact.content}</pre>
              <div className="application-result-artifact-actions">
                <button type="button" className="secondary-action" onClick={() => onOpenRun?.(readResult.artifact?.runRef.runId ?? "")}>Open exact run</button>
                <button type="button" disabled={Boolean(pending)} onClick={() => void transitionSelectedArtifact()}>
                  {pending === "transition" ? "Updating…" : currentLifecycle?.lifecycleState === "active" ? "Archive artifact" : "Unarchive artifact"}
                </button>
              </div>
            </>
          ) : selectedSummary ? (
            <p className="empty-state">选择记录后只读取该 artifact 的正文；列表、URL 与浏览器持久存储都不包含正文。</p>
          ) : (
            <p className="empty-state">从当前 Session 的 {lifecycleState} 列表中选择一个结果资产。</p>
          )}
          {readResult?.failureCode ? <p className="failure-summary" role="alert">{readResult.failureCode}</p> : null}
        </article>
      </div>

      <p className={listing.failureCode || operationSummary.includes("conflict") ? "failure-summary" : "boundary-note"} aria-live="polite">
        {operationSummary || listing.summary} Session、application、workspace、identity 或页面切换会清除正文并拒绝迟到响应。
      </p>
    </section>
  );
}

function mergeArtifactSummaries(
  current: ApplicationResultArtifactSummary[],
  next: ApplicationResultArtifactSummary[],
): ApplicationResultArtifactSummary[] {
  const byID = new Map(current.map((item) => [item.artifactId, item]));
  for (const item of next) byID.set(item.artifactId, item);
  return [...byID.values()].sort((left, right) =>
    right.createdAt.localeCompare(left.createdAt) || right.artifactId.localeCompare(left.artifactId)
  );
}
