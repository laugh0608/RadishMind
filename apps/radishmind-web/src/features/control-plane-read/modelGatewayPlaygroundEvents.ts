import type { ModelGatewayPlaygroundProtocol } from "./modelGatewayPlaygroundConsumer.ts";

export const MODEL_GATEWAY_REQUEST_REVIEW_EVENT = "radishmind:model-gateway-request-review";
export const MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT = "radishmind:model-gateway-playground-handoff";

export type ModelGatewayRequestReviewEventDetail = {
  requestId: string;
  applicationId: string;
};

export type ModelGatewayPlaygroundHandoffEventDetail = {
  applicationId: string;
  protocol: ModelGatewayPlaygroundProtocol;
  model: string;
};

export function createModelGatewayPlaygroundHandoffDetail(
  applicationId: string,
  protocol: ModelGatewayPlaygroundProtocol,
  model: string,
): ModelGatewayPlaygroundHandoffEventDetail {
  const detail = { applicationId: applicationId.trim(), protocol, model: model.trim() };
  if (!/^[A-Za-z0-9._:-]{1,160}$/u.test(detail.applicationId)) throw new Error("Playground application handoff is invalid.");
  if (!/^[A-Za-z0-9._:/-]{1,160}$/u.test(detail.model)) throw new Error("Playground model handoff is invalid.");
  if (!(detail.protocol === "chat_completions" || detail.protocol === "responses" || detail.protocol === "messages")) {
    throw new Error("Playground protocol handoff is invalid.");
  }
  return detail;
}

export function requestModelGatewayPlaygroundHandoff(
  applicationId: string,
  protocol: ModelGatewayPlaygroundProtocol,
  model: string,
) {
  window.dispatchEvent(new CustomEvent<ModelGatewayPlaygroundHandoffEventDetail>(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, {
    detail: createModelGatewayPlaygroundHandoffDetail(applicationId, protocol, model),
  }));
}

export function createModelGatewayRequestReviewDetail(
  requestId: string,
  applicationId: string,
): ModelGatewayRequestReviewEventDetail {
  const detail = { requestId: requestId.trim(), applicationId: applicationId.trim() };
  if (!/^[A-Za-z0-9._:-]{8,160}$/u.test(detail.requestId)) throw new Error("Gateway history request handoff is invalid.");
  if (!/^[A-Za-z0-9._:-]{1,160}$/u.test(detail.applicationId)) throw new Error("Gateway history application handoff is invalid.");
  return detail;
}

export function requestGatewayRequestHistoryReview(requestId: string, applicationId: string) {
  window.dispatchEvent(new CustomEvent<ModelGatewayRequestReviewEventDetail>(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, {
    detail: createModelGatewayRequestReviewDetail(requestId, applicationId),
  }));
}
