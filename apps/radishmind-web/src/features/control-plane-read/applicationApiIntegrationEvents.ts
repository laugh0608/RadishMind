import type { ApplicationApiProtocol } from "./applicationApiIntegrationConsumer.ts";

export const APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT = "radishmind:application-api-integration-draft-handoff";

export type ApplicationApiIntegrationDraftHandoffDetail = {
  applicationId: string;
  protocol: ApplicationApiProtocol;
  model: string;
};

export function createApplicationApiIntegrationDraftHandoffDetail(
  applicationId: string,
  protocol: ApplicationApiProtocol,
  model: string,
): ApplicationApiIntegrationDraftHandoffDetail {
  return { applicationId: applicationId.trim(), protocol, model: model.trim() };
}

export function requestApplicationApiIntegrationDraftHandoff(
  applicationId: string,
  protocol: ApplicationApiProtocol,
  model: string,
): void {
  window.dispatchEvent(new CustomEvent(APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT, {
    detail: createApplicationApiIntegrationDraftHandoffDetail(applicationId, protocol, model),
  }));
}
