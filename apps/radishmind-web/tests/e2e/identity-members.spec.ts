import { randomUUID } from "node:crypto";
import { test, expect, api, language, register, holdRequest, assertPrivateState } from "./identity-fixtures";
import {
  members, invitations, claim, confirmation, adminPath, bootstrap, seed, scopedRequest, registerApi,
  openAdmin, openClaim, enterCode, createInvitation, inspectMemberLayout, assertInvitationPrivacy, switchWorkspace,
} from "./identity-member-fixtures";

const workspaceId = () => `workspace_${randomUUID()}`;
const confirmMutation = /^(Confirm mutation|确认变更)$/;
const clearCode = /^(Done & clear code|完成并清除邀请码)$/;
const verify = /^(Verify and preview|验证并预览)$/;

test("member and role review preserve exact targets through language changes, real conflicts and atomic revocation", async ({ page, account, playwright }, info) => {
  const lang = language(page, info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  const workspace = workspaceId();
  await register(page, account);
  const administrator = await bootstrap(page, workspace);
  const targetClient = await playwright.request.newContext();
  try {
    const target = await registerApi(targetClient);
    await openAdmin(page, "user-directory");
    await expect(members(page).getByRole("button", { name: /^(Review membership revoke|审查成员关系撤销)$/ })).toBeDisabled();
    await members(page).getByRole("button", { name: /^(Add exact user|添加精确用户)$/ }).click();
    await members(page).getByLabel(/^(Exact user_id|精确 user_id)$/).fill("invalid-user");
    await members(page).getByRole("button", { name: /^(Review membership|审查成员关系)$/ }).click();
    await expect(members(page).getByRole("alert")).toContainText("local_identity_input_invalid");
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByRole("alert")).toContainText(lang.text("Enter one exact user_id", "请输入一个精确 user_id"));
    await members(page).getByLabel(/^(Exact user_id|精确 user_id)$/).fill(target.userId);
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByLabel(/^(Exact user_id|精确 user_id)$/)).toHaveValue(target.userId);
    await members(page).getByRole("button", { name: /^(Review membership|审查成员关系)$/ }).click();
    await expect(confirmation(page)).toContainText(target.userId);
    const held = await holdRequest(page, `${adminPath(workspace)}/memberships`);
    try {
      await confirmation(page).getByRole("button", { name: confirmMutation }).click(); await held.started;
      await lang.switch(page.locator(".product-nav-desktop"));
      await expect(confirmation(page).getByRole("button", { name: /Applying…|正在应用…/ })).toBeDisabled();
      await expect(confirmation(page)).toContainText(target.userId);
      held.release();
      await expect(members(page).locator(".local-identity-member-inspector")).toContainText(target.account.displayName);
    } finally { await held.dispose(); }
    await expect(members(page).getByRole("status")).toContainText(lang.text("Membership", "成员关系"));
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByRole("status")).toContainText(lang.text("was created", "已创建"));
    await inspectMemberLayout(page, members(page), info, "member-directory");
    await openAdmin(page, "role-policy");
    await expect(members(page).locator(".local-identity-role-member")).toContainText(target.userId);
    await members(page).locator(".local-identity-role-definitions button").filter({ hasText: "workspace_builder" }).click();
    const expiry = new Date(Date.now() + 86_400_000).toISOString();
    await members(page).getByLabel(/Assignment expiry|角色分配到期时间/).fill(expiry);
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByLabel(/Assignment expiry|角色分配到期时间/)).toHaveValue(expiry);
    await members(page).getByRole("button", { name: /^(Review role assignment|审查角色分配)$/ }).click();
    await expect(confirmation(page)).toContainText(target.userId);
    await lang.switch(page.locator(".product-nav-desktop"));
    await inspectMemberLayout(page, members(page), info, "role-confirmation");
    await confirmation(page).getByRole("button", { name: confirmMutation }).click();
    const assignments = members(page).locator(".local-identity-current-assignments");
    await expect(assignments).toContainText("workspace_builder");
    const detailResponse = await scopedRequest(page.request, workspace, `${adminPath(workspace)}/members/${target.userId}`);
    expect(detailResponse.status()).toBe(200);
    const detail = (await detailResponse.json()).member;
    const assignment = detail.role_assignments.find((row: { role_key: string }) => row.role_key === "workspace_builder");
    const membership = detail.memberships.find((row: { effective: boolean }) => row.effective);
    await assignments.getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
    // Revoke via the real handler while the browser reviews the old version.
    expect((await scopedRequest(page.request, workspace, `${adminPath(workspace)}/role-assignments/${assignment.assignment_id}/revoke`, { confirmed: true, expected_record_version: assignment.record_version })).status()).toBe(200);
    await confirmation(page).getByRole("button", { name: confirmMutation }).click();
    await expect(members(page).getByRole("alert")).toContainText("local_identity_role_assignment_conflict");
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByRole("alert")).toContainText(lang.text("Version conflict", "版本冲突"));
    await members(page).getByRole("button", { name: /^(Reload member detail|重新加载成员详情)$/ }).click();
    await expect(assignments.getByRole("button", { name: /^(Review revoke|审查撤销)$/ })).toHaveCount(0);
    await members(page).locator(".local-identity-role-definitions button").filter({ hasText: "workspace_reader" }).click();
    await members(page).getByRole("button", { name: /^(Review role assignment|审查角色分配)$/ }).click();
    await confirmation(page).getByRole("button", { name: confirmMutation }).click();
    await expect(assignments).toContainText("workspace_reader");
    await openAdmin(page, "user-directory");
    await members(page).getByRole("button", { name: /^(Review membership revoke|审查成员关系撤销)$/ }).click();
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(confirmation(page)).toContainText(membership.membership_id);
    await confirmation(page).getByRole("button", { name: confirmMutation }).click();
    await expect(members(page).getByRole("status")).toContainText(lang.text("revoked atomically: 1", "原子撤销 1 条"));
    const finalResponse = await scopedRequest(page.request, workspace, `${adminPath(workspace)}/members/${target.userId}`);
    const final = (await finalResponse.json()).member;
    expect(final.memberships.every((row: { effective: boolean }) => !row.effective)).toBeTruthy();
    expect(final.role_assignments.every((row: { effective: boolean }) => !row.effective)).toBeTruthy();
    expect(lang.requests.filter(value => value === `POST ${adminPath(workspace)}/memberships`)).toHaveLength(1);
    expect(administrator).not.toBe(target.userId);
    await openAdmin(page, "role-policy");
    const adminAssignment = members(page).locator(".local-identity-current-assignments article").filter({ hasText: "workspace_admin" });
    await adminAssignment.getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
    await confirmation(page).getByRole("button", { name: confirmMutation }).click();
    await expect(members(page).getByRole("alert")).toContainText("local_identity_last_admin_removal_denied");
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(members(page).getByRole("alert")).toContainText(lang.text("Administrator protection", "管理员保护"));
    await assertPrivateState(page, account);
  } finally { await targetClient.dispose(); }
});

