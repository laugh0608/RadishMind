import { randomUUID } from "node:crypto";
import type { Page, Locator, TestInfo } from "@playwright/test";
import { test, expect, isEndpoint, selectApplication, finishResponseRender, type Application } from "./workflow-fixtures";
import { uiText, setTestLanguage } from "./ui-language";
import { prepareApplication, session, sessionPath, template, templatePath, navigate, submit, publishAndActivate } from "./prompt-fixtures";

test.use({ applicationKind: "prompt_application", trace: "off", screenshot: "off" });
const input = (page: Page) => page.locator(".prompt-variable-input");
const reference = (page: Page) => input(page).locator("[data-reference-state]");
const read = async (page: Page) => {
  await input(page).getByRole("button", { name: uiText(page, "Read input template"), exact: true }).click();
  await expect(reference(page)).toHaveAttribute("data-reference-state", "ready");
};
async function fillVariable(page: Page, name: string, value: string) {
  const field = input(page).locator(`[data-variable="${name}"]`);
  await field.getByRole("combobox").first().selectOption("provided");
  await field.getByRole("textbox", { name, exact: true }).fill(value);
}
async function layout(page: Page, info: TestInfo, name: string, widths = [1440, 1200, 390]) {
  for (const width of widths) {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
    await input(page).scrollIntoViewIfNeeded();
    expect(await input(page).evaluate(element => ({
      viewport: document.documentElement.scrollWidth <= window.innerWidth,
      fields: [...element.querySelectorAll("input, textarea, select, button")].every(field => {
        const rect = field.getBoundingClientRect(), parent = element.getBoundingClientRect();
        return rect.left >= parent.left - 1 && rect.right <= parent.right + 1;
      }),
    }))).toEqual({ viewport: true, fields: true });
    // Nested desktop panels can remain inside a narrow owner even at 1200px.
    // A non-overflowing field is still unusable if the legacy side column crushes it.
    if (width >= 1200) {
      expect(await input(page).evaluate(element => element.getBoundingClientRect().width)).toBeGreaterThanOrEqual(300);
    }
    const readButton = input(page).getByRole("button", { name: uiText(page, "Read input template"), exact: true });
    await readButton.evaluate(element => element.scrollIntoView({ block: "center" }));
    expect(await readButton.evaluate(element => {
      const box = element.getBoundingClientRect();
      return element.contains(document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2));
    })).toBe(true);
    // Capture document coordinates with the fixed header at the document top,
    // rather than letting element screenshots scroll a tall editor underneath it.
    await page.evaluate(() => window.scrollTo(0, 0));
    const clip = await input(page).boundingBox();
    expect(clip).not.toBeNull();
    await page.screenshot({ path: info.outputPath(`${name}-${width}.png`), fullPage: true, clip: clip! });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
async function replaceTemplate(page: Page, application: Application, draftId: string, templateId: string, info: TestInfo) {
  const sibling = await page.context().newPage();
  setTestLanguage(sibling, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  try {
    await sibling.goto("/#workspace-applications");
    await selectApplication(sibling, application);
    await navigate(sibling, "application-configuration-draft");
    await navigate(sibling, "prompt-application-template-workspace");
    await template(sibling).getByRole("button", { name: uiText(sibling, "Refresh"), exact: true }).first().click();
    const restored = sibling.waitForResponse(response => isEndpoint(response, `${templatePath}/${templateId}`, "GET"));
    await template(sibling).locator(".prompt-template-summary").filter({ hasText: templateId }).click();
    await finishResponseRender(sibling, await restored);
    // Change a real default and digest without changing the fixture provider's transport contract.
    await template(sibling).locator(".prompt-template-variable").nth(1).getByRole("textbox").last().fill("E2E_PROMPT_SYSTEM_V1");
    await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Save with CAS"), exact: true }), templatePath);
    const version = await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Create immutable version"), exact: true }), `${templatePath}/${templateId}/versions`);
    expect(version.version.template_version).toBe(2);
    await template(sibling).getByRole("button", { name: uiText(sibling, "Load drafts"), exact: true }).click();
    await template(sibling).getByRole("combobox", { name: uiText(sibling, "Valid Prompt Application draft"), exact: true }).selectOption(draftId);
    await template(sibling).getByRole("combobox", { name: uiText(sibling, "Immutable template version"), exact: true }).selectOption("2");
    await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Bind and open Publish Review"), exact: true }), `/v1/user-workspace/application-configuration-drafts/${draftId}/prompt-template-binding`);
    await publishAndActivate(sibling, application.id, true);
  } finally { await sibling.close(); }
}
async function templateIdentity(page: Page) {
  const text = await reference(page).locator("code").innerText();
  return text.split(" · ")[0];
}

test("Prompt structured Session inputs preserve values and reject a changed binding before dispatch", async ({ page, application }, info) => {
  const copy = (en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
  const { sessionId, draftId } = await prepareApplication(page, application);
  const execute = session(page).getByRole("button", { name: uiText(page, "Execute Prompt turn"), exact: true });
  await expect(execute).toBeDisabled();
  await read(page);
  const templateId = await templateIdentity(page);
  const caseId = randomUUID();
  await fillVariable(page, "question", `E2E_CASE:${caseId}:valid`);
  await fillVariable(page, "tone", "E2E_PROMPT_SYSTEM_V1");
  await layout(page, info, "session-input");
  const nextLocale = info.project.use.locale === "zh-CN" ? "en-US" : "zh-CN";
  const before: string[] = [];
  const requests = (request: { method(): string; url(): string }) => { if (request.url().includes("17000")) before.push(request.method()); };
  page.on("request", requests);
  await page.locator(".ui-language-selector select:visible").first().selectOption(nextLocale);
  setTestLanguage(page, nextLocale);
  await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveValue(`E2E_CASE:${caseId}:valid`);
  expect(before).toEqual([]);
  page.off("request", requests);
  await page.locator(".ui-language-selector select:visible").first().selectOption(info.project.use.locale!);
  setTestLanguage(page, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  await session(page).getByRole("textbox", { name: "client_turn_key", exact: true }).fill(`input-${caseId}`);
  const result = await submit(page, execute, `${sessionPath}/${sessionId}/turns`);
  expect(result.turn.status).toBe("succeeded");
  await replaceTemplate(page, application, draftId, templateId, info);
  let dispatched = 0;
  page.on("request", request => { if (request.method() === "POST" && new URL(request.url()).pathname.endsWith("/turns")) dispatched += 1; });
  await execute.click();
  await expect(reference(page)).toHaveAttribute("data-reference-state", "unavailable");
  await expect(reference(page)).toContainText("prompt_input_session_authority_changed");
  expect(dispatched).toBe(0);
  await layout(page, info, "session-binding-changed", [1440, 390]);
  await expect(input(page).getByRole("textbox", { name: "question", exact: true })).toHaveCount(0);
  // Retained values remain available in advanced JSON; execution itself still has the old Session authority.
  await input(page).getByRole("combobox", { name: uiText(page, "Variable input mode"), exact: true }).selectOption("json");
  await expect(input(page).getByRole("textbox")).toHaveValue(new RegExp(caseId));
  const created = await submit(page, session(page).getByRole("button", { name: uiText(page, "Create Session v2"), exact: true }), sessionPath);
  await expect(input(page).getByRole("combobox", { name: uiText(page, "Variable input mode"), exact: true })).toHaveValue("form");
  await read(page);
  await fillVariable(page, "question", `E2E_CASE:${randomUUID()}:valid`);
  await expect(input(page).locator('[data-variable="tone"] select')).toHaveValue("omitted");
  const request = page.waitForRequest(req => req.method() === "POST" && new URL(req.url()).pathname.endsWith("/turns"));
  const saved = await submit(page, execute, `${sessionPath}/${created.session.session_id}/turns`);
  expect(saved.turn.status).toBe("succeeded");
  expect((await request).postDataJSON().variables).not.toHaveProperty("tone");
  // Cancel an in-flight source read by changing workspace; the old source must never repopulate the editor.
  let release!: () => void, arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const reached = new Promise<void>(resolve => { arrived = resolve; });
  await page.route(url => url.pathname === `${templatePath}/${templateId}/versions/2`, async route => {
    const response = await route.fetch(); arrived(); await gate; await route.fulfill({ response });
  });
  try {
    await input(page).getByRole("button", { name: uiText(page, "Read input template"), exact: true }).click();
    await reached;
    const workspace = page.locator("#desktop-active-workspace-input");
    await workspace.fill("workspace_prompt_input_other");
    await page.locator(".product-nav-desktop").getByRole("button", { name: copy("Switch workspace", "切换工作区"), exact: true }).click();
    await expect(input(page)).toHaveCount(0);
    const delivered = page.waitForResponse(response => isEndpoint(response, `${templatePath}/${templateId}/versions/2`, "GET"));
    release(); await finishResponseRender(page, await delivered);
    await expect(input(page)).toHaveCount(0);
    await workspace.fill("workspace_demo");
    await page.locator(".product-nav-desktop").getByRole("button", { name: copy("Switch workspace", "切换工作区"), exact: true }).click();
    await selectApplication(page, application);
    await navigate(page, "prompt-application-session");
    await expect(reference(page)).toHaveAttribute("data-reference-state", "idle");
  } finally { release(); await page.unrouteAll({ behavior: "wait" }); }
});

test("Prompt structured direct invocation requires explicit adoption and keeps advanced JSON available without source access", async ({ page, application }, info) => {
  const copy = (en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
  const { draftId } = await prepareApplication(page, application);
  await navigate(page, "workspace-api-keys");
  const keys = page.locator("#workspace-api-keys");
  await keys.locator("summary").filter({ hasText: uiText(page, "Issue new credential") }).click();
  await keys.getByRole("textbox", { name: uiText(page, "Display name"), exact: true }).fill("Prompt input fixture");
  await keys.getByRole("checkbox", { name: /prompt_application:invoke/ }).check();
  await keys.getByRole("button", { name: uiText(page, "Issue API key"), exact: true }).click();
  await keys.getByRole("button", { name: copy("Use in Prompt Application", "在 Prompt 应用中使用"), exact: true }).click();
  const panel = page.locator("#prompt-application-invocation");
  await read(page);
  const templateId = await templateIdentity(page);
  await fillVariable(page, "question", `E2E_CASE:${randomUUID()}:valid`);
  await fillVariable(page, "tone", "E2E_PROMPT_SYSTEM_V1");
  await layout(page, info, "invocation-input");
  await replaceTemplate(page, application, draftId, templateId, info);
  let calls = 0;
  page.on("request", request => { if (request.method() === "POST" && new URL(request.url()).pathname === "/v1/prompt-applications/invocations") calls += 1; });
  const invoke = panel.getByRole("button", { name: copy("Execute controlled invocation", "执行受控调用"), exact: true });
  await invoke.click();
  await expect(reference(page)).toHaveAttribute("data-reference-state", "changed");
  expect(calls).toBe(0);
  await layout(page, info, "invocation-reference-changed", [1440, 390]);
  await input(page).getByRole("button", { name: copy("Adopt new input reference", "采用新输入参考"), exact: true }).click();
  await expect(input(page).getByRole("textbox", { name: "tone", exact: true })).toHaveValue("E2E_PROMPT_SYSTEM_V1");
  const result = await submit(page, invoke, "/v1/prompt-applications/invocations");
  expect(result.run.status).toBe("succeeded");
  // A deterministic denial of source reads must not grant access or disable the existing JSON execution path.
  await page.route(url => url.pathname === `${templatePath}/${templateId}/versions/2`, route => route.fulfill({ status: 403, json: { failure_code: "prompt_template_scope_denied" } }));
  await input(page).getByRole("button", { name: uiText(page, "Read input template"), exact: true }).click();
  await expect(reference(page)).toHaveAttribute("data-reference-state", "unavailable");
  await layout(page, info, "invocation-source-denied", [1440, 390]);
  await input(page).getByRole("combobox", { name: uiText(page, "Variable input mode"), exact: true }).selectOption("json");
  const json = input(page).getByRole("textbox");
  await json.fill('{"question":"a","question":"b"}');
  await expect(invoke).toBeDisabled();
  await json.fill(JSON.stringify({ question: `E2E_CASE:${randomUUID()}:valid` }));
  await panel.getByRole("textbox", { name: "client_invocation_key", exact: true }).fill(`input-${randomUUID()}`);
  expect((await submit(page, invoke, "/v1/prompt-applications/invocations")).run.status).toBe("succeeded");
  expect(calls).toBe(2);
  await page.unrouteAll({ behavior: "wait" });
  await page.reload();
  await selectApplication(page, application);
  await navigate(page, "prompt-application-invocation");
  await input(page).getByRole("combobox", { name: uiText(page, "Variable input mode"), exact: true }).selectOption("json");
  await input(page).getByRole("textbox").fill('{"question":"valid input without a credential"}');
  await expect(invoke).toBeDisabled();
});
