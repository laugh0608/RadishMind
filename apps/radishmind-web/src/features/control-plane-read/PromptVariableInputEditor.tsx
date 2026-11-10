import "../../i18n/promptResources.ts";
import "./promptEvaluationPlan.css";
import { useTranslation } from "react-i18next";
import PromptVariableFields from "./PromptVariableFields.tsx";
import type { usePromptVariableInput } from "./usePromptVariableInput.ts";

export default function PromptVariableInputEditor({ editor, disabled, jsonLabel }: {
  editor: ReturnType<typeof usePromptVariableInput>; disabled: boolean; jsonLabel: string;
}) {
  const { t } = useTranslation("prompt");
  const { t: plan } = useTranslation("evaluation");
  const { input, reference } = editor;
  const locked = disabled || editor.busy;
  const source = reference.status === "ready" ? reference.current : null;
  return <section className="prompt-variable-input" aria-label={t($ => $.variableInput.title)}>
    <div className="prompt-plan-inline-actions">
      <button type="button" className="secondary-action" onClick={() => void editor.refresh()} disabled={locked}>{editor.busy ? t($ => $.variableInput.loading) : t($ => $.variableInput.read)}</button>
      <button type="button" className="secondary-action" onClick={editor.reset} disabled={locked}>{t($ => $.variableInput.clear)}</button>
    </div>
    <div className="prompt-plan-reference" role="status" data-reference-state={reference.status}>
      {reference.status === "ready" && source ? <><strong>{t($ => $.variableInput.ready)}</strong><code>{source.source.templateId} · v{source.source.templateVersion}</code><span>{t($ => $.variableInput.assignment, { version: source.assignmentVersion })}</span></> :
        reference.status === "changed" ? <><p>{t($ => $.variableInput.changed)}</p><code>{reference.candidate?.source.templateId} · v{reference.candidate?.source.templateVersion}</code><button type="button" className="secondary-action" onClick={editor.adopt} disabled={locked}>{t($ => $.variableInput.adopt)}</button></> :
        reference.status === "unavailable" ? <><p>{reference.code === "prompt_input_session_authority_changed" ? t($ => $.variableInput.sessionChanged) : reference.code === "prompt_input_session_required" ? t($ => $.variableInput.sessionRequired) : t($ => $.variableInput.unavailable)}</p><code>{reference.code}</code></> :
        <p>{editor.busy ? t($ => $.variableInput.loading) : t($ => $.variableInput.idle)}</p>}
    </div>
    <label><span>{t($ => $.variableInput.mode)}</span><select value={input.mode} disabled={locked} onChange={event => editor.changeMode(event.target.value as "form" | "json")}>
      <option value="form">{plan($ => $.promptPlan.formMode)}</option><option value="json">{plan($ => $.promptPlan.jsonMode)}</option>
    </select></label>
    {input.mode === "json" ? <><label>{jsonLabel}<textarea rows={8} value={input.text} disabled={locked} onChange={event => editor.change({ mode: "json", text: event.target.value })} /></label><p>{t($ => $.variableInput.jsonBoundary)}</p></> :
      source ? <PromptVariableFields values={input.values} declarations={source.source.variables} disabled={locked}
        issues={editor.issues.map(issue => ({ field: issue.field, message: plan($ => $.promptPlan.issues[issue.code]) }))}
        onChange={values => editor.change({ mode: "form", values })} /> : null}
    {editor.issues.length ? <div role="alert" className="prompt-plan-field-errors"><strong>{t($ => $.variableInput.issues)}</strong><ul>{editor.issues.map((issue, index) => <li key={index}>{issue.field ? <code>{issue.field} · </code> : null}{plan($ => $.promptPlan.issues[issue.code])}{issue.line ? ` (${issue.line}:${issue.column})` : ""}</li>)}</ul></div> : null}
    <p>{t($ => $.variableInput.boundary)}</p>
  </section>;
}
