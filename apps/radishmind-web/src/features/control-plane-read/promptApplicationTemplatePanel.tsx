import { useTranslation } from "react-i18next";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  bindApplicationConfigurationDraftPromptTemplate,
  initialApplicationConfigurationDraftListState,
  listApplicationConfigurationDrafts,
  readApplicationConfigurationDraftConfig,
  type ApplicationConfigurationDraftListState,
} from "./applicationConfigurationDraftConsumer.ts";
import {
  createPromptTemplateDraftInput,
  createPromptTemplateVersion,
  listPromptTemplateDrafts,
  listPromptTemplateVersions,
  mergePromptTemplateValidationOperation,
  readPromptTemplateConfig,
  readPromptTemplateDraft,
  readPromptTemplateVersion,
  renderPromptTemplatePreview,
  savePromptTemplateDraft,
  validatePromptTemplateLocally,
  validatePromptTemplateRemote,
  type PromptTemplateDraftInput,
  type PromptTemplateListResult,
  type PromptTemplateOperation,
  type PromptTemplateOutputKind,
  type PromptTemplatePreview,
  type PromptTemplateRole,
  type PromptTemplateVariableType,
  type PromptTemplateVersionListResult,
} from "./promptApplicationTemplateConsumer.ts";
import { formatPromptOutputSchema, parsePromptOutputSchema } from "./promptTemplateOutputSchema.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const templateConfig = readPromptTemplateConfig();
const draftConfig = readApplicationConfigurationDraftConfig();

