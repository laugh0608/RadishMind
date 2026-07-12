export const MODEL_GATEWAY_REQUEST_REVIEW_EVENT = "radishmind:model-gateway-request-review";

export type ModelGatewayRequestReviewEventDetail = {
  requestId: string;
};

export function requestGatewayRequestHistoryReview(requestId: string) {
  window.dispatchEvent(new CustomEvent<ModelGatewayRequestReviewEventDetail>(MODEL_GATEWAY_REQUEST_REVIEW_EVENT, {
    detail: { requestId: requestId.trim() },
  }));
}
