import type { Page, Locator, TestInfo } from "@playwright/test";
import { expect, isEndpoint } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";

// Independent copy expectations, not product dictionary imports.
export const agentEn = {
  validate: "Validate", name: "Profile name", defaultLocale: "Default request locale", allowedLocales: "Allowed request locales", save: "Save with CAS", version: "Create immutable version", refresh: "Refresh", loadDrafts: "Load drafts", validDraft: "Valid Agent draft", immutableVersion: "Immutable Profile version", bind: "Bind and open Publish Review", load: "Load assignment and approved candidates", create: "Create exact-authority Session v3", execute: "Execute one controlled turn", cancel: "Cancel current request", locale: "Request locale", context: "Transient context", key: "Client turn key", openRun: "Open Run v7 evidence", invalidContext: "Context must be a JSON object.", invalidName: "Profile names must contain 2–80 characters.", conflict: "The saved version changed. Reload the exact Profile before retrying.", replay: "The existing turn metadata was returned; the model was not called again and the answer was not restored.", canceled: "The browser request was canceled; late responses will be discarded.", activate: "Activate with CAS v0", revoke: "Revoke with CAS v1",
};
export const agentZh: typeof agentEn = {
  validate: "校验", name: "Profile 名称", defaultLocale: "默认请求语言", allowedLocales: "允许的请求语言", save: "按 CAS 保存", version: "创建不可变版本", refresh: "刷新", loadDrafts: "加载草案", validDraft: "有效 Agent 草案", immutableVersion: "不可变 Profile 版本", bind: "绑定并打开发布审查", load: "加载运行分配与已批准候选", create: "按精确权威创建 Session v3", execute: "执行一轮受控调用", cancel: "取消当前请求", locale: "请求语言", context: "易失上下文", key: "客户端轮次键", openRun: "打开 Run v7 证据", invalidContext: "上下文必须是 JSON 对象。", invalidName: "Profile 名称必须为 2 至 80 个字符。", conflict: "已保存版本发生变化。请重读精确 Profile 后重试。", replay: "已返回已有轮次的元数据；未重复调用模型，也未恢复回答。", canceled: "当前浏览器请求已取消；迟到响应会被丢弃。", activate: "激活（CAS v0）", revoke: "撤销（CAS v1）",
};
export function agentLanguage(page: Page, testInfo: TestInfo) {
  let locale: "zh-CN" | "en-US" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  return {
    copy: () => locale === "zh-CN" ? agentZh : agentEn,
    requests,
    async switch() {
      const before = [...requests]; locale = locale === "zh-CN" ? "en-US" : "zh-CN";
      await page.locator(".ui-language-selector select:visible").first().selectOption(locale);
      setTestLanguage(page, locale);
      await expect(page.locator("html")).toHaveAttribute("lang", locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests).toEqual(before);
    },
  };
}
export async function navigateAgent(page: Page, anchor: string) {
  await page.locator(`a[href="#${anchor}"]:visible`).first().click();
  await expect(page.locator(`#${anchor}`)).toBeVisible();
}
export async function agentSubmit(page: Page, button: Locator, path: string, allowFailure = false) {
  const pending = page.waitForResponse(response => isEndpoint(response, path, "POST"));
  await button.click(); const response = await pending;
  expect(response.ok()).toBe(true);
  const body = await response.json(); if (!allowFailure) expect(body.failure_code ?? "").toBe("");
  return { body, request: response.request() };
}
export async function agentLayout(page: Page, panel: Locator, testInfo: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
    const controls = panel.locator("input:not([type=checkbox]), textarea, select, button");
    expect(await controls.evaluateAll(elements => elements.flatMap(element => {
      const box = element.getBoundingClientRect();
      return !box.width || box.left >= 0 && box.right <= innerWidth + 1 ? [] : [{ tag: element.tagName, text: element.textContent, left: box.left, right: box.right }];
    })), `Controls must fit the ${width}px viewport`).toEqual([]);
    expect(await panel.locator(".prompt-template-summary > small, .prompt-template-scope article > strong, .prompt-template-scope article > code").evaluateAll(elements => elements.every(element => element.scrollWidth <= element.clientWidth + 1)), "Exact IDs and Profile digests must remain fully readable").toBe(true);
    if (testInfo.repeatEachIndex === 0) await panel.screenshot({ path: testInfo.outputPath(`${name}-${width}.png`), style: ".product-nav { visibility: hidden !important; }" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
