import { randomUUID } from "node:crypto";
import { test, expect, createDraft, saveDraft, draftField, holdDraftResponse, isEndpoint } from "./workflow-fixtures";
import { uiText, setTestLanguage, draftVersionText } from "./ui-language";

test("review and template language switching preserves pending decisions and explicit derivation", async ({ page, application }, testInfo) => {
  const locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next = locale === "zh-CN" ? "en-US" : "zh-CN";
  async function switchLanguage(value: "en-US" | "zh-CN") {
    await page.evaluate(locale => {
      localStorage.setItem("radishmind.uiLocale.v1", locale);
      window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: locale, storageArea: localStorage }));
    }, value);
    setTestLanguage(page, value);
    await expect(page.locator("html")).toHaveAttribute("lang", value);
  }
  const writes: string[] = [];
  page.on("request", request => {
    if (request.method() === "POST" && new URL(request.url()).port === "17000") writes.push(new URL(request.url()).pathname);
  });
  await createDraft(page);
  await saveDraft(page, 1);
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  const promotion = page.locator("#workflow-definition-promotion");
  const definitionId = await promotion.getByRole("textbox", { name: uiText(page, "Definition ID"), exact: true }).inputValue();
  const candidateId = await promotion.getByRole("textbox", { name: uiText(page, "Candidate ID"), exact: true }).inputValue();
  await promotion.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
  const reason = "Reviewed source <literal> 保留理由";
  await promotion.getByRole("textbox", { name: uiText(page, "Reason"), exact: true }).fill(reason);
  const reviewPath = `/v1/user-workspace/workflow-definition-candidates/${candidateId}/decisions`;
  const held = await holdDraftResponse(page, reviewPath, "POST");
  try {
    await promotion.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
    await held.arrived;
    const before = [...writes];
    await switchLanguage(next);
    expect(writes).toEqual(before);
    await expect(promotion.getByRole("textbox", { name: uiText(page, "Reason"), exact: true })).toHaveValue(reason);
    await expect(promotion.getByRole("button", { name: uiText(page, "Append review v1"), exact: true })).toBeDisabled();
    await held.deliver();
    await expect(promotion.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true })).toBeEnabled();
    expect(writes.filter(path => path === reviewPath)).toHaveLength(1);
  } finally { await held.dispose(); }

  const catalog = page.locator("#workspace-workflow-template-catalog");
  const field = (name: string) => catalog.getByRole("textbox", { name: uiText(page, name), exact: true });
  const button = (name: string) => catalog.getByRole("button", { name: uiText(page, name), exact: true });
  const token = randomUUID().replaceAll("-", "");
  const templateId = `template_${token}`, templateCandidate = `candidate_${token}`, draftId = `draft_${token}`;
  await field("Candidate ID").fill(templateCandidate);
  await field("Template ID").fill(templateId);
  await field("Definition ID").fill(definitionId);
  await field("Title").fill("Source template 源模板 <literal>");
  await field("Summary").fill("Synthetic template review and derivation evidence.");
  await button("Create candidate").click();
  await expect(catalog.locator(".workflow-template-records")).toContainText(templateCandidate);
  const tasks = catalog.locator(".workflow-template-catalog__tasks button");
  await tasks.nth(1).click();
  await catalog.getByRole("combobox", { name: uiText(page, "Pending candidate"), exact: true }).selectOption(templateCandidate);
  await field("Reason").fill(reason);
  const beforeReviewSwitch = [...writes];
  await switchLanguage(locale);
  expect(writes).toEqual(beforeReviewSwitch);
  await expect(field("Reason")).toHaveValue(reason);
  await expect(catalog.getByRole("combobox", { name: uiText(page, "Pending candidate"), exact: true })).toHaveValue(templateCandidate);
  await button("Submit review").click();
  await expect(button("Submit review")).toBeDisabled();
  await tasks.nth(2).click();
  await catalog.getByRole("combobox", { name: uiText(page, "Template"), exact: true }).selectOption(templateId);
  await field("Reason").fill("List exact reviewed version for isolated test.");
  const listed = page.waitForResponse(response => response.request().method() === "POST" && new URL(response.url()).pathname.endsWith("/listing-decisions"));
  await button("Apply pointer decision").click();
  expect((await listed).ok()).toBeTruthy();
  await expect(button("Refresh")).toBeEnabled();
  await tasks.nth(3).click();
  await catalog.getByRole("combobox", { name: uiText(page, "Listed template"), exact: true }).selectOption(templateId);
  await field("Saved draft ID").fill(draftId);
  await field("Draft name").fill("Derived draft 派生草案 <literal>");
  await expect(button("Derive and open saved draft")).toBeDisabled();
  await catalog.getByRole("checkbox").check();
  const beforeDeriveSwitch = [...writes];
  await switchLanguage(next);
  expect(writes).toEqual(beforeDeriveSwitch);
  await expect(catalog.getByRole("checkbox")).toBeChecked();
  await expect(field("Saved draft ID")).toHaveValue(draftId);
  await expect(field("Draft name")).toHaveValue("Derived draft 派生草案 <literal>");
  await expect(catalog.getByRole("combobox", { name: uiText(page, "Listed template"), exact: true })).toHaveValue(templateId);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await catalog.scrollIntoViewIfNeeded();
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
    if (testInfo.repeatEachIndex === 0) await testInfo.attach(`template-confirm-${width}`, { body: await catalog.screenshot(), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  const derived = page.waitForResponse(response => isEndpoint(response, `/v1/user-workspace/workflow-templates/${templateId}/derivations`, "POST"));
  await button("Derive and open saved draft").click();
  expect((await derived).ok()).toBeTruthy();
  await expect(draftField(page, "Draft")).toHaveText(draftId);
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 1, 1));
  expect(writes.filter(path => path.endsWith("/derivations"))).toHaveLength(1);
  expect(writes.filter(path => path.endsWith("workflow-definition-runs"))).toHaveLength(0);
  await expect(page.getByRole("region", { name: uiText(page, "Application development context") })).toContainText(application.id);
});
