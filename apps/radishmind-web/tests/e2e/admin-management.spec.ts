import { providerEn, providerZh, quotaEn, quotaZh, pricingEn, pricingZh, readEn, readZh } from "./admin-ui-language";
import { randomUUID } from "node:crypto";
import type { Locator, Page, TestInfo } from "@playwright/test";
import { test, expect, holdDraftResponse, isEndpoint } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";

function language(page: Page, info: TestInfo) {
  let locale = info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  return {
    provider: () => locale === "zh-CN" ? providerZh : providerEn,
    quota: () => locale === "zh-CN" ? quotaZh : quotaEn,
    pricing: () => locale === "zh-CN" ? pricingZh : pricingEn,
    read: () => locale === "zh-CN" ? readZh : readEn,
    requests,
    async switch() {
      const before = [...requests];
      locale = locale === "zh-CN" ? "en-US" : "zh-CN";
      await page.getByRole("combobox", { name: /Interface language|界面语言/ }).selectOption(locale);
      setTestLanguage(page, locale as "en-US" | "zh-CN");
      await expect(page.locator("html")).toHaveAttribute("lang", locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests, "Language changes must not trigger business requests").toEqual(before);
    },
  };
}

async function open(page: Page, anchor: string) {
  await page.evaluate(value => { location.hash = value; }, anchor);
}

