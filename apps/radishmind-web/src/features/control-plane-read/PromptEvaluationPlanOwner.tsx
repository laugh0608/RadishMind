import { useRef } from "react";
import { useTranslation } from "react-i18next";
import "../../i18n/promptEvaluationPlanResources.ts";
import "./promptEvaluationPlan.css";
import { APPLICATION_EVALUATION_PROFILES, type ApplicationEvaluationConfig, type ApplicationEvaluationPlan, type ApplicationEvaluationPlanVersion, type ApplicationEvaluationProfile } from "./applicationEvaluationCampaignConsumer.ts";
import { addPromptPlanSample, movePromptPlanSample, removePromptPlanSample, updatePromptPlanSample, type PromptPlanIssue } from "./promptEvaluationPlanDraft.ts";
import { usePromptEvaluationPlanEditor } from "./usePromptEvaluationPlanEditor.ts";
import PromptEvaluationVariables from "./PromptEvaluationVariables.tsx";
import { campaignState } from "./applicationEvaluationMessages.ts";

export default function PromptEvaluationPlanOwner(props: {
  config: ApplicationEvaluationConfig;
  selectedPlan: ApplicationEvaluationPlan | null;
  selectedVersion: ApplicationEvaluationPlanVersion | null;
  applicationActive: boolean;
  onSaved: (planId: string) => Promise<void>;
  onReload: () => Promise<void>;
  onBusyChange: (busy: boolean) => void;
  onDirtyChange: (dirty: boolean) => void;
  onProfileChange: (profile: ApplicationEvaluationProfile) => void;
  onLoadOtherProfile: () => void;
  onOpenCampaign: () => void;
}) {
  const { t } = useTranslation("evaluation");
  const owner = usePromptEvaluationPlanEditor(props);
  const { editor, reference, busy, dirty, confirmation, validation, conflict } = owner;
  const region = useRef<HTMLElement>(null);
  const samples = editor.content.mode === "form" ? editor.content.samples : [];
  const selected = samples.find(item => item.editorKey === editor.selectedKey) ?? samples[0];
  const sampleIndex = selected ? samples.indexOf(selected) : -1;
  const currentReference = reference.current;
  const itemCount = editor.content.mode === "form" ? samples.length : validation.draft?.items.length ?? 0;
  const savedSelection = editor.base && !dirty && props.selectedPlan?.planId === editor.base.planId &&
    props.selectedVersion?.planVersion === editor.base.planVersion && props.selectedVersion.planDigest === editor.base.planDigest;

  function loadSelected() {
    if (!props.selectedPlan || !props.selectedVersion) return;
    if (props.selectedVersion.executionProfile !== "prompt_application_invocation_v1") props.onLoadOtherProfile();
    else owner.replaceDraft({ kind: "load", plan: props.selectedPlan, version: props.selectedVersion });
  }

  function locateIssue(issue: PromptPlanIssue) {
    const sample = samples.find(item => item.itemKey === issue.itemKey);
    if (sample) owner.change({ ...editor, selectedKey: sample.editorKey });
    window.requestAnimationFrame(() => {
      const root = region.current;
      const field = [...(root?.querySelectorAll<HTMLElement>("[data-variable]") ?? [])].find(node => node.dataset.variable === issue.field);
      const input = issue.code === "planName" ? root?.querySelector<HTMLInputElement>('[name="prompt-plan-name"]')
        : editor.content.mode === "json" ? root?.querySelector<HTMLTextAreaElement>('[name="prompt-plan-json"]')
          : field?.querySelector<HTMLElement>("textarea, input, select") ?? root?.querySelector<HTMLElement>(`[name="${issue.field === "item_key" ? "sample-key" : "sample-name"}"]`);
      input?.focus();
    });
  }

  return <section ref={region} className="application-evaluation-plan-owner prompt-plan-owner" id="application-evaluation-plan" data-edit-mode={editor.content.mode}>
    <header className="prompt-plan-heading"><div><h4>{t($ => $.promptPlan.title)}</h4>
      <p>{editor.base ? t($ => $.promptPlan.editingVersion, { version: editor.base.planVersion, record: editor.base.recordVersion }) : t($ => $.promptPlan.newDraft)}</p>
      {editor.base ? <code>{editor.base.planId}</code> : null}</div>
      <span className="status-badge neutral">{dirty ? t($ => $.promptPlan.unsaved) : editor.base ? t($ => $.promptPlan.saved) : t($ => $.promptPlan.newDraft)}</span>
    </header>
    <div className="prompt-plan-selected-context">
      {props.selectedPlan ? <div><strong>{props.selectedPlan.name}</strong><span>{t($ => $.promptPlan.selectedVersion, { version: props.selectedPlan.latestPlanVersion, record: props.selectedPlan.recordVersion })}</span><code>{props.selectedPlan.planId}</code></div> : null}
      <div className="prompt-plan-inline-actions">
        <button type="button" className="secondary-action" disabled={busy} onClick={() => owner.replaceDraft({ kind: "new" })}>{t($ => $.promptPlan.newPlan)}</button>
        <button type="button" className="secondary-action" disabled={busy || !props.selectedVersion || !props.selectedPlan} onClick={loadSelected}>{t($ => $.promptPlan.loadSelected)}</button>
      </div>
    </div>
    <div className="prompt-plan-meta">
      <label><span>{t($ => $.promptPlan.planName)}</span><input name="prompt-plan-name" value={editor.name} disabled={busy} onChange={event => owner.change({ ...editor, name: event.target.value })} /></label>
      <label><span>{t($ => $.promptPlan.profile)}</span><select value="prompt_application_invocation_v1" disabled={busy} onChange={event => props.onProfileChange(event.target.value as ApplicationEvaluationProfile)}>
        {APPLICATION_EVALUATION_PROFILES.map(profile => <option key={profile} value={profile}>{profile}</option>)}
      </select></label>
    </div>
    <section className="prompt-plan-reference" aria-label={t($ => $.promptPlan.referenceLabel)} data-reference-state={reference.status}>
      <div><strong>{t($ => $.promptPlan.referenceLabel)}{currentReference ? ` · ${currentReference.source.templateName} · v${currentReference.source.templateVersion}` : ""}</strong>
        <p>{reference.status === "loading" ? t($ => $.promptPlan.readingReference) : reference.status === "unavailable" ? t($ => $.promptPlan.referenceUnavailable) : currentReference ? t($ => $.promptPlan.referenceBoundary) : t($ => $.promptPlan.referenceEmpty)}</p>
        {currentReference ? <code>{currentReference.source.templateId} · {currentReference.source.templateDigest}</code> : null}
        {reference.code ? <code>{reference.code}</code> : null}
      </div>
      <button type="button" className="secondary-action" disabled={busy} onClick={() => void owner.refreshReference()}>{currentReference ? t($ => $.promptPlan.refreshReference) : t($ => $.promptPlan.readReference)}</button>
      <small>{t($ => $.promptPlan.referencePermission)}</small>
    </section>
    {reference.status === "changed" && reference.candidate ? <section className="prompt-plan-warning" role="status">
      <strong>{t($ => $.promptPlan.referenceChanged)}</strong>
      <p>{t($ => $.promptPlan.previousReference)}: {currentReference?.source.templateName} v{currentReference?.source.templateVersion} → {t($ => $.promptPlan.nextReference)}: {reference.candidate.source.templateName} v{reference.candidate.source.templateVersion}</p>
      <p>{t($ => $.promptPlan.referenceChangeBoundary)}</p>
      <button type="button" onClick={owner.adoptReference}>{t($ => $.promptPlan.adoptReference)}</button>
    </section> : null}
    <div className="prompt-plan-modes" role="group" aria-label={t($ => $.promptPlan.modeLabel)}>
      <button type="button" className="secondary-action" aria-pressed={editor.content.mode === "form"} disabled={busy} onClick={() => owner.changeMode("form")}>{t($ => $.promptPlan.formMode)}</button>
      <button type="button" className="secondary-action" aria-pressed={editor.content.mode === "json"} disabled={busy} onClick={() => owner.changeMode("json")}>{t($ => $.promptPlan.jsonMode)}</button>
      <span className={validation.matched ? "prompt-plan-match" : ""}>{validation.matched ? t($ => $.promptPlan.matched) : t($ => $.promptPlan.unmatched)}</span>
    </div>
    {editor.content.mode === "json" ? <label className="prompt-plan-json"><span>{t($ => $.promptPlan.jsonLabel)}</span>
      <textarea name="prompt-plan-json" value={editor.content.text} spellCheck={false} rows={20} disabled={busy}
        onChange={event => { if (editor.content.mode === "json") owner.change({ ...editor, content: { ...editor.content, text: event.target.value } }); }} />
    </label> : <div className="prompt-plan-workbench">
      <aside className="prompt-plan-samples"><strong>{t($ => $.promptPlan.samples, { count: samples.length })}</strong>
        <button type="button" className="secondary-action" disabled={busy || samples.length >= 20} onClick={() => owner.change(addPromptPlanSample(editor))}>{t($ => $.promptPlan.addSample)}</button>
        <select className="prompt-plan-sample-select" aria-label={t($ => $.promptPlan.sampleSelector)} value={selected?.editorKey ?? ""} disabled={busy || !samples.length} onChange={event => owner.change({ ...editor, selectedKey: event.target.value })}>
          {!samples.length ? <option value="">{t($ => $.promptPlan.emptySamples)}</option> : null}
          {samples.map((sample, index) => <option key={sample.editorKey} value={sample.editorKey}>{index + 1}. {sample.name || t($ => $.promptPlan.unnamedSample)}</option>)}
        </select>
        <div className="prompt-plan-sample-list" role="listbox" aria-label={t($ => $.promptPlan.sampleSelector)}>{samples.map((sample, index) => <button key={sample.editorKey} type="button" role="option" aria-selected={sample.editorKey === selected?.editorKey} disabled={busy}
          onClick={() => owner.change({ ...editor, selectedKey: sample.editorKey })}><strong>{String(index + 1).padStart(2, "0")} {sample.name || t($ => $.promptPlan.unnamedSample)}</strong><code>{sample.itemKey}</code></button>)}</div>
        <p>{t($ => $.promptPlan.sampleBoundary)}</p>
      </aside>
      <div className="prompt-plan-sample-editor">{selected ? <>
        <div className="prompt-plan-sample-heading"><h5>{selected.name || t($ => $.promptPlan.unnamedSample)}</h5><div className="prompt-plan-inline-actions">
          <button type="button" className="secondary-action" disabled={busy || samples.length >= 20} onClick={() => owner.change(addPromptPlanSample(editor, selected.editorKey))}>{t($ => $.promptPlan.copySample)}</button>
          <button type="button" className="secondary-action" disabled={busy || sampleIndex === 0} onClick={() => owner.change(movePromptPlanSample(editor, selected.editorKey, -1))}>{t($ => $.promptPlan.moveUp)}</button>
          <button type="button" className="secondary-action" disabled={busy || sampleIndex === samples.length - 1} onClick={() => owner.change(movePromptPlanSample(editor, selected.editorKey, 1))}>{t($ => $.promptPlan.moveDown)}</button>
          <button type="button" className="secondary-action" disabled={busy} onClick={() => owner.change(removePromptPlanSample(editor, selected.editorKey))}>{t($ => $.promptPlan.deleteSample)}</button>
        </div></div>
        <details className="prompt-plan-sample-details" open><summary>{t($ => $.promptPlan.sampleDetails)}</summary><div className="prompt-plan-sample-meta">
          <label><span>{t($ => $.promptPlan.sampleName)}</span><input name="sample-name" value={selected.name} disabled={busy} onChange={event => owner.change(updatePromptPlanSample(editor, selected.editorKey, { name: event.target.value }))} /></label>
          <label><span>{t($ => $.promptPlan.sampleKey)}</span><input name="sample-key" value={selected.itemKey} disabled={busy} onChange={event => owner.change(updatePromptPlanSample(editor, selected.editorKey, { itemKey: event.target.value }))} /></label>
          <label><span>{t($ => $.promptPlan.expectedClassification)}</span><select value={selected.expectedClassification} disabled={busy} onChange={event => owner.change(updatePromptPlanSample(editor, selected.editorKey, { expectedClassification: event.target.value as typeof selected.expectedClassification }))}>
            {(["regression", "improvement", "changed", "unchanged", "inconclusive"] as const).map(value => <option key={value} value={value}>{campaignState(t, value)} · {value}</option>)}
          </select></label>
        </div></details>
        {currentReference ? <PromptEvaluationVariables sample={selected} declarations={currentReference.source.variables} disabled={busy} issues={owner.issues}
          onChange={variables => owner.change(updatePromptPlanSample(editor, selected.editorKey, { variables }))} /> : <p>{t($ => $.promptPlan.referenceEmpty)}</p>}
        <p className="prompt-plan-note">{t($ => $.promptPlan.classificationBoundary)}</p>
      </> : <p>{t($ => $.promptPlan.emptySamples)}</p>}</div>
    </div>}
    {owner.issues.length ? <section className="prompt-plan-errors" role="alert"><strong>{t($ => $.promptPlan.issuesTitle)}</strong><ul>{owner.issues.map((issue, index) => <li key={`${issue.code}-${index}`}>
      <button type="button" className="secondary-action" onClick={() => locateIssue(issue)}>{issue.itemKey ? <code>{issue.itemKey} {issue.field}</code> : null}
        {issue.line && issue.column ? <span>{t($ => $.promptPlan.issueLocation, { line: issue.line, column: issue.column })}</span> : null}<span>{t($ => $.promptPlan.issues[issue.code])}</span></button>
    </li>)}</ul></section> : null}
    {conflict ? <section className="prompt-plan-warning" role="status"><strong>{t($ => $.promptPlan.conflictTitle)}</strong><p>{t($ => $.promptPlan.conflictVersions, conflict)}</p><p>{t($ => $.promptPlan.conflictBoundary)}</p>
      <button type="button" className="secondary-action" disabled={busy} onClick={() => void props.onReload()}>{t($ => $.promptPlan.reloadRemote)}</button></section> : null}
    {owner.discard ? <div className="prompt-plan-warning" role="alertdialog" aria-label={t($ => $.promptPlan.discardTitle)} aria-modal="false"><strong>{t($ => $.promptPlan.discardTitle)}</strong><p>{t($ => $.promptPlan.discardBoundary)}</p>
      <div className="prompt-plan-inline-actions"><button type="button" className="secondary-action" onClick={owner.cancelDiscard}>{t($ => $.promptPlan.keepEditing)}</button><button type="button" onClick={owner.confirmDiscard}>{t($ => $.promptPlan.discardContinue)}</button></div></div> : null}
    {confirmation ? <div className="application-evaluation-confirmation prompt-plan-confirmation">
      <strong>{confirmation.kind === "archive" ? t($ => $.promptPlan.reviewArchiveTitle) : confirmation.review.base ? t($ => $.promptPlan.reviewReviseTitle) : t($ => $.promptPlan.reviewCreateTitle)}</strong>
      <span>{confirmation.kind === "archive" ? confirmation.plan.name : confirmation.review.draft.name}</span>
      <p>{t($ => $.promptPlan.reviewCAS, { version: confirmation.kind === "archive" ? confirmation.plan.recordVersion : confirmation.review.base?.recordVersion ?? 0 })}</p>
      {confirmation.kind === "save" ? <><p>{t($ => $.promptPlan.reviewCount, { count: confirmation.review.draft.items.length })}</p><p>{confirmation.review.matched ? t($ => $.promptPlan.matched) : t($ => $.promptPlan.unmatched)}</p><small>{t($ => $.promptPlan.matchingBoundary)}</small></> : null}
      <p>{confirmation.kind === "archive" ? t($ => $.promptPlan.archiveBoundary) : t($ => $.promptPlan.reviewBoundary)}</p>
      <div className="prompt-plan-inline-actions"><button type="button" className="secondary-action" disabled={busy} onClick={owner.cancelConfirmation}>{t($ => $.promptPlan.cancel)}</button>
        <button type="button" disabled={owner.writeBlocked} onClick={() => void owner.confirm()}>{confirmation.kind === "archive" ? t($ => $.promptPlan.confirmArchive) : confirmation.review.base ? t($ => $.promptPlan.confirmRevise) : t($ => $.promptPlan.confirmCreate)}</button></div>
    </div> : <footer className="prompt-plan-footer"><span>{editor.content.mode === "json" && !validation.draft ? t($ => $.promptPlan.unparsedSamples) : t($ => $.promptPlan.samples, { count: itemCount })}</span><div className="prompt-plan-inline-actions">
      <button type="button" className="secondary-action" disabled={owner.writeBlocked || dirty || !props.selectedPlan} onClick={() => { if (props.selectedPlan) owner.reviewArchive(props.selectedPlan); }}>{t($ => $.promptPlan.reviewArchive)}</button>
      <button type="button" disabled={owner.writeBlocked} onClick={owner.reviewSave}>{t($ => $.promptPlan.reviewSave)}</button>
    </div></footer>}
    {savedSelection ? <button type="button" className="secondary-action prompt-plan-execute-link" disabled={busy} onClick={props.onOpenCampaign}>{t($ => $.promptPlan.openCampaign)}</button> : null}
    {owner.notice ? <div className={`application-evaluation-operation ${owner.notice.kind === "failed" ? "is-failure" : ""}`} role="status">
      {owner.notice.kind === "failed" ? <><span>{t($ => $.promptPlan.failedNotice)}</span><code>{owner.notice.code}</code></>
        : owner.notice.kind === "archived" ? <span>{t($ => $.promptPlan.archivedNotice, { id: owner.notice.id })}</span>
          : <span>{t($ => $.promptPlan[owner.notice!.kind === "saved" ? "savedNotice" : "loadedNotice"], { id: owner.notice.id, version: owner.notice.version! })}</span>}
    </div> : null}
  </section>;
}
