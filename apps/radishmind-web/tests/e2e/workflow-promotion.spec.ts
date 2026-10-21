import { randomUUID } from "node:crypto";
import type { Page, Route } from "@playwright/test";
import { test, expect, createDraft, saveDraft, isEndpoint, finishResponseRender, selectApplication } from "./workflow-fixtures";
import { promotionRequest } from "./workflow-rag-promotion-fixtures";
import { uiText } from "./ui-language";

const promotion = (page: Page) => page.locator("#workflow-definition-promotion");
const owners = (page: Page) => promotion(page).locator(".workflow-definition-promotion-grid > article");
const authority = (page: Page) => owners(page).nth(3).locator("dd").last();
const runButton = (page: Page) => promotion(page).getByRole("button", { name: uiText(page, "Start exact version run"), exact: true });
const input = (page: Page) => promotion(page).getByRole("textbox", { name: uiText(page, "One-time input"), exact: true });
const choose = (page: Page, id: string) => owners(page).nth(0).getByRole("button").filter({ hasText: id }).click();
const reload = (page: Page) => promotion(page).getByRole("button", { name: /Reload candidates and exact authority|重新读取候选与精确执行依据/ });

async function seedLineages(page: Page, applicationId: string) {
  const draftId = await createDraft(page);
  await saveDraft(page, 1);
  const suffix = randomUUID().replaceAll("-", "");
  const lineages = ["a", "b"].map(key => ({ definition: `definition_${key}_${suffix}`, candidate: `candidate_${key}_${suffix}` }));
  for (const item of lineages) {
    await promotionRequest(page, applicationId, "/v1/user-workspace/workflow-definition-candidates", ["workflow_definitions:write"], { candidate_id: item.candidate, definition_id: item.definition, draft_id: draftId, expected_draft_version: 1, expected_lifecycle_version: 1 });
    await promotionRequest(page, applicationId, `/v1/user-workspace/workflow-definition-candidates/${item.candidate}/decisions`, ["workflow_definitions:review"], { expected_review_version: 0, decision: "approve", reason: "Synthetic lineage review." });
    await promotionRequest(page, applicationId, `/v1/user-workspace/workflow-definitions/${item.definition}/activation-decisions`, ["workflow_definitions:activate"], { expected_pointer_version: 0, decision: "activate", version: 1, reason: "Synthetic activation." });
  }
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  await expect(reload(page)).toBeEnabled();
  return lineages;
}

async function holdResponse(page: Page, path: string, method = "GET", fail = false) {
  let release!: () => void;
  let acknowledge!: () => void;
  let intercepted = false;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const arrived = new Promise<void>(resolve => { acknowledge = resolve; });
  const pattern = new RegExp(`${path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}(?:\\?|$)`);
  const handler = async (route: Route) => {
    if (intercepted || route.request().method() !== method) return route.fallback();
    intercepted = true;
    const response = await route.fetch();
    acknowledge();
    await gate;
    if (fail) await route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ failure_code: "synthetic_read_unavailable" }) });
    else await route.fulfill({ response });
  };
  await page.route(pattern, handler);
  return {
    arrived,
    async deliver() {
      const received = page.waitForResponse(response => isEndpoint(response, path, method));
      release();
      await finishResponseRender(page, await received);
    },
    async dispose() { release(); await page.unroute(pattern, handler); },
  };
}

const activationPath = (definition: string) => `/v1/user-workspace/workflow-definitions/${definition}/activation`;

test("Workflow promotion isolates A to B to A reads and late failures", async ({ page, application }) => {
  const [a, b] = await seedLineages(page, application.id);
  await choose(page, b.candidate);
  await expect(authority(page)).toHaveText(`${b.definition} · v1`);
  await input(page).fill("Synthetic input owned by B");
  const lateA = await holdResponse(page, activationPath(a.definition));
  try {
    await choose(page, a.candidate); await lateA.arrived;
    await expect(authority(page)).not.toContainText(b.definition);
    await expect(runButton(page)).toBeDisabled();
    await expect(input(page)).toHaveValue("");
    await choose(page, b.candidate);
    await expect(authority(page)).toHaveText(`${b.definition} · v1`);
    await input(page).fill("New B input survives stale A");
    await lateA.deliver();
    await expect(authority(page)).toHaveText(`${b.definition} · v1`);
    await expect(input(page)).toHaveValue("New B input survives stale A");
    await expect(runButton(page)).toBeEnabled();
  } finally { await lateA.dispose(); }

  const firstA = await holdResponse(page, activationPath(a.definition), "GET", true);
  try {
    await choose(page, a.candidate); await firstA.arrived;
    await choose(page, b.candidate); await expect(authority(page)).toHaveText(`${b.definition} · v1`);
    await choose(page, a.candidate); await expect(authority(page)).toHaveText(`${a.definition} · v1`);
    await input(page).fill("Latest A input");
    await firstA.deliver();
    await expect(authority(page)).toHaveText(`${a.definition} · v1`);
    await expect(input(page)).toHaveValue("Latest A input");
    await expect(promotion(page).getByRole("alert")).toHaveCount(0);
    await expect(runButton(page)).toBeEnabled();
  } finally { await firstA.dispose(); }
});

