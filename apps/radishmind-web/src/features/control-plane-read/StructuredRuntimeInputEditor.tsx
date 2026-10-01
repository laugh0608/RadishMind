import type { TFunction } from "i18next";
import type { StructuredRuntimeInputFieldError } from "./structuredRuntimeInput.ts";
import "../../i18n/workflowInputResources.ts";
import { useTranslation } from "react-i18next";
import type {
  StructuredRuntimeInputContract,
  StructuredRuntimeInputDrafts,
} from "./structuredRuntimeInput.ts";

type Props = {
  contract: StructuredRuntimeInputContract;
  drafts: StructuredRuntimeInputDrafts;
  fieldErrors: Record<string, StructuredRuntimeInputFieldError>;
  disabled?: boolean;
  onChange: (drafts: StructuredRuntimeInputDrafts) => void;
};

export default function StructuredRuntimeInputEditor({ contract, drafts, fieldErrors, disabled = false, onChange }: Props) {
  const { t } = useTranslation("workflow");
  function update(name: string, value: string | boolean | undefined) {
    const next = { ...drafts };
    if (value === undefined) delete next[name];
    else next[name] = value;
    onChange(next);
  }

  return <fieldset className="structured-runtime-input" disabled={disabled}>
    <legend>{t($ => $.input.structuredRuntimeInput)}</legend>
    <div className="structured-runtime-input-contract">
      <strong>{contract.contractId}</strong>
      <span>{contract.summary}</span>
      <code>{shortDigest(contract.contractDigest)}</code>
    </div>
    <div className="structured-runtime-input-fields">
      {contract.fields.map((field) => <div className={`structured-runtime-input-field ${fieldErrors[field.name] ? "invalid" : ""}`} key={field.name}>
        <div className="structured-runtime-input-label">
          <label
            id={`structured-runtime-input-${field.name}-label`}
            htmlFor={field.valueType === "boolean" ? undefined : `structured-runtime-input-${field.name}`}
          >{field.label}{field.required ? <span aria-hidden="true"> *</span> : null}</label>
          <code>{field.name} · {field.valueType}</code>
        </div>
        {field.description ? <p>{field.description}</p> : null}
        {field.valueType === "boolean" ? <div className="structured-runtime-boolean" role="group" aria-labelledby={`structured-runtime-input-${field.name}-label`}>
          <span>{t($ => $.input.notSet)}</span>
          <label><input type="radio" name={`structured-runtime-input-${field.name}`} checked={drafts[field.name] === true} onChange={() => update(field.name, true)} />true</label>
          <label><input type="radio" name={`structured-runtime-input-${field.name}`} checked={drafts[field.name] === false} onChange={() => update(field.name, false)} />false</label>
          {Object.hasOwn(drafts, field.name) ? <button type="button" className="text-button" onClick={() => update(field.name, undefined)}>{t($ => $.input.clear)}</button> : null}
        </div> : <input
          id={`structured-runtime-input-${field.name}`}
          type="text"
          inputMode={field.valueType === "integer" || field.valueType === "number" ? "decimal" : "text"}
          value={typeof drafts[field.name] === "string" ? String(drafts[field.name]) : ""}
          aria-invalid={Boolean(fieldErrors[field.name])}
          aria-describedby={fieldErrors[field.name] ? `structured-runtime-input-${field.name}-error` : undefined}
          onChange={(event) => update(field.name, event.currentTarget.value)}
        />}
        {fieldErrors[field.name] ? <p className="structured-runtime-input-error" id={`structured-runtime-input-${field.name}-error`} role="alert">{structuredInputErrorMessage(t, fieldErrors[field.name])}</p> : null}
      </div>)}
    </div>
    <p className="structured-runtime-input-boundary">{t($ => $.input.inputsAreRetainedOnlyForTheCurrentRequestDurableRecords)}</p>
  </fieldset>;
}

function shortDigest(value: string): string {
  return value.length > 24 ? `${value.slice(0, 23)}…` : value;
}

export function structuredInputErrorMessage(t: TFunction<"workflow">, error: StructuredRuntimeInputFieldError): string {
  switch (error) {
    case "unknown_field": return t($ => $.input.error_unknown_field);
    case "required": return t($ => $.input.error_required);
    case "boolean": return t($ => $.input.error_boolean);
    case "string": return t($ => $.input.error_string);
    case "integer": return t($ => $.input.error_integer);
    case "number": return t($ => $.input.error_number);
    case "string_budget": return t($ => $.input.error_string_budget);
    case "secret_material": return t($ => $.input.error_secret_material);
    case "integer_syntax": return t($ => $.input.error_integer_syntax);
    case "integer_range": return t($ => $.input.error_integer_range);
    case "number_syntax": return t($ => $.input.error_number_syntax);
  }
}
