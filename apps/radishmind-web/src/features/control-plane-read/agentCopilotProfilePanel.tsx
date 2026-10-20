import { profileFindingMessage, profileOperationMessage } from "./agentCopilotMessages.ts";
import "../../i18n/agentResources.ts";
import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useState } from "react";

import {
  bindApplicationConfigurationDraftAgentCopilotProfile,
  initialApplicationConfigurationDraftListState,
  listApplicationConfigurationDrafts,
  readApplicationConfigurationDraftConfig,
  type ApplicationConfigurationDraftListState,
} from "./applicationConfigurationDraftConsumer.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import {
  AGENT_COPILOT_TASKS,
  createAgentCopilotProfileInput,
  createAgentCopilotProfileVersion,
  listAgentCopilotProfiles,
  listAgentCopilotProfileVersions,
  readAgentCopilotProfile,
  readAgentCopilotProfileConfig,
  saveAgentCopilotProfile,
  validateAgentCopilotProfileLocally,
  validateAgentCopilotProfileRemote,
  type AgentCopilotProfileInput,
  type AgentCopilotProfileList,
  type AgentCopilotProfileOperation,
  type AgentCopilotProfileVersionList,
  type AgentCopilotProject,
} from "./agentCopilotProfileConsumer.ts";

const profileConfig = readAgentCopilotProfileConfig();
const draftConfig = readApplicationConfigurationDraftConfig();

