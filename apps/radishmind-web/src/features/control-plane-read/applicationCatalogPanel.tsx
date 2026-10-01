import "../../i18n/catalogResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useState } from "react";

import {
  archiveApplicationCatalogRecord,
  createApplicationCatalogRecord,
  listApplicationCatalogRecords,
  readApplicationCatalogConfig,
  readApplicationCatalogRecord,
  unarchiveApplicationCatalogRecord,
  updateApplicationCatalogRecord,
  validateApplicationCatalogMutableFields,
  type ApplicationCatalogKind,
  type ApplicationCatalogLifecycleState,
  type ApplicationCatalogMutableFields,
  type ApplicationCatalogOperationResult,
  type ApplicationCatalogRecord,
} from "./applicationCatalogConsumer.ts";

const defaultConfig = readApplicationCatalogConfig();
const EMPTY_FIELDS: ApplicationCatalogMutableFields = {
  displayName: "",
  description: "",
  applicationKind: "agent",
};

export type ApplicationCatalogSnapshot = {
  status: "offline" | "loading" | "ready" | "failed";
  records: ApplicationCatalogRecord[];
  failureCode: string;
  summary: string;
};

export function ApplicationCatalogPanel({
  workspaceId,
  selectedApplicationId,
  onSelectRecord,
  onSnapshotChange,
}: {
  workspaceId: string;
  selectedApplicationId: string | null;
  onSelectRecord: (record: ApplicationCatalogRecord) => void;
  onSnapshotChange: (snapshot: ApplicationCatalogSnapshot) => void;
}) {
  const { t } = useTranslation("applications");
  const config = useMemo(() => ({ ...defaultConfig, workspaceId }), [workspaceId]);
  const [snapshot, setSnapshot] = useState<ApplicationCatalogSnapshot>(() => initialSnapshot(config.mode));
  const [filter, setFilter] = useState<ApplicationCatalogLifecycleState>("active");
  const [nextCursors, setNextCursors] = useState<Record<ApplicationCatalogLifecycleState, string>>({ active: "", archived: "" });
  const [selectedRecord, setSelectedRecord] = useState<ApplicationCatalogRecord | null>(null);
  const [fields, setFields] = useState<ApplicationCatalogMutableFields>(EMPTY_FIELDS);
  const [createFields, setCreateFields] = useState<ApplicationCatalogMutableFields>(EMPTY_FIELDS);
  const [showCreate, setShowCreate] = useState(false);
  const [showArchiveConfirm, setShowArchiveConfirm] = useState(false);
  const [showUnarchiveConfirm, setShowUnarchiveConfirm] = useState(false);
  const [pending, setPending] = useState<"" | "loading" | "creating" | "reading" | "updating" | "archiving" | "unarchiving">("");
  const [operation, setOperation] = useState<ApplicationCatalogOperationResult | null>(null);

  const lifecycleLabel = (state: ApplicationCatalogLifecycleState) => state === "active"
    ? t($ => $.catalog.activeLifecycleStatus) : t($ => $.catalog.archivedLifecycleStatus);
  const catalogSummary = snapshot.status === "failed"
    ? t($ => $.catalog.catalogLoadFailureNoFallback)
    : snapshot.status === "loading" ? t($ => $.catalog.loadingCatalogRecords)
      : t($ => $.catalog.loadedApplicationCounts, {
        activeCount: snapshot.records.filter((record) => record.lifecycleState === "active").length,
        archivedCount: snapshot.records.filter((record) => record.lifecycleState === "archived").length,
      });
  const operationSummary = operation?.failureCode
    ? operation.status === "version_conflict" ? t($ => $.catalog.versionConflictNotice)
      : t($ => $.catalog.catalogOperationFailed)
    : operation?.status === "created" ? t($ => $.catalog.authoritativeApplicationCreated)
      : operation?.status === "updated" ? t($ => $.catalog.metadataUpdatedWithCas)
        : operation?.status === "archived" ? t($ => $.catalog.archivedAndDisabledActions)
          : operation?.status === "unarchived" ? t($ => $.catalog.unarchivedAfterAccessReview)
            : operation?.status === "loaded" ? t($ => $.catalog.exactRecordLoaded)
              : t($ => $.catalog.catalogOperationPending);

  const visibleRecords = useMemo(
    () => snapshot.records.filter((record) => record.lifecycleState === filter),
    [filter, snapshot.records],
  );

  useEffect(() => {
    onSnapshotChange(snapshot);
  }, [onSnapshotChange, snapshot]);

  useEffect(() => {
    if (config.mode === "offline") return;
    let cancelled = false;
    setPending("loading");
    Promise.all([
      listApplicationCatalogRecords(config, "active"),
      listApplicationCatalogRecords(config, "archived"),
    ]).then(([active, archived]) => {
      if (cancelled) return;
      setPending("");
      if (active.status === "failed" || archived.status === "failed") {
        setSnapshot({
          status: "failed",
          records: [],
          failureCode: active.failureCode || archived.failureCode,
          summary: "Application catalog failed without falling back to offline fixtures.",
        });
        return;
      }
      const records = mergeRecords(active.records, archived.records);
      setSnapshot({
        status: "ready",
        records,
        failureCode: "",
        summary: `Loaded ${active.records.length} active and ${archived.records.length} archived applications.`,
      });
      setNextCursors({ active: active.nextCursor, archived: archived.nextCursor });
      const preferred = records.find((record) => record.applicationId === selectedApplicationId) ?? active.records[0] ?? archived.records[0] ?? null;
      if (preferred) {
        setFilter(preferred.lifecycleState);
        selectAndRead(preferred, false);
      }
    }).catch(() => {
      if (cancelled) return;
      setPending("");
      setSnapshot({ status: "failed", records: [], failureCode: "application_catalog_store_unavailable", summary: "Application catalog failed without fallback." });
    });
    return () => { cancelled = true; };
  }, [config]);

  const replaceRecord = (record: ApplicationCatalogRecord, summary: string) => {
    setSnapshot((current) => ({
      status: "ready",
      records: mergeRecords(
        current.records.filter((item) => item.applicationId !== record.applicationId),
        [record],
      ),
      failureCode: "",
      summary,
    }));
    setSelectedRecord(record);
    setFields(fieldsFromRecord(record));
    onSelectRecord(record);
  };

  const selectAndRead = (record: ApplicationCatalogRecord, notifyImmediately = true) => {
    setSelectedRecord(record);
    setFields(fieldsFromRecord(record));
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setShowUnarchiveConfirm(false);
    setOperation(null);
    if (notifyImmediately) onSelectRecord(record);
    if (config.mode === "offline") return;
    setPending("reading");
    readApplicationCatalogRecord(config, record.applicationId).then((result) => {
      setPending("");
      setOperation(result);
      if (result.record) replaceRecord(result.record, "Loaded an exact application catalog record.");
    });
  };

  const handleCreate = () => {
    const validationFailure = validateApplicationCatalogMutableFields(createFields);
    if (validationFailure) {
      setOperation(localFailure(validationFailure));
      return;
    }
    setPending("creating");
    createApplicationCatalogRecord(config, createFields).then((result) => {
      setPending("");
      setOperation(result);
      if (!result.record) return;
      replaceRecord(result.record, "Created an application in the authoritative catalog.");
      setCreateFields(EMPTY_FIELDS);
      setShowCreate(false);
      setFilter("active");
    });
  };

  const handleUpdate = () => {
    if (!selectedRecord) return;
    const validationFailure = validateApplicationCatalogMutableFields(fields);
    if (validationFailure) {
      setOperation(localFailure(validationFailure));
      return;
    }
    setPending("updating");
    updateApplicationCatalogRecord(config, selectedRecord.applicationId, selectedRecord.recordVersion, fields).then((result) => {
      setPending("");
      setOperation(result);
      if (result.record) replaceRecord(result.record, "Updated application metadata through record-version CAS.");
    });
  };

  const handleRefreshAfterConflict = () => {
    if (!selectedRecord) return;
    setPending("reading");
    readApplicationCatalogRecord(config, selectedRecord.applicationId).then((result) => {
      setPending("");
      setOperation(result);
      if (result.record) replaceRecord(result.record, "Refreshed the current stored version after explicit conflict review.");
    });
  };

  const handleArchive = () => {
    if (!selectedRecord) return;
    setPending("archiving");
    archiveApplicationCatalogRecord(config, selectedRecord.applicationId, selectedRecord.recordVersion).then((result) => {
      setPending("");
      setOperation(result);
      if (!result.record) return;
      replaceRecord(result.record, "Archived application and disabled new downstream actions.");
      setFilter("archived");
      setShowArchiveConfirm(false);
    });
  };

  const handleUnarchive = () => {
    if (!selectedRecord) return;
    setPending("unarchiving");
    unarchiveApplicationCatalogRecord(config, selectedRecord.applicationId, selectedRecord.recordVersion).then((result) => {
      setPending("");
      setOperation(result);
      if (!result.record) return;
      replaceRecord(result.record, "Unarchived application after explicit access reactivation review.");
      setFilter("active");
      setShowUnarchiveConfirm(false);
    });
  };

  const handleLoadMore = () => {
    const cursor = nextCursors[filter];
    if (!cursor) return;
    setPending("loading");
    listApplicationCatalogRecords(config, filter, cursor).then((result) => {
      setPending("");
      setOperation(null);
      if (result.status === "failed") {
        setSnapshot((current) => ({ ...current, status: "failed", failureCode: result.failureCode, summary: result.summary }));
        return;
      }
      setSnapshot((current) => ({
        status: "ready",
        records: mergeRecords(current.records, result.records),
        failureCode: "",
        summary: `Loaded another ${filter} application page.`,
      }));
      setNextCursors((cursors) => ({ ...cursors, [filter]: result.nextCursor }));
    });
  };

  if (config.mode === "offline") {
    return (
      <section className="application-catalog-panel offline" aria-label={t($ => $.catalog.catalogManagement)}>
        <div className="application-catalog-heading">
          <div>
            <p className="eyebrow">{t($ => $.catalog.catalogTitle)}</p>
            <h4>{t($ => $.catalog.readOnlyFixtureMode)}</h4>
          </div>
          <span className="status-badge neutral">{t($ => $.catalog.offlineStatus)}</span>
        </div>
        <p>{t($ => $.catalog.offlineVerificationNotice)}</p>
      </section>
    );
  }

  return (
    <section className="application-catalog-panel" aria-label={t($ => $.catalog.catalogManagement)}>
      <div className="application-catalog-heading">
        <div>
          <p className="eyebrow">{t($ => $.catalog.developmentCatalogTitle)}</p>
          <h4>{t($ => $.catalog.manageWorkspaceApplications)}</h4>
        </div>
        <span className={`status-badge ${snapshot.status === "ready" ? "good" : snapshot.status === "failed" ? "bad" : "neutral"}`}>
          {snapshot.status === "ready" ? t($ => $.catalog.readyStatus) : snapshot.status === "failed" ? t($ => $.catalog.failedStatus) : t($ => $.catalog.loadingStatus)}
        </span>
      </div>

      <div className="application-catalog-toolbar">
        <div className="application-catalog-filter" role="group" aria-label={t($ => $.catalog.lifecycleFilter)}>
          <button type="button" className={filter === "active" ? "selected" : ""} onClick={() => setFilter("active")}>{t($ => $.catalog.activeFilter)}</button>
          <button type="button" className={filter === "archived" ? "selected" : ""} onClick={() => setFilter("archived")}>{t($ => $.catalog.archivedFilter)}</button>
        </div>
        <button type="button" className="primary-action" onClick={() => { setShowCreate(true); setOperation(null); }} disabled={Boolean(pending)}>
          {t($ => $.catalog.createApplication)}</button>
      </div>

      <p className="application-catalog-summary" role="status">{pending ? t($ => $.catalog.catalogOperationPending) : catalogSummary}</p>
      {snapshot.failureCode ? <p className="application-catalog-failure">{snapshot.failureCode}</p> : null}

      {showCreate ? (
        <article className="application-catalog-editor create-editor">
          <div className="card-title-row">
            <div><p className="eyebrow">{t($ => $.catalog.newApplication)}</p><h5>{t($ => $.catalog.serverGeneratedIdentity)}</h5></div>
            <button type="button" onClick={() => setShowCreate(false)} disabled={Boolean(pending)}>{t($ => $.catalog.cancel)}</button>
          </div>
          <ApplicationFields fields={createFields} onChange={setCreateFields} prefix="catalog-create" disabled={Boolean(pending)} />
          <button type="button" className="primary-action" onClick={handleCreate} disabled={Boolean(pending)}>{t($ => $.catalog.createAndSelect)}</button>
        </article>
      ) : null}

      <div className="application-catalog-list" aria-label={t($ => $.catalog.applicationListLabel, { lifecycle: lifecycleLabel(filter) })}>
        {visibleRecords.length === 0 ? <p className="empty-state">{t($ => $.catalog.noApplications, { lifecycle: lifecycleLabel(filter) })}</p> : visibleRecords.map((record) => (
          <button
            type="button"
            className={`application-catalog-row${selectedRecord?.applicationId === record.applicationId ? " selected" : ""}`}
            key={record.applicationId}
            onClick={() => selectAndRead(record)}
          >
            <span><strong>{record.displayName}</strong><small>{record.applicationKind === "workflow_copilot" ? t($ => $.catalog.workflowCopilot) : record.applicationKind === "docs_qa" ? t($ => $.catalog.docsQa) : record.applicationKind === "agent" ? t($ => $.catalog.agent) : t($ => $.catalog.promptApplication)}</small></span>
            <span><code>{record.applicationId}</code><small>v{record.recordVersion}</small></span>
          </button>
        ))}
      </div>
      {nextCursors[filter] ? <button type="button" onClick={handleLoadMore} disabled={Boolean(pending)}>{t($ => $.catalog.loadMore)}</button> : null}

      {selectedRecord ? (
        <article className="application-catalog-editor detail-editor">
          <div className="card-title-row">
            <div>
              <p className="eyebrow">{t($ => $.catalog.selectedLifecycleVersion, { lifecycle: lifecycleLabel(selectedRecord.lifecycleState), version: selectedRecord.recordVersion })}</p>
              <h5>{selectedRecord.applicationId}</h5>
            </div>
            <span className={`status-badge ${selectedRecord.lifecycleState === "active" ? "good" : "neutral"}`}>{lifecycleLabel(selectedRecord.lifecycleState)}</span>
          </div>
          <dl className="application-catalog-meta">
            <div><dt>{t($ => $.catalog.owner)}</dt><dd>{selectedRecord.ownerSubjectRef}</dd></div>
            <div><dt>{t($ => $.catalog.updated)}</dt><dd>{selectedRecord.updatedAt}</dd></div>
            <div><dt>{t($ => $.catalog.audit)}</dt><dd>{selectedRecord.auditRef || t($ => $.catalog.listProjection)}</dd></div>
            <div><dt>{t($ => $.catalog.archivedFilter)}</dt><dd>{selectedRecord.archivedAt ?? t($ => $.catalog.notArchived)}</dd></div>
          </dl>
          <ApplicationFields fields={fields} onChange={setFields} prefix="catalog-edit" disabled={Boolean(pending) || selectedRecord.lifecycleState === "archived"} />
          {selectedRecord.lifecycleState === "active" ? (
            <div className="application-catalog-actions">
              <button type="button" className="primary-action" onClick={handleUpdate} disabled={Boolean(pending)}>{t($ => $.catalog.saveMetadata)}</button>
              <button type="button" className="danger-action" onClick={() => setShowArchiveConfirm(true)} disabled={Boolean(pending)}>{t($ => $.catalog.archiveApplication)}</button>
            </div>
          ) : (
            <div className="application-catalog-actions">
              <p className="application-catalog-archive-note">{t($ => $.catalog.archivedReadabilityNotice)}</p>
              <button type="button" className="primary-action" onClick={() => setShowUnarchiveConfirm(true)} disabled={Boolean(pending)}>{t($ => $.catalog.unarchiveApplication)}</button>
            </div>
          )}
          {showArchiveConfirm ? (
            <div className="application-catalog-confirm" role="alert">
              <strong>{t($ => $.catalog.archiveVersionQuestion, { version: selectedRecord.recordVersion })}</strong>
              <p>{t($ => $.catalog.archiveImpactNotice)}</p>
              <button type="button" className="danger-action" onClick={handleArchive} disabled={Boolean(pending)}>{t($ => $.catalog.confirmArchive)}</button>
              <button type="button" onClick={() => setShowArchiveConfirm(false)} disabled={Boolean(pending)}>{t($ => $.catalog.keepActive)}</button>
            </div>
          ) : null}
          {showUnarchiveConfirm && selectedRecord.lifecycleState === "archived" ? (
            <div className="application-catalog-confirm" role="alert">
              <strong>{t($ => $.catalog.reactivateVersionQuestion, { version: selectedRecord.recordVersion })}</strong>
              <p>{t($ => $.catalog.reactivationImpactNotice)}</p>
              <p>{t($ => $.catalog.historicalVersionChecksNotice)}</p>
              <button type="button" className="primary-action" onClick={handleUnarchive} disabled={Boolean(pending)}>{t($ => $.catalog.understandAndUnarchive)}</button>
              <button type="button" onClick={() => setShowUnarchiveConfirm(false)} disabled={Boolean(pending)}>{t($ => $.catalog.keepArchived)}</button>
            </div>
          ) : null}
        </article>
      ) : null}

      {operation ? (
        <div className={`application-catalog-operation ${operation.failureCode ? "failed" : "succeeded"}`} role="status">
          <strong>{operation.failureCode ? t($ => $.catalog.failedStatus) : t($ => $.catalog.operationStatus)}</strong>
          <p>{operationSummary}</p>
          {operation.failureCode ? <code>{operation.failureCode}</code> : null}
          {operation.status === "version_conflict" ? (
            <button type="button" onClick={handleRefreshAfterConflict} disabled={Boolean(pending)}>
              {t($ => $.catalog.refreshStoredVersion, { version: operation.currentRecordVersion })}
            </button>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

function ApplicationFields({
  fields,
  onChange,
  prefix,
  disabled,
}: {
  fields: ApplicationCatalogMutableFields;
  onChange: (fields: ApplicationCatalogMutableFields) => void;
  prefix: string;
  disabled: boolean;
}) {
  const { t } = useTranslation("applications");
  return (
    <div className="application-catalog-fields">
      <label htmlFor={`${prefix}-name`}>{t($ => $.catalog.displayName)}</label>
      <input id={`${prefix}-name`} value={fields.displayName} maxLength={120} disabled={disabled} onChange={(event) => onChange({ ...fields, displayName: event.target.value })} />
      <label htmlFor={`${prefix}-kind`}>{t($ => $.catalog.applicationKind)}</label>
      <select id={`${prefix}-kind`} value={fields.applicationKind} disabled={disabled} onChange={(event) => onChange({ ...fields, applicationKind: event.target.value as ApplicationCatalogKind })}>
        <option value="workflow_copilot">{t($ => $.catalog.workflowCopilot)}</option>
        <option value="docs_qa">{t($ => $.catalog.docsQa)}</option>
        <option value="agent">{t($ => $.catalog.agent)}</option>
        <option value="prompt_application">{t($ => $.catalog.promptApplication)}</option>
      </select>
      <label htmlFor={`${prefix}-description`}>{t($ => $.catalog.description)}</label>
      <textarea id={`${prefix}-description`} value={fields.description} maxLength={1000} disabled={disabled} onChange={(event) => onChange({ ...fields, description: event.target.value })} />
    </div>
  );
}

function initialSnapshot(mode: typeof defaultConfig.mode): ApplicationCatalogSnapshot {
  return mode === "offline"
    ? { status: "offline", records: [], failureCode: "", summary: "Offline fixture mode sends no application catalog requests." }
    : { status: "loading", records: [], failureCode: "", summary: "Loading active and archived application catalog records." };
}

function fieldsFromRecord(record: ApplicationCatalogRecord): ApplicationCatalogMutableFields {
  return { displayName: record.displayName, description: record.description, applicationKind: record.applicationKind };
}

function mergeRecords(...recordSets: ApplicationCatalogRecord[][]): ApplicationCatalogRecord[] {
  const byId = new Map<string, ApplicationCatalogRecord>();
  recordSets.flat().forEach((record) => {
    const current = byId.get(record.applicationId);
    if (!current || record.recordVersion >= current.recordVersion) byId.set(record.applicationId, record);
  });
  return [...byId.values()].sort((left, right) => right.updatedAt.localeCompare(left.updatedAt) || right.applicationId.localeCompare(left.applicationId));
}

function localFailure(failureCode: string): ApplicationCatalogOperationResult {
  return {
    status: "failed", record: null, failureCode, currentRecordVersion: 0, currentLifecycleState: "",
    requestId: "", auditRef: "", summary: "Application metadata was rejected before any request was sent.",
  };
}
