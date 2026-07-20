export const WORKFLOW_RAG_APPLICATION_CREDENTIAL_HANDOFF_EVENT =
  "radishmind:workflow-rag-application-credential-handoff";

export type WorkflowRAGApplicationCredentialHandoffDetail = {
  applicationId: string;
  apiKeyId: string;
  token: string;
};

const APPLICATION_ID_PATTERN = /^app_[a-z0-9]{16}$/u;
const API_KEY_ID_PATTERN = /^key_[a-z2-7]{16}$/u;
const TOKEN_PATTERN = /^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u;

export function createWorkflowRAGApplicationCredentialHandoffDetail(
  applicationId: string,
  apiKeyId: string,
  token: string,
): WorkflowRAGApplicationCredentialHandoffDetail {
  const detail = {
    applicationId: applicationId.trim(),
    apiKeyId: apiKeyId.trim(),
    token: token.trim(),
  };
  if (!APPLICATION_ID_PATTERN.test(detail.applicationId) ||
    !API_KEY_ID_PATTERN.test(detail.apiKeyId) ||
    !TOKEN_PATTERN.test(detail.token) ||
    !detail.token.includes(detail.apiKeyId)) {
    throw new Error("Application RAG credential handoff is invalid.");
  }
  return detail;
}

export function requestWorkflowRAGApplicationCredentialHandoff(
  applicationId: string,
  apiKeyId: string,
  token: string,
): void {
  window.dispatchEvent(new CustomEvent<WorkflowRAGApplicationCredentialHandoffDetail>(
    WORKFLOW_RAG_APPLICATION_CREDENTIAL_HANDOFF_EVENT,
    { detail: createWorkflowRAGApplicationCredentialHandoffDetail(applicationId, apiKeyId, token) },
  ));
}
