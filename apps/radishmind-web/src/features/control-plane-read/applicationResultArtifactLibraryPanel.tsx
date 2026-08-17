import { useEffect, useRef, useState } from "react";

import { readApplicationInteractionSessionConfig } from "./applicationInteractionSessionConsumer.ts";
import {
  applicationResultArtifactExportFilename,
  applicationResultArtifactLibraryResponseMatchesScope,
  exportApplicationResultArtifact,
  listApplicationResultArtifactsByApplication,
  readApplicationResultArtifact,
  serializeApplicationResultArtifactExport,
  transitionApplicationResultArtifactLifecycle,
  type ApplicationResultArtifactConfig,
  type ApplicationResultArtifactContentType,
  type ApplicationResultArtifactExecutionProfile,
  type ApplicationResultArtifactLibraryRequestScope,
  type ApplicationResultArtifactLifecycleState,
  type ApplicationResultArtifactListResult,
  type ApplicationResultArtifactReadResult,
  type ApplicationResultArtifactSummary,
} from "./applicationResultArtifactConsumer.ts";

const config: ApplicationResultArtifactConfig = readApplicationInteractionSessionConfig();

type LibraryFilters = {
  lifecycleState: ApplicationResultArtifactLifecycleState;
  executionProfile: ApplicationResultArtifactExecutionProfile | "";
  contentType: ApplicationResultArtifactContentType | "";
};

const INITIAL_FILTERS: LibraryFilters = {
  lifecycleState: "active",
  executionProfile: "",
  contentType: "",
};

