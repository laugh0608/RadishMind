import "../../i18n/workflowDraftResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import type { TFunction } from "i18next";
import { workflowTemplateMessage, type WorkflowTemplateMessage, type WorkflowTemplateAction } from "./workflowOperationMessages.ts";
import "../../i18n/workflowTemplateResources.ts";
import { useTranslation } from "react-i18next";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  WorkflowTemplateRequestCoordinator,
  createWorkflowTemplateCandidate,
  decideWorkflowTemplateListing,
  deriveWorkflowTemplateDraft,
  listWorkflowTemplateCandidates,
  listWorkflowTemplates,
  listWorkflowTemplateVersions,
  readWorkflowTemplate,
  readWorkflowTemplateCatalogConfig,
  reviewWorkflowTemplateCandidate,
  type WorkflowTemplateCandidate,
  type WorkflowTemplateLineage,
  type WorkflowTemplateOperationResult,
  type WorkflowTemplateVersion,
} from "./workflowTemplateCatalogConsumer.ts";

type Props = {
  workspaceId: string;
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
  onDerivedDraft: (
    draft: WorkflowDraftDesignerDraft,
    authority: NonNullable<WorkflowTemplateOperationResult["draftAuthority"]>,
  ) => void;
};

type Task = "catalog" | "review" | "listing" | "derive";

const tasks: Task[] = ["catalog", "review", "listing", "derive"];