export default function AgentCopilotProfilePanel({
  applicationId,
  applicationName,
  applicationKind,
  applicationActive,
  onOpenPublishReview,
  onEvidenceChange,
}: {
  applicationId: string;
  applicationName: string;
  applicationKind: string;
  applicationActive: boolean;
  onOpenPublishReview?: (draftId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
}) {
  const { t } = useTranslation("agent");
  const [input, setInput] = useState(() => createAgentCopilotProfileInput(profileConfig, applicationId));
  const [operation, setOperation] = useState<AgentCopilotProfileOperation>(() => initialOperation());
  const [drafts, setDrafts] = useState<AgentCopilotProfileList>(() => initialProfileList());
  const [versions, setVersions] = useState<AgentCopilotProfileVersionList>(() => initialVersionList());
  const [validationSource, setValidationSource] = useState<"local" | "remote">("local");
  const [selectedProfileVersion, setSelectedProfileVersion] = useState(0);
  const [applicationDrafts, setApplicationDrafts] = useState<ApplicationConfigurationDraftListState>(
    () => initialApplicationConfigurationDraftListState(draftConfig),
  );
  const [selectedDraftId, setSelectedDraftId] = useState("");
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(true);
  const [bindingStatus, setBindingStatus] = useState<{ failureCode: string } | { draftVersion: number; profileId: string; profileVersion: number } | null>(null);
  const [boundProfileRef, setBoundProfileRef] = useState<{ profileId: string; profileVersion: number } | null>(null);
  const localValidation = useMemo(() => validateAgentCopilotProfileLocally(input), [input]);
  const enabled = applicationActive && profileConfig.mode === "dev_agent_copilot_http";
  const selectedVersion = versions.summaries.find((version) => version.profileVersion === selectedProfileVersion);
  const selectedDraft = applicationDrafts.summaries.find((draft) => draft.draftId === selectedDraftId);
  const canBind = enabled && Boolean(selectedVersion) && Boolean(selectedDraft) &&
    selectedDraft?.applicationKind === "agent" && selectedDraft.validationState === "valid";

  useEffect(() => {
    setInput(createAgentCopilotProfileInput(profileConfig, applicationId));
    setOperation(initialOperation());
    setValidationSource("local");
    setDrafts(initialProfileList());
    setVersions(initialVersionList());
    setApplicationDrafts(initialApplicationConfigurationDraftListState(draftConfig));
    setSelectedDraftId("");
    setSelectedProfileVersion(0);
    setHasUnsavedChanges(true);
    setBindingStatus(null);
    setBoundProfileRef(null);
  }, [applicationId]);

  useEffect(() => {
    const ready = Boolean(boundProfileRef);
    onEvidenceChange?.({
      contributionId: "agent_profile",
      status: ready ? "available" : operation.failureCode ? "blocked" : "incomplete",
      coverage: ready ? "complete" : operation.failureCode ? "partial" : "none",
      evidenceRefs: boundProfileRef
        ? [{ kind: "profile", id: boundProfileRef.profileId, version: boundProfileRef.profileVersion }]
        : [],
      missingEvidence: ready ? [] : ["Create and bind an immutable Agent Copilot Profile version."],
      blockers: operation.failureCode ? [{ code: operation.failureCode, summary: operation.summary }] : [],
      failureCodes: operation.failureCode ? [operation.failureCode] : [],
    });
  }, [boundProfileRef, onEvidenceChange, operation]);

  function edit(patch: Partial<AgentCopilotProfileInput>) {
    setInput((current) => ({ ...current, ...patch }));
    setHasUnsavedChanges(true);
    setOperation((current) => ({
      ...current,
      status: profileConfig.mode === "offline" ? "offline" : "idle",
      failureCode: "",
      summary: "Profile 包含未保存的内存编辑。",
    }));
  }

  function selectProject(project: AgentCopilotProject) {
    edit({
      project,
      allowedTasks: [AGENT_COPILOT_TASKS[project][0] ?? ""],
      contextPolicy: {
        ...input.contextPolicy,
        allowedFields: project === "radishflow"
          ? ["selected_unit_ids", "diagnostics"]
          : ["question", "document_refs"],
      },
    });
  }

  function toggleTask(task: string) {
    const allowedTasks = input.allowedTasks.includes(task)
      ? input.allowedTasks.filter((value) => value !== task)
      : [...input.allowedTasks, task];
    edit({ allowedTasks });
  }

  async function validateRemote() {
    if (!localValidation.isValid) {
      setValidationSource("local");
      setOperation({
        ...initialOperation(),
        status: "invalid",
        validation: localValidation,
        failureCode: localValidation.findings[0]?.code ?? "agent_copilot_profile_payload_invalid",
        summary: "请先解决本地确定性校验阻塞项。",
      });
      return;
    }
    setOperation(await validateAgentCopilotProfileRemote(profileConfig, input));
    setValidationSource("remote");
  }

  async function save() {
    if (!enabled || !localValidation.isValid) return;
    const result = await saveAgentCopilotProfile(profileConfig, input, operation.currentDraftVersion);
    setOperation(result);
    setValidationSource("remote");
    if (result.draft) {
      setInput(draftToInput(result.draft));
      setHasUnsavedChanges(false);
      await refreshDrafts();
    }
  }

  async function createVersion() {
    if (!enabled || hasUnsavedChanges || !operation.draft) return;
    const result = await createAgentCopilotProfileVersion(
      profileConfig,
      applicationId,
      operation.draft.profileId,
      operation.currentDraftVersion,
    );
    setOperation(result);
    setValidationSource("remote");
    if (result.version) {
      setSelectedProfileVersion(result.version.profileVersion);
      await refreshVersions(result.version.profileId);
    }
  }

  async function refreshDrafts() {
    const result = await listAgentCopilotProfiles(profileConfig, applicationId);
    setDrafts(result);
  }

  async function restore(profileId: string) {
    const result = await readAgentCopilotProfile(profileConfig, applicationId, profileId);
    setOperation(result);
    setValidationSource("remote");
    if (result.draft) {
      setInput(draftToInput(result.draft));
      setHasUnsavedChanges(false);
      await refreshVersions(profileId);
    }
  }

  async function refreshVersions(profileId = input.profileId) {
    const result = await listAgentCopilotProfileVersions(profileConfig, applicationId, profileId);
    setVersions(result);
    setSelectedProfileVersion(result.summaries[0]?.profileVersion ?? 0);
  }

  async function loadApplicationDrafts() {
    const result = await listApplicationConfigurationDrafts(draftConfig, applicationId);
    setApplicationDrafts(result);
    setSelectedDraftId(result.summaries.find((draft) =>
      draft.applicationKind === "agent" && draft.validationState === "valid"
    )?.draftId ?? "");
  }

  async function bindVersion() {
    if (!canBind || !selectedDraft || !selectedVersion) return;
    const result = await bindApplicationConfigurationDraftAgentCopilotProfile(
      draftConfig,
      applicationId,
      selectedDraft.draftId,
      selectedDraft.draftVersion,
      selectedVersion.profileId,
      selectedVersion.profileVersion,
    );
    if (!result.draft || result.state.status !== "saved") {
      setBindingStatus({ failureCode: result.state.failureCode || "agent_copilot_profile_binding_ineligible" });
      return;
    }
    setBindingStatus({ draftVersion: result.state.currentDraftVersion, profileId: result.draft.agentCopilotProfileRef?.profileId ?? selectedVersion.profileId, profileVersion: result.draft.agentCopilotProfileRef?.profileVersion ?? selectedVersion.profileVersion });
    setBoundProfileRef({
      profileId: result.draft.agentCopilotProfileRef?.profileId ?? selectedVersion.profileId,
      profileVersion: result.draft.agentCopilotProfileRef?.profileVersion ?? selectedVersion.profileVersion,
    });
    await loadApplicationDrafts();
    onOpenPublishReview?.(result.draft.draftId);
  }

  if (applicationKind !== "agent") {
    return null;
  }

  return (
    <section className="agent-copilot-profile-panel" id="agent-copilot-profile-workspace" aria-labelledby="agent-profile-title">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.profile.eyebrow)}</p>
          <h4 id="agent-profile-title">{t($ => $.profile.title)}</h4>
        </div>
        <span className={`status-badge ${operation.status === "saved" || operation.status === "versioned" ? "good" : operation.failureCode ? "bad" : "neutral"}`}>
          {t($ => $.states[operation.status])}
        </span>
      </div>

      <div className="prompt-template-scope">
        <article><span>{t($ => $.profile.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.profile.owner)}</span><strong>{profileConfig.mode}</strong><code>{input.profileId}</code></article>
        <article><span>{t($ => $.profile.safety)}</span><strong>{t($ => $.profile.advisory)}</strong><p>{t($ => $.profile.safetyNote)}</p></article>
      </div>

      <div className="prompt-template-layout">
        <article className="prompt-template-editor">
          <label>{t($ => $.profile.id)}<input value={input.profileId} onChange={(event) => edit({ profileId: event.target.value })} disabled={operation.currentDraftVersion > 0} /></label>
          <label>{t($ => $.profile.name)}<input value={input.profileName} maxLength={80} onChange={(event) => edit({ profileName: event.target.value })} /></label>
          <label>{t($ => $.profile.description)}<textarea value={input.description} maxLength={512} rows={3} onChange={(event) => edit({ description: event.target.value })} /></label>
          <label>{t($ => $.profile.project)}<select value={input.project} onChange={(event) => selectProject(event.target.value as AgentCopilotProject)}><option value="radishflow">radishflow</option><option value="radish">radish</option></select></label>
          <fieldset>
            <legend>{t($ => $.profile.tasks)}</legend>
            {AGENT_COPILOT_TASKS[input.project].map((task) => (
              <label key={task}><input type="checkbox" checked={input.allowedTasks.includes(task)} onChange={() => toggleTask(task)} />{task}</label>
            ))}
          </fieldset>
          <p className="boundary-note" id="agent-profile-locale-note">{t($ => $.profile.localeNote)}</p>
          <label>{t($ => $.profile.defaultLocale)}<input aria-describedby="agent-profile-locale-note" value={input.defaultLocale} onChange={(event) => edit({ defaultLocale: event.target.value })} /></label>
          <label>{t($ => $.profile.allowedLocales)}<input value={input.allowedLocales.join(", ")} onChange={(event) => edit({ allowedLocales: csv(event.target.value) })} /></label>
          <label>{t($ => $.profile.contextFields)}<textarea rows={3} value={input.contextPolicy.allowedFields.join(", ")} onChange={(event) => edit({ contextPolicy: { ...input.contextPolicy, allowedFields: csv(event.target.value) } })} /></label>
          <label>{t($ => $.profile.contextBudget)}<input type="number" min={1} max={131072} value={input.contextPolicy.maxBytes} onChange={(event) => edit({ contextPolicy: { ...input.contextPolicy, maxBytes: Number(event.target.value) } })} /></label>
          <div className="application-draft-actions">
            <button type="button" onClick={() => void validateRemote()} disabled={!applicationActive}>{t($ => $.profile.validate)}</button>
            <button type="button" onClick={() => void save()} disabled={!enabled || !localValidation.isValid}>{t($ => $.profile.save)}</button>
            <button type="button" onClick={() => void createVersion()} disabled={!enabled || hasUnsavedChanges || !operation.draft}>{t($ => $.profile.version)}</button>
          </div>
        </article>

        <article className="prompt-template-review">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.profile.review)}</p><h5>{profileOperationMessage(t, operation, hasUnsavedChanges)}</h5></div><span className={`status-badge ${localValidation.isValid ? "good" : "bad"}`}>{t($ => $.states[localValidation.state])}</span></div>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          {(operation.validation.findings.length ? operation.validation.findings : localValidation.findings).map((finding) => (
            <p className="failure-summary" key={`${finding.code}-${finding.field}`}><strong>{finding.field}</strong> · {finding.code} · {profileFindingMessage(t, finding)}{validationSource === "remote" && operation.validation.findings.length ? <small>{finding.summary}</small> : null}</p>
          ))}
          <dl className="tenant-meta">
            <div><dt>{t($ => $.profile.safetyMode)}</dt><dd>{t($ => $.profile.advisory)}</dd></div>
            <div><dt>{t($ => $.profile.confirmation)}</dt><dd>{t($ => $.profile.required)}</dd></div>
            <div><dt>{t($ => $.profile.retrieval)}</dt><dd>{t($ => $.profile.disabled)}</dd></div>
            <div><dt>{t($ => $.profile.tools)}</dt><dd>{t($ => $.profile.disabled)}</dd></div>
            <div><dt>{t($ => $.profile.images)}</dt><dd>{t($ => $.profile.disabled)}</dd></div>
          </dl>
          <p className="boundary-note">{t($ => $.profile.boundary)}</p>
        </article>
      </div>

      <div className="prompt-template-lower-grid">
        <article>
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.profile.drafts)}</p><h5>{t($ => $.list[drafts.status], { count: drafts.summaries.length })}</h5></div><button type="button" onClick={() => void refreshDrafts()}>{t($ => $.profile.refresh)}</button></div>
          {drafts.failureCode ? <p className="failure-summary">{drafts.failureCode}</p> : null}
          {drafts.summaries.map((draft) => <button type="button" className="prompt-template-summary" key={draft.profileId} onClick={() => void restore(draft.profileId)}><strong>{draft.profileName}</strong><span>{draft.profileId} · v{draft.draftVersion}</span><small>{draft.project} · {draft.allowedTasks.join(", ")}</small></button>)}
        </article>
        <article>
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.profile.versions)}</p><h5>{t($ => $.list[versions.status], { count: versions.summaries.length })}</h5></div><button type="button" onClick={() => void refreshVersions()}>{t($ => $.profile.refresh)}</button></div>
          {versions.failureCode ? <p className="failure-summary">{versions.failureCode}</p> : null}
          {versions.summaries.map((version) => <button type="button" className={selectedProfileVersion === version.profileVersion ? "prompt-template-summary selected" : "prompt-template-summary"} key={version.profileVersion} onClick={() => setSelectedProfileVersion(version.profileVersion)}><strong>{t($ => $.profile.versionLabel, { version: version.profileVersion })}</strong><span>{version.profileId} · {t($ => $.profile.sourceVersion, { version: version.sourceDraftVersion })}</span><small>{version.profileDigest}</small></button>)}
        </article>
      </div>

      <article className="prompt-template-binding">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.profile.binding)}</p><h5>{t($ => $.profile.bindingTitle)}</h5></div><button type="button" onClick={() => void loadApplicationDrafts()} disabled={!enabled}>{t($ => $.profile.loadDrafts)}</button></div>
        <label>{t($ => $.profile.validDraft)}<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)}><option value="">{t($ => $.profile.noDraft)}</option>{applicationDrafts.summaries.map((draft) => <option key={draft.draftId} value={draft.draftId} disabled={draft.applicationKind !== "agent" || draft.validationState !== "valid"}>{draft.draftId} · v{draft.draftVersion}{draft.agentCopilotProfileRef ? ` · profile v${draft.agentCopilotProfileRef.profileVersion}` : ""}</option>)}</select></label>
        <label>{t($ => $.profile.immutableVersion)}<select value={selectedProfileVersion} onChange={(event) => setSelectedProfileVersion(Number(event.target.value))}><option value={0}>{t($ => $.profile.noVersion)}</option>{versions.summaries.map((version) => <option key={version.profileVersion} value={version.profileVersion}>{version.profileId} · v{version.profileVersion}</option>)}</select></label>
        <button type="button" onClick={() => void bindVersion()} disabled={!canBind}>{t($ => $.profile.bind)}</button>
        <p className="boundary-note">{!bindingStatus ? t($ => $.profile.bindingNote) : "failureCode" in bindingStatus ? <>{t($ => $.profile.bindingFailed)} · {bindingStatus.failureCode}</> : t($ => $.profile.bound, bindingStatus)}</p>
      </article>
    </section>
  );
}

