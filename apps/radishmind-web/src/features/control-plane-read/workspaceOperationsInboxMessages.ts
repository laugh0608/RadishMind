import type { TFunction } from "i18next";
import type { WorkspaceOperationsInboxItem } from "./workspaceOperationsInbox.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";

export function inboxItemCopy(t: TFunction<"gateway">, item: WorkspaceOperationsInboxItem) {
  const reason = item.reason;
  const values = { name: item.displayName ?? item.resourceRef, status: gatewayReviewState(t, item.resourceStatus ?? "unavailable"), code: item.failureCode ?? "none" };
  return { title: t($ => $.operationsInbox.reasons[reason].title, values), summary: t($ => $.operationsInbox.reasons[reason].summary, values) };
}