test("invitation handoff and atomic claim remain private and preserve confirmation while switching languages", async ({ page, account, browser }, info) => {
  const initial = info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const lang = language(page, initial);
  const workspace = workspaceId();
  await register(page, account); await bootstrap(page, workspace); await openAdmin(page, "workspace-invitations");
  const panel = invitations(page);
  await panel.getByRole("button", { name: /^(Create invitation|创建邀请)$/ }).click();
  await panel.getByRole("combobox", { name: /^(Role|角色)/ }).selectOption("workspace_builder");
  await panel.getByRole("combobox", { name: /^(TTL|有效期)/ }).selectOption("24h");
  await panel.getByRole("checkbox").check();
  await lang.switch(page.locator(".product-nav-desktop"));
  await expect(panel.getByRole("combobox", { name: /^(Role|角色)/ })).toHaveValue("workspace_builder");
  await expect(panel.getByRole("checkbox")).toBeChecked();
  await inspectMemberLayout(page, panel, info, "invitation-create");
  const held = await holdRequest(page, `${adminPath(workspace)}/invitations`);
  try {
    await panel.getByRole("button", { name: /^(Confirm & create invitation|确认并创建邀请)$/ }).click(); await held.started;
    await lang.switch(page.locator(".product-nav-desktop"));
    await expect(panel.getByRole("button", { name: /^(Creating…|正在创建…)$/ })).toBeDisabled();
    held.release();
    await expect(panel.getByRole("textbox", { name: /^(One-time invitation code|一次性邀请码)$/ })).toBeVisible();
  } finally { await held.dispose(); }
  const codeField = panel.getByRole("textbox", { name: /^(One-time invitation code|一次性邀请码)$/ });
  const code = await codeField.inputValue();
  await lang.switch(page.locator(".product-nav-desktop"));
  expect(await codeField.evaluate((field, expected) => (field as HTMLTextAreaElement).value === expected, code)).toBeTruthy();
  await inspectMemberLayout(page, panel, info, "invitation-handoff");
  await assertInvitationPrivacy(page, [code]);
  const claimantContext = await browser.newContext({ locale: initial, viewport: { width: 1440, height: 900 } });
  try {
    const claimant = await claimantContext.newPage();
    const target = await registerApi(claimantContext.request);
    await claimant.goto("/"); await expect(claimant.locator(".local-identity-account-trigger")).toBeVisible();
    const claimantLang = language(claimant, initial);
    await openClaim(claimant); await enterCode(claimant, code);
    await claimantLang.switch(claim(claimant));
    expect(await claim(claimant).locator('input[type="password"]').evaluate((field, expected) => (field as HTMLInputElement).value === expected, code)).toBeTruthy();
    const preview = await holdRequest(claimant, "/v1/auth/workspace-invitations/preview");
    try {
      await claim(claimant).getByRole("button", { name: verify }).click(); await preview.started;
      await claimantLang.switch(claim(claimant)); preview.release();
      await expect(claim(claimant).locator(".invitation-preview")).toContainText(workspace);
    } finally { await preview.dispose(); }
    await claim(claimant).getByRole("checkbox").check();
    await claimantLang.switch(claim(claimant));
    await expect(claim(claimant).getByRole("checkbox")).toBeChecked();
    await inspectMemberLayout(claimant, claim(claimant), info, "invitation-preview");
    const claimRequest = await holdRequest(claimant, "/v1/auth/workspace-invitations/claim");
    try {
      await claim(claimant).getByRole("button", { name: /^(Confirm & claim workspace|确认并认领工作区)$/ }).click(); await claimRequest.started;
      await claimantLang.switch(claim(claimant));
      claimRequest.release();
      await expect(claim(claimant).locator(".invitation-claim-success")).toContainText(workspace);
    } finally { await claimRequest.dispose(); }
    // Claim success starts a required session refresh. Establish its completion
    // before measuring requests caused by the next language change.
    await expect(claim(claimant).locator(".invitation-available-workspaces")).toContainText(workspace);
    await claimantLang.switch(claim(claimant));
    await expect(claim(claimant)).toContainText(claimantLang.text("The invitation code has been cleared.", "邀请码已清除"));
    await inspectMemberLayout(claimant, claim(claimant), info, "invitation-claimed");
    await claim(claimant).getByRole("button", { name: /^(Done|完成)$/ }).click();
    await expect(claimant.locator("#desktop-active-workspace-input")).toHaveValue("workspace_demo");
    await switchWorkspace(claimant, workspace);
    const detail = await scopedRequest(page.request, workspace, `${adminPath(workspace)}/members/${target.userId}`);
    expect(detail.status()).toBe(200);
    const canonical = (await detail.json()).member;
    expect(canonical.memberships.filter((row: { effective: boolean }) => row.effective)).toHaveLength(1);
    expect(canonical.role_assignments.filter((row: { effective: boolean }) => row.effective)).toHaveLength(1);
    expect(canonical.role_assignments[0].role_key).toBe("workspace_builder");
    expect(claimantLang.requests.filter(value => value === "POST /v1/auth/workspace-invitations/claim")).toHaveLength(1);
    await openClaim(claimant); await enterCode(claimant, code); await claim(claimant).getByRole("button", { name: verify }).click();
    await expect(claim(claimant).getByRole("alert")).toContainText("workspace_invitation_not_claimable");
    await claimantLang.switch(claim(claimant));
    expect(await claim(claimant).locator('input[type="password"]').evaluate(field => (field as HTMLInputElement).value === "")).toBeTruthy();
    await assertInvitationPrivacy(claimant, [code]); await assertPrivateState(claimant, target.account);
  } finally { await claimantContext.close(); }
  await panel.getByRole("button", { name: clearCode }).click();
  await expect(codeField).toHaveCount(0);
  expect(lang.requests.filter(value => value === `POST ${adminPath(workspace)}/invitations`)).toHaveLength(1);
});

