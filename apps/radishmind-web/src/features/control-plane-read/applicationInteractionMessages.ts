import type { TFunction } from "i18next";
import { interaction } from "../../i18n/locales/en-US/interaction.ts";

const failureKeys = {
  application_session_http_disabled: "offline",
  application_session_payload_invalid: "invalid",
  application_session_workflow_options_invalid: "invalid",
  application_runtime_input_invalid: "invalid",
  application_session_secret_material_forbidden: "secret",
  application_session_request_canceled: "canceled",
  application_session_version_conflict: "conflict",
  application_session_store_contract_mismatch: "contract",
  application_session_scope_denied: "permission",
  application_session_environment_denied: "permission",
  application_session_store_unavailable: "unavailable",
  application_session_authority_not_found: "authorityMissing",
  application_session_authority_changed: "authorityChanged",
  application_session_profile_ineligible: "ineligible",
} as const;

export function interactionFailureCode(code: string): string {
  return /^[a-z][a-z0-9_]{2,95}$/u.test(code) ? code : "";
}
export function interactionFailureMessage(t: TFunction<"applications">, code: string): string {
  const key = Object.hasOwn(failureKeys, code) ? failureKeys[code as keyof typeof failureKeys] : "generic";
  return t($ => $.interaction.failures[key]);
}
export function interactionStatus(t: TFunction<"applications">, state: string): string {
  return Object.hasOwn(interaction.states, state) ? t($ => $.interaction.states[state as keyof typeof interaction.states]) : state;
}
