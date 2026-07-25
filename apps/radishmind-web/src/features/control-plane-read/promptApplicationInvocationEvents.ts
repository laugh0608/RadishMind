export const PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT =
  "radishmind:prompt-application-credential-handoff";

export type PromptApplicationCredentialHandoffDetail = {
  applicationId: string;
  apiKeyId: string;
  token: string;
};

const APPLICATION_ID_PATTERN = /^app_[a-z2-7]{16}$/u;
const API_KEY_ID_PATTERN = /^key_[a-z2-7]{16}$/u;
const TOKEN_PATTERN = /^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u;

export function createPromptApplicationCredentialHandoffDetail(
  applicationId: string,
  apiKeyId: string,
  token: string,
): PromptApplicationCredentialHandoffDetail {
  const detail = {
    applicationId: applicationId.trim(),
    apiKeyId: apiKeyId.trim(),
    token: token.trim(),
  };
  if (!APPLICATION_ID_PATTERN.test(detail.applicationId) ||
    !API_KEY_ID_PATTERN.test(detail.apiKeyId) ||
    !TOKEN_PATTERN.test(detail.token) ||
    !detail.token.includes(detail.apiKeyId)) {
    throw new Error("Prompt Application credential handoff is invalid.");
  }
  return detail;
}

export function requestPromptApplicationCredentialHandoff(
  applicationId: string,
  apiKeyId: string,
  token: string,
): void {
  window.dispatchEvent(new CustomEvent<PromptApplicationCredentialHandoffDetail>(
    PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT,
    { detail: createPromptApplicationCredentialHandoffDetail(applicationId, apiKeyId, token) },
  ));
}
