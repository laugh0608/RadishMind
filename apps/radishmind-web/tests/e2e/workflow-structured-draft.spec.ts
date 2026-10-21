import { test, expect, createDraft, saveDraft, designer, draftField, openDraft, promptLabel, selectApplication, isEndpoint } from "./workflow-fixtures";
import { reviewDefinitionVersions } from "./workflow-version-review-fixtures";
import { uiText } from "./ui-language";

test("Workflow v2 web authoring preserves mixed lists, revisions and two immutable versions", async ({ page, application }, testInfo) => {
  test.setTimeout(150_000);
  const label = (en: string, zh: string) => testInfo.project.use.locale === "zh-CN" ? zh : en;
  const v1 = await createDraft(page);
  await saveDraft(page, 1);
  await page.getByRole("button", { name: label("New structured input draft", "新建结构化输入草案"), exact: true }).click();
  const draftId = await draftField(page, "Draft").innerText();
  const editor = page.getByRole("region", { name: label("Structured input contract", "结构化输入合同"), exact: true });
  await expect(editor).toBeVisible();
  await editor.getByRole("textbox", { name: label("Contract ID", "合同标识"), exact: true }).fill("diagnostics_input_v1");
  await editor.getByRole("textbox", { name: label("Contract summary", "合同说明"), exact: true }).fill("Synthetic port diagnosis inputs");
  const field = (index: number) => editor.locator("fieldset").nth(index);
  await field(0).getByRole("textbox", { name: label("Display label", "显示标签"), exact: true }).fill("Log context");
  for (const [index, name, type, display] of [[1, "attempts", "integer", "Attempts"], [2, "threshold", "number", "Threshold"], [3, "confirmed", "boolean", "Confirmed"]] as const) {
    await editor.getByRole("button", { name: label("Add field", "添加字段"), exact: true }).click();
    await field(index).getByRole("textbox", { name: label("Field name", "字段名称"), exact: true }).fill(name);
    await field(index).getByRole("combobox", { name: label("Field type", "字段类型"), exact: true }).selectOption(type);
    await field(index).getByRole("textbox", { name: label("Display label", "显示标签"), exact: true }).fill(display);
    await field(index).getByRole("checkbox", { name: label("Required", "必填"), exact: true }).check();
  }
  await field(1).getByRole("textbox", { name: label("Field name", "字段名称"), exact: true }).fill("input_text");
  await expect(editor).toContainText(label("Field name is duplicated.", "字段名重复。"));
  await expect(designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true })).toBeDisabled();
  await field(1).getByRole("textbox", { name: label("Field name", "字段名称"), exact: true }).fill("attempts");
  await field(3).getByRole("button", { name: label("Move up", "上移"), exact: true }).click();
  await expect(field(2).getByRole("textbox", { name: label("Field name", "字段名称"), exact: true })).toHaveValue("confirmed");
  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await expect(editor).toBeVisible();
    expect(await editor.evaluate(element => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
    await editor.screenshot({ style: ".product-nav { visibility: hidden !important; }", path: testInfo.outputPath(`contract-${width}.png`) });
  }
  await field(0).getByRole("textbox", { name: label("Field name", "字段名称"), exact: true }).focus();
  await page.keyboard.press("Tab");
  await expect(field(0).getByRole("combobox")).toBeFocused();
  await page.setViewportSize({ width: 1440, height: 900 });
  const response = await saveDraft(page, 1);
  const saved = await response.json();
  expect(saved.draft.schema_version).toBe("saved_workflow_draft.v2");
  expect(saved.draft.input_contract.fields.map((item: { name: string }) => item.name)).toEqual(["input_text", "attempts", "confirmed", "threshold"]);
  const contractDigest = saved.draft.input_contract.contract_digest;
  expect(contractDigest).toMatch(/^sha256:[a-f0-9]{64}$/);
  await editor.getByRole("textbox", { name: label("Contract summary", "合同说明"), exact: true }).fill("Revised contract summary");
  await saveDraft(page, 2);
  const history = page.getByRole("region", { name: uiText(page, "Draft revision history and restore"), exact: true });
  await history.getByRole("button", { name: uiText(page, "Refresh history"), exact: true }).click();
  await history.getByRole("button", { name: /^v1 · / }).click();
  await history.getByRole("button", { name: uiText(page, "Prepare restore from this version"), exact: true }).click();
  await history.getByRole("button", { name: uiText(page, "Confirm new revision"), exact: true }).click();
  await expect(editor.getByRole("textbox", { name: label("Contract summary", "合同说明"), exact: true })).toHaveValue("Synthetic port diagnosis inputs");
  await page.reload();
  await selectApplication(page, application);
  await openDraft(page, v1);
  await expect(editor).toHaveCount(0);
  await openDraft(page, draftId);
  await expect(editor).toContainText(contractDigest);
  const promotion = page.locator("#workflow-definition-promotion");
  const runIds: string[] = [];
  for (const version of [1, 2]) {
    await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
    await promotion.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
    await promotion.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
    if (version === 2) {
      await promotion.locator(".workflow-definition-list button").filter({ has: page.locator("strong", { hasText: /^v2$/ }) }).click();
      await promotion.locator("article").nth(2).getByRole("combobox").selectOption("replace");
    }
    const pointerAction = version === 1 ? uiText(page, "Activate · expected pointer v0") : label("Replace · expected pointer v1", "替换 · 预期指针 v1");
    await promotion.getByRole("button", { name: pointerAction, exact: true }).click();
    await expect(promotion.locator(".structured-runtime-input")).toBeVisible();
    await promotion.getByRole("textbox", { name: "Log context", exact: true }).fill("Synthetic backend port 7100, frontend port 7000.");
    await promotion.getByRole("textbox", { name: "Attempts", exact: true }).fill("2");
    await promotion.getByRole("textbox", { name: "Threshold", exact: true }).fill("0.5");
    await promotion.getByRole("radio", { name: "false", exact: true }).check();
    const running = page.waitForResponse(response => isEndpoint(response, "/v1/user-workspace/workflow-definition-runs", "POST"));
    await promotion.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true }).click();
    const run = (await (await running).json()).run;
    expect(run.schema_version).toBe("workflow_run_record.v8");
    expect(run.definition_authority.definition_version).toBe(version);
    expect(run.input_contract_digest).toBe(contractDigest);
    runIds.push(run.run_id);
    if (version === 1) {
      await promotion.getByRole("button", { name: label("Derive new draft from v1", "从 v1 派生新草案"), exact: true }).click();
      await expect(editor).toContainText(contractDigest);
      await (await promptLabel(page)).fill("Second immutable version prompt");
      await saveDraft(page, 1);
    }
  }
  expect(new Set(runIds).size).toBe(2);
  await promotion.getByRole("button", { name: uiText(page, "Open Run History"), exact: true }).click();
  await expect(page.getByRole("region", { name: uiText(page, "runs workflow review owner"), exact: true })).toContainText(runIds[1]!);
  await reviewDefinitionVersions(page, application, runIds, testInfo);
});


