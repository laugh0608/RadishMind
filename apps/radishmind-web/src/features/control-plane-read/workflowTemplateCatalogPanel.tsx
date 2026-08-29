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

const tasks: Array<{ id: Task; label: string; summary: string }> = [
  { id: "catalog", label: "Catalog", summary: "Create candidates and inspect workspace templates." },
  { id: "review", label: "Review", summary: "Resolve one pending candidate through review CAS." },
  { id: "listing", label: "Listing", summary: "Move the single workspace listing pointer through CAS." },
  { id: "derive", label: "Derive", summary: "Create one independent Saved Draft from a listed version." },
];

export default function WorkflowTemplateCatalogPanel({
  workspaceId,
  applicationId,
  applicationName,
  applicationActive,
  onDerivedDraft,
}: Props) {
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
  const [status, setStatus] = useState(
    config.mode === "offline"
      ? "Offline source: catalog HTTP is disabled and no request will be sent."
      : "Loading workspace template authority…",
  );
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
      setStatus("Offline source: catalog HTTP is disabled and no request was sent.");
      return;
    }
    const ticket = coordinatorRef.current.current();
    setBusy(true);
    setStatus("Loading workspace candidates and template pointers…");
    try {
      const [candidatePage, lineagePage] = await Promise.all([
        listWorkflowTemplateCandidates(config, { limit: 50, signal: ticket.signal }),
        listWorkflowTemplates(config, { limit: 50, signal: ticket.signal }),
      ]);
      if (!coordinatorRef.current.accepts(ticket)) return;
      setCandidates(candidatePage.records);
      setLineages(lineagePage.records);
      setStatus(candidatePage.failureCode || lineagePage.failureCode
        ? `Catalog load failed closed: ${candidatePage.failureCode ?? lineagePage.failureCode}`
        : `Loaded ${candidatePage.records.length} candidates and ${lineagePage.records.length} template pointers.`);
    } catch (error) {
      if (!coordinatorRef.current.accepts(ticket)) return;
      setStatus(error instanceof Error ? error.message : "Workflow template catalog load failed.");
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
      .catch((error: unknown) => {
        if (coordinatorRef.current.accepts(ticket)) setStatus(error instanceof Error ? error.message : "Template version load failed.");
      });
  }, [config, lineages, selectedTemplateId]);

  async function runOperation(label: string, operation: (signal: AbortSignal) => Promise<WorkflowTemplateOperationResult>) {
    if (config.mode === "offline" || busy) return;
    const ticket = coordinatorRef.current.current();
    setBusy(true);
    setStatus(`${label}…`);
    try {
      const result = await operation(ticket.signal);
      if (!coordinatorRef.current.accepts(ticket)) return;
      if (result.failureCode) {
        setStatus(`${label} failed closed: ${result.failureCode} (review v${result.currentReviewVersion}, pointer v${result.currentPointerVersion}).`);
        return;
      }
      if (result.draft && result.draftAuthority) {
        onDerivedDraft(result.draft, result.draftAuthority);
        setStatus(`Derived and opened saved draft ${result.draftAuthority.draftId} v${result.draftAuthority.draftVersion}.`);
      } else {
        setStatus(`${label} completed. Audit ${result.auditRef}.`);
      }
      await refresh();
    } catch (error) {
      if (coordinatorRef.current.accepts(ticket)) setStatus(error instanceof Error ? error.message : `${label} failed.`);
    } finally {
      if (coordinatorRef.current.accepts(ticket)) setBusy(false);
    }
  }

  return (
    <section className="workflow-template-catalog" id="workspace-workflow-template-catalog" aria-labelledby="workflow-template-catalog-title">
      <header className="workflow-template-catalog__heading">
        <div>
          <p className="eyebrow">Workflow Templates · dev / test</p>
          <h3 id="workflow-template-catalog-title">Workspace Template Catalog</h3>
          <p>Review immutable Definition references, move one listing pointer, then derive one independent Saved Draft.</p>
        </div>
        <div className="workflow-template-catalog__badges" aria-label="Template catalog boundaries">
          <span className={`status-badge ${config.mode === "offline" ? "neutral" : "ready"}`}>{config.mode === "offline" ? "offline" : "dev HTTP"}</span>
          <span className="status-badge neutral">workspace scoped</span>
          <button type="button" className="secondary-action" disabled={busy || config.mode === "offline"} onClick={() => void refresh()}>Refresh</button>
        </div>
      </header>

      <p className="workflow-template-catalog__status" role="status">{status}</p>
      {!applicationActive ? <p className="workflow-template-catalog__warning">The selected application is not active. Catalog inspection remains read-only.</p> : null}

      <div className="workflow-template-catalog__layout">
        <nav className="workflow-template-catalog__tasks" aria-label="Workflow template tasks">
          {tasks.map((item) => (
            <button key={item.id} type="button" aria-current={task === item.id ? "page" : undefined} onClick={() => setTask(item.id)}>
              <strong>{item.label}</strong><small>{item.summary}</small>
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
              onCreate={() => void runOperation("Create candidate", (signal) => createWorkflowTemplateCandidate(config, {
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
              onSubmit={() => selectedCandidate && void runOperation("Review candidate", (signal) => reviewWorkflowTemplateCandidate(config, selectedCandidate.candidateId, { expectedReviewVersion: selectedCandidate.reviewVersion, decision: reviewDecision, reason: reviewReason.trim() }, signal))} />
          ) : null}
          {task === "listing" ? (
            <ListingTask lineages={lineages} versions={versions} selected={selectedLineage} selectedId={selectedTemplateId} onSelect={setSelectedTemplateId}
              decision={listingDecision} onDecision={(value) => { setListingDecision(value); setUnlistConfirmed(false); }} version={listingVersion} onVersion={setListingVersion}
              reason={listingReason} onReason={setListingReason} confirmed={unlistConfirmed} onConfirmed={setUnlistConfirmed}
              disabled={busy || config.mode === "offline" || !selectedLineage || (listingDecision === "unlist" && !unlistConfirmed)}
              onSubmit={() => selectedLineage && void runOperation("Update listing", (signal) => decideWorkflowTemplateListing(config, selectedLineage.templateId, { expectedPointerVersion: selectedLineage.pointerVersion, decision: listingDecision, version: listingDecision === "unlist" ? 0 : Number(listingVersion), reason: listingReason.trim() }, signal))} />
          ) : null}
          {task === "derive" ? (
            <DeriveTask lineages={lineages.filter((lineage) => lineage.lifecycle === "listed")} selected={selectedLineage?.lifecycle === "listed" ? selectedLineage : null}
              selectedId={selectedTemplateId} onSelect={setSelectedTemplateId} targetApplicationId={targetApplicationId} onTargetApplicationId={setTargetApplicationId}
              draftId={draftId} onDraftId={setDraftId} name={draftName} onName={setDraftName} confirmed={deriveConfirmed} onConfirmed={setDeriveConfirmed}
              version={selectedVersion} disabled={busy || config.mode === "offline" || !selectedLineage || selectedLineage.lifecycle !== "listed" || !deriveConfirmed}
              onSubmit={() => selectedLineage?.lifecycle === "listed" && void runOperation("Derive saved draft", (signal) => deriveWorkflowTemplateDraft(config, selectedLineage.templateId, { expectedPointerVersion: selectedLineage.pointerVersion, templateVersion: selectedLineage.listedVersion, targetApplicationId: targetApplicationId.trim(), draftId: draftId.trim(), name: draftName.trim(), confirmed: deriveConfirmed }, signal))} />
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
  const field = (key: keyof CandidateForm) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onFormChange({ ...form, [key]: event.target.value });
  return <div className="workflow-template-task">
    <div><p className="eyebrow">Catalog source</p><h4>{applicationName}</h4><p><code>{applicationId}</code> contributes only an exact released Definition reference and digest.</p></div>
    <form onSubmit={(event) => { event.preventDefault(); onCreate(); }} className="workflow-template-form">
      <label>Candidate ID<input value={form.candidateId} onChange={field("candidateId")} placeholder="candidate_support_v1" required /></label>
      <label>Template ID<input value={form.templateId} onChange={field("templateId")} placeholder="template_support" required /></label>
      <label>Definition ID<input value={form.sourceDefinitionId} onChange={field("sourceDefinitionId")} placeholder="definition_support" required /></label>
      <label>Definition version<input type="number" min="1" value={form.sourceDefinitionVersion} onChange={field("sourceDefinitionVersion")} required /></label>
      <label>Title<input value={form.title} minLength={2} maxLength={120} onChange={field("title")} required /></label>
      <label className="span-2">Summary<textarea value={form.summary} minLength={4} maxLength={1000} onChange={field("summary")} required /></label>
      <label className="span-2">Usage notes<textarea value={form.usageNotes} maxLength={2000} onChange={field("usageNotes")} /></label>
      <label className="span-2">Labels, comma separated<input value={form.labels} onChange={field("labels")} placeholder="support,reviewed" /></label>
      <button type="submit" className="primary-action" disabled={disabled}>Create candidate</button>
    </form>
    <RecordLists candidates={candidates} lineages={lineages} selectedCandidateId={selectedCandidateId} selectedTemplateId={selectedTemplateId} onSelectCandidate={onSelectCandidate} onSelectTemplate={onSelectTemplate} />
  </div>;
}

function RecordLists({ candidates, lineages, selectedCandidateId, selectedTemplateId, onSelectCandidate, onSelectTemplate }: {
  candidates: WorkflowTemplateCandidate[]; lineages: WorkflowTemplateLineage[]; selectedCandidateId: string; selectedTemplateId: string;
  onSelectCandidate: (value: string) => void; onSelectTemplate: (value: string) => void;
}) {
  return <div className="workflow-template-records">
    <div><h4>Candidates</h4>{candidates.length === 0 ? <p>No candidates in this source.</p> : candidates.map((candidate) => <button type="button" key={candidate.candidateId} aria-pressed={selectedCandidateId === candidate.candidateId} onClick={() => onSelectCandidate(candidate.candidateId)}><strong>{candidate.title}</strong><small>{candidate.state} · review v{candidate.reviewVersion} · {candidate.candidateId}</small></button>)}</div>
    <div><h4>Templates</h4>{lineages.length === 0 ? <p>No template lineages in this source.</p> : lineages.map((lineage) => <button type="button" key={lineage.templateId} aria-pressed={selectedTemplateId === lineage.templateId} onClick={() => onSelectTemplate(lineage.templateId)}><strong>{lineage.templateId}</strong><small>{lineage.lifecycle} · pointer v{lineage.pointerVersion} · listed v{lineage.listedVersion}</small></button>)}</div>
  </div>;
}

function ReviewTask({ candidates, selected, selectedId, onSelect, decision, onDecision, reason, onReason, disabled, onSubmit }: {
  candidates: WorkflowTemplateCandidate[]; selected: WorkflowTemplateCandidate | null; selectedId: string; onSelect: (value: string) => void;
  decision: WorkflowTemplateDecisionOption; onDecision: (value: WorkflowTemplateDecisionOption) => void; reason: string; onReason: (value: string) => void; disabled: boolean; onSubmit: () => void;
}) {
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">Candidate review</p><h4>One review decision, one CAS</h4><p>The current candidate review version is sent exactly. A conflict never retries or falls back.</p></div>
    <label className="span-2">Pending candidate<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">Select candidate</option>{candidates.map((candidate) => <option key={candidate.candidateId} value={candidate.candidateId}>{candidate.title} · {candidate.state} · v{candidate.reviewVersion}</option>)}</select></label>
    <label>Decision<select value={decision} onChange={(event) => onDecision(event.target.value as WorkflowTemplateDecisionOption)}><option value="approve">Approve</option><option value="reject">Reject</option><option value="request_changes">Request changes</option><option value="withdraw">Withdraw</option></select></label>
    <label>Expected review version<input readOnly value={selected?.reviewVersion ?? ""} /></label>
    <label className="span-2">Reason<textarea value={reason} minLength={4} maxLength={500} onChange={(event) => onReason(event.target.value)} required /></label>
    <button className="primary-action" disabled={disabled}>Submit review</button>
  </form>;
}
type WorkflowTemplateDecisionOption = "approve" | "reject" | "request_changes" | "withdraw";

function ListingTask({ lineages, versions, selected, selectedId, onSelect, decision, onDecision, version, onVersion, reason, onReason, confirmed, onConfirmed, disabled, onSubmit }: {
  lineages: WorkflowTemplateLineage[]; versions: WorkflowTemplateVersion[]; selected: WorkflowTemplateLineage | null; selectedId: string; onSelect: (value: string) => void;
  decision: "list" | "replace" | "unlist"; onDecision: (value: "list" | "replace" | "unlist") => void; version: string; onVersion: (value: string) => void;
  reason: string; onReason: (value: string) => void; confirmed: boolean; onConfirmed: (value: boolean) => void; disabled: boolean; onSubmit: () => void;
}) {
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">Listing pointer</p><h4>Explicit list, replace, or unlist</h4><p>Only the listed pointer is derivable. Version records remain immutable.</p></div>
    <label className="span-2">Template<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">Select template</option>{lineages.map((lineage) => <option key={lineage.templateId} value={lineage.templateId}>{lineage.templateId} · {lineage.lifecycle} · pointer v{lineage.pointerVersion}</option>)}</select></label>
    <label>Decision<select value={decision} onChange={(event) => onDecision(event.target.value as "list" | "replace" | "unlist")}><option value="list">List</option><option value="replace">Replace</option><option value="unlist">Unlist</option></select></label>
    <label>Version<select value={version} disabled={decision === "unlist"} onChange={(event) => onVersion(event.target.value)}>{versions.length === 0 ? <option value={version}>v{version}</option> : versions.map((item) => <option key={item.version} value={item.version}>v{item.version} · {item.title}</option>)}</select></label>
    <label className="span-2">Reason<textarea value={reason} minLength={4} maxLength={500} onChange={(event) => onReason(event.target.value)} required /></label>
    {decision === "unlist" ? <label className="workflow-template-confirm span-2"><input type="checkbox" checked={confirmed} onChange={(event) => onConfirmed(event.target.checked)} />I understand this removes the derivable workspace pointer.</label> : null}
    <button className="primary-action" disabled={disabled}>Apply pointer decision</button>
  </form>;
}

function DeriveTask({ lineages, selected, selectedId, onSelect, targetApplicationId, onTargetApplicationId, draftId, onDraftId, name, onName, confirmed, onConfirmed, version, disabled, onSubmit }: {
  lineages: WorkflowTemplateLineage[]; selected: WorkflowTemplateLineage | null; selectedId: string; onSelect: (value: string) => void;
  targetApplicationId: string; onTargetApplicationId: (value: string) => void; draftId: string; onDraftId: (value: string) => void; name: string; onName: (value: string) => void;
  confirmed: boolean; onConfirmed: (value: boolean) => void; version: WorkflowTemplateVersion | null; disabled: boolean; onSubmit: () => void;
}) {
  return <form className="workflow-template-task workflow-template-form" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <div className="span-2"><p className="eyebrow">Controlled derivation</p><h4>One listed version → one Saved Draft</h4><p>The catalog never copies Definition graphs. The server resolves the exact digest and validates target bindings before the single draft write.</p></div>
    <label className="span-2">Listed template<select value={selectedId} onChange={(event) => onSelect(event.target.value)}><option value="">Select listed template</option>{lineages.map((lineage) => <option key={lineage.templateId} value={lineage.templateId}>{lineage.templateId} · v{lineage.listedVersion} · pointer v{lineage.pointerVersion}</option>)}</select></label>
    <label>Target application ID<input value={targetApplicationId} onChange={(event) => onTargetApplicationId(event.target.value)} required /></label>
    <label>Saved Draft ID<input value={draftId} onChange={(event) => onDraftId(event.target.value)} placeholder="draft_from_template" required /></label>
    <label className="span-2">Draft name<input value={name} minLength={2} maxLength={120} onChange={(event) => onName(event.target.value)} required /></label>
    <div className="workflow-template-authority span-2"><strong>Authority</strong><span>{selected ? `${selected.templateId} v${selected.listedVersion} · pointer v${selected.pointerVersion}` : "No listed template selected"}</span><code>{version?.templateDigest ?? selected?.listedDigest ?? "digest unavailable"}</code></div>
    <label className="workflow-template-confirm span-2"><input type="checkbox" checked={confirmed} onChange={(event) => onConfirmed(event.target.checked)} />I confirm the exact listed version and target application for this one-time draft creation.</label>
    <button className="primary-action" disabled={disabled}>Derive and open Saved Draft</button>
  </form>;
}