export default function WorkflowTemplateCatalogPanel({
  workspaceId,
  applicationId,
  applicationName,
  applicationActive,
  onDerivedDraft,
}: Props) {
  const { t } = useTranslation("workflow");
  const baseConfig = useMemo(readWorkflowTemplateCatalogConfig, []);
  const config = useMemo(() => ({ ...baseConfig, workspaceId }), [baseConfig, workspaceId]);
  const scopeKey = `${workspaceId}:${applicationId}:${config.subjectRef}`;
  const coordinatorRef = useRef(new WorkflowTemplateRequestCoordinator());
  const [task, setTask] = useState<Task>("catalog");
  const [candidates, setCandidates] = useState<WorkflowTemplateCandidate[]>([]);
  const [lineages, setLineages] = useState<WorkflowTemplateLineage[]>([]);
  const [versions, setVersions] = useState<WorkflowTemplateVersion[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = useState("");
  const [selectedTemplateId, setSelectedTemplateId] = useState("");
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState<WorkflowTemplateMessage>({ code: config.mode === "offline" ? "offline" : "loading" });
  const [candidateForm, setCandidateForm] = useState({
    candidateId: "", templateId: "", sourceDefinitionId: "", sourceDefinitionVersion: "1",
    title: "", summary: "", usageNotes: "", labels: "",
  });
  const [reviewDecision, setReviewDecision] = useState<"approve" | "reject" | "request_changes" | "withdraw">("approve");
  const [reviewReason, setReviewReason] = useState("");
  const [listingDecision, setListingDecision] = useState<"list" | "replace" | "unlist">("list");
  const [listingVersion, setListingVersion] = useState("1");
  const [listingReason, setListingReason] = useState("");
  const [unlistConfirmed, setUnlistConfirmed] = useState(false);
  const [targetApplicationId, setTargetApplicationId] = useState(applicationId);
  const [draftId, setDraftId] = useState("");
  const [draftName, setDraftName] = useState("");
  const [deriveConfirmed, setDeriveConfirmed] = useState(false);

  const selectedCandidate = candidates.find((candidate) => candidate.candidateId === selectedCandidateId) ?? null;
  const selectedLineage = lineages.find((lineage) => lineage.templateId === selectedTemplateId) ?? null;
  const selectedVersion = versions.find((version) => version.version === Number(listingVersion)) ?? null;

  const refresh = useCallback(async () => {
    if (config.mode === "offline") {
      setStatus({ code: "offline" });
      return;
    }
    const ticket = coordinatorRef.current.current();
    setBusy(true);
    setStatus({ code: "loading" });
    try {
      const [candidatePage, lineagePage] = await Promise.all([
        listWorkflowTemplateCandidates(config, { limit: 50, signal: ticket.signal }),
        listWorkflowTemplates(config, { limit: 50, signal: ticket.signal }),
      ]);
      if (!coordinatorRef.current.accepts(ticket)) return;
      const catalogTemplateIds = new Set(lineagePage.records.map((lineage) => lineage.templateId));
      const approvedUnlistedIds = [...new Set(candidatePage.records
        .filter((candidate) => candidate.state === "approved" && !catalogTemplateIds.has(candidate.templateId))
        .map((candidate) => candidate.templateId))];
      const unlistedResults = await Promise.all(approvedUnlistedIds.map((templateId) =>
        readWorkflowTemplate(config, templateId, ticket.signal)));
      if (!coordinatorRef.current.accepts(ticket)) return;
      const failedUnlisted = unlistedResults.find((result) => result.failureCode || !result.lineage);
      if (failedUnlisted) {
        setCandidates([]);
        setLineages([]);
        setStatus({ code: "loadFailed", failureCode: failedUnlisted.failureCode ?? "workflow_template_store_unavailable" });
        return;
      }
      const reviewLineages = unlistedResults.map((result) => result.lineage as WorkflowTemplateLineage);
      const allLineages = [...lineagePage.records, ...reviewLineages];
      setCandidates(candidatePage.records);
      setLineages(allLineages);
      setStatus(candidatePage.failureCode || lineagePage.failureCode
        ? { code: "loadFailed", failureCode: candidatePage.failureCode || lineagePage.failureCode || "workflow_template_store_unavailable" }
        : { code: "ready", candidates: candidatePage.records.length, templates: allLineages.length });
    } catch {
      if (!coordinatorRef.current.accepts(ticket)) return;
      setStatus({ code: "loadFailed", failureCode: "workflow_template_store_unavailable" });
    } finally {
      if (coordinatorRef.current.accepts(ticket)) setBusy(false);
    }
  }, [config]);

  useEffect(() => {
    coordinatorRef.current.reset(scopeKey);
    setTask("catalog");
    setCandidates([]);
    setLineages([]);
    setVersions([]);
    setSelectedCandidateId("");
    setSelectedTemplateId("");
    setTargetApplicationId(applicationId);
    setDraftId("");
    setDraftName("");
    setDeriveConfirmed(false);
    setUnlistConfirmed(false);
    void refresh();
    return () => coordinatorRef.current.abort();
  }, [applicationId, refresh, scopeKey]);

  useEffect(() => {
    if (config.mode === "offline" || !selectedTemplateId) {
      setVersions([]);
      return;
    }
    const ticket = coordinatorRef.current.current();
    listWorkflowTemplateVersions(config, selectedTemplateId, { limit: 50, signal: ticket.signal })
      .then((page) => {
        if (!coordinatorRef.current.accepts(ticket)) return;
        setVersions(page.records);
        const listed = lineages.find((lineage) => lineage.templateId === selectedTemplateId)?.listedVersion ?? 0;
        setListingVersion(String(listed || page.records.at(-1)?.version || 1));
      })
      .catch(() => {
        if (coordinatorRef.current.accepts(ticket)) setStatus({ code: "versionFailed" });
      });
  }, [config, lineages, selectedTemplateId]);

  async function runOperation(action: WorkflowTemplateAction, operation: (signal: AbortSignal) => Promise<WorkflowTemplateOperationResult>) {
    if (config.mode === "offline" || busy) return;
    const ticket = coordinatorRef.current.current();
    setBusy(true);
    setStatus({ code: "working", action });
    try {
      const result = await operation(ticket.signal);
      if (!coordinatorRef.current.accepts(ticket)) return;
      if (result.failureCode) {
        setStatus({ code: "failed", action, failureCode: result.failureCode, review: result.currentReviewVersion, pointer: result.currentPointerVersion });
        return;
      }
      if (result.draft && result.draftAuthority) {
        onDerivedDraft(result.draft, result.draftAuthority);
        setStatus({ code: "derived", id: result.draftAuthority.draftId, version: result.draftAuthority.draftVersion });
      } else {
        setStatus({ code: "completed", action, audit: result.auditRef });
      }
      await refresh();
    } catch {
      if (coordinatorRef.current.accepts(ticket)) setStatus({ code: "failed", action, failureCode: "workflow_template_operation_failed", review: 0, pointer: 0 });
    } finally {
      if (coordinatorRef.current.accepts(ticket)) setBusy(false);
    }
  }

  return (
    <section className="workflow-template-catalog" id="workspace-workflow-template-catalog" aria-labelledby="workflow-template-catalog-title">
      <header className="workflow-template-catalog__heading">
        <div>
          <p className="eyebrow">{t($ => $.template.workflowTemplatesDevTest)}</p>
          <h3 id="workflow-template-catalog-title">{t($ => $.template.workspaceTemplateCatalog)}</h3>
          <p>{t($ => $.template.reviewImmutableDefinitionReferencesMoveOneListingPointerThenDerive)}</p>
        </div>
        <div className="workflow-template-catalog__badges" aria-label={t($ => $.template.templateCatalogBoundaries)}>
          <span className={`status-badge ${config.mode === "offline" ? "neutral" : "ready"}`}>{workflowDraftStatusLabel(t, config.mode === "offline" ? "offline" : "dev_http")}</span>
          <span className="status-badge neutral">{t($ => $.template.workspaceScoped)}</span>
          <button type="button" className="secondary-action" disabled={busy || config.mode === "offline"} onClick={() => void refresh()}>{t($ => $.template.refresh)}</button>
        </div>
      </header>

      <p className="workflow-template-catalog__status" role="status">{workflowTemplateMessage(t, status)}</p>
      {!applicationActive ? <p className="workflow-template-catalog__warning">{t($ => $.template.theSelectedApplicationIsNotActiveCatalogInspectionRemainsRead)}</p> : null}

      <div className="workflow-template-catalog__layout">
        <nav className="workflow-template-catalog__tasks" aria-label={t($ => $.template.workflowTemplateTasks)}>
          {tasks.map((item) => (
            <button key={item} type="button" aria-current={task === item ? "page" : undefined} onClick={() => setTask(item)}>
              <strong>{templateTaskLabel(t, item)}</strong><small>{templateTaskSummary(t, item)}</small>
            </button>
          ))}
        </nav>

        <div className="workflow-template-catalog__workspace">
          {task === "catalog" ? (
            <CatalogTask
              applicationId={applicationId}
              applicationName={applicationName}
              disabled={busy || config.mode === "offline" || !applicationActive}
              form={candidateForm}
              onFormChange={setCandidateForm}
              candidates={candidates}
              lineages={lineages}
              selectedCandidateId={selectedCandidateId}
              selectedTemplateId={selectedTemplateId}
              onSelectCandidate={setSelectedCandidateId}
              onSelectTemplate={setSelectedTemplateId}
              onCreate={() => void runOperation("create", (signal) => createWorkflowTemplateCandidate(config, {
                candidateId: candidateForm.candidateId.trim(), templateId: candidateForm.templateId.trim(), sourceApplicationId: applicationId,
                sourceDefinitionId: candidateForm.sourceDefinitionId.trim(), sourceDefinitionVersion: Number(candidateForm.sourceDefinitionVersion),
                title: candidateForm.title.trim(), summary: candidateForm.summary.trim(), usageNotes: candidateForm.usageNotes.trim(),
                labels: candidateForm.labels.split(",").map((label) => label.trim().toLowerCase()).filter(Boolean),
              }, signal))}
            />
          ) : null}
          {task === "review" ? (
            <ReviewTask candidates={candidates} selected={selectedCandidate} selectedId={selectedCandidateId} onSelect={setSelectedCandidateId}
              decision={reviewDecision} onDecision={setReviewDecision} reason={reviewReason} onReason={setReviewReason}
              disabled={busy || config.mode === "offline" || !selectedCandidate || selectedCandidate.state !== "pending"}
              onSubmit={() => selectedCandidate && void runOperation("review", (signal) => reviewWorkflowTemplateCandidate(config, selectedCandidate.candidateId, { expectedReviewVersion: selectedCandidate.reviewVersion, decision: reviewDecision, reason: reviewReason.trim() }, signal))} />
          ) : null}
          {task === "listing" ? (
            <ListingTask lineages={lineages} versions={versions} selected={selectedLineage} selectedId={selectedTemplateId} onSelect={setSelectedTemplateId}
              decision={listingDecision} onDecision={(value) => { setListingDecision(value); setUnlistConfirmed(false); }} version={listingVersion} onVersion={setListingVersion}
              reason={listingReason} onReason={setListingReason} confirmed={unlistConfirmed} onConfirmed={setUnlistConfirmed}
              disabled={busy || config.mode === "offline" || !selectedLineage || (listingDecision === "unlist" && !unlistConfirmed)}
              onSubmit={() => selectedLineage && void runOperation("listing", (signal) => decideWorkflowTemplateListing(config, selectedLineage.templateId, { expectedPointerVersion: selectedLineage.pointerVersion, decision: listingDecision, version: listingDecision === "unlist" ? 0 : Number(listingVersion), reason: listingReason.trim() }, signal))} />
          ) : null}
          {task === "derive" ? (
            <DeriveTask lineages={lineages.filter((lineage) => lineage.lifecycle === "listed")} selected={selectedLineage?.lifecycle === "listed" ? selectedLineage : null}
              selectedId={selectedTemplateId} onSelect={setSelectedTemplateId} targetApplicationId={targetApplicationId} onTargetApplicationId={setTargetApplicationId}
              draftId={draftId} onDraftId={setDraftId} name={draftName} onName={setDraftName} confirmed={deriveConfirmed} onConfirmed={setDeriveConfirmed}
              version={selectedVersion} disabled={busy || config.mode === "offline" || !selectedLineage || selectedLineage.lifecycle !== "listed" || !deriveConfirmed}
              onSubmit={() => selectedLineage?.lifecycle === "listed" && void runOperation("derive", (signal) => deriveWorkflowTemplateDraft(config, selectedLineage.templateId, { expectedPointerVersion: selectedLineage.pointerVersion, templateVersion: selectedLineage.listedVersion, targetApplicationId: targetApplicationId.trim(), draftId: draftId.trim(), name: draftName.trim(), confirmed: deriveConfirmed }, signal))} />
          ) : null}
        </div>
      </div>
    </section>
  );
}

type CandidateForm = { candidateId: string; templateId: string; sourceDefinitionId: string; sourceDefinitionVersion: string; title: string; summary: string; usageNotes: string; labels: string };

function CatalogTask({ applicationId, applicationName, disabled, form, onFormChange, candidates, lineages, selectedCandidateId, selectedTemplateId, onSelectCandidate, onSelectTemplate, onCreate }: {
  applicationId: string; applicationName: string; disabled: boolean; form: CandidateForm; onFormChange: (value: CandidateForm) => void;
  candidates: WorkflowTemplateCandidate[]; lineages: WorkflowTemplateLineage[]; selectedCandidateId: string; selectedTemplateId: string;
  onSelectCandidate: (value: string) => void; onSelectTemplate: (value: string) => void; onCreate: () => void;
}) {
  const { t } = useTranslation("workflow");
  const field = (key: keyof CandidateForm) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onFormChange({ ...form, [key]: event.target.value });
  return <div className="workflow-template-task">
    <div><p className="eyebrow">{t($ => $.template.catalogSource)}</p><h4>{applicationName}</h4><p><code>{applicationId}</code>{" "}{t($ => $.template.contributesOnlyAnExactReleasedDefinitionReferenceAndDigest)}</p></div>
    <form onSubmit={(event) => { event.preventDefault(); onCreate(); }} className="workflow-template-form">
      <label>{t($ => $.template.candidateID)}<input value={form.candidateId} onChange={field("candidateId")} placeholder="candidate_support_v1" required /></label>
      <label>{t($ => $.template.templateID)}<input value={form.templateId} onChange={field("templateId")} placeholder="template_support" required /></label>
      <label>{t($ => $.template.definitionID)}<input value={form.sourceDefinitionId} onChange={field("sourceDefinitionId")} placeholder="definition_support" required /></label>
      <label>{t($ => $.template.definitionVersion)}<input type="number" min="1" value={form.sourceDefinitionVersion} onChange={field("sourceDefinitionVersion")} required /></label>
      <label>{t($ => $.template.title)}<input value={form.title} minLength={2} maxLength={120} onChange={field("title")} required /></label>
      <label className="span-2">{t($ => $.template.summary)}<textarea value={form.summary} minLength={4} maxLength={1000} onChange={field("summary")} required /></label>
      <label className="span-2">{t($ => $.template.usageNotes)}<textarea value={form.usageNotes} maxLength={2000} onChange={field("usageNotes")} /></label>
      <label className="span-2">{t($ => $.template.labelsCommaSeparated)}<input value={form.labels} onChange={field("labels")} placeholder="support,reviewed" /></label>
      <button type="submit" className="primary-action" disabled={disabled}>{t($ => $.template.createCandidate)}</button>
    </form>
    <RecordLists candidates={candidates} lineages={lineages} selectedCandidateId={selectedCandidateId} selectedTemplateId={selectedTemplateId} onSelectCandidate={onSelectCandidate} onSelectTemplate={onSelectTemplate} />
  </div>;
}

function RecordLists({ candidates, lineages, selectedCandidateId, selectedTemplateId, onSelectCandidate, onSelectTemplate }: {
  candidates: WorkflowTemplateCandidate[]; lineages: WorkflowTemplateLineage[]; selectedCandidateId: string; selectedTemplateId: string;
  onSelectCandidate: (value: string) => void; onSelectTemplate: (value: string) => void;
}) {
  const { t } = useTranslation("workflow");
  return <div className="workflow-template-records">
    <div><h4>{t($ => $.template.candidates)}</h4>{candidates.length === 0 ? <p>{t($ => $.template.noCandidatesInThisSource)}</p> : candidates.map((candidate) => <button type="button" key={candidate.candidateId} aria-pressed={selectedCandidateId === candidate.candidateId} onClick={() => onSelectCandidate(candidate.candidateId)}><strong>{candidate.title}</strong><small>{workflowDraftStatusLabel(t, candidate.state)}{t($ => $.template.reviewV)}{candidate.reviewVersion} · {candidate.candidateId}</small></button>)}</div>
    <div><h4>{t($ => $.template.templates)}</h4>{lineages.length === 0 ? <p>{t($ => $.template.noTemplateLineagesInThisSource)}</p> : lineages.map((lineage) => <button type="button" key={lineage.templateId} aria-pressed={selectedTemplateId === lineage.templateId} onClick={() => onSelectTemplate(lineage.templateId)}><strong>{lineage.templateId}</strong><small>{workflowDraftStatusLabel(t, lineage.lifecycle)}{t($ => $.template.pointerV)}{lineage.pointerVersion}{t($ => $.template.listedV)}{lineage.listedVersion}</small></button>)}</div>
  </div>;
}

function ReviewTask({ candidates, selected, selectedId, onSelect, decision, onDecision, reason, onReason, disabled, onSubmit }: {
  candidates: WorkflowTemplateCandidate[]; selected: WorkflowTemplateCandidate | null; selectedId: string; onSelect: (value: string) => void;
  decision: WorkflowTemplateDecisionOption; onDecision: (value: WorkflowTemplateDecisionOption) => void; reason: string; onReason: (value: string) => void; disabled: boolean; onSubmit: () => void;
}) {
  const { t } = useTranslation("workflow");
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">{t($ => $.template.candidateReview)}</p><h4>{t($ => $.template.oneReviewDecisionOneCAS)}</h4><p>{t($ => $.template.theCurrentCandidateReviewVersionIsSentExactlyAConflict)}</p></div>
    <label className="span-2">{t($ => $.template.pendingCandidate)}<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">{t($ => $.template.selectCandidate)}</option>{candidates.map((candidate) => <option key={candidate.candidateId} value={candidate.candidateId}>{candidate.title} · {workflowDraftStatusLabel(t, candidate.state)} · v{candidate.reviewVersion}</option>)}</select></label>
    <label>{t($ => $.template.decision)}<select value={decision} onChange={(event) => onDecision(event.target.value as WorkflowTemplateDecisionOption)}><option value="approve">{t($ => $.template.approve)}</option><option value="reject">{t($ => $.template.reject)}</option><option value="request_changes">{t($ => $.template.requestChanges)}</option><option value="withdraw">{t($ => $.template.withdraw)}</option></select></label>
    <label>{t($ => $.template.expectedReviewVersion)}<input readOnly value={selected?.reviewVersion ?? ""} /></label>
    <label className="span-2">{t($ => $.template.reason)}<textarea value={reason} minLength={4} maxLength={500} onChange={(event) => onReason(event.target.value)} required /></label>
    <button className="primary-action" disabled={disabled}>{t($ => $.template.submitReview)}</button>
  </form>;
}
type WorkflowTemplateDecisionOption = "approve" | "reject" | "request_changes" | "withdraw";

function ListingTask({ lineages, versions, selected, selectedId, onSelect, decision, onDecision, version, onVersion, reason, onReason, confirmed, onConfirmed, disabled, onSubmit }: {
  lineages: WorkflowTemplateLineage[]; versions: WorkflowTemplateVersion[]; selected: WorkflowTemplateLineage | null; selectedId: string; onSelect: (value: string) => void;
  decision: "list" | "replace" | "unlist"; onDecision: (value: "list" | "replace" | "unlist") => void; version: string; onVersion: (value: string) => void;
  reason: string; onReason: (value: string) => void; confirmed: boolean; onConfirmed: (value: boolean) => void; disabled: boolean; onSubmit: () => void;
}) {
  const { t } = useTranslation("workflow");
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">{t($ => $.template.listingPointer)}</p><h4>{t($ => $.template.explicitListReplaceOrUnlist)}</h4><p>{t($ => $.template.onlyTheListedPointerIsDerivableVersionRecordsRemainImmutable)}</p></div>
    <label className="span-2">{t($ => $.template.template)}<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">{t($ => $.template.selectTemplate)}</option>{lineages.map((lineage) => <option key={lineage.templateId} value={lineage.templateId}>{lineage.templateId} · {workflowDraftStatusLabel(t, lineage.lifecycle)}{t($ => $.template.pointerV)}{lineage.pointerVersion}</option>)}</select></label>
    <label>{t($ => $.template.decision)}<select value={decision} onChange={(event) => onDecision(event.target.value as "list" | "replace" | "unlist")}><option value="list">{t($ => $.template.list)}</option><option value="replace">{t($ => $.template.replace)}</option><option value="unlist">{t($ => $.template.unlist)}</option></select></label>
    <label>{t($ => $.template.version)}<select value={version} disabled={decision === "unlist"} onChange={(event) => onVersion(event.target.value)}>{versions.length === 0 ? <option value={version}>v{version}</option> : versions.map((item) => <option key={item.version} value={item.version}>v{item.version} · {item.title}</option>)}</select></label>
    <label className="span-2">{t($ => $.template.reason)}<textarea value={reason} minLength={4} maxLength={500} onChange={(event) => onReason(event.target.value)} required /></label>
    {decision === "unlist" ? <label className="workflow-template-confirm span-2"><input type="checkbox" checked={confirmed} onChange={(event) => onConfirmed(event.target.checked)} />{t($ => $.template.iUnderstandThisRemovesTheDerivableWorkspacePointer)}</label> : null}
    <button className="primary-action" disabled={disabled}>{t($ => $.template.applyPointerDecision)}</button>
  </form>;
}

function DeriveTask({ lineages, selected, selectedId, onSelect, targetApplicationId, onTargetApplicationId, draftId, onDraftId, name, onName, confirmed, onConfirmed, version, disabled, onSubmit }: {
  lineages: WorkflowTemplateLineage[]; selected: WorkflowTemplateLineage | null; selectedId: string; onSelect: (value: string) => void;
  targetApplicationId: string; onTargetApplicationId: (value: string) => void; draftId: string; onDraftId: (value: string) => void; name: string; onName: (value: string) => void;
  confirmed: boolean; onConfirmed: (value: boolean) => void; version: WorkflowTemplateVersion | null; disabled: boolean; onSubmit: () => void;
}) {
  const { t } = useTranslation("workflow");
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">{t($ => $.template.controlledDerivation)}</p><h4>{t($ => $.template.oneListedVersionOneSavedDraft)}</h4><p>{t($ => $.template.theCatalogNeverCopiesDefinitionGraphsTheServerResolvesThe)}</p></div>
    <label className="span-2">{t($ => $.template.listedTemplate)}<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">{t($ => $.template.selectListedTemplate)}</option>{lineages.map((lineage) => <option key={lineage.templateId} value={lineage.templateId}>{lineage.templateId} · v{lineage.listedVersion}{t($ => $.template.pointerV)}{lineage.pointerVersion}</option>)}</select></label>
    <label>{t($ => $.template.targetApplicationID)}<input value={targetApplicationId} onChange={(event) => onTargetApplicationId(event.target.value)} required /></label>
    <label>{t($ => $.template.savedDraftID)}<input value={draftId} onChange={(event) => onDraftId(event.target.value)} placeholder="draft_from_template" required /></label>
    <label className="span-2">{t($ => $.template.draftName)}<input value={name} minLength={2} maxLength={120} onChange={(event) => onName(event.target.value)} required /></label>
    <div className="workflow-template-authority span-2"><strong>{t($ => $.template.authority)}</strong><span>{selected ? t($ => $.template.exactPointer, { id: selected.templateId, version: selected.listedVersion, pointer: selected.pointerVersion }) : t($ => $.template.noTemplate)}</span><code>{version?.templateDigest ?? selected?.listedDigest ?? t($ => $.template.noDigest)}</code></div>
    <label className="workflow-template-confirm span-2"><input type="checkbox" checked={confirmed} onChange={(event) => onConfirmed(event.target.checked)} />{t($ => $.template.iConfirmTheExactListedVersionAndTargetApplicationFor)}</label>
    <button className="primary-action" disabled={disabled}>{t($ => $.template.deriveAndOpenSavedDraft)}</button>
  </form>;
}

function templateTaskLabel(t: TFunction<"workflow">, task: Task): string {
  switch (task) {
    case "catalog": return t($ => $.template.catalogTask);
    case "review": return t($ => $.template.reviewTask);
    case "listing": return t($ => $.template.listingTask);
    case "derive": return t($ => $.template.deriveTask);
  }
}
function templateTaskSummary(t: TFunction<"workflow">, task: Task): string {
  switch (task) {
    case "catalog": return t($ => $.template.catalogTaskSummary);
    case "review": return t($ => $.template.reviewTaskSummary);
    case "listing": return t($ => $.template.listingTaskSummary);
    case "derive": return t($ => $.template.deriveTaskSummary);
  }
}
