import type { ModelGatewayPlaygroundProtocol } from "./modelGatewayPlaygroundConsumer.ts";

export const MODEL_GATEWAY_REQUEST_REVIEW_EVENT = "radishmind:model-gateway-request-review";
export const MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT = "radishmind:model-gateway-playground-handoff";

export type ModelGatewayRequestReviewEventDetail = {
  requestId: string;
  applicationId: string;
  consumerRef?: string;
};

export type ModelGatewayPlaygroundHandoffEventDetail = {
  applicationId: string;
  protocol: ModelGatewayPlaygroundProtocol;
  model: string;
  apiKeyCredential?: {
    apiKeyId: string;
    token: string;
  };
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

export function createAPIKeyModelGatewayPlaygroundHandoffDetail(
  applicationId: string,
  apiKeyId: string,
  token: string,
  model: string,
): ModelGatewayPlaygroundHandoffEventDetail {
  const detail = createModelGatewayPlaygroundHandoffDetail(applicationId, "responses", model);
  const normalizedAPIKeyId = apiKeyId.trim();
  const normalizedToken = token.trim();
  if (!/^key_[a-z2-7]{16}$/u.test(normalizedAPIKeyId) ||
    !/^rmd_dev_key_[a-z2-7]{16}\.[A-Za-z0-9_-]{43}$/u.test(normalizedToken) ||
    !normalizedToken.includes(normalizedAPIKeyId)) {
    throw new Error("Playground API key handoff is invalid.");
  }
  return { ...detail, apiKeyCredential: { apiKeyId: normalizedAPIKeyId, token: normalizedToken } };
}

export function requestAPIKeyModelGatewayPlaygroundHandoff(
  applicationId: string,
  apiKeyId: string,
  token: string,
  model: string,
) {
  window.dispatchEvent(new CustomEvent<ModelGatewayPlaygroundHandoffEventDetail>(MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT, {
    detail: createAPIKeyModelGatewayPlaygroundHandoffDetail(applicationId, apiKeyId, token, model),
  }));
}

export function createModelGatewayRequestReviewDetail(
  requestId: string,
  applicationId: string,
  consumerRef = "",
): ModelGatewayRequestReviewEventDetail {
  const detail = { requestId: requestId.trim(), applicationId: applicationId.trim(), consumerRef: consumerRef.trim() };
  if (!/^[A-Za-z0-9._:-]{8,160}$/u.test(detail.requestId)) throw new Error("Gateway history request handoff is invalid.");
  if (!/^[A-Za-z0-9._:-]{1,160}$/u.test(detail.applicationId)) throw new Error("Gateway history application handoff is invalid.");
  if (detail.consumerRef && !/^[A-Za-z0-9._:-]{1,160}$/u.test(detail.consumerRef)) {
    throw new Error("Gateway history consumer handoff is invalid.");
  }
  return detail.consumerRef
    ? detail
    : { requestId: detail.requestId, applicationId: detail.applicationId };
}

export function requestGatewayRequestHistoryReview(requestId: string, applicationId: string, consumerRef = "") {
  window.dispatchEvent(new CustomEvent<ModelGatewayRequestReviewEventDetail>(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, {
    detail: createModelGatewayRequestReviewDetail(requestId, applicationId, consumerRef),
  }));
}
