import type { TFunction } from "i18next";
import type { artifactFeedback } from "../../i18n/locales/en-US/artifactFeedback.ts";

const failureKeys: Record<string, keyof typeof artifactFeedback> = {
  application_session_http_disabled: "offline",
  application_session_scope_denied: "permission",
  application_session_environment_denied: "permission",
  application_result_artifact_payload_invalid: "invalid",
  application_result_artifact_store_unavailable: "unavailable",
  application_result_artifact_not_found: "notFound",
  application_result_artifact_store_contract_mismatch: "contract",
  application_result_artifact_lifecycle_state_conflict: "conflict",
  application_result_artifact_lifecycle_version_conflict: "conflict",
  application_result_artifact_conflict: "conflict",
  application_result_artifact_request_canceled: "canceled",
  application_result_artifact_content_too_large: "content",
  application_result_artifact_source_unavailable: "content",
};

export function artifactFailureMessage(t: TFunction<"applications">, code: string): string {
  const known = Object.hasOwn(failureKeys, code);
  const message = t($ => $.artifactFeedback[known ? failureKeys[code] : "generic"]);
  return known ? `${message} (${code})` : message;
}
