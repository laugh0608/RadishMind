import { randomUUID } from "node:crypto";
import type { Locator, Page, TestInfo } from "@playwright/test";
import { test, expect, isEndpoint } from "./workflow-fixtures";
import { uiText } from "./ui-language";
import { prepareApplication, session, sessionPath, templatePath, navigate, submit } from "./prompt-fixtures";

test.use({ applicationKind: "prompt_application", trace: "off", screenshot: "off" });
const input = (page: Page) => page.locator(".prompt-variable-input");
const copy = (info: TestInfo, en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
async function fillInputs(page: Page) {
  await input(page).getByRole("button", { name: uiText(page, "Read input template"), exact: true }).click();
  await expect(input(page).locator("[data-reference-state]")).toHaveAttribute("data-reference-state", "ready");
  for (const [name, value] of [["question", `E2E_CASE:${randomUUID()}:valid`], ["tone", "E2E_PROMPT_SYSTEM_V1"]]) {
    const field = input(page).locator(`[data-variable="${name}"]`);
    await field.getByRole("combobox").first().selectOption("provided");
    await field.getByRole("textbox", { name, exact: true }).fill(value);
  }
}
async function layout(page: Page, panel: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 390]) {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
    expect(await panel.evaluate(element => ({
      viewport: document.documentElement.scrollWidth <= window.innerWidth,
      controls: [...element.querySelectorAll("input, textarea, button, select")].every(field => {
        const a = field.getBoundingClientRect(), b = element.getBoundingClientRect();
        return a.left >= b.left - 1 && a.right <= b.right + 1;
      }),
    }))).toEqual({ viewport: true, controls: true });
    expect(await panel.locator(".prompt-session-controls > button, .prompt-turn-controls > button, .prompt-execution-actions > button").evaluateAll(buttons =>
      buttons.length > 0 && buttons.every(button => button.getBoundingClientRect().height >= 36)
    )).toBe(true);
    await page.evaluate(() => window.scrollTo(0, 0));
    const clip = await panel.boundingBox();
    await page.screenshot({ path: info.outputPath(`${name}-${width}.png`), fullPage: true, clip: clip! });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}