export default function ApplicationResultArtifactLibraryPanel({
  applicationId,
  applicationName,
  active,
  onOpenRun,
}: {
  applicationId: string;
  applicationName: string;
  active: boolean;
  onOpenRun?: (runId: string) => void;
}) {
  const [draftFilters, setDraftFilters] = useState<LibraryFilters>(INITIAL_FILTERS);
  const [appliedFilters, setAppliedFilters] = useState<LibraryFilters>(INITIAL_FILTERS);
  const [listing, setListing] = useState<ApplicationResultArtifactListResult>(() => emptyListing());
  const [items, setItems] = useState<ApplicationResultArtifactSummary[]>([]);
  const [selected, setSelected] = useState<ApplicationResultArtifactSummary | null>(null);
  const [readResult, setReadResult] = useState<ApplicationResultArtifactReadResult | null>(null);
  const [pending, setPending] = useState<"" | "list" | "read" | "lifecycle" | "export">("");
  const [operationSummary, setOperationSummary] = useState("");
  const abortRef = useRef<AbortController | null>(null);
  const generationRef = useRef(0);
  const requestScopeRef = useRef<ApplicationResultArtifactLibraryRequestScope>(emptyScope());

  useEffect(() => {
    abortRef.current?.abort();
    generationRef.current += 1;
    requestScopeRef.current = emptyScope(generationRef.current, applicationId);
    setDraftFilters(INITIAL_FILTERS);
    setAppliedFilters(INITIAL_FILTERS);
    setListing(emptyListing());
    setItems([]);
    setSelected(null);
    setReadResult(null);
    setPending("");
    setOperationSummary("");
    if (!active || !applicationId || config.mode === "offline") return;
    void loadArtifacts(INITIAL_FILTERS, "", false);
    return () => {
      abortRef.current?.abort();
      generationRef.current += 1;
    };
  }, [active, applicationId]);

  function beginOperation(next: typeof pending): AbortController {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    setPending(next);
    return controller;
  }

  async function loadArtifacts(filters: LibraryFilters, cursor: string, append: boolean) {
    if (!active || !applicationId || config.mode === "offline") return;
    const generation = ++generationRef.current;
    const expected = libraryScope(generation, applicationId, filters, cursor);
    requestScopeRef.current = expected;
    setSelected(null);
    setReadResult(null);
    if (!append) setItems([]);
    const controller = beginOperation("list");
    const result = await listApplicationResultArtifactsByApplication(config, {
      applicationId,
      lifecycleState: filters.lifecycleState,
      executionProfile: filters.executionProfile,
      contentType: filters.contentType,
      limit: 50,
      cursor,
    }, controller.signal);
    if (!applicationResultArtifactLibraryResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setListing(result);
    setItems((current) => append && result.status === "ready" ? mergeArtifacts(current, result.items) : result.items);
    setOperationSummary(result.failureCode ? result.summary : "");
  }

  async function selectArtifact(summary: ApplicationResultArtifactSummary) {
    const generation = ++generationRef.current;
    const expected = libraryScope(generation, applicationId, appliedFilters, "", summary.sessionId, summary.artifactId);
    requestScopeRef.current = expected;
    setSelected(summary);
    setReadResult(null);
    setOperationSummary("");
    const controller = beginOperation("read");
    const result = await readApplicationResultArtifact(config, {
      applicationId,
      sessionId: summary.sessionId,
      artifactId: summary.artifactId,
    }, controller.signal);
    if (!applicationResultArtifactLibraryResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    if (!result.artifact || !result.lifecycle || result.artifact.contentDigest !== summary.contentDigest ||
      result.lifecycle.lifecycleVersion !== summary.lifecycleVersion || result.lifecycle.lifecycleState !== summary.lifecycleState) {
      setReadResult(null);
      setOperationSummary(result.failureCode ? result.summary : "精确读取与列表摘要不一致；已拒绝展示正文。");
      return;
    }
    setReadResult(result);
  }

  async function changeLifecycle() {
    const artifact = readResult?.artifact;
    const lifecycle = readResult?.lifecycle;
    if (!artifact || !lifecycle || pending) return;
    const targetState: ApplicationResultArtifactLifecycleState = lifecycle.lifecycleState === "active" ? "archived" : "active";
    const generation = ++generationRef.current;
    const expected = libraryScope(generation, applicationId, appliedFilters, "", artifact.sessionId, artifact.artifactId);
    requestScopeRef.current = expected;
    const controller = beginOperation("lifecycle");
    const result = await transitionApplicationResultArtifactLifecycle(config, {
      applicationId,
      sessionId: artifact.sessionId,
      artifactId: artifact.artifactId,
      expectedLifecycleVersion: lifecycle.lifecycleVersion,
      targetState,
    }, controller.signal);
    if (!applicationResultArtifactLibraryResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    setOperationSummary(result.summary);
    if (result.status !== "ready") return;
    setSelected(null);
    setReadResult(null);
    setItems((current) => current.filter((item) => item.artifactId !== artifact.artifactId));
  }

  async function downloadExport() {
    const summary = selected;
    const artifact = readResult?.artifact;
    const lifecycle = readResult?.lifecycle;
    if (!summary || !artifact || !lifecycle || pending) return;
    const generation = ++generationRef.current;
    const expected = libraryScope(generation, applicationId, appliedFilters, "", artifact.sessionId, artifact.artifactId);
    requestScopeRef.current = expected;
    const controller = beginOperation("export");
    const result = await exportApplicationResultArtifact(config, {
      applicationId,
      artifactId: artifact.artifactId,
    }, controller.signal);
    if (!applicationResultArtifactLibraryResponseMatchesScope(expected, requestScopeRef.current)) return;
    abortRef.current = null;
    setPending("");
    const exported = result.exportDocument;
    if (!exported || exported.artifact.artifactId !== summary.artifactId ||
      exported.artifact.sessionId !== summary.sessionId || exported.artifact.contentDigest !== summary.contentDigest ||
      exported.artifact.contentDigest !== artifact.contentDigest ||
      exported.lifecycle.lifecycleVersion !== lifecycle.lifecycleVersion ||
      exported.lifecycle.lifecycleState !== lifecycle.lifecycleState) {
      setOperationSummary(result.failureCode ? result.summary : "导出响应与当前精确资产不一致；已拒绝下载。");
      return;
    }
    const objectURL = URL.createObjectURL(new Blob(
      [serializeApplicationResultArtifactExport(exported)],
      { type: "application/json;charset=utf-8" },
    ));
    const anchor = document.createElement("a");
    anchor.href = objectURL;
    anchor.download = applicationResultArtifactExportFilename(exported);
    anchor.hidden = true;
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(objectURL);
    setOperationSummary(result.summary);
  }

  function applyFilters() {
    setAppliedFilters(draftFilters);
    void loadArtifacts(draftFilters, "", false);
  }

  function clearFilters() {
    setDraftFilters(INITIAL_FILTERS);
    setAppliedFilters(INITIAL_FILTERS);
    void loadArtifacts(INITIAL_FILTERS, "", false);
  }

  return (
    <section className="application-result-artifact-library" id="application-result-artifact-library" aria-label="Application result artifact library">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Application Result Workspace</p>
          <h3>Saved results across Sessions</h3>
          <p>发现当前 Application 下的 metadata-only 结果；仅在选择单项后精确读取正文。</p>
        </div>
        <span className={`status-badge ${listing.failureCode ? "bad" : items.length ? "good" : "neutral"}`}>
          {config.mode === "offline" ? "offline" : listing.failureCode ? "failed" : `${items.length} loaded`}
        </span>
      </div>

      <div className="application-result-artifact-toolbar application-result-artifact-library-filters">
        <label>Lifecycle
          <select
            value={draftFilters.lifecycleState}
            onChange={(event) => setDraftFilters((current) => ({ ...current, lifecycleState: event.target.value as ApplicationResultArtifactLifecycleState }))}
            disabled={Boolean(pending)}
          >
            <option value="active">Active</option>
            <option value="archived">Archived</option>
          </select>
        </label>
        <label>Execution profile
          <select
            value={draftFilters.executionProfile}
            onChange={(event) => setDraftFilters((current) => ({ ...current, executionProfile: event.target.value as LibraryFilters["executionProfile"] }))}
            disabled={Boolean(pending)}
          >
            <option value="">All profiles</option>
            <option value="workflow_definition_executor_v1">Workflow v1</option>
            <option value="workflow_definition_executor_v2">Workflow v2</option>
            <option value="application_rag_invocation_v1">Application RAG</option>
            <option value="prompt_application_invocation_v1">Prompt Application</option>
            <option value="agent_copilot_suggestion_v1">Agent Copilot</option>
          </select>
        </label>
        <label>Content type
          <select
            value={draftFilters.contentType}
            onChange={(event) => setDraftFilters((current) => ({ ...current, contentType: event.target.value as LibraryFilters["contentType"] }))}
            disabled={Boolean(pending)}
          >
            <option value="">All content types</option>
            <option value="text/markdown">Markdown</option>
            <option value="application/json">JSON</option>
          </select>
        </label>
        <button type="button" onClick={applyFilters} disabled={config.mode === "offline" || Boolean(pending)}>Apply filters</button>
        <button type="button" className="secondary-action" onClick={clearFilters} disabled={config.mode === "offline" || Boolean(pending)}>Clear</button>
      </div>

      <div className="application-result-artifact-layout">
        <div className="application-result-artifact-list" aria-label={`${appliedFilters.lifecycleState} application result artifacts`}>
          <header><strong>{applicationName}</strong><small>{appliedFilterSummary(appliedFilters)}</small></header>
          {items.length === 0 ? <p className="empty-state">{pending === "list" ? "正在读取结果资产元数据…" : listing.summary}</p> : items.map((item) => (
            <button
              type="button"
              key={item.artifactId}
              className={selected?.artifactId === item.artifactId ? "selected" : ""}
              onClick={() => void selectArtifact(item)}
              disabled={Boolean(pending)}
            >
              <span>
                <strong>{item.executionProfile}</strong>
                <code>{item.artifactId}</code>
                <small>{item.sessionId} · {item.contentType}</small>
              </span>
              <small>{new Date(item.createdAt).toLocaleString()}<br />lifecycle v{item.lifecycleVersion}</small>
            </button>
          ))}
          {listing.nextCursor ? (
            <button type="button" className="secondary-action" onClick={() => void loadArtifacts(appliedFilters, listing.nextCursor, true)} disabled={Boolean(pending)}>
              {pending === "list" ? "Loading…" : "Load more"}
            </button>
          ) : null}
        </div>

        <article className="application-result-artifact-inspector">
          {!selected ? <p className="empty-state">选择一条 metadata summary 后读取精确正文、来源与当前 lifecycle。</p> : null}
          {selected && pending === "read" ? <p>正在精确读取 <code>{selected.artifactId}</code>…</p> : null}
          {readResult?.artifact && readResult.lifecycle ? (
            <>
              <div className="application-api-card-heading">
                <div><p className="eyebrow">Exact artifact</p><h4>{readResult.artifact.artifactId}</h4></div>
                <span className={`status-badge ${readResult.lifecycle.lifecycleState === "active" ? "good" : "neutral"}`}>
                  {readResult.lifecycle.lifecycleState} · v{readResult.lifecycle.lifecycleVersion}
                </span>
              </div>
              <dl className="application-result-artifact-library-facts">
                <div><dt>Session</dt><dd><code>{readResult.artifact.sessionId}</code></dd></div>
                <div><dt>Turn</dt><dd><code>{readResult.artifact.turnId}</code></dd></div>
                <div><dt>Run</dt><dd><code>{readResult.artifact.runRef.runId}</code></dd></div>
                <div><dt>Digest</dt><dd><code>{readResult.artifact.contentDigest}</code></dd></div>
              </dl>
              <pre>{readResult.artifact.content}</pre>
              <div className="application-result-artifact-actions">
                <button type="button" onClick={() => onOpenRun?.(readResult.artifact!.runRef.runId)}>Open Run evidence</button>
                <button type="button" className="secondary-action" onClick={() => void changeLifecycle()} disabled={Boolean(pending)}>
                  {pending === "lifecycle" ? "Updating…" : readResult.lifecycle.lifecycleState === "active" ? "Archive" : "Unarchive"}
                </button>
                <button type="button" className="secondary-action" onClick={() => void downloadExport()} disabled={Boolean(pending)}>
                  {pending === "export" ? "Verifying export…" : "Export verified JSON"}
                </button>
              </div>
            </>
          ) : null}
          {operationSummary ? <p className={listing.failureCode || readResult?.failureCode ? "failure-summary" : "boundary-note"}>{operationSummary}</p> : null}
        </article>
      </div>

      <p className="boundary-note">
        列表不含正文；导出在下载前重读并校验 scope、artifact/content digest、lifecycle version 与 export digest。不会创建分享链接、导出记录或浏览器持久化副本。
      </p>
    </section>
  );
}

function emptyListing(): ApplicationResultArtifactListResult {
  return {
    status: config.mode === "offline" ? "offline" : "ready",
    items: [],
    nextCursor: "",
    failureCode: config.mode === "offline" ? "application_session_http_disabled" : "",
    requestId: "",
    auditRef: "",
    summary: config.mode === "offline"
      ? "Offline mode sends zero result artifact library requests."
      : "当前筛选条件下还没有已保存结果。",
  };
}

function emptyScope(generation = 0, applicationId = ""): ApplicationResultArtifactLibraryRequestScope {
  return libraryScope(generation, applicationId, INITIAL_FILTERS, "");
}

function libraryScope(
  generation: number,
  applicationId: string,
  filters: LibraryFilters,
  cursor: string,
  sessionId = "",
  artifactId = "",
): ApplicationResultArtifactLibraryRequestScope {
  return {
    generation,
    applicationId,
    lifecycleState: filters.lifecycleState,
    executionProfile: filters.executionProfile,
    contentType: filters.contentType,
    cursor,
    sessionId,
    artifactId,
  };
}

function mergeArtifacts(
  current: ApplicationResultArtifactSummary[],
  incoming: ApplicationResultArtifactSummary[],
): ApplicationResultArtifactSummary[] {
  const known = new Set(current.map((item) => item.artifactId));
  return [...current, ...incoming.filter((item) => !known.has(item.artifactId))];
}

function appliedFilterSummary(filters: LibraryFilters): string {
  return [filters.lifecycleState, filters.executionProfile || "all profiles", filters.contentType || "all content types"].join(" · ");
}
