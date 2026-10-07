import { execFile } from "node:child_process";
import { randomUUID } from "node:crypto";
import { readFile, rm } from "node:fs/promises";
import { basename, join, resolve } from "node:path";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import type { APIRequestContext, Page, Locator, TestInfo } from "@playwright/test";
import { api, expect, security, openSecurity, fillPassword, type Account } from "./identity-fixtures";

export const tenant = "tenant_demo";
export const members = (page: Page) => page.locator(".local-identity-administration");
export const invitations = (page: Page) => page.locator(".workspace-invitation-admin");
export const claim = (page: Page) => page.locator(".workspace-invitation-claim");
export const confirmation = (page: Page) => page.locator(".local-identity-mutation-confirmation");
export const adminPath = (workspace: string) => `/v1/admin/local-identity/workspaces/${workspace}`;

export async function seed(input: { kind: "bootstrap" | "expired-invitation"; userId: string; workspaceId: string }) {
  const runtime = process.env.RADISHMIND_IDENTITY_E2E_RUNTIME;
  if (!runtime || !basename(runtime).startsWith("radishmind-workflow-e2e-")) throw new Error("Owned identity runtime is required");
  const output = join(runtime, `identity-seed-${randomUUID()}.json`);
  const platform = fileURLToPath(new URL("../../../../services/platform/", import.meta.url));
  try {
    await promisify(execFile)("go", ["test", "-tags=identity_browser_test", "./internal/httpapi", "-run", "^TestLocalIdentityBrowserSeed$", "-count=1"], {
      cwd: resolve(platform), timeout: 60_000,
      env: { ...process.env, GOCACHE: process.env.GOCACHE ?? "/tmp/radishmind-go-build-cache", RADISHMIND_IDENTITY_BROWSER_TEST: "1",
        RADISHMIND_SQLITE_DEV_DATABASE_PATH: join(runtime, "workflow.db"), RADISHMIND_IDENTITY_BROWSER_SEED_OUTPUT: output,
        RADISHMIND_IDENTITY_BROWSER_SEED: JSON.stringify({ ...input, tenantRef: tenant }) },
    });
    return JSON.parse(await readFile(output, "utf8"));
  } catch {
    // Do not attach seed results, subprocess buffers or credentials to the report.
    throw new Error(`Identity ${input.kind} fixture failed`);
  } finally { await rm(output, { force: true }); }
}

export async function currentUser(page: Page): Promise<string> {
  const response = await page.request.get(`${api}/v1/auth/session`);
  expect(response.status()).toBe(200);
  return (await response.json()).account.user_id;
}

export async function switchWorkspace(page: Page, workspace: string) {
  const navigation = page.locator(".product-nav-desktop");
  await navigation.locator("#desktop-active-workspace-input").fill(workspace);
  await navigation.getByRole("button", { name: /^(Switch workspace|切换工作区)$/ }).click();
  await expect(navigation.locator("#desktop-active-workspace-input")).toHaveValue(workspace);
}

export async function openAdmin(page: Page, surface: "user-directory" | "role-policy" | "workspace-invitations") {
  await page.evaluate(hash => { location.hash = hash; }, `#admin-${surface}`);
  await expect(page.locator(".admin-control-plane-workspace")).toBeVisible();
  if (surface === "workspace-invitations") await expect(invitations(page).getByRole("button", { name: /^(Create invitation|创建邀请)$/ })).toBeEnabled();
  else await expect(members(page).locator(".local-identity-member-rows, .local-identity-role-definitions").first()).toBeVisible();
}

export async function bootstrap(page: Page, workspace: string) {
  const userId = await currentUser(page);
  await seed({ kind: "bootstrap", userId, workspaceId: workspace });
  await page.reload();
  await expect(page.locator(".local-identity-account-trigger")).toBeVisible();
  await switchWorkspace(page, workspace);
  return userId;
}

export async function scopedRequest(client: APIRequestContext, workspace: string, path: string, data?: object) {
  const state = await client.storageState();
  const csrf = state.cookies.find(cookie => cookie.name === "radishmind_csrf_dev")?.value;
  if (!csrf) throw new Error("Scoped test request requires an authenticated session");
  const headers = { Origin: "http://127.0.0.1:4100", "X-RadishMind-CSRF-Token": csrf,
    "X-RadishMind-Active-Tenant": tenant, "X-RadishMind-Active-Workspace": workspace };
  return data ? client.post(`${api}${path}`, { headers, data }) : client.get(`${api}${path}`, { headers });
}

export async function registerApi(client: APIRequestContext): Promise<{ account: Account; userId: string }> {
  const account = { identifier: `member-${randomUUID()}@example.invalid`, displayName: "Member 原文 <literal>", password: `test-only-${randomUUID()}` };
  const response = await client.post(`${api}/v1/auth/local/register`, { headers: { Origin: "http://127.0.0.1:4100", "X-RadishMind-CSRF-Token": "bootstrap" },
    data: { login_identifier: account.identifier, display_name: account.displayName, password: account.password, return_to: "/" } });
  expect(response.status()).toBe(201);
  return { account, userId: (await response.json()).account.user_id };
}

export async function openClaim(page: Page) {
  await openSecurity(page);
  await security(page).getByRole("button", { name: /^(Claim invitation|认领邀请)$/ }).click();
  await expect(claim(page)).toBeVisible();
}

export async function enterCode(page: Page, code: string) {
  await fillPassword(claim(page).locator('input[type="password"]'), code);
}

export async function createInvitation(page: Page, role = "workspace_reader") {
  const panel = invitations(page);
  await panel.getByRole("button", { name: /^(Create invitation|创建邀请)$/ }).click();
  await panel.getByRole("combobox", { name: /^(Role|角色)/ }).selectOption(role);
  await panel.getByRole("checkbox").check();
  const response = page.waitForResponse(response => response.request().method() === "POST" && /\/invitations$/.test(new URL(response.url()).pathname));
  await panel.getByRole("button", { name: /^(Confirm & create invitation|确认并创建邀请)$/ }).click();
  const canonical = await (await response).json();
  const field = panel.getByRole("textbox", { name: /^(One-time invitation code|一次性邀请码)$/ });
  await expect(field).toBeVisible();
  const code = await field.inputValue();
  expect(code === canonical.invitation_code, "One-time handoff must match the creation response").toBeTruthy();
  return { code, invitation: canonical.invitation as { invitation_id: string; record_version: number; role_key: string } };
}

export async function assertInvitationPrivacy(page: Page, codes: string[]) {
  const clean = await page.evaluate(values => {
    const saved = JSON.stringify({ url: location.href, local: { ...localStorage }, session: { ...sessionStorage }, cookie: document.cookie });
    return values.every(value => !saved.includes(value));
  }, codes);
  expect(clean, "Invitation codes must stay outside URL, cookies and browser storage").toBeTruthy();
}

export async function inspectMemberLayout(page: Page, surface: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const clipped = await surface.locator("h2, h3, h4, h5, p, label, dl > div, button, .invitation-progress, .local-identity-member-rows").evaluateAll(elements => elements.filter(element => element.clientWidth > 0 && element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName));
    expect(clipped, `${name} must fit at ${width}px`).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    await info.attach(`${name}-${width}`, { body: await surface.screenshot({ mask: [page.locator(".invitation-code-handoff textarea"), page.locator('input[type="password"]')] }), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