for (const fail of [false, true]) {
  test(`Workflow promotion ignores late mutation ${fail ? "failure" : "success"} and its finalizer`, async ({ page, application }) => {
    const [a, b] = await seedLineages(page, application.id);
    await choose(page, a.candidate); await expect(authority(page)).toHaveText(`${a.definition} · v1`);
    await owners(page).nth(2).getByRole("combobox").selectOption("deactivate");
    const oldMutation = await holdResponse(page, `${activationPath(a.definition)}-decisions`, "POST", fail);
    const newRead = await holdResponse(page, activationPath(b.definition));
    const requests: string[] = [];
    page.on("request", request => { if (request.method() === "GET") requests.push(new URL(request.url()).pathname); });
    try {
      await owners(page).nth(2).getByRole("button", { name: /Deactivate · expected pointer v1|停用 · 预期指针 v1/ }).click();
      await oldMutation.arrived;
      await choose(page, b.candidate); await newRead.arrived;
      const before = requests.filter(path => path.startsWith(`/v1/user-workspace/workflow-definitions/${a.definition}/`)).length;
      await oldMutation.deliver();
      await expect(reload(page)).toBeDisabled();
      await expect(runButton(page)).toBeDisabled();
      await expect(promotion(page).getByRole("alert")).toHaveCount(0);
      expect(requests.filter(path => path.startsWith(`/v1/user-workspace/workflow-definitions/${a.definition}/`))).toHaveLength(before);
      await newRead.deliver();
      await expect(authority(page)).toHaveText(`${b.definition} · v1`);
      await expect(reload(page)).toBeEnabled();
      await expect(owners(page).nth(1)).toContainText(b.definition);
    } finally { await oldMutation.dispose(); await newRead.dispose(); }
  });
}

test("Workflow promotion conflict retains reason and requires explicit fresh decision", async ({ page, application }, info) => {
  const [, b] = await seedLineages(page, application.id);
  await choose(page, b.candidate); await expect(authority(page)).toHaveText(`${b.definition} · v1`);
  await owners(page).nth(2).getByRole("combobox").selectOption("deactivate");
  const reason = owners(page).nth(1).getByRole("textbox");
  await reason.fill("Keep this explicit local review reason");
  await promotionRequest(page, application.id, `${activationPath(b.definition)}-decisions`, ["workflow_definitions:activate"], { expected_pointer_version: 1, decision: "deactivate", version: 0, reason: "Concurrent tab deactivation." });
  const posts: string[] = [];
  page.on("request", request => { if (request.method() === "POST") posts.push(new URL(request.url()).pathname); });
  const received = page.waitForResponse(response => isEndpoint(response, `${activationPath(b.definition)}-decisions`, "POST"));
  await owners(page).nth(2).getByRole("button", { name: /Deactivate · expected pointer v1|停用 · 预期指针 v1/ }).click();
  expect((await (await received).json()).failure_code).toBe("workflow_definition_version_conflict");
  await expect(promotion(page).getByRole("alert")).toContainText("workflow_definition_version_conflict");
  await expect(runButton(page)).toBeDisabled();
  await expect(reason).toHaveValue("Keep this explicit local review reason");
  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
    expect(await promotion(page).evaluate(element => [...element.querySelectorAll(".workflow-definition-list, .workflow-definition-list button, .workflow-definition-list strong, .workflow-definition-failure")].filter(node => node.scrollWidth > node.clientWidth + 1).map(node => node.className))).toEqual([]);
    await promotion(page).screenshot({ path: info.outputPath(`conflict-${width}.png`), style: ".product-nav { visibility: hidden !important; }" });
  }
  const failedRead = await holdResponse(page, activationPath(b.definition), "GET", true);
  try {
    await reload(page).click(); await failedRead.arrived; await failedRead.deliver();
    await expect(promotion(page).getByRole("alert")).toBeVisible();
    await expect(reason).toHaveValue("Keep this explicit local review reason");
    await expect(runButton(page)).toBeDisabled();
  } finally { await failedRead.dispose(); }
  await reload(page).click();
  const activate = owners(page).nth(2).getByRole("button", { name: /Activate · expected pointer v2|激活 · 预期指针 v2/ });
  await expect(activate).toBeEnabled();
  await expect(reason).toHaveValue("Keep this explicit local review reason");
  expect(posts).toEqual([`${activationPath(b.definition)}-decisions`]);
  await activate.click();
  await expect(authority(page)).toHaveText(`${b.definition} · v1`);
  await expect(owners(page).nth(2)).toContainText(/pointer v3|指针 v3/);
  expect(posts).toHaveLength(2);
  await expect(input(page)).toHaveValue("");
  await promotion(page).screenshot({ path: info.outputPath("recovered-390.png"), style: ".product-nav { visibility: hidden !important; }" });
});

test("Workflow promotion discards pending responses after workspace exit", async ({ page, application }) => {
  const [a, b] = await seedLineages(page, application.id);
  await choose(page, b.candidate); await expect(authority(page)).toHaveText(`${b.definition} · v1`);
  const oldRead = await holdResponse(page, activationPath(a.definition), "GET", true);
  try {
    await choose(page, a.candidate); await oldRead.arrived;
    await page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true }).fill("workspace_e2e_other");
    await page.getByRole("button", { name: uiText(page, "Switch workspace"), exact: true }).click();
    await expect(page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true })).toHaveValue("workspace_e2e_other");
    await oldRead.deliver();
    await page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true }).fill("workspace_demo");
    await page.getByRole("button", { name: uiText(page, "Switch workspace"), exact: true }).click();
    await selectApplication(page, application);
    await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
    await expect(reload(page)).toBeEnabled();
    await expect(promotion(page).getByRole("alert")).toHaveCount(0);
    await expect(input(page)).toHaveValue("");
  } finally { await oldRead.dispose(); }
});
