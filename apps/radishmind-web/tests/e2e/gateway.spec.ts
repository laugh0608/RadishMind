import { randomUUID } from "node:crypto";
import { test, expect, holdDraftResponse, isEndpoint, selectApplication } from "./workflow-fixtures";
import { gatewayLanguage, gatewayLayout, gatewayNavigate, historyPanel, playground, prepareGateway } from "./gateway-ui";

const protocols = [ ["chat_completions", "/v1/chat/completions"], ["responses", "/v1/responses"], ["messages", "/v1/messages"] ] as const;
// Credentials are only issued in the isolated SQLite environment. Disable automatic
// trace/screenshots in the runner, and clear all transient surfaces before teardown.
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("Gateway three protocols and API examples keep input and code invariant across languages", async ({ page, application }, info) => {
  test.setTimeout(150_000);
  const language = gatewayLanguage(page, info), copy = language.copy;
  const keyId = await prepareGateway(page, language);
  const panel = playground(page);
  const input = `Synthetic 原始请求 ${randomUUID()}`;
  let lastRequestId = "";
  for (const [protocol, path] of protocols) {
    await panel.getByRole("combobox", { name: copy().protocol, exact: true }).selectOption(protocol);
    await panel.getByRole("textbox", { name: copy().input, exact: true }).fill(input);
    for (const stream of [false, true]) {
      await panel.getByRole("checkbox", { name: copy().stream, exact: true }).setChecked(stream);
      const held = await holdDraftResponse(page, path, "POST");
      try {
        await panel.getByRole("button", { name: copy().send, exact: true }).click(); await held.arrived;
        await language.switch();
        await expect(panel.getByRole("textbox", { name: copy().input, exact: true })).toHaveValue(input);
        await expect(panel.getByRole("combobox", { name: copy().protocol, exact: true })).toHaveValue(protocol);
        const finished = page.waitForResponse(response => isEndpoint(response, path, "POST"));
        await held.deliver(); await finished;
        await expect(panel.locator(".gateway-playground-result .status-badge").first()).toHaveText(copy().success);
        await expect(panel.locator(".gateway-playground-output")).not.toBeEmpty();
        lastRequestId = (await panel.locator(".gateway-playground-result h4").innerText()).trim();
      } finally { await held.dispose(); }
    }
  }
  await gatewayLayout(page, panel, info, "gateway-result");
  await panel.getByRole("button", { name: copy().history, exact: true }).click();
  const detail = historyPanel(page).locator(".gateway-request-history-detail");
  await expect(detail).toContainText(lastRequestId); await expect(detail).toContainText(`api_key:${keyId}`);
  await expect(detail).not.toContainText(input); await expect(detail).toContainText(application.id);
  await language.switch(); await expect(detail).toContainText(lastRequestId);
  await gatewayLayout(page, historyPanel(page), info, "gateway-history");
  await prepareGateway(page, language);
  await gatewayNavigate(page, "application-api-integration");
  const integration = page.locator("#application-api-integration");
  for (const [protocol] of protocols) {
    await integration.getByRole("combobox", { name: copy().protocol, exact: true }).selectOption(protocol);
    for (const example of ["curl", "python", "typescript"]) {
      await integration.getByRole("combobox", { name: copy().example, exact: true }).selectOption(example);
      const code = await integration.getByLabel(copy().code).textContent();
      expect(code).toContain("RADISHMIND_API_KEY"); expect(code).not.toContain(input);
      await language.switch();
      await expect(integration.getByLabel(copy().code)).toHaveText(code!);
    }
  }
  await gatewayLayout(page, integration, info, "gateway-api-examples");
  await gatewayNavigate(page, "model-gateway-playground");
  await expect(panel.getByRole("button", { name: copy().send, exact: true })).toBeDisabled();
  const stored = await page.evaluate(() => JSON.stringify({ local: { ...localStorage }, session: { ...sessionStorage } }));
  expect(stored).not.toContain(input); expect(stored).not.toMatch(/rmd_dev_/);
});

