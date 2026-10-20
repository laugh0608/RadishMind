import { readFile } from "node:fs/promises";
import { test, expect, isEndpoint } from "./workflow-fixtures";
import { evidenceLanguage, evidenceLayout, library, openLibrary, readFixtureResult, resultId, resultReadPath, resultExportPath, resultContent, fixtureApplication, holdEvidenceResponse } from "./results-evidence-fixtures";

test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("result library exports verified original JSON and preserves bilingual selection, filters and lifecycle", async ({ page, application: _application }, info) => {
  const ui = evidenceLanguage(page, info), copy = ui.copy;
  await openLibrary(page); const panel = library(page);
  const held = await holdEvidenceResponse(page, resultReadPath);
  try {
    await panel.locator(".application-result-artifact-list").getByRole("button").filter({ hasText: resultId }).click();
    await held.arrived; await ui.switch();
    await expect(panel).toContainText(copy("Reading exact content for", "正在精确读取"));
    await held.deliver(); await expect(panel.locator("pre")).toHaveText(resultContent);
  } finally { await held.dispose(); }
  const exportHeld = await holdEvidenceResponse(page, resultExportPath);
  let exportJSON = "";
  try {
    await panel.getByRole("button", { name: copy("Prepare verified JSON", "准备已校验的 JSON"), exact: true }).click();
    await exportHeld.arrived; await ui.switch();
    await expect(panel.getByRole("button", { name: copy("Verifying export…", "正在校验导出…"), exact: true })).toBeDisabled();
    const response = page.waitForResponse(value => isEndpoint(value, resultExportPath, "GET"));
    await exportHeld.deliver(); exportJSON = JSON.stringify((await (await response).json()).export);
  } finally { await exportHeld.dispose(); }
  const downloadLink = panel.locator(".application-result-artifact-download");
  await expect(downloadLink).toBeVisible();
  const filename = await downloadLink.getAttribute("download"), objectURL = await downloadLink.getAttribute("href");
  await ui.switch();
  await expect(downloadLink).toHaveAttribute("download", filename!); await expect(downloadLink).toHaveAttribute("href", objectURL!);
  const downloaded = page.waitForEvent("download"); await downloadLink.click(); const download = await downloaded;
  expect(download.suggestedFilename()).toBe(filename);
  const content = await readFile((await download.path())!, "utf8");
  expect(JSON.parse(content)).toEqual(JSON.parse(exportJSON)); expect(content).toContain(resultContent.replaceAll("\n", "\\n"));
  await expect(panel.locator("[role=status]")).toContainText(filename!);
  await evidenceLayout(page, panel, info, "result-library");
  const archived = page.waitForResponse(value => isEndpoint(value, `${resultReadPath}/archive`, "POST"));
  await panel.getByRole("button", { name: copy("Archive", "归档"), exact: true }).click(); expect((await archived).ok()).toBeTruthy();
  await expect(downloadLink).toHaveCount(0);
  await panel.getByRole("combobox", { name: copy("Lifecycle", "生命周期"), exact: true }).selectOption("archived");
  await ui.switch();
  await expect(panel.getByRole("combobox", { name: copy("Lifecycle", "生命周期"), exact: true })).toHaveValue("archived");
  await panel.getByRole("button", { name: copy("Apply filters", "应用筛选"), exact: true }).click(); await readFixtureResult(page);
  const restored = page.waitForResponse(value => isEndpoint(value, `${resultReadPath}/unarchive`, "POST"));
  await panel.getByRole("button", { name: copy("Unarchive", "取消归档"), exact: true }).click(); expect((await restored).ok()).toBeTruthy();
  await page.reload(); await openLibrary(page); await readFixtureResult(page);
  await expect(panel.locator("pre")).toHaveText(resultContent);
});

