import "../../i18n/artifactResources.ts";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation("applications");
  const [lifecycleState, setLifecycleState] = useState<ApplicationResultArtifactLifecycleState>("active");
  const [listing, setListing] = useState(() => initialApplicationResultArtifactListResult(config));
  const [items, setItems] = useState<ApplicationResultArtifactSummary[]>([]);
  const [selectedSummary, setSelectedSummary] = useState<ApplicationResultArtifactSummary | null>(null);
  const [readResult, setReadResult] = useState<ApplicationResultArtifactReadResult | null>(null);
  const [pending, setPending] = useState<"" | "list" | "more" | "read" | "transition">("");
  const [operationNotice, setOperationNotice] = useState<
    | { kind: "saved"; artifactId: string }
    | { kind: "read"; artifactId: string }
    | { kind: "transitioned"; state: ApplicationResultArtifactLifecycleState }
    | { kind: "conflict"; code: string; state: string; version: string }
    | { kind: "failure"; code: string }
    | null
  >(null);
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
    setOperationNotice(null);
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
    setOperationNotice({ kind: "saved", artifactId: latestArtifact.artifactId });
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
    setOperationNotice(result.status === "ready" ? null : { kind: "failure", code: result.failureCode });
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
    setOperationNotice(null);
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
    setOperationNotice(result.status === "ready" && result.artifact
      ? { kind: "read", artifactId: result.artifact.artifactId }
      : { kind: "failure", code: result.failureCode });
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
    setOperationNotice(result.status === "version_conflict" || result.status === "state_conflict"
      ? { kind: "conflict", code: result.failureCode, state: result.currentLifecycleState || "unknown", version: String(result.currentLifecycleVersion || "?") }
      : result.status === "ready" && result.lifecycle ? { kind: "transitioned", state: result.lifecycle.lifecycleState }
      : { kind: "failure", code: result.failureCode });
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
      summary: current.summary,
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
  const lifecycleLabel = (state: ApplicationResultArtifactLifecycleState) => state === "active" ? t($ => $.artifact.active) : t($ => $.artifact.archived);
  const listMessage = listing.status === "failed"
    ? t($ => $.artifact.listUnavailable, { code: listing.failureCode || t($ => $.artifact.unknownFailure) })
    : config.mode === "offline" ? t($ => $.artifact.offlineUnavailable)
    : pending === "list" ? t($ => $.artifact.loading)
    : listing.requestId ? t($ => $.artifact.loadedMetadata, { count: items.length, state: lifecycleLabel(lifecycleState) })
    : t($ => $.artifact.selectSession);
  const operationMessage = operationNotice?.kind === "saved"
    ? t($ => $.artifact.resultSavedArtifact, { artifactId: operationNotice.artifactId })
    : operationNotice?.kind === "read"
      ? t($ => $.artifact.readSucceeded, { artifactId: operationNotice.artifactId })
      : operationNotice?.kind === "transitioned"
        ? operationNotice.state === "active" ? t($ => $.artifact.unarchiveSucceeded) : t($ => $.artifact.archiveSucceeded)
        : operationNotice?.kind === "conflict"
          ? t($ => $.artifact.lifecycleConflict, { failureCode: operationNotice.code, state: operationNotice.state === "active" || operationNotice.state === "archived" ? lifecycleLabel(operationNotice.state) : t($ => $.artifact.unknown), version: operationNotice.version })
          : operationNotice?.kind === "failure"
            ? t($ => $.artifact.operationUnavailable, { code: operationNotice.code || t($ => $.artifact.unknownFailure) })
            : "";

  return (
    <section className="application-result-artifact-owner" aria-label={t($ => $.artifact.artifactRegion)}>
      <div className="application-api-card-heading">
        <div><p className="eyebrow">{t($ => $.artifact.explicitRetention)}</p><h4>{t($ => $.artifact.savedArtifacts)}</h4></div>
        <span className={`status-badge ${latestArtifact ? "good" : latestArtifactFailureCode ? "bad" : "neutral"}`}>
          {latestArtifact ? t($ => $.artifact.saved) : latestArtifactFailureCode ? t($ => $.artifact.saveFailed) : t($ => $.artifact.defaultOff)}
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
          <strong>{t($ => $.artifact.saveNextSuccess)}</strong>
          <small>{t($ => $.artifact.saveBoundary)}</small>
        </span>
      </label>

      {latestArtifactFailureCode ? <p className="failure-summary" role="alert">{t($ => $.artifact.resultSucceededSaveFailed)}{latestArtifactFailureCode}</p> : null}
      {latestArtifact ? (
        <p className="boundary-note">
          {t($ => $.artifact.savedPrefix)} <code>{latestArtifact.artifactId}</code> · {latestArtifact.contentType} · {t($ => $.artifact.contentBytes, { count: latestArtifact.contentBytes })}{t($ => $.artifact.bytesExactRead)}</p>
      ) : null}

      <div className="application-result-artifact-toolbar">
        <label>{t($ => $.artifact.lifecycle)}<select
            value={lifecycleState}
            disabled={!sessionId || Boolean(pending)}
            onChange={(event) => changeLifecycleState(event.target.value as ApplicationResultArtifactLifecycleState)}
          >
            <option value="active">{t($ => $.artifact.active)}</option>
            <option value="archived">{t($ => $.artifact.archived)}</option>
          </select>
        </label>
        <button
          type="button"
          className="secondary-action"
          disabled={!sessionId || Boolean(pending) || config.mode === "offline"}
          onClick={() => void loadArtifacts(lifecycleState, "", false)}
        >
          {pending === "list" ? t($ => $.artifact.loading) : t($ => $.artifact.refreshArtifacts)}
        </button>
        <span>{t($ => $.artifact.metadataRecordCount, { count: items.length })}</span>
      </div>

      <div className="application-result-artifact-layout">
        <div className="application-result-artifact-list" aria-label={t($ => $.artifact.resultArtifactListLabel, { state: lifecycleLabel(lifecycleState) })}>
          {items.length === 0 ? <p className="empty-state">{listMessage}</p> : items.map((item) => (
            <button
              type="button"
              key={item.artifactId}
              className={selectedSummary?.artifactId === item.artifactId ? "selected" : ""}
              disabled={Boolean(pending)}
              onClick={() => void openArtifact(item)}
            >
              <span><strong>{t($ => $.artifact.turnWithId, { turnId: item.turnId })}</strong><code>{item.artifactId}</code></span>
              <span><small>{item.contentType} · {t($ => $.artifact.contentBytes, { count: item.contentBytes })}</small><small>{t($ => $.artifact.lifecycleVersionValue, { state: lifecycleLabel(item.lifecycleState), version: item.lifecycleVersion })}</small></span>
            </button>
          ))}
          {listing.nextCursor ? (
            <button
              type="button"
              className="secondary-action"
              disabled={Boolean(pending)}
              onClick={() => void loadArtifacts(lifecycleState, listing.nextCursor, true)}
            >
              {pending === "more" ? t($ => $.artifact.loading) : t($ => $.artifact.loadMore)}
            </button>
          ) : null}
        </div>

        <article className="application-result-artifact-inspector">
          <div className="application-api-card-heading">
            <div><p className="eyebrow">{t($ => $.artifact.exactContentRead)}</p><h5>{selectedSummary?.artifactId ?? t($ => $.artifact.noArtifactSelected)}</h5></div>
            <span className={`status-badge ${readResult?.artifact ? "good" : readResult?.failureCode ? "bad" : "neutral"}`}>
              {pending === "read" ? t($ => $.artifact.reading) : currentLifecycle ? lifecycleLabel(currentLifecycle.lifecycleState) : t($ => $.artifact.metadataOnly)}
            </span>
          </div>
          {readResult?.artifact ? (
            <>
              <dl className="tenant-meta">
                <div><dt>{t($ => $.artifact.source)}</dt><dd>{readResult.artifact.executionProfile}</dd></div>
                <div><dt>{t($ => $.artifact.run)}</dt><dd>{readResult.artifact.runRef.schemaVersion}</dd></div>
                <div><dt>{t($ => $.artifact.digest)}</dt><dd><code>{readResult.artifact.contentDigest}</code></dd></div>
                <div><dt>{t($ => $.artifact.lifecycle)}</dt><dd>{readResult.lifecycle ? t($ => $.artifact.lifecycleVersionValue, { state: lifecycleLabel(readResult.lifecycle.lifecycleState), version: readResult.lifecycle.lifecycleVersion }) : ""}</dd></div>
              </dl>
              <pre>{readResult.artifact.content}</pre>
              <div className="application-result-artifact-actions">
                <button type="button" className="secondary-action" onClick={() => onOpenRun?.(readResult.artifact?.runRef.runId ?? "")}>{t($ => $.artifact.openExactRun)}</button>
                <button type="button" disabled={Boolean(pending)} onClick={() => void transitionSelectedArtifact()}>
                  {pending === "transition" ? t($ => $.artifact.updating) : currentLifecycle?.lifecycleState === "active" ? t($ => $.artifact.archiveArtifact) : t($ => $.artifact.unarchiveArtifact)}
                </button>
              </div>
            </>
          ) : selectedSummary ? (
            <p className="empty-state">{t($ => $.artifact.contentBoundary)}</p>
          ) : (
            <p className="empty-state">{t($ => $.artifact.selectFromSession, { state: lifecycleLabel(lifecycleState) })}</p>
          )}
          {readResult?.failureCode ? <p className="failure-summary" role="alert">{readResult.failureCode}</p> : null}
        </article>
      </div>

      <p className={listing.failureCode || operationNotice?.kind === "conflict" || operationNotice?.kind === "failure" ? "failure-summary" : "boundary-note"} aria-live="polite">
        {operationMessage || listMessage} {t($ => $.artifact.clearOnScopeChange)}</p>
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