test("Gateway validation, cancel, exact filters, pagination and read failures preserve their state", async ({ page, application: _application }, info) => {
  test.setTimeout(180_000);
  const language = gatewayLanguage(page, info), copy = language.copy;
  await prepareGateway(page, language); const panel = playground(page);
  await panel.getByRole("combobox", { name: copy().protocol, exact: true }).selectOption("chat_completions");
  await panel.getByRole("textbox", { name: copy().input, exact: true }).fill("   ");
  await panel.getByRole("button", { name: copy().send, exact: true }).click();
  await expect(panel).toContainText(copy().invalid); await language.switch(); await expect(panel).toContainText(copy().invalid);
  const input = "Synthetic cancel input 保留";
  await panel.getByRole("textbox", { name: copy().input, exact: true }).fill(input);
  const held = await holdDraftResponse(page, "/v1/chat/completions", "POST");
  try {
    await panel.getByRole("button", { name: copy().send, exact: true }).click(); await held.arrived;
    await language.switch(); await panel.getByRole("button", { name: copy().cancel, exact: true }).click();
    await expect(panel.locator(".gateway-playground-result .status-badge").first()).toHaveText(copy().canceled);
    await expect(panel.getByRole("textbox", { name: copy().input, exact: true })).toHaveValue(input);
  } finally { await held.dispose(true); }
  const completed = page.waitForResponse(response => isEndpoint(response, "/v1/chat/completions", "POST"));
  await panel.getByRole("button", { name: copy().send, exact: true }).click(); const response = await completed;
  await expect(panel.locator(".gateway-playground-result .status-badge").first()).toHaveText(copy().success);
  const lastId = (await panel.locator(".gateway-playground-result h4").innerText()).trim();
  // Seed earlier records via the same public Gateway route, never by writing SQLite.
  // Headers stay in memory; trace is disabled and only paths/statuses are attached.
  const headers = await response.request().allHeaders(), data = response.request().postDataJSON();
  for (let i = 0; i < 25; i++) {
    const seeded = await page.request.post(response.url(), { headers: { ...headers, "x-request-id": `gateway-page-${randomUUID()}` }, data });
    expect(seeded.status()).toBe(200);
  }
  await panel.getByRole("button", { name: copy().history, exact: true }).click(); const history = historyPanel(page);
  await expect(history.locator(".gateway-request-history-detail")).toContainText(lastId);
  await expect(history.locator(".gateway-request-history-row")).toHaveCount(25);
  await language.switch(); await history.getByRole("button", { name: copy().earlier, exact: true }).click();
  await expect(history.locator(".gateway-request-history-row")).toHaveCount(28);
  await history.locator("summary").filter({ hasText: copy().filters }).click();
  await history.getByRole("textbox", { name: copy().route, exact: true }).fill("/v1/responses");
  await language.switch(); await expect(history.getByRole("textbox", { name: copy().route, exact: true })).toHaveValue("/v1/responses");
  await history.getByRole("button", { name: copy().apply, exact: true }).click(); await expect(history).toContainText(copy().empty);
  await history.getByRole("textbox", { name: copy().route, exact: true }).fill("");
  await page.route("**/v1/model-gateway/requests?*", route => route.abort("failed"));
  await history.getByRole("button", { name: copy().apply, exact: true }).click(); await expect(history).toContainText(copy().failure);
  await language.switch(); await expect(history).toContainText(copy().failure);
  await gatewayLayout(page, history, info, "gateway-filters-failure");
  await page.unroute("**/v1/model-gateway/requests?*");
  await history.getByRole("button", { name: copy().apply, exact: true }).click(); await expect(history.locator(".gateway-request-history-row")).toHaveCount(25);
});

