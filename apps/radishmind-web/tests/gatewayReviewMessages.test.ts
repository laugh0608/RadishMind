import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { formatMicroUSD } from "../src/i18n/formatters.ts";
import { gatewayReview as en } from "../src/i18n/locales/en-US/gatewayReview.ts";
import { gatewayReview as zh } from "../src/i18n/locales/zh-CN/gatewayReview.ts";
import { operationsInbox as inboxEn } from "../src/i18n/locales/en-US/operationsInbox.ts";
import { operationsInbox as inboxZh } from "../src/i18n/locales/zh-CN/operationsInbox.ts";
import { gatewayReviewState, gatewayHistoryFailure } from "../src/features/control-plane-read/gatewayReviewMessages.ts";
import { inboxItemCopy } from "../src/features/control-plane-read/workspaceOperationsInboxMessages.ts";
import type { WorkspaceOperationsInboxItem } from "../src/features/control-plane-read/workspaceOperationsInbox.ts";

test("Gateway money preserves micro-units and refuses uncertain numeric precision", () => {
  for (const locale of ["en-US", "zh-CN"] as const) {
    assert.equal(formatMicroUSD(0, locale), "USD 0.000000");
    assert.equal(formatMicroUSD(17, locale), "USD 0.000017");
    assert.equal(formatMicroUSD(Number.MAX_SAFE_INTEGER, locale), "USD 9,007,199,254.740991");
    assert.equal(formatMicroUSD(9007199254740993n, locale), "USD 9,007,199,254.740993");
    for (const value of [-1, NaN, Infinity, .5, Number.MAX_SAFE_INTEGER + 1, -1n]) assert.equal(formatMicroUSD(value, locale), null);
  }
});

test("Gateway messages retranslate stable diagnostics and preserve unknown codes", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "gateway", { review: en, operationsInbox: inboxEn });
  instance.addResourceBundle("zh-CN", "gateway", { review: zh, operationsInbox: inboxZh });
  const item: WorkspaceOperationsInboxItem = { itemId: "synthetic-item", sourceId: "applications", reason: "application_run_attention", severity: "high", title: "ignored materialized English title", summary: "ignored materialized English summary", displayName: "原始 App Name", resourceStatus: "failed", resourceRef: "app_exact", applicationRef: "app_exact", workflowDefinitionId: "", runId: "", targetAnchor: "#workspace-applications", occurredAt: "2026-10-07T00:00:00Z", requestId: "req_exact", auditRef: "audit_exact" };
  const original = structuredClone(item);
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale); const t = instance.getFixedT(locale, "gateway");
    assert.equal(gatewayReviewState(t, "failed"), locale === "en-US" ? "Failed" : "失败");
    assert.equal(gatewayReviewState(t, "future_owner_state"), "future_owner_state");
    const known = gatewayHistoryFailure(t, "gateway_request_cursor_invalid");
    assert.match(known, locale === "en-US" ? /cursor/ : /游标/);
    assert.equal(gatewayHistoryFailure(t, "future_code"), locale === "en-US" ? en.failures.unknown : zh.failures.unknown);
    const copy = inboxItemCopy(t, item);
    assert.ok(copy.title.includes("原始 App Name")); assert.ok(!copy.title.includes("ignored"));
    assert.ok(!copy.summary.includes("ignored"));
    assert.deepEqual(item, original);
    const run = inboxItemCopy(t, { ...item, reason: "run_failure", failureCode: "synthetic_exact_failure" });
    assert.ok(run.summary.includes("synthetic_exact_failure"));
  }
});