type Props = {
  applicationId: string;
  applicationName: string;
  applicationKind: string;
  applicationActive: boolean;
  onOpenPublishReview?: (draftId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
};

export default function PromptApplicationTemplatePanel({
  applicationId,
  applicationName,
  applicationKind,
  applicationActive,
  onOpenPublishReview,
  onEvidenceChange,
}: Props) {
  const { t } = useTranslation("prompt");
  const [input, setInput] = useState(() => createPromptTemplateDraftInput(templateConfig, applicationId));
  const [schemaSource, setSchemaSource] = useState(() => formatPromptOutputSchema());
  const parsedSchema = useMemo(() => parsePromptOutputSchema(schemaSource), [schemaSource]);
  const editRevision = useRef(0);
  const [saving, setSaving] = useState(false);
  const [operation, setOperation] = useState<PromptTemplateOperation>(() => initialOperation());
  const [drafts, setDrafts] = useState<PromptTemplateListResult>(() => initialDraftList());
  const [versions, setVersions] = useState<PromptTemplateVersionListResult>(() => initialVersionList());
  const [draftsLoading, setDraftsLoading] = useState(false);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [previewInput, setPreviewInput] = useState('{"question":"如何审查本次发布？","tone":"清晰"}');
  const [preview, setPreview] = useState<PromptTemplatePreview | null>(null);
  const [applicationDrafts, setApplicationDrafts] = useState<ApplicationConfigurationDraftListState>(
    () => initialApplicationConfigurationDraftListState(draftConfig),
  );
  const [selectedDraftId, setSelectedDraftId] = useState("");
  const [selectedTemplateVersion, setSelectedTemplateVersion] = useState(0);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(true);
  const [bindingStatus, setBindingStatus] = useState<"" | "pending" | "failed" | "saved">("");
  const [bindingFailure, setBindingFailure] = useState("");

  useEffect(() => {
    setInput(createPromptTemplateDraftInput(templateConfig, applicationId));
    setSchemaSource(formatPromptOutputSchema());
    setOperation(initialOperation());
    setDrafts(initialDraftList());
    setVersions(initialVersionList());
    setPreviewInput('{"question":"如何审查本次发布？","tone":"清晰"}');
    setPreview(null);
    setApplicationDrafts(initialApplicationConfigurationDraftListState(draftConfig));
    setSelectedDraftId("");
    setSelectedTemplateVersion(0);
    setHasUnsavedChanges(true);
    setBindingStatus("");
    setBindingFailure("");
  }, [applicationId]);

  const localValidation = useMemo(() => validatePromptTemplateLocally(input), [input]);
  const enabled = templateConfig.mode === "dev_prompt_application_http" &&
    draftConfig.mode === "dev_application_draft_http" &&
    applicationActive && applicationKind === "prompt_application";
  const selectedDraft = applicationDrafts.summaries.find((draft) => draft.draftId === selectedDraftId) ?? null;
  const selectedVersion = versions.summaries.find((version) => version.templateVersion === selectedTemplateVersion) ?? null;
  const canBind = Boolean(
    enabled && selectedDraft && selectedDraft.validationState === "valid" &&
    selectedDraft.applicationKind === "prompt_application" && selectedVersion,
  );

  useEffect(() => {
    if (!onEvidenceChange) return;
    const failed = operation.status === "failed" || operation.status === "scope_denied" ||
      operation.status === "invalid" || operation.status === "version_conflict";
    onEvidenceChange({
      contributionId: "prompt_template",
      status: selectedVersion ? "available" : failed ? "blocked" : "incomplete",
      coverage: selectedVersion || failed ? "complete" : "none",
      evidenceRefs: selectedVersion
        ? [{ kind: "template", id: selectedVersion.templateId, version: selectedVersion.templateVersion }]
        : [],
      missingEvidence: selectedVersion ? [] : ["Create and review an immutable Prompt Template version."],
      blockers: failed
        ? [{ code: operation.failureCode || "prompt_template_blocked", summary: operation.summary }]
        : [],
      failureCodes: failed && operation.failureCode ? [operation.failureCode] : [],
    });
  }, [onEvidenceChange, operation, selectedVersion]);

  function edit(patch: Partial<PromptTemplateDraftInput>) {
    editRevision.current += 1;
    setInput((current) => ({ ...current, ...patch }));
    setHasUnsavedChanges(true);
    setOperation((current) => ({ ...current, status: templateConfig.mode === "offline" ? "offline" : "idle", failureCode: "", summary: "模板包含未保存的内存编辑。" }));
    setPreview(null);
  }

  async function validateRemote() {
    if (!localValidation.isValid) {
      setOperation((current) => ({
        ...current,
        status: "invalid",
        validation: localValidation,
        failureCode: localValidation.findings[0]?.code ?? "prompt_template_payload_invalid",
        summary: "请先解决本地确定性校验阻塞项。",
      }));
      return;
    }
    setOperation((current) => ({ ...current, status: "idle", failureCode: "", summary: "正在执行服务端确定性校验。" }));
    const result = await validatePromptTemplateRemote(templateConfig, input);
    setOperation((current) => mergePromptTemplateValidationOperation(current, result));
  }

  async function saveDraft() {
    if (saving || !enabled || !localValidation.isValid) return;
    const revision = editRevision.current;
    setSaving(true);
    const result = await savePromptTemplateDraft(templateConfig, input, operation.currentDraftVersion);
    setOperation(result);
    setSaving(false);
    if (result.draft) {
      // Preserve newer source edits while still advancing the persisted CAS lineage.
      if (revision === editRevision.current) {
        setInput(draftToInput(result.draft));
        if (result.draft.outputContract.jsonSchema) setSchemaSource(formatPromptOutputSchema(result.draft.outputContract.jsonSchema));
        setHasUnsavedChanges(false);
      }
      await refreshDrafts();
    }
  }

  async function refreshDrafts() {
    if (templateConfig.mode !== "dev_prompt_application_http") return;
    setDraftsLoading(true);
    try { setDrafts(await listPromptTemplateDrafts(templateConfig, applicationId)); }
    finally { setDraftsLoading(false); }
  }

  async function restoreDraft(templateId: string) {
    const result = await readPromptTemplateDraft(templateConfig, applicationId, templateId);
    setOperation(result);
    if (result.draft) {
      setInput(draftToInput(result.draft));
      setSchemaSource(formatPromptOutputSchema(result.draft.outputContract.jsonSchema));
      setHasUnsavedChanges(false);
      await refreshVersions(templateId);
    }
  }

  async function createVersion() {
    if (!enabled || !localValidation.isValid || hasUnsavedChanges || !operation.draft || operation.currentDraftVersion < 1) return;
    const result = await createPromptTemplateVersion(
      templateConfig,
      applicationId,
      operation.draft.templateId,
      operation.currentDraftVersion,
    );
    setOperation(result);
    if (result.version) {
      setSelectedTemplateVersion(result.version.templateVersion);
      await refreshVersions(result.version.templateId);
    }
  }

  async function refreshVersions(templateId = input.templateId) {
    if (templateConfig.mode !== "dev_prompt_application_http" || !templateId) return;
    setVersionsLoading(true);
    try {
      const result = await listPromptTemplateVersions(templateConfig, applicationId, templateId);
      setVersions(result);
      setSelectedTemplateVersion(result.summaries[0]?.templateVersion ?? 0);
    } finally { setVersionsLoading(false); }
  }

  async function openVersion(version: number) {
    const result = await readPromptTemplateVersion(templateConfig, applicationId, input.templateId, version);
    setOperation(result);
    if (result.version) setSelectedTemplateVersion(result.version.templateVersion);
  }

  function renderPreview() {
    try {
      const values: unknown = JSON.parse(previewInput);
      if (!values || typeof values !== "object" || Array.isArray(values)) throw new Error("invalid object");
      setPreview(renderPromptTemplatePreview(input, values as Record<string, unknown>));
    } catch {
      setPreview({
        status: "invalid",
        messages: [],
        findings: [{ code: "prompt_template_variable_invalid", field: "input", messageKey: "inputObject", summary: "合成变量必须是 JSON object。" }],
      });
    }
  }

  async function loadApplicationDrafts() {
    if (!enabled) return;
    setApplicationDrafts((current) => ({ ...current, status: "loading", summaries: [], failureCode: "", summary: "正在加载可绑定配置草案。" }));
    const result = await listApplicationConfigurationDrafts(draftConfig, applicationId);
    setApplicationDrafts(result);
    const first = result.summaries.find((draft) =>
      draft.applicationKind === "prompt_application" && draft.validationState === "valid"
    );
    setSelectedDraftId(first?.draftId ?? "");
  }

  async function bindVersion() {
    if (!canBind || !selectedDraft || !selectedVersion) return;
    setBindingFailure("");
    setBindingStatus("pending");
    const result = await bindApplicationConfigurationDraftPromptTemplate(
      draftConfig,
      applicationId,
      selectedDraft.draftId,
      selectedDraft.draftVersion,
      selectedVersion.templateId,
      selectedVersion.templateVersion,
    );
    if (!result.draft || result.state.status !== "saved") {
      setBindingFailure(result.state.failureCode || "prompt_template_binding_ineligible");
      setBindingStatus("failed");
      return;
    }
    setBindingStatus("saved");
    await loadApplicationDrafts();
    onOpenPublishReview?.(result.draft.draftId);
  }

  if (applicationKind !== "prompt_application") {
    return (
      <section className="prompt-application-template-panel not-applicable" aria-label={t($ => $.templateWorkspace)}>
        <div className="section-heading compact-heading">
          <div><p className="eyebrow">{t($ => $.applicationType)}</p><h4>{t($ => $.notApplicableTitle)}</h4></div>
          <span className="status-badge neutral">{t($ => $.notApplicable)}</span>
        </div>
        <p>{t($ => $.onlyPrompt)}</p>
      </section>
    );
  }

  return (
    <section className="prompt-application-template-panel" id="prompt-application-template-workspace" aria-labelledby="prompt-template-title">
      <div className="section-heading compact-heading">
        <div><p className="eyebrow">{t($ => $.templateOwnerHeading)}</p><h4 id="prompt-template-title">{t($ => $.templateTitle)}</h4></div>
        <span className={`status-badge ${operation.status === "saved" || operation.status === "versioned" ? "good" : operation.failureCode ? "bad" : "neutral"}`}>
          {t($ => $.states[operation.status])}
        </span>
      </div>

      <div className="prompt-template-scope">
        <article><span>{t($ => $.application)}</span><strong>{applicationName}</strong><code>{applicationId}</code></article>
        <article><span>{t($ => $.templateOwner)}</span><strong>{templateConfig.mode}</strong><code>{input.templateId}</code></article>
        <article><span>{t($ => $.privacy)}</span><strong>{t($ => $.sourceOnly)}</strong><p>{t($ => $.templatePrivacy)}</p></article>
      </div>

      {!applicationActive ? <p className="failure-summary">{t($ => $.archivedTemplate)}</p> : null}

      <div className="prompt-template-layout">
        <article className="prompt-template-editor">
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.templateDraft)}</p><h5>{input.templateId}</h5></div><span className="status-badge neutral">{t($ => $.draftVersion, { version: operation.currentDraftVersion })}</span></div>
          <label>{t($ => $.templateId)}<input value={input.templateId} onChange={(event) => edit({ templateId: event.target.value })} disabled={operation.currentDraftVersion > 0} /></label>
          <label>{t($ => $.templateName)}<input value={input.templateName} maxLength={80} onChange={(event) => edit({ templateName: event.target.value })} /></label>
          <label>{t($ => $.description)}<textarea value={input.description} maxLength={512} rows={3} onChange={(event) => edit({ description: event.target.value })} /></label>

          <fieldset>
            <legend>{t($ => $.orderedMessages)}</legend>
            {input.messages.map((message, index) => (
              <div className="prompt-template-message" key={`${index}-${message.role}`}>
                <select value={message.role} onChange={(event) => editMessage(index, { role: event.target.value as PromptTemplateRole })}>
                  <option value="system">system</option><option value="developer">developer</option><option value="user">user</option>
                </select>
                <textarea value={message.content} rows={4} maxLength={16384} onChange={(event) => editMessage(index, { content: event.target.value })} />
                <button type="button" onClick={() => removeMessage(index)} disabled={input.messages.length === 1}>{t($ => $.remove)}</button>
              </div>
            ))}
            <button type="button" onClick={() => edit({ messages: [...input.messages, { role: "user", content: "{{ value }}" }] })} disabled={input.messages.length >= 16}>{t($ => $.addMessage)}</button>
          </fieldset>

          <fieldset>
            <legend>{t($ => $.variables)}</legend>
            {input.variables.map((variable, index) => (
              <div className="prompt-template-variable" key={`${index}-${variable.name}`}>
                <input aria-label={t($ => $.variableFields.name, { index: index + 1 })} value={variable.name} onChange={(event) => editVariable(index, { name: event.target.value })} />
                <select aria-label={t($ => $.variableFields.type, { index: index + 1 })} value={variable.type} onChange={(event) => editVariable(index, { type: event.target.value as PromptTemplateVariableType, defaultValue: undefined })}>
                  <option value="string">string</option><option value="integer">integer</option><option value="number">number</option><option value="boolean">boolean</option><option value="string_list">string_list</option>
                </select>
                <label><input type="checkbox" checked={variable.required} onChange={(event) => editVariable(index, { required: event.target.checked, defaultValue: undefined })} />{t($ => $.required)}</label>
                <input aria-label={t($ => $.variableFields.description, { index: index + 1 })} value={variable.description} maxLength={512} onChange={(event) => editVariable(index, { description: event.target.value })} />
                {!variable.required ? <input aria-label={t($ => $.variableFields.default, { index: index + 1 })} value={formatDefaultValue(variable.defaultValue)} placeholder={t($ => $.optionalDefault)} onChange={(event) => editVariable(index, { defaultValue: parseDefaultValue(event.target.value, variable.type) })} /> : null}
                <button type="button" onClick={() => removeVariable(index)}>{t($ => $.remove)}</button>
              </div>
            ))}
            <button type="button" onClick={() => edit({ variables: [...input.variables, { name: `value${input.variables.length + 1}`, type: "string", required: true, description: "" }] })} disabled={input.variables.length >= 64}>{t($ => $.addVariable)}</button>
          </fieldset>

          <fieldset>
            <legend>{t($ => $.outputContract)}</legend>
            <label>{t($ => $.kind)}<select value={input.outputContract.kind} onChange={(event) => setOutputKind(event.target.value as PromptTemplateOutputKind)}><option value="text">text</option><option value="json_object">json_object</option></select></label>
            <label><input type="checkbox" checked={input.outputContract.allowEmpty} onChange={(event) => edit({ outputContract: { ...input.outputContract, allowEmpty: event.target.checked } })} />{t($ => $.allowEmptyOutput)}</label>
            {input.outputContract.kind === "json_object" ? <>
              <label htmlFor="prompt-output-schema">{t($ => $.outputSchema)}</label>
              <textarea id="prompt-output-schema" rows={16} spellCheck={false} value={schemaSource}
                aria-invalid={Boolean(parsedSchema.error)} aria-describedby="prompt-output-schema-help prompt-output-schema-error"
                onChange={(event) => {
                  const source = event.target.value;
                  setSchemaSource(source);
                  edit({ outputContract: { ...input.outputContract, jsonSchema: parsePromptOutputSchema(source).schema } });
                }} />
              <p id="prompt-output-schema-help" className="boundary-note">{t($ => $.schemaHelp)}</p>
              <p id="prompt-output-schema-error" className="failure-summary" aria-live="polite">{parsedSchema.issue ? `${parsedSchema.issue.path}：${t($ => $.schemaIssues[parsedSchema.issue!.code])}` : parsedSchema.error}</p>
            </> : <p className="boundary-note">{t($ => $.schemaRetention)}</p>}
            <label>{t($ => $.maximumBytes)}<input type="number" min={1} max={65536} value={input.outputContract.maxBytes} onChange={(event) => edit({ outputContract: { ...input.outputContract, maxBytes: Number(event.target.value) } })} /></label>
          </fieldset>

          <div className="application-draft-actions">
            <button type="button" onClick={() => void validateRemote()} disabled={saving || !applicationActive}>{t($ => $.validate)}</button>
            <button type="button" onClick={() => void saveDraft()} disabled={saving || !enabled || !localValidation.isValid}>{t($ => $.saveCas)}</button>
            <button type="button" onClick={() => void createVersion()} disabled={saving || !enabled || !localValidation.isValid || hasUnsavedChanges || !operation.draft || operation.currentDraftVersion < 1}>{t($ => $.createVersion)}</button>
          </div>
        </article>

        <article className="prompt-template-review">
          {hasUnsavedChanges ? <p>{t($ => $.unsavedChanges)}</p> : null}
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.deterministicReview)}</p><h5>{t($ => $.templateStates[operation.status])}</h5></div><span className={`status-badge ${localValidation.isValid ? "good" : "bad"}`}>{t($ => $.states[localValidation.state])}</span></div>
          {operation.failureCode ? <p className="failure-summary">{operation.failureCode}</p> : null}
          {(operation.validation.findings.length ? operation.validation.findings : localValidation.findings).map((finding) => (
            <p className="failure-summary" key={`${finding.code}-${finding.field}`}><strong>{finding.field}</strong> · {finding.code} · {t($ => $.findings[finding.messageKey ?? "serverFinding"])}</p>
          ))}
          {operation.status === "version_conflict" ? <button type="button" onClick={() => void restoreDraft(input.templateId)}>{t($ => $.restoreDraft, { version: operation.currentDraftVersion })}</button> : null}

          <label>{t($ => $.syntheticVariables)}<textarea rows={7} value={previewInput} onChange={(event) => { setPreviewInput(event.target.value); setPreview(null); }} /></label>
          <button type="button" onClick={renderPreview}>{t($ => $.renderPreview)}</button>
          {preview ? <div className={`prompt-template-preview ${preview.status}`}>
            <strong>{preview.status === "valid" ? t($ => $.previewOnly) : t($ => $.previewBlocked)}</strong>
            {preview.messages.map((message, index) => <div key={`${message.role}-${index}`}><code>{message.role}</code><pre>{message.content}</pre></div>)}
            {preview.findings.map((finding) => <p className="failure-summary" key={`${finding.code}-${finding.field}`}>{finding.code} · {finding.field} · {t($ => $.findings[finding.messageKey ?? "serverFinding"])}</p>)}
          </div> : null}
          <p className="boundary-note">{t($ => $.previewPrivacy)}</p>
        </article>
      </div>

      <div className="prompt-template-lower-grid">
        <article>
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.savedDrafts)}</p><h5>{draftsLoading ? t($ => $.listLoading) : drafts.status === "offline" ? t($ => $.listOffline) : drafts.status === "failed" ? t($ => $.listFailed) : drafts.summaries.length ? t($ => $.listedCount, { count: drafts.summaries.length }) : t($ => $.listEmpty)}</h5></div><button type="button" onClick={() => void refreshDrafts()}>{t($ => $.refresh)}</button></div>
          {drafts.failureCode ? <p className="failure-summary">{drafts.failureCode}</p> : null}
          {drafts.summaries.map((draft) => <button type="button" className="prompt-template-summary" key={draft.templateId} disabled={saving} onClick={() => void restoreDraft(draft.templateId)}><strong>{draft.templateName}</strong><span>{draft.templateId} · v{draft.draftVersion}</span><small>{draft.messageRoles.join(" → ")} · {draft.variableNames.join(", ") || t($ => $.noVariables)}</small></button>)}
        </article>

        <article>
          <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.immutableVersions)}</p><h5>{versionsLoading ? t($ => $.listLoading) : versions.status === "offline" ? t($ => $.listOffline) : versions.status === "failed" ? t($ => $.listFailed) : versions.summaries.length ? t($ => $.listedCount, { count: versions.summaries.length }) : t($ => $.listEmpty)}</h5></div><button type="button" onClick={() => void refreshVersions()}>{t($ => $.refresh)}</button></div>
          {versions.failureCode ? <p className="failure-summary">{versions.failureCode}</p> : null}
          {versions.summaries.map((version) => <button type="button" className={selectedTemplateVersion === version.templateVersion ? "prompt-template-summary selected" : "prompt-template-summary"} key={version.templateVersion} onClick={() => void openVersion(version.templateVersion)}><strong>{t($ => $.version)}{version.templateVersion}</strong><span>{t($ => $.sourceDraftVersion, { version: version.sourceDraftVersion })} · {version.outputKind}</span><small>{version.templateDigest}</small></button>)}
          {operation.version ? <div className="prompt-template-version-detail"><strong>{operation.version.templateName} · {t($ => $.immutableVersion, { version: operation.version.templateVersion })}</strong><code>{operation.version.templateDigest}</code><p>{t($ => $.messageCount, { count: operation.version.messages.length })} · {t($ => $.variableCount, { count: operation.version.variables.length })}</p><pre aria-label={t($ => $.immutableContract)}>{JSON.stringify(operation.version.outputContract, null, 2)}</pre></div> : null}
        </article>
      </div>

      <article className="prompt-template-binding">
        <div className="application-api-card-heading"><div><p className="eyebrow">{t($ => $.configurationBinding)}</p><h5>{t($ => $.bindExactVersion)}</h5></div><button type="button" onClick={() => void loadApplicationDrafts()} disabled={!enabled}>{t($ => $.loadDrafts)}</button></div>
        <label>{t($ => $.validDraft)}<select value={selectedDraftId} onChange={(event) => setSelectedDraftId(event.target.value)}><option value="">{t($ => $.noDraft)}</option>{applicationDrafts.summaries.map((draft) => <option key={draft.draftId} value={draft.draftId} disabled={draft.applicationKind !== "prompt_application" || draft.validationState !== "valid"}>{draft.draftId} · v{draft.draftVersion}{draft.promptTemplateRef ? ` · template v${draft.promptTemplateRef.templateVersion}` : ""}</option>)}</select></label>
        <label>{t($ => $.templateVersion)}<select value={selectedTemplateVersion} onChange={(event) => setSelectedTemplateVersion(Number(event.target.value))}><option value={0}>{t($ => $.noVersion)}</option>{versions.summaries.map((version) => <option key={version.templateVersion} value={version.templateVersion}>{version.templateId} · v{version.templateVersion}</option>)}</select></label>
        <button type="button" onClick={() => void bindVersion()} disabled={!canBind}>{t($ => $.bindReview)}</button>
        {bindingFailure ? <p className="failure-summary">{bindingFailure}</p> : null}
        <p className="boundary-note">{bindingStatus === "pending" ? t($ => $.bindingPending) : bindingStatus === "failed" ? t($ => $.bindingFailed) : bindingStatus === "saved" ? t($ => $.bindingSaved) : t($ => $.bindingBoundary)}</p>
      </article>
    </section>
  );

  function editMessage(index: number, patch: Partial<PromptTemplateDraftInput["messages"][number]>) {
    edit({ messages: input.messages.map((message, current) => current === index ? { ...message, ...patch } : message) });
  }

  function removeMessage(index: number) {
    edit({ messages: input.messages.filter((_message, current) => current !== index) });
  }

  function editVariable(index: number, patch: Partial<PromptTemplateDraftInput["variables"][number]>) {
    edit({ variables: input.variables.map((variable, current) => current === index ? { ...variable, ...patch } : variable) });
  }

  function removeVariable(index: number) {
    edit({ variables: input.variables.filter((_variable, current) => current !== index) });
  }

  function setOutputKind(kind: PromptTemplateOutputKind) {
    edit({
      outputContract: kind === "text"
        ? { kind, allowEmpty: input.outputContract.allowEmpty, maxBytes: input.outputContract.maxBytes }
        : {
          kind,
          allowEmpty: input.outputContract.allowEmpty,
          maxBytes: input.outputContract.maxBytes,
          jsonSchema: parsedSchema.schema,
        },
    });
  }
}

