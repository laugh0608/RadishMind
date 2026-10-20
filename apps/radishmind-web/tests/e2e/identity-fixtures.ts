import { randomUUID } from "node:crypto";
import { test as base, expect, type APIRequestContext, type Page, type Locator, type Route, type TestInfo } from "@playwright/test";

export const api = "http://127.0.0.1:17000";
export type Locale = "en-US" | "zh-CN";
export type Account = { identifier: string; displayName: string; password: string };

export const test = base.extend<{ account: Account }>({
  account: async ({}, use) => {
    const id = randomUUID();
    await use({ identifier: `identity-${id}@example.invalid`, displayName: "Identity 原文 <literal>", password: `test-only-${randomUUID()}` });
  },
});
export { expect };

export function language(page: Page, initial: Locale) {
  let locale = initial;
  const requests: string[] = [];
  page.on("request", request => {
    const url = new URL(request.url());
    if (url.origin === api) requests.push(`${request.method()} ${url.pathname}`);
  });
  return {
    text: (en: string, zh: string) => locale === "zh-CN" ? zh : en,
    locale: () => locale,
    requests,
    async switch(surface?: Locator) {
      const count = requests.length;
      locale = locale === "zh-CN" ? "en-US" : "zh-CN";
      const selector = (surface ?? page).getByRole("combobox", { name: /Interface language|界面语言/ }).first();
      await selector.selectOption(locale);
      await expect(page.locator("html")).toHaveAttribute("lang", locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests.length, "A language change must not repeat identity requests").toBe(count);
    },
  };
}

export const security = (page: Page) => page.locator("section.local-identity-security-surface");
export const dialog = (page: Page) => page.getByRole("alertdialog");
export const authForm = (page: Page) => page.locator("form.local-identity-card");

export async function fillPassword(input: Locator, value: string) {
  await expect(input).toBeEditable();
  await expect(input).toHaveAttribute("type", "password");
  await input.focus();
  // Playwright's fill step title records its value even with trace disabled.
  // The native setter + input event updates React without a secret-bearing title.
  await input.evaluate((element, password) => {
    const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")?.set;
    if (!setter) throw new Error("Password input setter is unavailable.");
    setter.call(element, password);
    element.dispatchEvent(new Event("input", { bubbles: true }));
  }, value);
}

export async function openSecurity(page: Page) {
  await page.locator(".local-identity-account-trigger").click();
  await expect(security(page).locator(".local-identity-session-row.is-current")).toBeVisible();
}

export async function register(page: Page, account: Account) {
  await page.goto("/");
  await page.getByRole("tab", { name: /^(Register|注册)$/ }).click();
  await authForm(page).getByLabel(/^(Display name|显示名称)$/).fill(account.displayName);
  await authForm(page).getByLabel(/^(Local ID|本地 ID)$/).fill(account.identifier);
  await fillPassword(authForm(page).getByLabel(/^(Password|密码)$/), account.password);
  await authForm(page).locator('button[type="submit"]').click();
  await expect(page.locator(".local-identity-account-trigger")).toBeVisible();
}

export async function login(page: Page, identifier: string, password: string) {
  await authForm(page).getByLabel(/^(Local ID|本地 ID)$/).fill(identifier);
  await fillPassword(authForm(page).getByLabel(/^(Password|密码)$/), password);
  await authForm(page).locator('button[type="submit"]').click();
}

export async function apiLogin(client: APIRequestContext, account: Account, password = account.password) {
  return client.post(`${api}/v1/auth/local/login`, {
    headers: { Origin: "http://127.0.0.1:4100", "X-RadishMind-CSRF-Token": "bootstrap" },
    data: { login_identifier: account.identifier, password, return_to: "/" },
  });
}

export async function mutate(client: APIRequestContext, path: string, data: object) {
  const state = await client.storageState();
  const csrf = state.cookies.find(cookie => cookie.name === "radishmind_csrf_dev")?.value;
  if (!csrf) throw new Error("Test session has no CSRF cookie.");
  return client.post(`${api}${path}`, { headers: { Origin: "http://127.0.0.1:4100", "X-RadishMind-CSRF-Token": csrf }, data });
}

export async function holdRequest(page: Page, path: string, method = "POST") {
  let arrived!: () => void, release!: () => void;
  const started = new Promise<void>(resolve => { arrived = resolve; });
  const gate = new Promise<void>(resolve => { release = resolve; });
  const pattern = `${api}${path}`;
  const handler = async (route: Route) => {
    if (route.request().method() !== method) { await route.continue(); return; }
    arrived(); await gate; await route.continue();
  };
  await page.route(pattern, handler);
  return { started, release, async dispose() { release(); await page.unroute(pattern, handler); } };
}

export async function passwordsCleared(page: Page) {
  await expect.poll(() => page.locator('input[type="password"]').evaluateAll(inputs => inputs.every(input => (input as HTMLInputElement).value === ""))).toBeTruthy();
}

export async function assertPrivateState(page: Page, account: Account, extraPasswords: string[] = []) {
  const privateValues = [account.password, account.identifier, ...extraPasswords];
  // Return only a boolean: assertion failures must never print credentials or cookie values.
  const clean = await page.evaluate(values => {
    const persisted = JSON.stringify({ local: { ...localStorage }, session: { ...sessionStorage }, url: location.href });
    return values.every(value => !persisted.includes(value)) && !document.cookie.includes("radishmind_session_dev");
  }, privateValues);
  expect(clean, "Identity inputs and the HttpOnly session must remain outside browser storage and the URL").toBeTruthy();
}

export async function inspectLayout(page: Page, surface: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const clipped = await surface.locator("h2, h3, p, dl > div, .local-identity-session-row, .local-identity-security-targets, footer").evaluateAll(elements => elements.filter(element => element.clientWidth > 0 && element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName));
    expect(clipped, `${name} at ${width}px must fit`).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    if (name === "authentication-failure") {
      const overlaps = await page.locator(".local-identity-gateway-header").evaluate(header => {
        const boxes = [...header.children].map(element => element.getBoundingClientRect());
        return boxes.some((a, index) => boxes.slice(index + 1).some(b =>
          Math.min(a.right, b.right) - Math.max(a.left, b.left) > 1 &&
          Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top) > 1));
      });
      expect(overlaps, `Authentication header controls must not overlap at ${width}px`).toBeFalsy();
    }
    // Only call after clearing password inputs or while the fields are masked.
    await info.attach(`${name}-${width}`, { body: await page.screenshot(), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