function initialOperation(): AgentCopilotProfileOperation {
  return {
    status: profileConfig.mode === "offline" ? "offline" : "idle",
    draft: null,
    version: null,
    validation: { state: "invalid", isValid: false, findings: [] },
    currentDraftVersion: 0,
    currentProfileVersion: 0,
    failureCode: "",
    summary: profileConfig.mode === "offline" ? "Agent Copilot Web 未启用。" : "编辑 Profile 并执行本地与服务端校验。",
  };
}

function initialProfileList(): AgentCopilotProfileList {
  return { status: profileConfig.mode === "offline" ? "offline" : "empty", summaries: [], failureCode: "", summary: "加载当前应用的 Profile 草案。" };
}

function initialVersionList(): AgentCopilotProfileVersionList {
  return { status: profileConfig.mode === "offline" ? "offline" : "empty", summaries: [], failureCode: "", summary: "选择或保存 Profile 后加载不可变版本。" };
}

function draftToInput(draft: NonNullable<AgentCopilotProfileOperation["draft"]>): AgentCopilotProfileInput {
  const {
    draftVersion: _draftVersion,
    profileDigest: _profileDigest,
    policyDigest: _policyDigest,
    validation: _validation,
    updatedAt: _updatedAt,
    updatedByActorRef: _updatedByActorRef,
    ...input
  } = draft;
  return input;
}

function csv(value: string): string[] {
  return [...new Set(value.split(",").map((item) => item.trim()).filter(Boolean))];
}