function initialOperation(): PromptTemplateOperation {
  return {
    status: templateConfig.mode === "offline" ? "offline" : "idle",
    draft: null,
    version: null,
    validation: { state: "invalid", isValid: false, findings: [] },
    failureCode: "",
    currentDraftVersion: 0,
    currentTemplateVersion: 0,
    summary: templateConfig.mode === "offline" ? "Prompt Application Web 未启用。" : "编辑模板并执行本地与服务端校验。",
  };
}

function initialDraftList(): PromptTemplateListResult {
  return { status: templateConfig.mode === "offline" ? "offline" : "empty", summaries: [], failureCode: "", summary: "加载当前应用的模板草案。" };
}

function initialVersionList(): PromptTemplateVersionListResult {
  return { status: templateConfig.mode === "offline" ? "offline" : "empty", summaries: [], failureCode: "", summary: "选择或保存模板后加载不可变版本。" };
}

function draftToInput(draft: NonNullable<PromptTemplateOperation["draft"]>): PromptTemplateDraftInput {
  return {
    templateId: draft.templateId,
    workspaceId: draft.workspaceId,
    applicationId: draft.applicationId,
    templateName: draft.templateName,
    description: draft.description,
    messages: draft.messages.map((message) => ({ ...message })),
    variables: draft.variables.map((variable) => ({ ...variable, ...(Array.isArray(variable.defaultValue) ? { defaultValue: [...variable.defaultValue] } : {}) })),
    outputContract: { ...draft.outputContract },
  };
}

function formatDefaultValue(value: unknown): string {
  if (value === undefined) return "";
  return typeof value === "string" ? value : JSON.stringify(value);
}

function parseDefaultValue(value: string, type: PromptTemplateVariableType): string | number | boolean | string[] | undefined {
  if (!value.trim()) return undefined;
  if (type === "string") return value;
  try {
    const parsed: unknown = JSON.parse(value);
    if (type === "integer" || type === "number") return typeof parsed === "number" ? parsed : value;
    if (type === "boolean") return typeof parsed === "boolean" ? parsed : value;
    return Array.isArray(parsed) && parsed.every((item) => typeof item === "string") ? parsed : value;
  } catch {
    return value;
  }
}