test("Workflow v1 web iteration preserves two exact versions through Case and Suite review", async ({ page, application }, info) => {
  test.setTimeout(120_000);
  const copy = (en: string, zh: string) => info.project.use.locale === "zh-CN" ? zh : en;
  await createDraft(page); await saveDraft(page, 1);
  const promotion = page.locator("#workflow-definition-promotion");
  const ids: string[] = [];
  for (const version of [1, 2]) {
    await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
    await promotion.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
    await promotion.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
    if (version === 2) {
      await promotion.locator(".workflow-definition-list button").filter({ has: page.locator("strong", { hasText: /^v2$/ }) }).click();
      await promotion.locator("article").nth(2).getByRole("combobox").selectOption("replace");
    }
    await promotion.getByRole("button", { name: version === 1 ? uiText(page, "Activate · expected pointer v0") : copy("Replace · expected pointer v1", "替换 · 预期指针 v1"), exact: true }).click();
    await promotion.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true }).fill("Synthetic port mismatch diagnosis.");
    const running = page.waitForResponse(response => isEndpoint(response, "/v1/user-workspace/workflow-definition-runs", "POST"));
    await promotion.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true }).click();
    const run = (await (await running).json()).run;
    expect(run.schema_version).toBe("workflow_run_record.v5");
    expect(run.definition_authority.definition_version).toBe(version);
    ids.push(run.run_id);
    if (version === 1) {
      await promotion.getByRole("button", { name: copy("Derive new draft from v1", "从 v1 派生新草案"), exact: true }).click();
      await (await promptLabel(page)).fill("Second v1 contract version"); await saveDraft(page, 1);
    }
  }
  await promotion.getByRole("button", { name: uiText(page, "Open Run History"), exact: true }).click();
  await reviewDefinitionVersions(page, application, ids, info);
});