async function layouts(page: Page, surface: Locator, info: TestInfo, label: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await surface.scrollIntoViewIfNeeded();
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
    const clipped = await surface.locator("input, textarea, select, button").evaluateAll(elements => elements.filter(element => {
      const rect = element.getBoundingClientRect();
      return rect.width > 0 && (rect.left < -1 || rect.right > document.documentElement.clientWidth + 1);
    }).length);
    expect(clipped, "All fields and actions fit the viewport").toBe(0);
    const overflowingHistory = await surface.locator(".admin-provider-route-history-list article > p").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).length);
    expect(overflowingHistory, "Exact candidate IDs and reasons must wrap within history cards").toBe(0);
    // Exclude the fixed shell from a full component capture. Layout assertions
    // above still run with the real shell visible and unchanged.
    if (info.repeatEachIndex === 0) await info.attach(`${label}-${width}`, { body: await surface.screenshot({ style: ".product-nav { visibility: hidden !important; }" }), contentType: "image/png" });
    if (info.repeatEachIndex === 0 && label === "provider-lineage") await info.attach(`provider-history-${width}`, { body: await surface.locator(".admin-provider-route-history-list article").first().screenshot({ style: ".product-nav { visibility: hidden !important; }" }), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}

test("Provider draft, pending review, activation and rollback preserve exact lineage across languages", async ({ page, application }, info) => {
  expect(application.id).toMatch(/^app_/);
  const ui = language(page, info);
  await open(page, "admin-route-config");
  const surface = page.locator(".admin-provider-route-workspace");
  const button = (name: string) => surface.getByRole("button", { name, exact: true });
  const field = (name: string) => surface.getByRole("textbox", { name, exact: true });
  const operation = surface.locator(".admin-provider-route-operation");
  await expect(operation).toContainText(ui.provider().operations.workspaceLoaded);
  const base = "/v1/admin/provider-route-configurations/gateway-default";
  const display = `Route review ${randomUUID().slice(0, 8)} 原文`;
  await field(ui.provider().displayName).first().fill(display);
  await surface.getByRole("textbox", { name: /^(Provider ID|模型提供方 ID)$/ }).fill("openai-compatible");
  await surface.getByRole("textbox", { name: /^(Runtime profile ref|运行时配置档案引用)$/ }).fill("ref:radishmind/test/provider-profiles/radishmind-default-workflow");
  await surface.getByRole("textbox", { name: /^(Requested model|请求模型)$/ }).fill("mock-workflow-model");
  await ui.switch();
  await expect(field(ui.provider().displayName).first()).toHaveValue(display);
  const saved = page.waitForResponse(response => isEndpoint(response, base, "PUT"));
  await button(ui.provider().saveDraft).click();
  const draftBody = await (await saved).json();
  expect(draftBody.failure_code).toBeNull();
  await expect(operation).toContainText(ui.provider().operations.draftSaved);
  const candidate = `candidate_${randomUUID().replaceAll("-", "")}`;
  await field(ui.provider().candidateId).fill(candidate);
  const created = page.waitForResponse(response => isEndpoint(response, `${base}/candidates`, "POST"));
  await surface.getByRole("button", { name: /Create from revision|从修订.*创建/ }).click();
  expect((await created).ok()).toBeTruthy();
  await expect(operation).toContainText(ui.provider().operations.candidateCreated);
  const reason = "Reviewed exact metadata 保留原文 <literal>";
  await field(ui.provider().reviewReason).fill(reason);
  const held = await holdDraftResponse(page, `${base}/candidates/${candidate}/reviews`, "POST");
  try {
    await surface.getByRole("button", { name: /Record review|记录审查/ }).click();
    await held.arrived;
    await ui.switch();
    await expect(field(ui.provider().reviewReason)).toHaveValue(reason);
    await expect(surface.getByRole("button", { name: /Record review|记录审查/ })).toBeDisabled();
    await held.deliver();
    await expect(operation).toContainText(ui.provider().operations.reviewRecorded);
  } finally { await held.dispose(); }
  const activation = page.waitForResponse(response => isEndpoint(response, `${base}/candidates/${candidate}/activations`, "POST"));
  await surface.getByRole("button", { name: /Commit activation|提交激活/ }).click();
  const activated = await (await activation).json();
  expect(activated.failure_code).toBeNull();
  await expect(operation).toContainText(ui.provider().operations.activated);
  const generation = activated.snapshot.generation as number;
  // Another explicit activation makes the original candidate a historical rollback target.
  const nextCandidate = `${candidate}_next`;
  await field(ui.provider().candidateId).fill(nextCandidate);
  await surface.getByRole("button", { name: /Create from revision|从修订.*创建/ }).click();
  await expect(operation).toContainText(ui.provider().operations.candidateCreated);
  await surface.getByRole("button", { name: /Record review|记录审查/ }).click();
  await expect(operation).toContainText(ui.provider().operations.reviewRecorded);
  await surface.getByRole("button", { name: /Commit activation|提交激活/ }).click();
  await expect(operation).toContainText(ui.provider().operations.activated);
  // Locate the exact candidate in history without matching its suffixed successor.
  await field(ui.provider().candidateId).fill(candidate);
  await button(ui.provider().loadCandidate).click();
  await expect(operation).toContainText(candidate);
  await surface.getByRole("combobox", { name: ui.provider().action, exact: true }).selectOption("rollback");
  await field(ui.provider().activationReason).fill(reason);
  await ui.switch();
  await expect(field(ui.provider().candidateId)).toHaveValue(candidate);
  await expect(field(ui.provider().activationReason)).toHaveValue(reason);
  const rollback = page.waitForResponse(response => isEndpoint(response, `${base}/candidates/${candidate}/activations`, "POST"));
  await surface.getByRole("button", { name: /Commit rollback|提交回滚/ }).click();
  const rolledBack = await (await rollback).json();
  expect(rolledBack.failure_code).toBeNull();
  expect(rolledBack.snapshot.generation).toBe(generation + 2);
  expect(rolledBack.snapshot.candidate_id).toBe(candidate);
  await expect(operation).toContainText(ui.provider().operations.rolledBack);
  await layouts(page, surface, info, "provider-lineage");
  expect(ui.requests.filter(value => /chat\/completions|gateway-requests/.test(value))).toHaveLength(0);
});

test("Quota confirmation survives language changes and rejects a real stale version", async ({ page, application, request }, info) => {
  const ui = language(page, info);
  await open(page, "admin-gateway-request-quota");
  const surface = page.locator(".admin-gateway-quota-workspace");
  const button = (name: string) => surface.getByRole("button", { name, exact: true });
  await expect(surface).toContainText(ui.quota().missingTitle);
  const limit = () => surface.getByRole("spinbutton", { name: ui.quota().requestLimit, exact: true });
  await limit().fill("1234");
  await button(ui.quota().reviewUpdate).click();
  await ui.switch();
  await expect(limit()).toHaveValue("1234");
  await expect(surface.getByRole("region", { name: ui.quota().confirmationLabel })).toContainText("1,234");
  await layouts(page, surface, info, "quota-confirmation");
  const path = `/v1/admin/gateway-request-quotas/${application.id}`;
  const put = page.waitForRequest(value => new URL(value.url()).pathname === path && value.method() === "PUT");
  const response = page.waitForResponse(value => isEndpoint(value, path, "PUT"));
  await button(ui.quota().confirmUpdate).click();
  const initialRequest = await put;
  const initial = await (await response).json();
  expect(initial.policy.request_limit).toBe(1234);
  expect(initial.policy.record_version).toBe(1);
  await expect(surface.locator(".admin-gateway-quota-version")).toContainText("1");
  await limit().fill("2345");
  await button(ui.quota().reviewUpdate).click();
  const competing = await request.put(initialRequest.url(), { headers: initialRequest.headers(), data: { expected_version: 1, request_limit: 3456 } });
  expect(competing.ok()).toBeTruthy();
  const conflict = page.waitForResponse(value => isEndpoint(value, path, "PUT"));
  await button(ui.quota().confirmUpdate).click();
  expect((await conflict).status()).toBe(409);
  await expect(surface).toContainText(ui.quota().operations.conflict);
  await ui.switch();
  await expect(surface).toContainText(ui.quota().operations.conflict);
  const reloaded = page.waitForResponse(value => isEndpoint(value, path, "GET"));
  await button(ui.quota().reload).click();
  const current = await (await reloaded).json();
  expect(current.policy.record_version).toBe(2);
  expect(current.usage.admitted_request_count).toBe(initial.usage.admitted_request_count);
  await expect(limit()).toHaveValue("3456");
});

test("Pricing keeps exact micro-USD, reason and confirmation while handling scope changes and CAS", async ({ page, application, request }, info) => {
  expect(application.id).toMatch(/^app_/);
  const ui = language(page, info);
  await open(page, "admin-gateway-model-pricing");
  const surface = page.locator(".admin-gateway-pricing-workspace");
  const field = (name: string) => surface.getByRole("textbox", { name, exact: true });
  const button = (name: string) => surface.getByRole("button", { name, exact: true });
  await field(ui.pricing().providerId).fill("openai-compatible");
  await field(ui.pricing().profileId).fill("radishmind-default-workflow");
  const model = `price_${randomUUID().slice(0, 8)}`;
  await field(ui.pricing().modelId).fill(model);
  await button(ui.pricing().loadOwner).click();
  await expect(surface).toContainText(ui.pricing().missingTitle);
  await field(ui.pricing().inputRate).fill("1");
  await field(ui.pricing().outputRate).fill("123456789");
  const reason = "Reviewed exact prices 原始理由";
  await field(ui.pricing().sanitizedReason).fill(reason);
  await button(ui.pricing().reviewRevision).click();
  await ui.switch();
  await expect(field(ui.pricing().inputRate)).toHaveValue("1");
  await expect(field(ui.pricing().sanitizedReason)).toHaveValue(reason);
  const confirmation = surface.locator(".admin-gateway-pricing-confirm");
  await expect(confirmation).toContainText("USD 0.000001");
  await expect(confirmation).toContainText("USD 123.456789");
  await layouts(page, surface, info, "pricing-confirmation");
  const path = "/v1/admin/gateway-model-pricing-policy";
  const put = page.waitForRequest(value => new URL(value.url()).pathname === path && value.method() === "PUT");
  const updated = page.waitForResponse(value => isEndpoint(value, path, "PUT"));
  await button(ui.pricing().confirm).click();
  const initialRequest = await put;
  const first = await (await updated).json();
  expect(first.policy.input_price_micros_per_token_unit).toBe(1);
  expect(first.policy.output_price_micros_per_token_unit).toBe(123456789);
  await expect(surface.locator(".admin-gateway-pricing-evidence")).toContainText(reason);
  await field(ui.pricing().inputRate).fill("2");
  await button(ui.pricing().reviewRevision).click();
  const competitor = await request.put(initialRequest.url(), { headers: initialRequest.headers(), data: { ...initialRequest.postDataJSON(), expected_version: 1, input_price_micros_per_token_unit: 3 } });
  expect(competitor.ok()).toBeTruthy();
  const conflict = page.waitForResponse(value => isEndpoint(value, path, "PUT"));
  await button(ui.pricing().confirm).click();
  expect((await conflict).status()).toBe(409);
  await expect(surface.getByRole("alert")).toContainText("gateway_pricing_policy_version_conflict");
  await ui.switch();
  await expect(surface.getByRole("alert")).toContainText("2");
  await button(ui.pricing().reload).click();
  await expect(field(ui.pricing().inputRate)).toHaveValue("3");
  await field(ui.pricing().modelId).fill(`${model}_other`);
  await expect(surface.getByRole("alert")).toHaveText(ui.pricing().scopeChanged);
  await expect(button(ui.pricing().reviewRevision)).toHaveCount(0);
  await ui.switch();
  await expect(field(ui.pricing().modelId)).toHaveValue(`${model}_other`);
  await expect(surface.getByRole("alert")).toHaveText(ui.pricing().scopeChanged);
});

test("Tenant and audit summaries retain selection, raw references and UTC timestamps", async ({ page, application }, info) => {
  expect(application.id).toMatch(/^app_/);
  const ui = language(page, info);
  await open(page, "admin-tenant-overview");
  await expect(page.locator(".admin-control-tenant-owner")).toContainText(ui.read().tenantBoundary);
  await open(page, "admin-audit-log");
  const surface = page.locator(".admin-control-audit-owner");
  await expect(surface).toContainText(ui.read().auditTitle);
  const rows = surface.locator(".admin-control-audit-list > button");
  await expect(rows.first()).toBeVisible();
  await rows.last().click();
  const detail = surface.locator(".admin-control-audit-detail");
  const reference = await detail.locator("h5").innerText();
  const timestamp = await detail.locator("time").getAttribute("datetime");
  await ui.switch();
  await expect(detail.locator("h5")).toHaveText(reference);
  await expect(detail.locator("time")).toHaveAttribute("datetime", timestamp!);
  await expect(detail.locator("time")).toContainText("UTC");
  await expect(detail.locator(".admin-control-readonly-actions")).toContainText(ui.read().deleteBlocked);
  await expect(detail.getByRole("button")).toHaveCount(0);
  await layouts(page, surface, info, "audit-detail");
});
