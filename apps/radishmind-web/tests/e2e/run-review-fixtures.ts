import type { Page } from "@playwright/test";
import { expect, createDraft, saveDraft, isEndpoint } from "./workflow-fixtures";
import { uiText, setTestLanguage } from "./ui-language";
type Locale = "en-US" | "zh-CN";
const owner = (page: Page) => page.locator(".workflow-review-owner-surface");
export async function switchLanguage(page: Page, locale: Locale) {
  await page.evaluate(value => {
    localStorage.setItem("radishmind.uiLocale.v1", value);
    window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
  }, locale);
  setTestLanguage(page, locale);
  await expect(page.locator("html")).toHaveAttribute("lang", locale);
}
export async function prepareRuns(page: Page) {
  await createDraft(page); await saveDraft(page, 1);
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true }).click();
  const ids: string[] = [];
  for (let i = 0; i < 2; i++) {
    await page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true }).fill(`Synthetic run review ${i}`);
    const response = page.waitForResponse(r => isEndpoint(r, "/v1/user-workspace/workflow-definition-runs", "POST"));
    await page.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true }).click();
    const received = await response;
    expect(received.ok()).toBeTruthy();
    const body = await received.json();
    await expect(page.locator(".workflow-definition-run-result strong")).toHaveText(body.run.run_id);
    await expect(page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true })).toHaveValue("");
    ids.push(body.run.run_id);
  }
  expect(ids[0]).not.toBe(ids[1]);
  await page.getByRole("button", { name: uiText(page, "Open Run History"), exact: true }).click();
  await expect(owner(page).locator(".workflow-review-run-row")).toHaveCount(2);
  await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[1]);
  return ids;
}
