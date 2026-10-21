import { useId } from "react";
import { useTranslation } from "react-i18next";
import "../../i18n/workflowDraftResources.ts";
import { validateWorkflowDraftInputContract, type WorkflowDraftInputContract } from "./workflowDraftInputContract.ts";
import type { StructuredRuntimeInputField } from "./structuredRuntimeInput.ts";

export function WorkflowDraftInputContractEditor({ contract, disabled, onChange }: {
  contract: WorkflowDraftInputContract;
  disabled: boolean;
  onChange: (contract: WorkflowDraftInputContract) => void;
}) {
  const { t } = useTranslation("workflow");
  const id = useId();
  const issues = validateWorkflowDraftInputContract(contract);
  const messages = {
    identity: t($ => $.draft.contractIdentityError), summary: t($ => $.draft.contractTextError),
    count: t($ => $.draft.contractCountError), name: t($ => $.draft.contractNameError),
    duplicate: t($ => $.draft.contractDuplicateError), type: t($ => $.draft.contractTypeError),
    required: t($ => $.draft.contractRequiredError), label: t($ => $.draft.contractLabelError),
    description: t($ => $.draft.contractTextError), sensitive: t($ => $.draft.contractSensitiveError),
  };
  const change = (patch: Partial<WorkflowDraftInputContract>) => onChange({ contractId: contract.contractId, summary: contract.summary, fields: contract.fields, ...patch });
  const editField = (index: number, patch: Partial<StructuredRuntimeInputField>) => change({ fields: contract.fields.map((field, position) => position === index ? { ...field, ...patch } : field) });
  const move = (index: number, direction: -1 | 1) => {
    const fields = [...contract.fields];
    const target = index + direction;
    if (target < 0 || target >= fields.length) return;
    [fields[index], fields[target]] = [fields[target]!, fields[index]!];
    change({ fields });
  };
  const add = () => {
    if (contract.fields.length >= 16) return;
    let number = 1;
    while (contract.fields.some((field) => field.name === `field_${number}`)) number += 1;
    change({ fields: [...contract.fields, { name: `field_${number}`, valueType: "string", required: false, label: t($ => $.draft.contractNewField, { number }), description: "" }] });
  };
  const errors = (path: string) => issues.filter((issue) => issue.field === path);
  const textField = (path: string, label: string, value: string, maxLength: number, update: (value: string) => void) => <label className="workflow-contract-field" key={path}>
    <span id={`${id}-${path}-label`}>{label}</span>
    <input aria-labelledby={`${id}-${path}-label`} value={value} maxLength={maxLength} disabled={disabled} aria-invalid={errors(path).length > 0} aria-describedby={errors(path).length ? `${id}-${path}-error` : undefined} onChange={(event) => update(event.currentTarget.value)} />
    {errors(path).length ? <small id={`${id}-${path}-error`} className="workflow-contract-error">{errors(path).map((issue) => messages[issue.code]).join(" ")}</small> : null}
  </label>;
  return <section className="workflow-contract-editor" aria-labelledby={`${id}-title`}>
    <header><p className="eyebrow">{t($ => $.draft.contractSchema)}</p><h4 id={`${id}-title`}>{t($ => $.draft.contractTitle)}</h4><p>{t($ => $.draft.contractBoundary)}</p></header>
    {textField("contractId", t($ => $.draft.contractId), contract.contractId, 160, (contractId) => change({ contractId }))}
    {textField("summary", t($ => $.draft.contractSummary), contract.summary, 4000, (summary) => change({ summary }))}
    <h5>{t($ => $.draft.contractFields, { count: contract.fields.length })}</h5>
    {contract.fields.map((field, index) => <fieldset key={index} disabled={disabled} className="workflow-contract-row">
      <legend>{t($ => $.draft.contractFieldNumber, { number: index + 1 })}</legend>
      <div className="workflow-contract-columns">
        {textField(`fields.${index}.name`, t($ => $.draft.contractFieldName), field.name, 64, (name) => editField(index, { name }))}
        <label className="workflow-contract-field"><span>{t($ => $.draft.contractFieldType)}</span><select value={field.valueType} onChange={(event) => editField(index, { valueType: event.currentTarget.value as StructuredRuntimeInputField["valueType"] })}>{["string", "integer", "number", "boolean"].map((value) => <option key={value} value={value}>{value}</option>)}</select></label>
        <label className="workflow-contract-required"><input type="checkbox" checked={field.required} onChange={(event) => editField(index, { required: event.currentTarget.checked })} />{t($ => $.draft.contractFieldRequired)}</label>
      </div>
      <div className="workflow-contract-columns">
        {textField(`fields.${index}.label`, t($ => $.draft.contractFieldLabel), field.label, 160, (label) => editField(index, { label }))}
        {textField(`fields.${index}.description`, t($ => $.draft.contractFieldDescription), field.description, 4000, (description) => editField(index, { description }))}
      </div>
      <div className="workflow-contract-actions">
        <button type="button" disabled={disabled || index === 0} onClick={() => move(index, -1)}>{t($ => $.draft.contractMoveUp)}</button>
        <button type="button" disabled={disabled || index === contract.fields.length - 1} onClick={() => move(index, 1)}>{t($ => $.draft.contractMoveDown)}</button>
        <button type="button" disabled={disabled} onClick={() => change({ fields: contract.fields.filter((_, position) => position !== index) })}>{t($ => $.draft.contractRemove)}</button>
      </div>
    </fieldset>)}
    <div className="workflow-contract-actions"><button type="button" disabled={disabled || contract.fields.length >= 16} onClick={add}>{t($ => $.draft.contractAdd)}</button></div>
    {issues.length ? <p className="workflow-contract-error" role="status">{t($ => $.draft.contractBlocked, { count: issues.length })} {errors("fields").map((issue) => messages[issue.code]).join(" ")}</p> : <p>{t($ => $.draft.contractValid)}</p>}
    <small>{contract.contractDigest ? t($ => $.draft.contractServerDigest, { digest: contract.contractDigest }) : t($ => $.draft.contractDigestPending)}</small>
  </section>;
}