test("Operations source coverage, inbox selection and quota metadata retranslate without writes", async ({ page, application }, info) => {
  const language = gatewayLanguage(page, info), copy = language.copy;
  await prepareGateway(page, language);
  await page.reload(); await selectApplication(page, application);
  language.requests.length = 0;
  await gatewayNavigate(page, "workspace-overview");
  await expect(page.locator("#workspace-overview")).toContainText(copy().overview); await language.switch();
  await gatewayLayout(page, page.locator("#workspace-overview"), info, "operations-overview");
  await gatewayNavigate(page, "workspace-operations-inbox"); const inbox = page.locator("#workspace-operations-inbox");
  await expect(inbox).toContainText(copy().inbox);
  await inbox.locator(".workspace-operations-inbox-item").last().click();
  const expiringKey = await inbox.locator(".workspace-attention-resource strong").innerText();
  expect(expiringKey).toMatch(/^key_/);
  await language.switch(); await expect(inbox.locator(".workspace-attention-resource strong")).toHaveText(expiringKey);
  await gatewayLayout(page, inbox, info, "operations-inbox");
  await gatewayNavigate(page, "workspace-usage-quota"); const quota = page.locator("#workspace-usage-quota");
  await expect(quota).toContainText(copy().quota); await expect(quota).toContainText("quota_policy_missing");
  await expect(quota.locator(".usage-quota-limit")).toHaveCount(0);
  await language.switch(); await expect(quota).toContainText("quota_policy_missing");
  await gatewayLayout(page, quota, info, "operations-usage");
  // Open the existing evidence task through its user-facing rail.
  await gatewayNavigate(page, "model-gateway-playground");
  await gatewayNavigate(page, "application-operations"); const operations = page.locator("#application-operations");
  await expect(operations).toContainText(application.id); await expect(operations).toContainText(copy().operations);
  await expect(operations.getByRole("button", { name: copy().refresh, exact: true })).toBeEnabled();
  await page.route("**/v1/model-gateway/requests?*", route => route.abort("failed"));
  await operations.getByRole("button", { name: copy().refresh, exact: true }).click();
  await expect(operations.locator(".application-operations-actions .status-badge")).toHaveText(copy().partial);
  await language.switch(); await expect(operations).toContainText(copy().partial);
  await gatewayLayout(page, operations, info, "operations-partial");
  await page.unroute("**/v1/model-gateway/requests?*");
  await operations.getByRole("button", { name: copy().refresh, exact: true }).click();
  await expect(operations).not.toContainText("application_operations_gateway_unavailable");
  expect(language.requests.filter(request => request.startsWith("POST ") || request.startsWith("PATCH ") || request.startsWith("DELETE "))).toEqual([]);
  // The live quota owner intentionally returns unavailable. Exercise populated
  // display separately with an explicit synthetic read envelope, without enabling writes.
  await page.route("**/v1/user-workspace/usage/quota-summary", async route => {
    const response = await route.fetch(); const envelope = await response.json();
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({
      ...envelope, failure_code: null, next_cursor: null,
      items: [{ quota_id: "quota_synthetic_display", tenant_ref: "tenant_demo", period: "monthly", request_limit: 10000, token_limit: 1000000, cost_limit: 100, usage_snapshot: { request_count: 12, token_count: 3456, estimated_cost: 0.42 }, over_quota_failure_code: "quota_exceeded" }],
    }) });
  });
  await page.reload(); await gatewayNavigate(page, "workspace-usage-quota");
  await page.waitForLoadState("networkidle");
  await expect(quota).toContainText("quota_synthetic_display"); await expect(quota).toContainText(copy().missingCurrency);
  await language.switch(); await expect(quota).toContainText(copy().missingCurrency);
  await expect(quota).not.toContainText("$"); await expect(quota).not.toContainText("USD");
  await gatewayLayout(page, quota, info, "operations-quota-synthetic");
  expect(language.requests.filter(request => /^(POST|PATCH|DELETE) /.test(request))).toEqual([]);
});