test("Prompt recovery preserves save intent and input, reloads exact metadata and explicitly prepares new turns", async ({ page, application }, info) => {
  const { sessionId } = await prepareApplication(page, application);
  const panel = session(page), key = panel.getByRole("textbox", { name: "client_turn_key", exact: true });
  const execute = panel.getByRole("button", { name: uiText(page, "Execute Prompt turn"), exact: true });
  const save = panel.locator(".application-result-artifact-opt-in input");
  const fresh = panel.getByRole("button", { name: copy(info, "Prepare new turn", "准备新轮次"), exact: true });
  const refresh = panel.getByRole("button", { name: copy(info, "Reload selected Session metadata", "重读所选会话元数据"), exact: true });
  await fillInputs(page);
  const originalKey = await key.inputValue();
  const originalInput = await input(page).getByRole("textbox", { name: "question", exact: true }).inputValue();
  let turns = 0;
  page.on("request", req => { if (req.method() === "POST" && new URL(req.url()).pathname.endsWith("/turns")) turns++; });
  await save.check();
  await key.fill("bad");
  await expect(execute).toBeDisabled();
  await expect(save).toBeChecked();
  await key.fill(originalKey);
  const sourceMatcher = (url: URL) => url.pathname.startsWith(templatePath) && /\/versions\/\d+$/.test(url.pathname);
  await page.route(sourceMatcher, route => route.fulfill({ status: 503, json: {} }));
  await execute.click();
  await expect(input(page).locator("[data-reference-state]")).toHaveAttribute("data-reference-state", "unavailable");
  await expect(save).toBeChecked();
  expect(turns).toBe(0);
  await page.unroute(sourceMatcher);
  await fillInputs(page);
  // Save intent is consumed at dispatch, even when delivery fails.
  const turnMatcher = (url: URL) => url.pathname === `${sessionPath}/${sessionId}/turns`;
  await page.route(turnMatcher, route => route.abort("connectionfailed"));
  const retained = await input(page).getByRole("textbox", { name: "question", exact: true }).inputValue();
  await execute.click();
  await expect(panel.locator(".failure-summary")).toContainText("application_session_store_unavailable");
  await expect(save).not.toBeChecked();
  await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveValue(retained);
  await expect(key).toHaveValue(originalKey);
  await expect(key).toBeDisabled();
  await expect(execute).toBeDisabled();
  await expect(fresh).toBeDisabled();
  await layout(page, panel, info, "session-interrupted");
  await page.unroute(turnMatcher);
  const metadata = page.waitForResponse(response => isEndpoint(response, `${sessionPath}/${sessionId}`, "GET"));
  await refresh.click(); await metadata;
  await expect(execute).toBeEnabled();
  await expect(fresh).toBeDisabled(); // A metadata read does not resolve the original attempt.
  await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveValue(retained);
  const executed = await submit(page, execute, `${sessionPath}/${sessionId}/turns`);
  expect(executed.turn.status).toBe("succeeded");
  const replay = await submit(page, execute, `${sessionPath}/${sessionId}/turns`);
  expect(replay.idempotent_replay).toBe(true);
  expect(replay.prompt_output ?? "").toBe("");
  await save.check();
  const before = turns;
  await fresh.click();
  await expect(key).not.toHaveValue(originalKey);
  await expect(save).not.toBeChecked();
  expect(turns).toBe(before);
  await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveValue(retained);
  await input(page).getByRole("textbox", { name: "question", exact: true }).fill(originalInput);
  expect((await submit(page, execute, `${sessionPath}/${sessionId}/turns`)).turn.sequence).toBe(2);
  await layout(page, panel, info, "session-new-turn");
  // The server completes this request; cancellation only discards the delayed browser response.
  await fresh.click();
  const pendingKey = await key.inputValue();
  let release!: () => void, arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const reached = new Promise<void>(resolve => { arrived = resolve; });
  let delivered!: () => void;
  const delivery = new Promise<void>(resolve => { delivered = resolve; });
  await page.route(turnMatcher, async route => {
    const response = await route.fetch();
    expect((await response.json()).turn.sequence).toBe(3);
    arrived(); await gate;
    try {
      await route.fulfill({ response });
    } catch (error) {
      // An already aborted intercepted request may reject late delivery. Any
      // other routing error still fails the test; the server result is checked below.
      expect(error instanceof Error ? error.message : String(error)).toBe("route.fulfill: Route is already handled!");
    } finally { delivered(); }
  });
  try {
    await execute.click(); await reached;
    await panel.getByRole("button", { name: copy(info, "Cancel current request", "取消当前请求"), exact: true }).click();
    await expect(panel.locator(".failure-summary")).toContainText("application_session_request_canceled");
    await expect(key).toHaveValue(pendingKey);
    await expect(fresh).toBeDisabled();
    await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveValue(originalInput);
    release(); await delivery;
    await page.unrouteAll({ behavior: "wait" });
    await expect(panel.locator(".failure-summary")).toContainText("application_session_request_canceled");
    const reloaded = page.waitForResponse(response => isEndpoint(response, `${sessionPath}/${sessionId}`, "GET"));
    await refresh.click();
    expect((await (await reloaded).json()).session.turn_count).toBe(3);
    await expect(fresh).toBeDisabled();
    const checked = await submit(page, execute, `${sessionPath}/${sessionId}/turns`);
    expect(checked.idempotent_replay).toBe(true);
    expect(checked.turn.sequence).toBe(3);
    expect(checked.prompt_output ?? "").toBe("");
    await expect(fresh).toBeEnabled();
  } finally { release(); await page.unrouteAll({ behavior: "wait" }); }
});

