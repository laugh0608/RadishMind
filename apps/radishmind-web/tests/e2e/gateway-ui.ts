import type { Locator, Page, TestInfo } from "@playwright/test";
import { expect, isEndpoint } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";

const en = {
  issueSettings: "Issue new credential", issue: "Issue API key", handoff: "Use in Playground", load: "Load models", input: "Temporary input", protocol: "Protocol", stream: "Stream response", send: "Send request", cancel: "Cancel", history: "Review sanitized history", clear: "Clear credential", success: "succeeded", canceled: "canceled", invalid: "gateway_playground_input_invalid", filters: "Exact filters", apply: "Apply filters", route: "Route", status: "Status", earlier: "Load earlier requests", empty: "No request records match the current caller scope and filters.", example: "Example", code: "Generated API integration example", operations: "Runtime observations, usage and cost evidence", refresh: "Refresh observations", requests: "Gateway requests", runs: "Workflow runs", partial: "Partial failure", overview: "Workspace pulse", inbox: "Operations inbox", quota: "Usage quota", missingCurrency: "The source provides no currency", failure: "Request history is unavailable.", filterFailure: "Check the exact filters and local time interval.", openEvidence: "Open evidence",
};
const zh: typeof en = {
  issueSettings: "签发新凭据", issue: "签发 API 密钥", handoff: "在调试台中使用", load: "加载模型", input: "临时输入", protocol: "协议", stream: "流式响应", send: "发送请求", cancel: "取消", history: "查看脱敏历史", clear: "清除凭据", success: "成功", canceled: "已取消", invalid: "gateway_playground_input_invalid", filters: "精确筛选", apply: "应用筛选", route: "路由", status: "状态", earlier: "加载更早的请求", empty: "当前调用方作用域与筛选条件下没有请求记录。", example: "示例", code: "生成的 API 接入示例", operations: "运行观测、用量与成本证据", refresh: "刷新观测", requests: "网关请求", runs: "工作流运行", partial: "部分来源失败", overview: "工作区概况", inbox: "运营收件箱", quota: "用量与配额", missingCurrency: "来源未提供币种", failure: "请求历史不可用", filterFailure: "请检查精确筛选条件与本地时间区间", openEvidence: "打开证据",
};
export function gatewayLanguage(page: Page, info: TestInfo) {
  let locale: "zh-CN" | "en-US" = info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  return {
    copy: () => locale === "zh-CN" ? zh : en,
    requests,
    async switch() {
      const before = [...requests]; locale = locale === "zh-CN" ? "en-US" : "zh-CN";
      await page.locator(".ui-language-selector select:visible").first().selectOption(locale);
      setTestLanguage(page, locale); await expect(page.locator("html")).toHaveAttribute("lang", locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests).toEqual(before);
    },
  };
}
export const playground = (page: Page) => page.locator("#model-gateway-playground");
export const historyPanel = (page: Page) => page.locator("#model-gateway-request-history");
export async function gatewayNavigate(page: Page, anchor: string) {
  const link = page.locator(`a[href="#${anchor}"]:visible`).first();
  if (!await link.count()) {
    const more = page.locator(".product-nav-more:visible").filter({ has: page.locator(`a[href="#${anchor}"]`) }).first();
    if (await more.count() && !await more.getAttribute("open")) await more.locator("summary").click();
  }
  await link.click();
  await expect(page.locator(`#${anchor}`)).toBeVisible();
}
export async function prepareGateway(page: Page, language: ReturnType<typeof gatewayLanguage>) {
  await gatewayNavigate(page, "application-api-integration");
  await gatewayNavigate(page, "workspace-api-keys");
  const keys = page.locator("#workspace-api-keys");
  await keys.locator("details > summary").first().click();
  await keys.locator('input[type="number"]').first().fill("1");
  const issued = page.waitForResponse(response => isEndpoint(response, "/v1/user-workspace/api-keys", "POST"));
  await keys.getByRole("button", { name: language.copy().issue, exact: true }).click();
  const response = await issued;
  // Read only the sanitized ID. Never return or attach the one-time credential.
  const id: string = (await response.json()).record?.api_key_id ?? "";
  try { await keys.getByRole("button", { name: language.copy().handoff, exact: true }).click(); }
  finally { if (await keys.locator(".api-key-one-time-token").count()) await page.reload(); }
  expect(id).toMatch(/^key_[a-z2-7]{16}$/);
  await expect(playground(page)).toBeVisible();
  await playground(page).getByRole("button", { name: language.copy().load, exact: true }).click();
  await expect(playground(page).getByRole("button", { name: language.copy().send, exact: true })).toBeEnabled();
  return id;
}
export async function gatewayLayout(page: Page, panel: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
    expect(await panel.locator("input, textarea, select, button").evaluateAll(elements => elements.flatMap(element => {
      const r = element.getBoundingClientRect(); return !r.width || r.left >= 0 && r.right <= innerWidth + 1 ? [] : [{ tag: element.tagName, left: r.left, right: r.right }];
    })), `Controls fit ${width}px`).toEqual([]);
    expect(await panel.locator("article, dl > div, .workspace-attention-resource, .workspace-operations-inbox-item-copy").evaluateAll(elements => elements.filter(element => element.clientWidth > 0 && element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName)), `Evidence fits its container at ${width}px`).toEqual([]);
    if (info.repeatEachIndex === 0) await panel.screenshot({ path: info.outputPath(`${name}-${width}.png`), style: ".product-nav { visibility: hidden !important; }" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
