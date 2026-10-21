import { randomUUID } from "node:crypto";
import { test, expect, evaluation, planOwner, button, evaluationLanguage, openEvaluation, createStructuredPlan, evaluationLayout } from "./evaluation-fixtures";
import { finishResponseRender, isEndpoint, selectApplication } from "./workflow-fixtures";
import { prepareApplication, navigate, template, templatePath, submit, publishAndActivate } from "./prompt-fixtures";
import { setTestLanguage, uiText } from "./ui-language";

test.use({ applicationKind: "prompt_application" });
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("Prompt reference changes require adoption and workspace switches reject late source reads", async ({ page, application }, info) => {
  const ui = evaluationLanguage(page, info), copy = ui.copy;
  const prepared = await prepareApplication(page, application);
  const assignmentPath = `/v1/user-workspace/applications/${application.id}/prompt-runtime-assignment`;
  const reading = page.waitForResponse(response => isEndpoint(response, assignmentPath, "GET"));
  await createStructuredPlan(page, application.id, ui, randomUUID());
  const assignment = (await (await reading).json()).assignment;
  const templateId = assignment.prompt_template_ref.template_id;
  await ui.switch(); // Restore the initial locale before opening a sibling tab.
  const panel = planOwner(page);
  const reference = panel.locator(".prompt-plan-reference");
  const name = panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true });
  await name.fill("Unsaved reference review 原文");
  await evaluationLayout(page, panel, info, "prompt-plan-form");

  // Publish a real second template version in another tab while preserving this draft.
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
    const newMessage = "{{ tone }}\nSynthetic reference version two";
    await template(sibling).locator(".prompt-template-message textarea").nth(0).fill(newMessage);
    const saved = await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Save with CAS"), exact: true }), templatePath);
    expect(saved.draft.messages[0].content).toBe(newMessage);
    const version = await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Create immutable version"), exact: true }), `${templatePath}/${templateId}/versions`);
    expect(version.version.template_version).toBe(2);
    expect(version.version.template_digest).not.toBe(assignment.prompt_template_ref.template_digest);
    await template(sibling).getByRole("button", { name: uiText(sibling, "Load drafts"), exact: true }).click();
    await template(sibling).getByRole("combobox", { name: uiText(sibling, "Valid Prompt Application draft"), exact: true }).selectOption(prepared.draftId);
    await template(sibling).getByRole("combobox", { name: uiText(sibling, "Immutable template version"), exact: true }).selectOption("2");
    await submit(sibling, template(sibling).getByRole("button", { name: uiText(sibling, "Bind and open Publish Review"), exact: true }), `/v1/user-workspace/application-configuration-drafts/${prepared.draftId}/prompt-template-binding`);
    await publishAndActivate(sibling, application.id, true);
  } finally { await sibling.close(); }

  await button(panel, copy("Review save", "审查保存")).click();
  await expect(reference).toHaveAttribute("data-reference-state", "changed");
  await expect(panel.locator(".application-evaluation-confirmation")).toHaveCount(0);
  await expect(name).toHaveValue("Unsaved reference review 原文");
  await evaluationLayout(page, panel, info, "prompt-reference-changed");
  await button(panel, copy("Adopt new reference", "采用新参考")).click();
  await expect(reference).toHaveAttribute("data-reference-state", "ready");
  await expect(reference.locator("strong")).toContainText("v2");
  await expect(panel.locator('[data-variable="tone"] textarea')).toHaveValue("E2E_PROMPT_SYSTEM_V1");
  await button(panel, copy("Review save", "审查保存")).click();
  await expect(panel.locator(".application-evaluation-confirmation")).toBeVisible();
  await evaluationLayout(page, panel, info, "prompt-plan-review");
  await name.fill("Scope-private unsaved text");
  await expect(panel.locator(".application-evaluation-confirmation")).toHaveCount(0);

  const sourcePath = `${templatePath}/${templateId}/versions/2`;
  let acknowledge!: () => void, release!: () => void;
  const arrived = new Promise<void>(resolve => { acknowledge = resolve; });
  const gate = new Promise<void>(resolve => { release = resolve; });
  const sourceURL = (url: URL) => url.origin === "http://127.0.0.1:17000" && url.pathname === sourcePath;
  await page.route(sourceURL, async route => {
    const response = await route.fetch();
    acknowledge();
    await gate;
    await route.fulfill({ response });
  });
  try {
    await button(panel, copy("Refresh template", "刷新模板")).click();
    await arrived;
    const workspace = page.locator("#desktop-active-workspace-input");
    await workspace.fill("workspace_prompt_reference_other");
    await page.locator(".product-nav-desktop").getByRole("button", { name: copy("Switch workspace", "切换工作区"), exact: true }).click();
    // The generic evaluation owner remains, but the Prompt editor is unmounted
    // when an empty workspace clears the selected application.
    await expect(workspace).toHaveValue("workspace_prompt_reference_other");
    await expect(evaluation(page).locator(".application-evaluation-scope")).not.toContainText(application.id);
    await expect(page.locator(".prompt-plan-owner")).toHaveCount(0);
    await expect(evaluation(page)).not.toContainText("Scope-private unsaved text");
    const delivered = page.waitForResponse(response => isEndpoint(response, sourcePath, "GET"));
    release();
    await finishResponseRender(page, await delivered);
    await expect(page.locator(".prompt-plan-owner")).toHaveCount(0);
    await expect(evaluation(page)).not.toContainText("Scope-private unsaved text");
    await workspace.fill("workspace_demo");
    await page.locator(".product-nav-desktop").getByRole("button", { name: copy("Switch workspace", "切换工作区"), exact: true }).click();
    await selectApplication(page, application);
    await openEvaluation(page, "plan");
    await expect(reference).toHaveAttribute("data-reference-state", "idle");
    await expect(name).toHaveValue("");
    await expect(evaluation(page)).not.toContainText("Scope-private unsaved text");
  } finally { release(); await page.unrouteAll({ behavior: "wait" }); }
});