test("Prompt recovery distinguishes new invocation, terminal recheck and credential rejection", async ({ page, application }, info) => {
  await prepareApplication(page, application);
  await navigate(page, "workspace-api-keys");
  const keys = page.locator("#workspace-api-keys");
  await keys.locator("summary").filter({ hasText: uiText(page, "Issue new credential") }).click();
  await keys.getByRole("textbox", { name: uiText(page, "Display name"), exact: true }).fill("Prompt recovery fixture");
  await keys.getByRole("checkbox", { name: /prompt_application:invoke/ }).check();
  await keys.getByRole("button", { name: uiText(page, "Issue API key"), exact: true }).click();
  await keys.getByRole("button", { name: copy(info, "Use in Prompt Application", "在 Prompt 应用中使用"), exact: true }).click();
  const panel = page.locator("#prompt-application-invocation");
  const key = panel.getByRole("textbox", { name: "client_invocation_key", exact: true });
  const fresh = panel.getByRole("button", { name: copy(info, "Prepare new invocation", "准备新调用"), exact: true });
  const execute = panel.getByRole("button", { name: copy(info, "Execute controlled invocation", "执行受控调用"), exact: true });
  await fillInputs(page);
  const oldKey = await key.inputValue();
  const path = "/v1/prompt-applications/invocations";
  expect((await submit(page, execute, path)).run.status).toBe("succeeded");
  const replay = await submit(page, execute, path);
  expect(replay.idempotent_replay).toBe(true);
  expect(replay.output).toBe("");
  let calls = 0;
  page.on("request", req => { if (req.method() === "POST" && new URL(req.url()).pathname === path) calls++; });
  await fresh.click();
  await expect(key).not.toHaveValue(oldKey);
  expect(calls).toBe(0);
  await input(page).getByRole("textbox", { name: "question", exact: true }).fill(`E2E_CASE:${randomUUID()}:valid`);
  expect((await submit(page, execute, path)).run.status).toBe("succeeded");
  const matcher = (url: URL) => url.pathname === path;
  for (const code of ["api_key_invalid", "api_key_expired", "api_key_revoked", "api_key_scope_denied"]) {
    await page.route(matcher, route => route.fulfill({ status: code === "api_key_invalid" ? 401 : 403, json: { error: {
      code, message: "PRIVATE_DIAGNOSTIC_NOT_FOR_DISPLAY", type: code === "api_key_scope_denied" ? "permission_error" : "authentication_error",
      request_id: "credential-rejection", route: `POST ${path}`, failure_boundary: "gateway_authentication", metadata: { latency_ms: 0 },
    } } }));
    await execute.click();
    await expect(panel.locator(".failure-summary")).toContainText(code);
    await expect(panel.getByRole("link", { name: copy(info, "Open current application API keys", "打开当前应用 API 密钥"), exact: true })).toHaveAttribute("href", "#workspace-api-keys");
    await expect(panel).not.toContainText("PRIVATE_DIAGNOSTIC_NOT_FOR_DISPLAY");
    await page.unroute(matcher);
  }
  await layout(page, panel, info, "invocation-credential-rejected");
  await page.route(matcher, route => route.abort("connectionfailed"));
  await execute.click();
  await expect(panel.locator(".failure-summary")).toContainText("prompt_invocation_transport_failed");
  await expect(fresh).toBeDisabled();
  await expect(key).toBeDisabled();
  await expect(panel.locator(".controlled-use-failure-guidance")).toHaveCount(0);
  await layout(page, panel, info, "invocation-interrupted");
  await page.unroute(matcher);
  // A later pre-dispatch denial cannot establish the earlier attempt's outcome.
  await page.route(matcher, route => route.fulfill({ status: 403, json: { error: {
    code: "api_key_expired", message: "private", type: "authentication_error", request_id: "expired-after-interruption",
    route: `POST ${path}`, failure_boundary: "gateway_authentication", metadata: { latency_ms: 0 },
  } } }));
  await execute.click();
  await expect(panel.locator(".failure-summary")).toContainText("api_key_expired");
  await expect(fresh).toBeDisabled();
  await expect(key).toBeDisabled();
  await page.unroute(matcher);
  expect((await submit(page, execute, path)).idempotent_replay).toBe(true);
  await expect(fresh).toBeEnabled();
});