test("result failures, real lifecycle conflict and stale export remain scoped across language and application changes", async ({ page, application }, info) => {
  const ui = evidenceLanguage(page, info), copy = ui.copy;
  await page.evaluate(() => {
    const originalCreate = URL.createObjectURL.bind(URL), originalRevoke = URL.revokeObjectURL.bind(URL);
    const urls = { created: [] as string[], revoked: [] as string[] };
    Object.assign(window, { resultTestURLs: urls });
    URL.createObjectURL = blob => { const url = originalCreate(blob); urls.created.push(url); return url; };
    URL.revokeObjectURL = url => { urls.revoked.push(url); originalRevoke(url); };
  });
  await openLibrary(page); const panel = library(page);
  const readResponse = page.waitForResponse(value => isEndpoint(value, resultReadPath, "GET"));
  await readFixtureResult(page); const original = await readResponse; const originalBody = await original.json();
  const headers = await original.request().allHeaders(); delete headers["content-length"];
  headers["x-radishmind-dev-read-scopes"] = headers["x-radishmind-dev-read-membership-permissions"] = "application_sessions:read,application_result_artifacts:archive";
  const concurrent = await page.request.post(`http://127.0.0.1:17000${resultReadPath}/archive`, { headers, data: { workspace_id: "workspace_demo", application_id: fixtureApplication.id, expected_lifecycle_version: originalBody.lifecycle.lifecycle_version } });
  expect(concurrent.ok(), await concurrent.text()).toBeTruthy();
  await panel.getByRole("button", { name: copy("Archive", "归档"), exact: true }).click();
  await expect(panel.locator("[role=status]")).toContainText("application_result_artifact_lifecycle_version_conflict");
  await ui.switch(); await expect(panel.locator("[role=status]")).toContainText(copy("The lifecycle changed", "生命周期已变化"));
  await expect(panel.locator("pre")).toHaveText(resultContent);
  await panel.getByRole("combobox", { name: copy("Lifecycle", "生命周期"), exact: true }).selectOption("archived");
  await panel.getByRole("button", { name: copy("Apply filters", "应用筛选"), exact: true }).click(); await readFixtureResult(page);
  await panel.getByRole("button", { name: copy("Unarchive", "取消归档"), exact: true }).click();
  await expect(panel.locator("pre")).toHaveCount(0);
  await panel.getByRole("button", { name: copy("Clear", "清除"), exact: true }).click(); await readFixtureResult(page);
  // A structurally valid export with a tampered digest must not create a download.
  const exportPattern = `http://127.0.0.1:17000${resultExportPath}?*`;
  await page.route(exportPattern, async route => {
    const response = await route.fetch(); const body = await response.json(); body.export.export_digest = "sha256:" + "0".repeat(64);
    await route.fulfill({ response, json: body });
  }, { times: 1 });
  await panel.getByRole("button", { name: copy("Prepare verified JSON", "准备已校验的 JSON"), exact: true }).click();
  await expect(panel.locator("[role=status]")).toContainText("application_result_artifact_store_contract_mismatch");
  await expect(panel.locator(".application-result-artifact-download")).toHaveCount(0); await ui.switch();
  await expect(panel.locator("[role=status]")).toContainText(copy("scope or content verification", "作用域或内容校验"));
  await panel.getByRole("button", { name: copy("Prepare verified JSON", "准备已校验的 JSON"), exact: true }).click();
  await expect(panel.locator(".application-result-artifact-download")).toBeVisible();
  const objectURL = await panel.locator(".application-result-artifact-download").getAttribute("href");
  const held = await holdEvidenceResponse(page, resultExportPath);
  try {
    await panel.getByRole("button", { name: copy("Prepare verified JSON", "准备已校验的 JSON"), exact: true }).click(); await held.arrived;
    await expect(panel.locator(".application-result-artifact-download")).toHaveCount(0);
    expect(await page.evaluate(url => (window as unknown as { resultTestURLs: { revoked: string[] } }).resultTestURLs.revoked.includes(url!), objectURL)).toBeTruthy();
    await openLibrary(page, application.id); await held.deliver();
    await expect(panel.locator(".empty-state").first()).toContainText(copy("No saved results match", "当前筛选条件下还没有已保存结果"));
    await expect(panel.locator(".application-result-artifact-download")).toHaveCount(0); await expect(panel).not.toContainText(resultContent);
    await ui.switch(); await expect(panel.locator(".empty-state").first()).toContainText(copy("No saved results match", "当前筛选条件下还没有已保存结果"));
  } finally { await held.dispose(); }
});
