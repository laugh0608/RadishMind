import { test, expect } from "./workflow-fixtures";
import { gatewayNavigate } from "./gateway-ui";
import { evidenceLanguage, evidenceLayout } from "./results-evidence-fixtures";

const pages = [
  ["model-gateway-overview", "API distribution evidence", "API 分发证据"],
  ["model-gateway-route-evidence", "Provider/profile route evidence", "提供方 / 档案路由证据"],
  ["model-gateway-usage-audit-evidence", "Usage, trace, and audit evidence", "用量、追踪与审计证据"],
  ["model-gateway-evidence-review", "Evidence review and readiness", "证据审查与就绪状态"],
  ["admin-operations-review", "Operations review and readiness", "运维审查与就绪状态"],
  ["admin-provider-deployment-review", "Provider and deployment review", "提供方与部署审查"],
  ["workflow-workspace-review", "Context, scenario, draft, blockers", "上下文、场景、草案与阻塞"],
] as const;

test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });
test("all reachable auxiliary evidence views retranslate fixed explanations and preserve raw refs across three widths", async ({ page, application: _application }, info) => {
  test.setTimeout(180_000);
  const ui = evidenceLanguage(page, info);
  for (const [id, en, zh] of pages) {
    await gatewayNavigate(page, id);
    const panel = page.locator(`#${id}`);
    await expect(panel.getByRole("heading", { name: ui.copy(en, zh), exact: true })).toBeVisible();
    const refs = await panel.locator("code").allTextContents();
    const explanation = await panel.locator("p").allTextContents();
    await ui.switch();
    await expect(panel.getByRole("heading", { name: ui.copy(en, zh), exact: true })).toBeVisible();
    expect(await panel.locator("code").allTextContents()).toEqual(refs);
    expect(await panel.locator("p").allTextContents()).not.toEqual(explanation);
    await expect(panel).not.toContainText("{{");
    await expect(panel).not.toContainText("auxiliary.");
    await evidenceLayout(page, panel, info, id);
  }
  expect(ui.requests.filter(request => request.startsWith("POST "))).toEqual([]);
});
