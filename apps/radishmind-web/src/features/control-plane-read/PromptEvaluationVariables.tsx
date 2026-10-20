import { useId } from "react";
import { useTranslation } from "react-i18next";
import type { PromptTemplateVariable } from "./promptApplicationTemplateConsumer.ts";
import type { PromptPlanIssue, PromptPlanSample, PromptPlanValue } from "./promptEvaluationPlanDraft.ts";

export default function PromptEvaluationVariables({ sample, declarations, disabled, issues, onChange }: {
  sample: PromptPlanSample;
  declarations: PromptTemplateVariable[];
  disabled: boolean;
  issues: PromptPlanIssue[];
  onChange: (variables: PromptPlanSample["variables"]) => void;
}) {
  const { t } = useTranslation("evaluation");
  const names = new Set(declarations.map(variable => variable.name));
  const unknown = Object.entries(sample.variables).filter(([name]) => !names.has(name));
  function change(name: string, value: PromptPlanValue | undefined) {
    const variables = { ...sample.variables };
    if (value === undefined) delete variables[name];
    else variables[name] = value;
    onChange(variables);
  }
  return <div className="prompt-plan-variables">
    {!declarations.length ? <p>{t($ => $.promptPlan.noVariables)}</p> : null}
    {declarations.map(variable => <VariableInput key={variable.name} variable={variable} value={sample.variables[variable.name]}
      disabled={disabled} issues={issues.filter(issue => issue.itemKey === sample.itemKey && issue.field === variable.name)}
      onChange={value => change(variable.name, value)} />)}
    {unknown.length ? <section className="prompt-plan-raw-values" aria-label={t($ => $.promptPlan.undeclaredValues)}>
      <h6>{t($ => $.promptPlan.undeclaredValues)}</h6>
      {unknown.map(([name, value]) => <div key={name}><code>{name}</code><pre>{rawValue(value)}</pre></div>)}
      <p>{t($ => $.promptPlan.rawBoundary)}</p>
    </section> : null}
  </div>;
}

