import { randomUUID } from "node:crypto";
import { test, expect, holdDraftResponse, isEndpoint } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";

const route = "/v1/user-workspace/workflow-retrieval-snapshots";
const labels = {
  create: ["New knowledge snapshot", "新建知识快照"], key: ["Snapshot key", "快照键"],
  name: ["Display name", "显示名称"], content: ["Content", "正文"], title: ["Title", "标题"],
  createV1: ["Create v1", "创建 v1"], archive: ["Archive snapshot", "归档快照"],
  confirmArchive: ["Confirm archive", "确认归档"], archived: ["Archived", "已归档"],
  replace: ["Replace all and create v2", "完整替换并创建 v2"],
  invalidType: ["is not a supported Markdown / Text file", "不是受支持的 Markdown / Text 文件"],
  conflict: ["Version conflict", "版本冲突"], failed: ["Failed", "失败"],
} as const;

test("RAG material review, pending create, version and archive retain bilingual state", async ({ page, application }, testInfo) => {
  let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const text = (key: keyof typeof labels) => labels[key][locale === "zh-CN" ? 1 : 0];
  await page.locator('a[href="#application-configuration-draft"]').first().click();
  await page.locator('button[aria-controls="application-development-owner-surface"]').click();
  const panel = page.locator("#workflow-rag-snapshot-panel");
  await expect(panel).toBeVisible();
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).pathname.startsWith(route)) requests.push(`${request.method()} ${request.url()}`); });
  async function switchLanguage() {
    locale = locale === "zh-CN" ? "en-US" : "zh-CN";
    await page.evaluate(value => {
      localStorage.setItem("radishmind.uiLocale.v1", value);
      window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
    }, locale);
    setTestLanguage(page, locale);
    await expect(page.locator("html")).toHaveAttribute("lang", locale);
  }
  await panel.getByRole("button", { name: text("create"), exact: true }).click();
  const key = `rag_${randomUUID().replaceAll("-", "").slice(0, 14)}`;
  const name = "知识资料 <literal> Original";
  await panel.getByRole("textbox", { name: text("key"), exact: true }).fill(key);
  await panel.getByRole("textbox", { name: text("name"), exact: true }).fill(name);
  const fileInput = panel.locator('input[type="file"]');
  await fileInput.setInputFiles({ name: "资料.pdf", mimeType: "application/pdf", buffer: Buffer.from("unsupported") });
  const findings = panel.locator(".workflow-rag-findings");
  await expect(findings).toContainText(text("invalidType"));
  const requestCount = requests.length;
  await switchLanguage();
  await expect(findings).toContainText(text("invalidType"));
  expect(requests).toHaveLength(requestCount);
  const material = "# 原始标题\nLiteral <b> content 保持原文。";
  await fileInput.setInputFiles({ name: "原始材料.md", mimeType: "text/markdown", buffer: Buffer.from(material) });
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toHaveValue(material);
  await expect(panel.getByRole("button", { name: text("createV1"), exact: true })).toBeEnabled();
  const held = await holdDraftResponse(page, route, "POST");
  let snapshotId = "";
  try {
    await panel.getByRole("button", { name: text("createV1"), exact: true }).click();
    await held.arrived;
    const pendingRequests = requests.length;
    await switchLanguage();
    await expect(panel.getByRole("textbox", { name: text("name"), exact: true })).toHaveValue(name);
    await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toHaveValue(material);
    expect(requests).toHaveLength(pendingRequests);
    const response = page.waitForResponse(value => isEndpoint(value, route, "POST"));
    await held.deliver();
    const result = await response;
    expect(result.ok()).toBeTruthy();
    snapshotId = (await result.json()).record.snapshot_id;
    const payload = result.request().postDataJSON();
    expect(payload.application_id).toBe(application.id);
    expect(payload.fragments[0].content).toBe(material);
    expect(Object.keys(payload.fragments[0]).sort()).toEqual(["content", "fragment_ref", "is_official", "page_slug", "source_ref", "source_type", "title"]);
    expect(JSON.stringify(payload)).not.toContain("原始材料.md");
  } finally { await held.dispose(); }
  await expect(panel.locator(".workflow-rag-record")).toContainText(`workflow.rag.${key}.v1`);
  const edited = `${material}\nHuman reviewed revision 保持原文。`;
  await panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) }).fill(edited);
  await switchLanguage();
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toHaveValue(edited);
  const versionPath = `${route}/${snapshotId}/versions`;
  const versionResponse = page.waitForResponse(value => isEndpoint(value, versionPath, "POST"));
  await panel.getByRole("button", { name: text("replace"), exact: true }).click();
  expect((await versionResponse).ok()).toBeTruthy();
  await expect(panel.locator(".workflow-rag-record")).toContainText(`workflow.rag.${key}.v2`);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await panel.scrollIntoViewIfNeeded();
    const dimensions = await page.evaluate(() => ({ scroll: document.documentElement.scrollWidth, client: document.documentElement.clientWidth }));
    expect(dimensions.scroll).toBeLessThanOrEqual(dimensions.client + 1);
    const clippedRegions = await panel.locator(".workflow-rag-fragment-owner, .workflow-rag-fragment-workspace, .workflow-rag-fragment-inspector").evaluateAll(elements =>
      elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className));
    expect(clippedRegions, "Fragment controls must fit inside their owner, including on desktop").toEqual([]);
    await testInfo.attach(`rag-${locale}-${width}`, { body: await panel.screenshot(), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await panel.getByRole("button", { name: text("archive"), exact: true }).click();
  const beforeArchiveSwitch = requests.length;
  await switchLanguage();
  await expect(panel.getByRole("button", { name: text("confirmArchive"), exact: true })).toBeVisible();
  expect(requests).toHaveLength(beforeArchiveSwitch);
  const archivePath = `${route}/${snapshotId}/archive`;
  const archived = page.waitForResponse(value => isEndpoint(value, archivePath, "POST"));
  await panel.getByRole("button", { name: text("confirmArchive"), exact: true }).click();
  expect((await archived).ok()).toBeTruthy();
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toBeDisabled();
  await page.reload();
  await page.locator('button[aria-controls="application-development-owner-surface"]').click();
  await panel.locator(".workflow-rag-filter").getByRole("button", { name: text("archived"), exact: true }).click();
  await panel.locator(".workflow-rag-list").getByRole("button").filter({ hasText: name }).click();
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toHaveValue(edited);
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toBeDisabled();
});

