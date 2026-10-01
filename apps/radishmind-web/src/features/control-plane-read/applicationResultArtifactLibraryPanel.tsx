import "../../i18n/artifactLibraryResources.ts";
import { useTranslation } from "react-i18next";
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

type PreparedExportDownload = {
  artifactId: string;
  lifecycleVersion: number;
  filename: string;
  objectURL: string;
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
  const { t } = useTranslation("applications");
  const [draftFilters, setDraftFilters] = useState<LibraryFilters>(INITIAL_FILTERS);
  const [appliedFilters, setAppliedFilters] = useState<LibraryFilters>(INITIAL_FILTERS);
  const [listing, setListing] = useState<ApplicationResultArtifactListResult>(() => emptyListing());
  const [items, setItems] = useState<ApplicationResultArtifactSummary[]>([]);
  const [selected, setSelected] = useState<ApplicationResultArtifactSummary | null>(null);
  const [readResult, setReadResult] = useState<ApplicationResultArtifactReadResult | null>(null);
  const [pending, setPending] = useState<"" | "list" | "read" | "lifecycle" | "export">("");
  const [operationNotice, setOperationNotice] = useState<
    | { kind: "failure"; code: string }
    | { kind: "readMismatch" }
    | { kind: "exportMismatch" }
    | { kind: "transitioned"; state: ApplicationResultArtifactLifecycleState }
    | { kind: "exportReady"; artifactId: string }
    | { kind: "downloadStarted"; filename: string }
    | null
  >(null);
  const [preparedExport, setPreparedExport] = useState<PreparedExportDownload | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const preparedExportRef = useRef<PreparedExportDownload | null>(null);
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
    setOperationNotice(null);
    discardPreparedExport();
    if (!active || !applicationId || config.mode === "offline") return;
    void loadArtifacts(INITIAL_FILTERS, "", false);
    return () => {
      abortRef.current?.abort();
      generationRef.current += 1;
      if (preparedExportRef.current) URL.revokeObjectURL(preparedExportRef.current.objectURL);
      preparedExportRef.current = null;
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
    discardPreparedExport();
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
    setOperationNotice(result.status === "ready" ? null : { kind: "failure", code: result.failureCode });
  }

  async function selectArtifact(summary: ApplicationResultArtifactSummary) {
    const generation = ++generationRef.current;
    const expected = libraryScope(generation, applicationId, appliedFilters, "", summary.sessionId, summary.artifactId);
    requestScopeRef.current = expected;
    discardPreparedExport();
    setSelected(summary);
    setReadResult(null);
    setOperationNotice(null);
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
      setOperationNotice(result.status !== "ready" ? { kind: "failure", code: result.failureCode } : { kind: "readMismatch" });
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
    discardPreparedExport();
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
    setOperationNotice(result.status === "ready" && result.lifecycle
      ? { kind: "transitioned", state: result.lifecycle.lifecycleState }
      : { kind: "failure", code: result.failureCode });
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
    if (result.status !== "ready" || !exported || exported.artifact.artifactId !== summary.artifactId ||
      exported.artifact.sessionId !== summary.sessionId || exported.artifact.contentDigest !== summary.contentDigest ||
      exported.artifact.contentDigest !== artifact.contentDigest ||
      exported.lifecycle.lifecycleVersion !== lifecycle.lifecycleVersion ||
      exported.lifecycle.lifecycleState !== lifecycle.lifecycleState) {
      setOperationNotice(result.status !== "ready" ? { kind: "failure", code: result.failureCode } : { kind: "exportMismatch" });
      return;
    }
    const prepared = {
      artifactId: exported.artifact.artifactId,
      lifecycleVersion: exported.lifecycle.lifecycleVersion,
      filename: applicationResultArtifactExportFilename(exported),
      objectURL: URL.createObjectURL(new Blob(
      [serializeApplicationResultArtifactExport(exported)],
      { type: "application/json;charset=utf-8" },
      )),
    } satisfies PreparedExportDownload;
    discardPreparedExport();
    preparedExportRef.current = prepared;
    setPreparedExport(prepared);
    setOperationNotice({ kind: "exportReady", artifactId: exported.artifact.artifactId });
  }

  function discardPreparedExport() {
    if (preparedExportRef.current) URL.revokeObjectURL(preparedExportRef.current.objectURL);
    preparedExportRef.current = null;
    setPreparedExport(null);
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

  const lifecycleLabel = (state: ApplicationResultArtifactLifecycleState) => state === "active" ? t($ => $.artifactLibrary.active) : t($ => $.artifactLibrary.archived);
  const profileLabel = (profile: ApplicationResultArtifactExecutionProfile | "") => {
    switch (profile) {
      case "workflow_definition_executor_v1": return t($ => $.artifactLibrary.workflowV1);
      case "workflow_definition_executor_v2": return t($ => $.artifactLibrary.workflowV2);
      case "application_rag_invocation_v1": return t($ => $.artifactLibrary.applicationRag);
      case "prompt_application_invocation_v1": return t($ => $.artifactLibrary.promptApplication);
      case "agent_copilot_suggestion_v1": return t($ => $.artifactLibrary.agentCopilot);
      default: return t($ => $.artifactLibrary.allProfiles);
    }
  };
  const contentTypeLabel = (contentType: ApplicationResultArtifactContentType | "") => contentType === "text/markdown"
    ? t($ => $.artifactLibrary.markdown) : contentType === "application/json" ? t($ => $.artifactLibrary.json) : t($ => $.artifactLibrary.allContentTypes);
  const listMessage = config.mode === "offline"
    ? t($ => $.artifactLibrary.offlineNoRequests)
    : listing.status === "failed" ? t($ => $.artifactLibrary.listUnavailable, { code: listing.failureCode || t($ => $.artifactLibrary.unknownFailure) })
    : listing.requestId ? t($ => $.artifactLibrary.listLoaded, { count: items.length, state: lifecycleLabel(appliedFilters.lifecycleState) })
    : t($ => $.artifactLibrary.noSavedResults);
  const operationMessage = operationNotice?.kind === "failure" ? t($ => $.artifactLibrary.operationUnavailable, { code: operationNotice.code || t($ => $.artifactLibrary.unknownFailure) })
    : operationNotice?.kind === "readMismatch" ? t($ => $.artifactLibrary.readMismatch)
    : operationNotice?.kind === "exportMismatch" ? t($ => $.artifactLibrary.exportMismatch)
    : operationNotice?.kind === "transitioned" ? operationNotice.state === "active" ? t($ => $.artifactLibrary.unarchiveSucceeded) : t($ => $.artifactLibrary.archiveSucceeded)
    : operationNotice?.kind === "exportReady" ? t($ => $.artifactLibrary.exportReady, { artifactId: operationNotice.artifactId })
    : operationNotice?.kind === "downloadStarted" ? t($ => $.artifactLibrary.downloadStarted, { filename: operationNotice.filename }) : "";

  return (
    <section className="application-result-artifact-library" id="application-result-artifact-library" aria-label={t($ => $.artifactLibrary.libraryRegion)}>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.artifactLibrary.workspace)}</p>
          <h3>{t($ => $.artifactLibrary.savedAcrossSessions)}</h3>
          <p>{t($ => $.artifactLibrary.metadataOnlyDiscovery)}</p>
        </div>
        <span className={`status-badge ${listing.failureCode ? "bad" : items.length ? "good" : "neutral"}`}>
          {config.mode === "offline" ? t($ => $.artifactLibrary.offline) : listing.failureCode ? t($ => $.artifactLibrary.failed) : t($ => $.artifactLibrary.loadedCount, { count: items.length })}
        </span>
      </div>

      <div className="application-result-artifact-toolbar application-result-artifact-library-filters">
        <label>{t($ => $.artifactLibrary.lifecycle)}<select
            value={draftFilters.lifecycleState}
            onChange={(event) => setDraftFilters((current) => ({ ...current, lifecycleState: event.target.value as ApplicationResultArtifactLifecycleState }))}
            disabled={Boolean(pending)}
          >
            <option value="active">{t($ => $.artifactLibrary.active)}</option>
            <option value="archived">{t($ => $.artifactLibrary.archived)}</option>
          </select>
        </label>
        <label>{t($ => $.artifactLibrary.executionProfile)}<select
            value={draftFilters.executionProfile}
            onChange={(event) => setDraftFilters((current) => ({ ...current, executionProfile: event.target.value as LibraryFilters["executionProfile"] }))}
            disabled={Boolean(pending)}
          >
            <option value="">{t($ => $.artifactLibrary.allProfiles)}</option>
            <option value="workflow_definition_executor_v1">{t($ => $.artifactLibrary.workflowV1)}</option>
            <option value="workflow_definition_executor_v2">{t($ => $.artifactLibrary.workflowV2)}</option>
            <option value="application_rag_invocation_v1">{t($ => $.artifactLibrary.applicationRag)}</option>
            <option value="prompt_application_invocation_v1">{t($ => $.artifactLibrary.promptApplication)}</option>
            <option value="agent_copilot_suggestion_v1">{t($ => $.artifactLibrary.agentCopilot)}</option>
          </select>
        </label>
        <label>{t($ => $.artifactLibrary.contentType)}<select
            value={draftFilters.contentType}
            onChange={(event) => setDraftFilters((current) => ({ ...current, contentType: event.target.value as LibraryFilters["contentType"] }))}
            disabled={Boolean(pending)}
          >
            <option value="">{t($ => $.artifactLibrary.allContentTypes)}</option>
            <option value="text/markdown">{t($ => $.artifactLibrary.markdown)}</option>
            <option value="application/json">{t($ => $.artifactLibrary.json)}</option>
          </select>
        </label>
        <button type="button" onClick={applyFilters} disabled={config.mode === "offline" || Boolean(pending)}>{t($ => $.artifactLibrary.applyFilters)}</button>
        <button type="button" className="secondary-action" onClick={clearFilters} disabled={config.mode === "offline" || Boolean(pending)}>{t($ => $.artifactLibrary.clear)}</button>
      </div>

      <div className="application-result-artifact-layout">
        <div className="application-result-artifact-list" aria-label={t($ => $.artifactLibrary.listLabel, { state: lifecycleLabel(appliedFilters.lifecycleState) })}>
          <header><strong>{applicationName}</strong><small>{[lifecycleLabel(appliedFilters.lifecycleState), profileLabel(appliedFilters.executionProfile), contentTypeLabel(appliedFilters.contentType)].join(" · ")}</small></header>
          {items.length === 0 ? <p className="empty-state">{pending === "list" ? t($ => $.artifactLibrary.loadingMetadata) : listMessage}</p> : items.map((item) => (
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
              <small>{new Date(item.createdAt).toLocaleString()}<br />{t($ => $.artifactLibrary.lifecycleVersionValue, { version: item.lifecycleVersion })}</small>
            </button>
          ))}
          {listing.nextCursor ? (
            <button type="button" className="secondary-action" onClick={() => void loadArtifacts(appliedFilters, listing.nextCursor, true)} disabled={Boolean(pending)}>
              {pending === "list" ? t($ => $.artifactLibrary.loading) : t($ => $.artifactLibrary.loadMore)}
            </button>
          ) : null}
        </div>

        <article className="application-result-artifact-inspector">
          {!selected ? <p className="empty-state">{t($ => $.artifactLibrary.selectToRead)}</p> : null}
          {selected && pending === "read" ? <p>{t($ => $.artifactLibrary.readingExact)}<code>{selected.artifactId}</code>…</p> : null}
          {readResult?.artifact && readResult.lifecycle ? (
            <>
              <div className="application-api-card-heading">
                <div><p className="eyebrow">{t($ => $.artifactLibrary.exactArtifact)}</p><h4>{readResult.artifact.artifactId}</h4></div>
                <span className={`status-badge ${readResult.lifecycle.lifecycleState === "active" ? "good" : "neutral"}`}>
                  {t($ => $.artifactLibrary.lifecycleStateVersion, { state: lifecycleLabel(readResult.lifecycle.lifecycleState), version: readResult.lifecycle.lifecycleVersion })}
                </span>
              </div>
              <dl className="application-result-artifact-library-facts">
                <div><dt>{t($ => $.artifactLibrary.session)}</dt><dd><code>{readResult.artifact.sessionId}</code></dd></div>
                <div><dt>{t($ => $.artifactLibrary.turn)}</dt><dd><code>{readResult.artifact.turnId}</code></dd></div>
                <div><dt>{t($ => $.artifactLibrary.run)}</dt><dd><code>{readResult.artifact.runRef.runId}</code></dd></div>
                <div><dt>{t($ => $.artifactLibrary.digest)}</dt><dd><code>{readResult.artifact.contentDigest}</code></dd></div>
              </dl>
              <pre>{readResult.artifact.content}</pre>
              <div className="application-result-artifact-actions">
                <button type="button" onClick={() => onOpenRun?.(readResult.artifact!.runRef.runId)}>{t($ => $.artifactLibrary.openRunEvidence)}</button>
                <button type="button" className="secondary-action" onClick={() => void changeLifecycle()} disabled={Boolean(pending)}>
                  {pending === "lifecycle" ? t($ => $.artifactLibrary.updating) : readResult.lifecycle.lifecycleState === "active" ? t($ => $.artifactLibrary.archive) : t($ => $.artifactLibrary.unarchive)}
                </button>
                <button type="button" className="secondary-action" onClick={() => void downloadExport()} disabled={Boolean(pending)}>
                  {pending === "export" ? t($ => $.artifactLibrary.verifyingExport) : t($ => $.artifactLibrary.prepareVerifiedJson)}
                </button>
                {preparedExport?.artifactId === readResult.artifact.artifactId &&
                preparedExport.lifecycleVersion === readResult.lifecycle.lifecycleVersion ? (
                  <a
                    className="application-result-artifact-download"
                    href={preparedExport.objectURL}
                    download={preparedExport.filename}
                    onClick={() => setOperationNotice({ kind: "downloadStarted", filename: preparedExport.filename })}
                  >
                    {t($ => $.artifactLibrary.downloadVerifiedJson)}</a>
                ) : null}
              </div>
            </>
          ) : null}
          {operationMessage ? <p className={operationNotice?.kind === "failure" || operationNotice?.kind === "readMismatch" || operationNotice?.kind === "exportMismatch" ? "failure-summary" : "boundary-note"}>{operationMessage}</p> : null}
        </article>
      </div>

      <p className="boundary-note">
        {t($ => $.artifactLibrary.exportBoundary)}</p>
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
    summary: "",
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