test("invitation expiry, stale revoke, denied scope and sibling invalidation keep bilingual recovery and clear credentials", async ({ page, account, browser }, info) => {
  const initial = info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const lang = language(page, initial);
  const workspace = workspaceId();
  await register(page, account); const userId = await bootstrap(page, workspace); await openAdmin(page, "workspace-invitations");
  const created = await createInvitation(page);
  await invitations(page).getByRole("button", { name: clearCode }).click();
  const sibling = await page.context().newPage();
  await sibling.goto("/"); await expect(sibling.locator(".local-identity-account-trigger")).toBeVisible();
  await switchWorkspace(sibling, workspace); await openAdmin(sibling, "workspace-invitations");
  const handoff = await createInvitation(sibling);
  // The sibling's mutation invalidates this tab; refresh finishes before selection.
  const row = invitations(page).locator(".invitation-rows button").filter({ has: page.locator(`code[title="${created.invitation.invitation_id}"]`) });
  await row.click(); await invitations(page).getByRole("button", { name: /^(Review revoke|审查撤销)$/ }).click();
  await invitations(page).getByRole("checkbox").check();
  await lang.switch(page.locator(".product-nav-desktop"));
  await expect(invitations(page).getByRole("checkbox")).toBeChecked();
  await invitations(page).getByRole("button", { name: /^(Confirm revoke|确认撤销)$/ }).click();
  await expect(invitations(page).getByRole("status").filter({ hasText: /Invitation revoked|邀请已撤销/ })).toBeVisible();
  await expect(invitations(sibling).getByRole("textbox", { name: /^(One-time invitation code|一次性邀请码)$/ })).toHaveCount(0);
  await assertInvitationPrivacy(sibling, [created.code, handoff.code]);
  await sibling.close();
  const stale = await createInvitation(page);
  await invitations(page).getByRole("button", { name: /^(Revoke & recreate|撤销并重新创建)$/ }).click();
  await invitations(page).getByRole("checkbox").check();
  expect((await scopedRequest(page.request, workspace, `${adminPath(workspace)}/invitations/${stale.invitation.invitation_id}/revoke`, { confirmed: true, expected_record_version: stale.invitation.record_version })).status()).toBe(200);
  await invitations(page).getByRole("button", { name: /^(Confirm revoke|确认撤销)$/ }).click();
  await expect(invitations(page).getByRole("alert")).toContainText("workspace_invitation_version_conflict");
  await lang.switch(page.locator(".product-nav-desktop"));
  await expect(invitations(page).getByRole("alert")).toContainText(lang.text("Another operation changed", "已被其他操作修改"));
  const expiring = await seed({ kind: "expired-invitation", userId, workspaceId: workspace });
  await invitations(page).getByRole("button", { name: /^(Expired|已到期)$/ }).click();
  await expect(invitations(page).locator(".invitation-rows")).toContainText("workspace_reader");
  await lang.switch(page.locator(".product-nav-desktop"));
  await inspectMemberLayout(page, invitations(page), info, "invitation-expired-directory");
  const claimantContext = await browser.newContext({ locale: initial, viewport: { width: 1440, height: 900 } });
  try {
    const target = await registerApi(claimantContext.request); const claimant = await claimantContext.newPage();
    await claimant.goto("/"); await expect(claimant.locator(".local-identity-account-trigger")).toBeVisible();
    const claimantLang = language(claimant, initial); await openClaim(claimant);
    for (const code of ["invalid", created.code, expiring.invitation_code]) {
      await enterCode(claimant, code); await claim(claimant).getByRole("button", { name: verify }).click();
      await expect(claim(claimant).getByRole("alert")).toContainText(code === "invalid" ? "workspace_invitation_invalid" : "workspace_invitation_not_claimable");
      await claimantLang.switch(claim(claimant));
      await expect(claim(claimant).getByRole("alert")).toContainText(claimantLang.text(code === "invalid" ? "The invitation code is invalid" : "can no longer be claimed", code === "invalid" ? "邀请码无效" : "已不能认领"));
      expect(await claim(claimant).locator('input[type="password"]').evaluate(field => (field as HTMLInputElement).value === "")).toBeTruthy();
    }
    await inspectMemberLayout(claimant, claim(claimant), info, "invitation-terminal-recovery");
    await claim(claimant).getByRole("button", { name: /Close invitation claim|关闭邀请认领/ }).click();
    await switchWorkspace(claimant, workspace);
    await claimant.evaluate(() => { location.hash = "#admin-workspace-invitations"; });
    await expect(invitations(claimant).getByRole("alert")).toContainText("workspace_membership_denied");
    await claimantLang.switch(claimant.locator(".product-nav-desktop"));
    await expect(invitations(claimant).getByRole("alert")).toContainText(claimantLang.text("no effective membership", "没有此工作区的有效成员资格"));
    await assertPrivateState(claimant, target.account); await assertInvitationPrivacy(claimant, [created.code, expiring.invitation_code]);
  } finally { await claimantContext.close(); }
});