test("RAG real version conflict preserves local edits and translates after the failure", async ({ page, application: _application }, testInfo) => {
  let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const text = (key: keyof typeof labels) => labels[key][locale === "zh-CN" ? 1 : 0];
  await page.locator('a[href="#application-configuration-draft"]').first().click();
  await page.locator('button[aria-controls="application-development-owner-surface"]').click();
  const panel = page.locator("#workflow-rag-snapshot-panel");
  await panel.getByRole("button", { name: text("create"), exact: true }).click();
  await panel.getByRole("textbox", { name: text("key"), exact: true }).fill(`cas_${randomUUID().replaceAll("-", "").slice(0, 14)}`);
  await panel.getByRole("textbox", { name: text("name"), exact: true }).fill("Conflict 原文");
  await panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) }).fill("Original content.");
  const creation = page.waitForResponse(value => isEndpoint(value, route, "POST"));
  await panel.getByRole("button", { name: text("createV1"), exact: true }).click();
  const response = await creation;
  expect(response.ok()).toBeTruthy();
  const { record } = await response.json();
  await expect(panel.locator(".workflow-rag-record")).toContainText(record.rag_ref);
  const versionPath = `${route}/${record.snapshot_id}/versions`;
  const { snapshot_key: _snapshotKey, ...body } = response.request().postDataJSON();
  const headers = await response.request().allHeaders();
  delete headers["content-length"];
  const concurrent = await page.request.post(`http://127.0.0.1:17000${versionPath}`, {
    headers, data: { ...body, expected_latest_version: 1 },
  });
  expect(concurrent.ok()).toBeTruthy();
  const local = "Local unsaved edit 本地未提交修改。";
  await panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) }).fill(local);
  const conflict = page.waitForResponse(value => isEndpoint(value, versionPath, "POST"));
  await panel.getByRole("button", { name: text("replace"), exact: true }).click();
  expect((await (await conflict).json()).failure_code).toBe("workflow_rag_snapshot_version_conflict");
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(text("conflict"));
  locale = locale === "zh-CN" ? "en-US" : "zh-CN";
  await page.evaluate(value => {
    localStorage.setItem("radishmind.uiLocale.v1", value);
    window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
  }, locale);
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(text("conflict"));
  await expect(panel.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) })).toHaveValue(local);
  await expect(panel.locator(".workflow-rag-operation")).toContainText("workflow_rag_snapshot_version_conflict");
});
