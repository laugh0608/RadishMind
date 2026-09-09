import { randomUUID } from "node:crypto";
import { test as base, expect, type Page, type Response, type Route } from "@playwright/test";

type Application = { id: string; name: string };

export const test = base.extend<{ application: Application; diagnostics: void }>({
  diagnostics: [async ({ context }, use, testInfo) => {
    const failures: string[] = [];
    const requests: { method: string; path: string; status: number }[] = [];
    const pageErrors: string[] = [];
    const watch = (page: Page) => {
      page.on("pageerror", (error) => pageErrors.push(error.message));
      page.on("response", (response) => {
        const url = new URL(response.url());
        if (url.port === "17000") requests.push({ method: response.request().method(), path: url.pathname, status: response.status() });
      });
    };
    context.pages().forEach(watch);
    context.on("page", watch);
    await context.route("**/*", async (route) => {
      const url = new URL(route.request().url());
      if (url.hostname === "127.0.0.1" && ["4100", "17000"].includes(url.port)) return route.continue();
      failures.push(`${route.request().method()} ${url.origin}${url.pathname}`);
      await route.abort("blockedbyclient");
    });
    try {
      await use();
      expect(failures, "Browser regression must not call external services").toEqual([]);
      expect(pageErrors, "No unhandled browser exceptions").toEqual([]);
    } finally {
      await testInfo.attach("request-statuses", { body: JSON.stringify({ requests, failures, pageErrors }, null, 2), contentType: "application/json" });
    }
  }, { auto: true }],
  application: async ({ page, diagnostics: _diagnostics }, use) => {
    const name = `Workflow E2E ${randomUUID().slice(0, 8)}`;
    await page.goto("/#workspace-applications");
    await page.getByRole("button", { name: "Create application", exact: true }).click();
    const creation = page.locator("article").filter({ has: page.getByRole("heading", { name: "Server-generated identity", exact: true }) });
    await creation.getByRole("textbox", { name: "Display name", exact: true }).fill(name);
    await creation.getByRole("combobox", { name: "Application kind", exact: true }).selectOption({ label: "Workflow Copilot" });
    const created = page.waitForResponse((response) => isEndpoint(response, "/v1/user-workspace/applications", "POST"));
    await page.getByRole("button", { name: "Create and select", exact: true }).click();
    const response = await created;
    expect(response.ok()).toBeTruthy();
    const { record } = await response.json();
    expect(record.application_id).toMatch(/^app_[a-z2-7]+$/);
    const application = { id: record.application_id as string, name };
    await expect(page.getByRole("region", { name: "Application development context" })).toContainText(application.id);
    await use(application);
  },
});

export { expect };

export function isEndpoint(response: Response, path: string, method: string) {
  const url = new URL(response.url());
  return url.origin === "http://127.0.0.1:17000" && url.pathname === path && response.request().method() === method;
}

export const draftRoute = "/v1/user-workspace/workflow-drafts";

export function designer(page: Page) {
  return page.getByRole("region", { name: "Workflow draft designer development workbench", exact: true });
}

export function draftField(page: Page, field: string) {
  return designer(page).locator(".workflow-designer-context > div")
    .filter({ has: page.locator("dt").filter({ hasText: new RegExp(`^${field}$`) }) }).locator("dd");
}

export async function createDraft(page: Page) {
  await page.getByRole("button", { name: "Create executor v0 draft", exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText("content 0 / lifecycle 0");
  const id = (await draftField(page, "Draft").innerText()).trim();
  expect(id).toMatch(/^draft_executor_v0_[a-f0-9-]+$/);
  return id;
}

export async function saveDraft(page: Page, expectedVersion: number) {
  const saved = page.waitForResponse((response) => isEndpoint(response, draftRoute, "POST"));
  await designer(page).getByRole("button", { name: "Save draft", exact: true }).click();
  const response = await saved;
  expect(response.ok()).toBeTruthy();
  await expect(draftField(page, "Version")).toHaveText(`content ${expectedVersion} / lifecycle 1`);
  return response;
}

export async function promptLabel(page: Page) {
  const selection = designer(page).getByRole("combobox", { name: "Inspect node", exact: true });
  if (await selection.inputValue() !== "node_executor_prompt") await selection.selectOption("node_executor_prompt");
  return designer(page).getByRole("textbox", { name: "Label", exact: true });
}

export async function selectApplication(page: Page, application: Application) {
  await page.getByRole("button", { name: `${application.name} workflow_copilot ${application.id} v1`, exact: true }).click();
  await expect(page.getByRole("region", { name: "Application development context" })).toContainText(application.id);
}

export async function openDraft(page: Page, draftId: string) {
  const row = page.getByLabel("Saved draft summaries").locator("article").filter({ hasText: draftId });
  await row.getByRole("button", { name: "打开草案", exact: true }).click();
  await expect(draftField(page, "Draft")).toHaveText(draftId);
  await expect(designer(page).getByRole("button", { name: "Save draft", exact: true })).toBeEnabled();
}

// Wait for the response body and a completed render, not a guessed network delay.
export async function finishResponseRender(page: Page, response: Response) {
  await response.finished();
  await page.evaluate(() => new Promise<void>((accept) => requestAnimationFrame(() => requestAnimationFrame(() => accept()))));
}

export async function holdDraftResponse(page: Page, path: string, method: string, injectFailure = false) {
  let acknowledge: () => void;
  let release: () => void;
  const arrived = new Promise<void>((accept) => { acknowledge = accept; });
  const gate = new Promise<void>((accept) => { release = accept; });
  const handler = async (route: Route) => {
    if (route.request().method() !== method) return route.fallback();
    const response = injectFailure ? null : await route.fetch();
    acknowledge();
    await gate;
    if (response) await route.fulfill({ response });
    else await route.fulfill({ status: 500, contentType: "application/json", body: JSON.stringify({ failure_code: "e2e_delayed_validation_failure" }) });
  };
  const url = `http://127.0.0.1:17000${path}`;
  await page.route(url, handler);
  return {
    arrived,
    async deliver() {
      const received = page.waitForResponse((response) => isEndpoint(response, path, method));
      release!();
      await finishResponseRender(page, await received);
    },
    async dispose() {
      release!();
      await page.unroute(url, handler);
    },
  };
}
