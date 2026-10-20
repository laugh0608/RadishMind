import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import type { Page } from "@playwright/test";
import { test, expect, isEndpoint, selectApplication, finishResponseRender } from "./workflow-fixtures";
import { uiText } from "./ui-language";
import { templatePath, sessionPath, template, session, artifacts, navigate, submit, prepareApplication } from "./prompt-fixtures";

test.use({ applicationKind: "prompt_application" });

async function saveTurn(page: Page, sessionId: string, mode: "valid" | "alternate") {
  const caseId = randomUUID();
  await session(page).getByRole("textbox", { name: uiText(page, "Turn variables"), exact: true }).fill(JSON.stringify({ question: `E2E_CASE:${caseId}:${mode}`, tone: "E2E_PROMPT_SYSTEM_V1" }));
  await session(page).getByRole("textbox", { name: "client_turn_key", exact: true }).fill(`e2e-${caseId}`);
  await artifacts(page).getByRole("checkbox").check();
  const body = await submit(page, session(page).getByRole("button", { name: uiText(page, "Execute Prompt turn"), exact: true }), `${sessionPath}/${sessionId}/turns`);
  expect(body.result_artifact_failure_code ?? "").toBe("");
  expect(body.turn.status).toBe("succeeded");
  return body;
}

for (const delayed of [false, true]) test(`Prompt result selection clears ${delayed ? "a delayed" : "an already read"} prior answer before reviewing a new result`, async ({ page, application }, info) => {
  const copy = (en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
  const { sessionId } = await prepareApplication(page, application);
  const a = await saveTurn(page, sessionId, "valid");
  const aId = a.result_artifact.artifact_id as string;
  const aPath = `${sessionPath}/${sessionId}/result-artifacts/${aId}`;
  const inspector = artifacts(page).locator(".application-result-artifact-inspector");
  let release!: () => void, arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const ready = new Promise<void>(resolve => { arrived = resolve; });
  const pattern = `**${aPath}?*`;
  if (delayed) {
    // Keep this read deliverable after abort to exercise generation checks, not only AbortController.
    await page.evaluate(path => {
      const original = window.fetch;
      window.fetch = (input, init) => typeof input === "string" && new URL(input).pathname === path
        ? original(input, { ...init, signal: undefined }) : original(input, init);
    }, aPath);
    await page.route(pattern, async route => {
      const response = await route.fetch(); arrived(); await gate; await route.fulfill({ response });
    });
  }
  try {
    await artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: aId }).click();
    if (delayed) await ready;
    else await expect(inspector.locator("pre")).toContainText("Synthetic timeout diagnosis");
    const b = await saveTurn(page, sessionId, "alternate");
    const bId = b.result_artifact.artifact_id as string;
    expect(bId).not.toBe(aId);
    await expect(inspector.locator("h5")).toHaveText(bId);
    await expect(inspector.locator("pre")).toHaveCount(0);
    await expect(inspector.getByRole("button")).toHaveCount(0);
    if (delayed) {
      const late = page.waitForResponse(response => isEndpoint(response, aPath, "GET"));
      release(); await finishResponseRender(page, await late);
      await expect(inspector.locator("h5")).toHaveText(bId);
      await expect(inspector.locator("pre")).toHaveCount(0);
      await expect(inspector.getByRole("button")).toHaveCount(0);
    }
    const bPath = `${sessionPath}/${sessionId}/result-artifacts/${bId}`;
    const readB = page.waitForResponse(response => isEndpoint(response, bPath, "GET"));
    await artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: bId }).click();
    await readB;
    await expect(inspector.locator("h5")).toHaveText(bId);
    await expect(inspector.locator("pre")).toContainText("Second synthetic diagnosis");
    await expect(inspector.locator("pre")).not.toContainText("Synthetic timeout diagnosis");
    const archived = await submit(page, inspector.getByRole("button", { name: copy("Archive artifact", "归档结果资产"), exact: true }), `${bPath}/archive`);
    expect(archived.lifecycle.artifact_id).toBe(bId);
    expect(archived.lifecycle.lifecycle_version).toBe(2);
    // Filter reload must not reapply the stale save receipt for B.
    const filter = artifacts(page).getByRole("combobox");
    await filter.selectOption("archived");
    await expect(artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: bId })).toBeVisible();
    await filter.selectOption("active");
    await expect(artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: aId })).toBeVisible();
    await expect(artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: bId })).toHaveCount(0);
    await filter.selectOption("archived");
    await artifacts(page).locator(".application-result-artifact-list button").filter({ hasText: bId }).click();
    await expect(inspector.locator("pre")).toContainText("Second synthetic diagnosis");
    await inspector.getByRole("button", { name: uiText(page, "Open exact run"), exact: true }).click();
    await expect(page.locator(".workflow-review-run-detail")).toContainText(b.turn.run_ref.run_id);
  } finally {
    release();
    if (delayed) await page.unroute(pattern);
  }
});

