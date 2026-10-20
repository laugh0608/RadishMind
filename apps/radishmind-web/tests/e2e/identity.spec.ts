import { randomUUID } from "node:crypto";
import { test, expect, api, language, security, dialog, authForm, register, login, openSecurity, apiLogin, mutate, holdRequest, passwordsCleared, fillPassword, assertPrivateState, inspectLayout } from "./identity-fixtures";
import type { Page } from "@playwright/test";

test("registration, login pending and authentication failure keep their meaning when language changes", async ({ page, account }, info) => {
  const lang = language(page, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  await page.goto("/");
  await page.getByRole("tab", { name: /^(Register|注册)$/ }).click();
  await authForm(page).getByLabel(/^(Display name|显示名称)$/).fill(account.displayName);
  await authForm(page).getByLabel(/^(Local ID|本地 ID)$/).fill(account.identifier);
  await fillPassword(authForm(page).getByLabel(/^(Password|密码)$/), account.password);
  await lang.switch();
  await expect(authForm(page).getByLabel(/^(Display name|显示名称)$/)).toHaveValue(account.displayName);
  const registration = await holdRequest(page, "/v1/auth/local/register");
  try {
    await authForm(page).locator('button[type="submit"]').click(); await registration.started;
    await lang.switch();
    await expect(authForm(page).locator('button[type="submit"]')).toHaveText(lang.text("Verifying…", "正在验证…"));
    registration.release();
    await expect(page.locator(".local-identity-account-trigger")).toBeVisible();
  } finally { await registration.dispose(); }
  expect(lang.requests.filter(value => value === "POST /v1/auth/local/register")).toHaveLength(1);
  await expect(page.locator(".local-identity-account-trigger strong")).toHaveText(account.displayName);
  await openSecurity(page);
  await security(page).getByRole("button", { name: /^(Sign out|退出登录)$/ }).click();
  await expect(authForm(page)).toBeVisible();
  const held = await holdRequest(page, "/v1/auth/local/login");
  try {
    await login(page, account.identifier, `wrong-${randomUUID()}`); await held.started;
    await lang.switch();
    await expect(authForm(page).locator('button[type="submit"]')).toHaveText(lang.text("Verifying…", "正在验证…"));
    held.release();
    await expect(authForm(page).getByRole("alert")).toContainText("LOCAL_IDENTITY_AUTHENTICATION_FAILED");
  } finally { await held.dispose(); }
  await passwordsCleared(page);
  await lang.switch();
  await expect(authForm(page).getByRole("alert")).toContainText(lang.text("The local ID or password is invalid.", "本地 ID 或密码无效"));
  await inspectLayout(page, page.locator(".local-identity-gateway"), info, "authentication-failure");
  await authForm(page).getByRole("button", { name: /Continue with Radish|通过 Radish 继续/ }).click();
  await expect(authForm(page).getByRole("alert")).toContainText("LOCAL_IDENTITY_OIDC_DISABLED");
  await lang.switch();
  await expect(authForm(page).getByRole("alert")).toContainText(lang.text("Radish sign-in is unavailable in this environment.", "此环境未开放 Radish 登录。"));
  await login(page, account.identifier, account.password);
  await expect(page.locator(".local-identity-account-trigger")).toBeVisible();
  await assertPrivateState(page, account);
});

test("session review preserves exact targets through pending, conflict recovery, bulk and cross-tab revocation", async ({ page, account, playwright }, info) => {
  const lang = language(page, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  await register(page, account);
  const first = await playwright.request.newContext();
  const second = await playwright.request.newContext();
  try {
    const firstLogin = await apiLogin(first, account), secondLogin = await apiLogin(second, account);
    expect(firstLogin.status()).toBe(200); expect(secondLogin.status()).toBe(200);
    const firstSession = (await firstLogin.json()).session.session_id as string;
    const secondSession = (await secondLogin.json()).session.session_id as string;
    const heldDirectory = await holdRequest(page, "/v1/auth/sessions?state=all&limit=100", "GET");
    try {
      await page.locator(".local-identity-account-trigger").click(); await heldDirectory.started;
      await expect(security(page)).toContainText(lang.text("Reading sessions", "正在读取会话"));
      await lang.switch(security(page)); heldDirectory.release();
      await expect(security(page).locator(".local-identity-session-row")).toHaveCount(3);
    } finally { await heldDirectory.dispose(); }
    await inspectLayout(page, security(page), info, "session-directory");
    const row = (id: string) => security(page).locator(".local-identity-session-row").filter({ hasText: id });
    await row(firstSession).getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
    await expect(dialog(page).locator("code")).toHaveText([firstSession]);
    await lang.switch(dialog(page));
    await expect(dialog(page).locator("code")).toHaveText([firstSession]);
    await inspectLayout(page, dialog(page), info, "session-confirmation");
    const held = await holdRequest(page, `/v1/auth/sessions/${firstSession}/revoke`);
    try {
      await dialog(page).getByRole("button", { name: /^(Revoke exact session|撤销此会话)$/ }).click(); await held.started;
      await lang.switch(dialog(page));
      await expect(dialog(page)).toContainText(lang.text("Committing security change…", "正在提交安全变更…"));
      held.release();
      await expect(dialog(page)).toBeHidden();
      await expect(security(page).locator(".local-identity-session-row.is-current")).toBeVisible();
    } finally { await held.dispose(); }
    expect((await first.get(`${api}/v1/auth/session`)).status()).toBe(401);
    const directory = await page.request.get(`${api}/v1/auth/sessions?state=all&limit=100`);
    const canonical = await directory.json();
    expect(canonical.sessions.find((value: { session_id: string }) => value.session_id === firstSession).effective_state).toBe("revoked");
    // Change the second target on the real service while the UI still reviews its old version.
    await row(secondSession).getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
    const target = canonical.sessions.find((value: { session_id: string }) => value.session_id === secondSession);
    expect((await mutate(second, `/v1/auth/sessions/${secondSession}/revoke`, { confirmed: true, expected_record_version: target.record_version })).status()).toBe(200);
    await dialog(page).getByRole("button", { name: /^(Revoke exact session|撤销此会话)$/ }).click();
    await expect(security(page).getByRole("alert")).toContainText("local_identity_session_version_conflict");
    await lang.switch(security(page));
    await expect(security(page).getByRole("alert")).toContainText(lang.text("Session or credential changed", "会话或凭据已变更"));
    expect((await apiLogin(first, account)).status()).toBe(200);
    expect((await apiLogin(second, account)).status()).toBe(200);
    await security(page).getByRole("button", { name: /^(Reload session directory|重新加载会话目录)$/ }).click();
    await expect(security(page).locator(".local-identity-session-row").filter({ has: page.getByRole("button", { name: /^(Review revoke|审查撤销)$/ }) })).toHaveCount(3);
    await security(page).getByRole("button", { name: /^(Review revoke others|审查撤销其他会话)$/ }).click();
    const targets = await dialog(page).locator("code").allTextContents();
    expect(targets).toHaveLength(2);
    await lang.switch(dialog(page));
    expect(await dialog(page).locator("code").allTextContents()).toEqual(targets);
    await dialog(page).getByRole("button", { name: /^(Revoke other sessions|撤销其他会话)$/ }).click();
    await expect(dialog(page)).toBeHidden();
    await expect(security(page)).toContainText(lang.text("No other active sessions in this snapshot.", "此快照中没有其他有效会话。"));
    expect((await first.get(`${api}/v1/auth/session`)).status()).toBe(401);
    expect((await second.get(`${api}/v1/auth/session`)).status()).toBe(401);
    const sibling = await page.context().newPage();
    await sibling.goto("/"); await expect(sibling.locator(".local-identity-account-trigger")).toBeVisible();
    await security(page).locator(".local-identity-session-row.is-current").getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
    await expect(dialog(page)).toContainText(lang.text("you must sign in again", "你需要重新登录"));
    await lang.switch(dialog(page));
    await dialog(page).getByRole("button", { name: /^(Revoke exact session|撤销此会话)$/ }).click();
    await expect(authForm(page)).toBeVisible(); await expect(authForm(sibling)).toBeVisible();
    expect(lang.requests.filter(value => value === "POST /v1/auth/sessions/revoke-others")).toHaveLength(1);
    await assertPrivateState(page, account); await sibling.close();
  } finally { await first.dispose(); await second.dispose(); }
});

async function credentialInput(page: Page, current: string, replacement: string, confirmation = replacement) {
  await fillPassword(security(page).getByLabel(/^(Current password|当前密码)$/), current);
  await fillPassword(security(page).getByLabel(/^(Replacement password|新密码)$/), replacement);
  await fillPassword(security(page).getByLabel(/^(Confirm replacement|确认新密码)$/), confirmation);
  await security(page).getByRole("checkbox").check();
  await security(page).getByRole("button", { name: /^(Review credential rotation|审查凭据轮换)$/ }).click();
}

test("credential review clears secrets on cancel and failure, then rotates and requires both tabs to sign in", async ({ page, account, playwright }, info) => {
  const lang = language(page, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  const replacement = `replacement-${randomUUID()}`, wrong = `wrong-${randomUUID()}`;
  await register(page, account); await openSecurity(page);
  await credentialInput(page, account.password, replacement, wrong);
  await expect(security(page).getByRole("alert")).toContainText(lang.text("does not match", "不一致"));
  await lang.switch(security(page));
  await expect(security(page).getByRole("alert")).toContainText(lang.text("does not match", "不一致"));
  const inputsRetained = await security(page).getByLabel(/^(Replacement password|新密码)$/).evaluate((input, value) => (input as HTMLInputElement).value === value, replacement);
  expect(inputsRetained).toBeTruthy();
  await credentialInput(page, account.password, replacement);
  await passwordsCleared(page);
  const targets = await dialog(page).locator("code").allTextContents();
  await lang.switch(dialog(page));
  expect(await dialog(page).locator("code").allTextContents()).toEqual(targets);
  await expect(dialog(page)).toContainText(lang.text("old credential and all sessions remain unchanged", "原凭据和所有会话均保持不变"));
  await inspectLayout(page, dialog(page), info, "credential-confirmation");
  await dialog(page).getByRole("button", { name: /^(Cancel and clear input|取消并清除输入)$/ }).click();
  await passwordsCleared(page);
  expect(lang.requests.filter(value => value === "POST /v1/auth/local/credential/rotate")).toHaveLength(0);
  await credentialInput(page, wrong, replacement);
  await dialog(page).getByRole("button", { name: /^(Rotate and revoke|轮换并撤销)$/ }).click();
  await expect(security(page).getByRole("alert")).toContainText("local_identity_credential_current_invalid");
  await passwordsCleared(page); await lang.switch(security(page));
  await expect(security(page).getByRole("alert")).toContainText(lang.text("Password verification rejected", "密码验证未通过"));
  const other = await playwright.request.newContext();
  try {
    expect((await apiLogin(other, account)).status()).toBe(200);
    await security(page).getByRole("button", { name: /^(Reload session directory|重新加载会话目录)$/ }).click();
    await expect(security(page).locator(".local-identity-session-row")).toHaveCount(2);
    const sibling = await page.context().newPage();
    await sibling.goto("/"); await expect(sibling.locator(".local-identity-account-trigger")).toBeVisible();
    await credentialInput(page, account.password, replacement);
    await expect(dialog(page).locator("code")).toHaveCount(2);
    const held = await holdRequest(page, "/v1/auth/local/credential/rotate");
    try {
      await dialog(page).getByRole("button", { name: /^(Rotate and revoke|轮换并撤销)$/ }).click(); await held.started;
      await passwordsCleared(page); await lang.switch(security(page));
      await expect(security(page).getByRole("status").filter({ hasText: lang.text("Committing security change…", "正在提交安全变更…") })).toBeVisible();
      held.release();
      await expect(authForm(page)).toBeVisible(); await expect(authForm(sibling)).toBeVisible();
    } finally { await held.dispose(); }
    expect((await other.get(`${api}/v1/auth/session`)).status()).toBe(401);
    expect((await apiLogin(other, account)).status()).toBe(401);
    await login(page, account.identifier, replacement);
    await expect(page.locator(".local-identity-account-trigger")).toBeVisible();
    await expect(sibling.locator(".local-identity-account-trigger")).toBeVisible();
    await openSecurity(page);
    await expect(security(page)).toContainText(lang.text("No other active sessions in this snapshot.", "此快照中没有其他有效会话。"));
    expect(lang.requests.filter(value => value === "POST /v1/auth/local/credential/rotate")).toHaveLength(2);
    await assertPrivateState(page, account, [replacement, wrong]); await sibling.close();
  } finally { await other.dispose(); }
});
