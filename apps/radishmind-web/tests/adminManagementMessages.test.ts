import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { adminProvider as providerEn } from "../src/i18n/locales/en-US/adminProvider.ts";
import { adminProvider as providerZh } from "../src/i18n/locales/zh-CN/adminProvider.ts";
import { adminQuota as quotaEn } from "../src/i18n/locales/en-US/adminQuota.ts";
import { adminQuota as quotaZh } from "../src/i18n/locales/zh-CN/adminQuota.ts";
import { adminPricing as pricingEn } from "../src/i18n/locales/en-US/adminPricing.ts";
import { adminPricing as pricingZh } from "../src/i18n/locales/zh-CN/adminPricing.ts";
import { adminDisplayDate, adminDisplayRate } from "../src/features/control-plane-read/adminManagementFormatters.ts";
import { providerOperationMessage, providerFailureMessage, providerFindingMessage, type ProviderMessage } from "../src/features/control-plane-read/adminProviderRouteMessages.ts";
import { quotaNoticeMessage, pricingNoticeMessage, pricingFailureMessage, type QuotaNotice, type PricingNoticeMessage } from "../src/features/control-plane-read/adminGatewayManagementMessages.ts";

async function messages() {
  const instance = createUiI18n();
  await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "admin", { provider: providerEn, quota: quotaEn, pricing: pricingEn }, true);
  instance.addResourceBundle("zh-CN", "admin", { provider: providerZh, quota: quotaZh, pricing: pricingZh }, true);
  return instance;
}

test("management notices retranslate pending and completed operations while retaining exact CAS references", async () => {
  const instance = await messages();
  const provider: ProviderMessage[] = [{ key: "candidateLoaded", candidateId: "candidate_exact" }, { key: "activating", generation: 123 }, { key: "recordingReview", decision: "approve" }];
  const quotas: QuotaNotice[] = [{ key: "updated", version: 0 }, { key: "updated", version: null }, { key: "conflict" }];
  const prices: PricingNoticeMessage[] = [{ key: "updated", version: 123 }, { key: "conflict", version: 124 }];
  const before = structuredClone({ provider, quotas, prices });
  const copies: string[][] = [];
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(null, "admin");
    const rendered = [...provider.map(value => providerOperationMessage(t, value, "")), ...quotas.map(value => quotaNoticeMessage(t, value)), ...prices.map(value => pricingNoticeMessage(t, value))];
    assert.match(rendered[0], /candidate_exact/);
    assert.match(rendered[1], /123/);
    assert.match(rendered[3], /0/);
    assert.match(rendered[6], /123/);
    assert.match(rendered[7], /124/);
    assert.doesNotMatch(rendered.join(" "), /\{\{|operations\./);
    copies.push(rendered);
  }
  copies[0].forEach((value, index) => assert.notEqual(value, copies[1][index]));
  assert.deepEqual({ provider, quotas, prices }, before);
});

test("all management failure codes and unknown diagnostics use local guidance", async () => {
  const instance = await messages();
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(null, "admin");
    const copies = [
      ...Object.keys(providerEn.failures).map(code => providerFailureMessage(t, code)),
      ...Object.keys(pricingEn.failures).map(code => pricingFailureMessage(t, code)),
      ...Object.keys(quotaEn.failures).map(code => quotaNoticeMessage(t, { key: "failure", code: code as keyof typeof quotaEn.failures })),
      ...["PRIVATE_DIAGNOSTIC", "__proto__", "toString"].flatMap(code => [providerFailureMessage(t, code), pricingFailureMessage(t, code)]),
    ];
    assert.doesNotMatch(copies.join(" "), /PRIVATE_DIAGNOSTIC|__proto__|toString|\{\{|failures\./);
    if (locale === "zh-CN") copies.forEach(copy => assert.match(copy, /[\u4e00-\u9fff]/u));
    assert.match(providerFindingMessage(t, { field: "route", kind: "duplicateBinding", protocol: "responses", modelId: "model_exact", summary: "PRIVATE_DIAGNOSTIC" }), /responses.*model_exact/);
  }
});

test("rate display preserves every micro-USD unit including zero and safe large integers", () => {
  for (const locale of ["en-US", "zh-CN"] as const) {
    for (const value of [0, 1, 999999, 1000000, 1234567890123, Number.MAX_SAFE_INTEGER]) {
      const display = adminDisplayRate(value, locale);
      const digits = display.slice(4).replaceAll(",", "").replace(".", "");
      assert.equal(BigInt(digits), BigInt(value));
      assert.match(display, /^USD [\d,]+\.\d{6}$/);
    }
    assert.match(adminDisplayDate("2026-10-07T23:59:00Z", locale), locale === "en-US" ? /Oct 7, 2026.*11:59 PM UTC/ : /2026年10月7日.*23:59 UTC/);
    assert.equal(adminDisplayDate("invalid_timestamp", locale), "invalid_timestamp");
  }
});
