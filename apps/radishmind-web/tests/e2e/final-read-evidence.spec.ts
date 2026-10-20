import { test, expect } from "./workflow-fixtures";
import { evidenceLanguage, evidenceLayout } from "./results-evidence-fixtures";

const surfaces = [
  ["#workflow-surface-overview", "Application, draft, plan, readiness", "应用、草案、计划与就绪情况"],
  ["#workflow-scenario-inspector", "Scenario, input, output, blockers", "场景、输入、输出与阻塞项"],
  ["#workflow-application-detail", "Application Detail", "应用详情"],
  [".workflow-definition-detail", "Workflow Definition Detail", "工作流定义详情"],
  ["#routes", "Route catalog", "路由目录"],
  ["#states", "Shared states", "共享状态"],
  ["#guard", "Forbidden output guard", "禁止输出防护"],
] as const;

test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });
test("final read evidence covers reachable shell states and scenario selection in both languages", async ({ page, application }, info) => {
  test.setTimeout(180_000);
  const ui = evidenceLanguage(page, info);
  await page.waitForLoadState("networkidle");
  const scenario = page.locator("#workflow-scenario-inspector");
  const choices = scenario.getByRole("button");
  await choices.last().click();
  const selected = await scenario.locator('[aria-pressed="true"] small').innerText();
  for (const [selector, en, zh] of surfaces) {
    const panel = page.locator(selector);
    await panel.scrollIntoViewIfNeeded();
    await expect(panel).toContainText(ui.copy(en, zh));
    const rawPaths = await panel.locator(".route-path, code").allTextContents();
    const explanations = await panel.locator("p").allTextContents();
    await ui.switch();
    await expect(panel).toContainText(ui.copy(en, zh));
    expect(await panel.locator(".route-path, code").allTextContents()).toEqual(rawPaths);
    if (selector !== "#guard") expect(await panel.locator("p").allTextContents()).not.toEqual(explanations);
    await expect(panel).not.toContainText("{{");
    await expect(panel).not.toContainText(ui.copy("This message is unavailable.", "此消息暂不可用"));
    await evidenceLayout(page, panel, info, selector.replace(/^[#.]/, ""));
  }
  await expect(scenario.locator('[aria-pressed="true"] small')).toHaveText(selected);
  await expect(page.locator("#workflow-application-detail")).toContainText(application.name);
  await expect(page.locator("#routes")).toContainText("/v1/user-workspace/applications");
  await expect(page.locator(".application-states")).toContainText(ui.copy("The page renders a denial state", "页面显示拒绝状态"));
  await expect(page.locator(".workflow-definition-states")).toContainText(ui.copy("Tenant binding failure", "租户绑定失败"));
  expect(ui.requests.filter(request => request.startsWith("POST "))).toEqual([]);
});