function VariableInput({ variable, value, disabled, issues, onChange }: {
  variable: PromptTemplateVariable; value: PromptPlanValue | undefined; disabled: boolean; issues: PromptPlanIssue[];
  onChange: (value: PromptPlanValue | undefined) => void;
}) {
  const { t } = useTranslation("evaluation");
  const id = useId();
  const invalidOriginal = value !== undefined && !controlMatches(variable, value);
  const text = value?.kind === "value" && typeof value.value === "string" ? value.value : "";
  const numericText = value?.kind === "number" ? value.text : value?.kind === "value" && typeof value.value === "number" ? String(value.value) : "";
  const errorId = `${id}-errors`;
  return <fieldset className="prompt-plan-variable" disabled={disabled} data-variable={variable.name}>
    <legend><code>{variable.name}</code><span>{variable.type} · {variable.required ? t($ => $.promptPlan.required) : t($ => $.promptPlan.optional)}</span></legend>
    {variable.description ? <p>{variable.description}</p> : null}
    {invalidOriginal ? <div className="prompt-plan-raw-values">
      <strong>{t($ => $.promptPlan.rawValue)}</strong><pre>{rawValue(value)}</pre><p>{t($ => $.promptPlan.rawBoundary)}</p>
    </div> : <>
      <div className="prompt-plan-presence">
        <select aria-label={t($ => $.promptPlan.presence, { name: variable.name })} value={value === undefined ? "omitted" : "provided"}
          onChange={event => onChange(event.target.value === "omitted" ? undefined : initialValue(variable))}>
          <option value="omitted">{t($ => $.promptPlan.omitted)}</option><option value="provided">{t($ => $.promptPlan.provided)}</option>
        </select>
        {value === undefined ? <small>{variable.defaultValue !== undefined ? t($ => $.promptPlan.defaultValue, { value: JSON.stringify(variable.defaultValue) }) : null} {t($ => $.promptPlan.omittedBoundary)}</small> : null}
      </div>
      {value !== undefined && variable.type === "string" ? <textarea aria-label={variable.name} value={text} rows={3}
        placeholder={t($ => $.promptPlan.emptyString)} aria-invalid={issues.length > 0} aria-describedby={issues.length ? errorId : undefined}
        onChange={event => onChange({ kind: "value", value: event.target.value })} /> : null}
      {value !== undefined && (variable.type === "integer" || variable.type === "number") ? <input aria-label={variable.name} type="text" inputMode="decimal" value={numericText}
        aria-invalid={issues.length > 0} aria-describedby={issues.length ? errorId : undefined}
        onChange={event => onChange({ kind: "number", integer: variable.type === "integer", text: event.target.value })} /> : null}
      {value?.kind === "value" && variable.type === "boolean" ? <select aria-label={variable.name} value={String(value.value)}
        onChange={event => onChange({ kind: "value", value: event.target.value === "true" })}>
        <option value="false">false</option><option value="true">true</option>
      </select> : null}
      {value?.kind === "list" && variable.type === "string_list" ? <div className="prompt-plan-string-list">
        {!value.rows.length ? <span>{t($ => $.promptPlan.emptyList)}</span> : null}
        {value.rows.map((row, index) => <div key={row.editorKey} className="prompt-plan-list-row">
          <span>{index + 1}</span><textarea aria-label={t($ => $.promptPlan.listItem, { name: variable.name, number: index + 1 })} value={row.value} rows={1}
            placeholder={t($ => $.promptPlan.emptyString)} aria-invalid={issues.length > 0}
            onChange={event => onChange({ kind: "list", rows: value.rows.map(item => item.editorKey === row.editorKey ? { ...item, value: event.target.value } : item) })} />
          <div className="prompt-plan-list-actions">
            <button type="button" className="secondary-action" aria-label={`${t($ => $.promptPlan.moveUp)} ${index + 1}`} disabled={index === 0}
              onClick={() => onChange({ kind: "list", rows: moveRow(value.rows, index, -1) })}>↑</button>
            <button type="button" className="secondary-action" aria-label={`${t($ => $.promptPlan.moveDown)} ${index + 1}`} disabled={index === value.rows.length - 1}
              onClick={() => onChange({ kind: "list", rows: moveRow(value.rows, index, 1) })}>↓</button>
            <button type="button" className="secondary-action" aria-label={`${t($ => $.promptPlan.deleteListItem)} ${variable.name} ${index + 1}`}
              onClick={() => onChange({ kind: "list", rows: value.rows.filter(item => item.editorKey !== row.editorKey) })}>×</button>
          </div>
        </div>)}
        <div className="prompt-plan-inline-actions"><button type="button" className="secondary-action" disabled={value.rows.length >= 128}
          onClick={() => onChange({ kind: "list", rows: [...value.rows, { editorKey: crypto.randomUUID(), value: "" }] })}>{t($ => $.promptPlan.addListItem)}</button>
          <button type="button" className="secondary-action" onClick={() => onChange({ kind: "list", rows: [] })}>{t($ => $.promptPlan.clearList)}</button></div>
      </div> : null}
    </>}
    {issues.length ? <ul id={errorId} className="prompt-plan-field-errors">{issues.map((issue, index) => <li key={`${issue.code}-${index}`}>{t($ => $.promptPlan.issues[issue.code])}</li>)}</ul> : null}
  </fieldset>;
}

function initialValue(variable: PromptTemplateVariable): PromptPlanValue {
  if (variable.type === "string_list") return { kind: "list", rows: [] };
  if (variable.type === "integer" || variable.type === "number") return { kind: "number", integer: variable.type === "integer", text: "" };
  return { kind: "value", value: variable.type === "boolean" ? false : "" };
}

function controlMatches(variable: PromptTemplateVariable, value: PromptPlanValue): boolean {
  if (value.kind === "number") return variable.type === "number" || variable.type === "integer";
  if (value.kind === "list") return variable.type === "string_list";
  if (variable.type === "integer") return typeof value.value === "number" && Number.isSafeInteger(value.value);
  if (variable.type === "number") return typeof value.value === "number";
  return typeof value.value === variable.type;
}

function rawValue(value: PromptPlanValue): string {
  return value.kind === "number" ? value.text : JSON.stringify(value.kind === "list" ? value.rows.map(row => row.value) : value.value, null, 2);
}

function moveRow<T>(values: T[], index: number, direction: -1 | 1): T[] {
  const next = [...values];
  [next[index], next[index + direction]] = [next[index + direction], next[index]];
  return next;
}