test("Prompt diagnostic variables retain focus while typing and survive saved draft recovery", async ({ page, application }, info) => {
  const copy = (en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
  const source = JSON.parse(readFileSync(new URL("../../../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.parts/diagnostics-trial-source.json", import.meta.url), "utf8"));
  await navigate(page, "application-configuration-draft");
  await navigate(page, "prompt-application-template-workspace");
  for (const [index, message] of source.messages.entries()) await template(page).locator(".prompt-template-message textarea").nth(index).fill(message.content);
  await template(page).getByRole("button", { name: copy("Add variable", "添加变量"), exact: true }).click();
  for (const [index, variable] of source.variables.entries()) {
    const row = template(page).locator(".prompt-template-variable").nth(index);
    const name = row.getByRole("textbox").first();
    await name.focus(); await name.press("ControlOrMeta+A"); await name.press("Backspace");
    await expect(name).toBeFocused();
    await name.pressSequentially(variable.name);
    await expect(name).toBeFocused(); await expect(name).toHaveValue(variable.name);
    await name.press("Backspace"); await expect(name).toBeFocused();
    await name.pressSequentially(variable.name.slice(-1)); await expect(name).toHaveValue(variable.name);
    await row.getByRole("checkbox").check();
    await row.getByRole("textbox").nth(1).fill(variable.description);
  }
  await template(page).getByRole("combobox", { name: uiText(page, "Kind"), exact: true }).selectOption("json_object");
  await template(page).locator("#prompt-output-schema").fill(JSON.stringify(source.output_contract.json_schema));
  await template(page).getByRole("spinbutton").fill("8192");
  const saved = await submit(page, template(page).getByRole("button", { name: uiText(page, "Save with CAS"), exact: true }), templatePath);
  expect(saved.draft.variables).toEqual(source.variables);
  expect(JSON.stringify(saved.draft)).not.toContain("editorKey");
  await page.reload(); await selectApplication(page, application);
  await navigate(page, "prompt-application-template-workspace");
  await template(page).getByRole("button", { name: uiText(page, "Refresh"), exact: true }).first().click();
  await template(page).locator(".prompt-template-summary").filter({ hasText: saved.draft.template_id }).click();
  for (const [index, variable] of source.variables.entries()) await expect(template(page).locator(".prompt-template-variable").nth(index).getByRole("textbox").first()).toHaveValue(variable.name);
  const samples = JSON.parse(readFileSync(new URL("../../../../docs/features/user-workspace/prompt-application-dev-test-usage-guide.parts/diagnostics-trial-samples.json", import.meta.url), "utf8")).samples;
  for (const sample of samples) {
    await template(page).locator(".prompt-template-review textarea").fill(JSON.stringify(sample.variables));
    await template(page).getByRole("button", { name: copy("Render synthetic preview", "生成合成预览"), exact: true }).click();
    await expect(template(page).locator(`.prompt-template-preview.${sample.expected_render_failure_code ? "invalid" : "valid"}`)).toBeVisible();
    if (sample.expected_render_failure_code) await expect(template(page).locator(".prompt-template-preview")).toContainText(sample.expected_render_failure_code);
  }
});
