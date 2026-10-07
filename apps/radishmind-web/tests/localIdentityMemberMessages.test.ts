import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { identityMembers as membersEn } from "../src/i18n/locales/en-US/identityMembers.ts";
import { identityMembers as membersZh } from "../src/i18n/locales/zh-CN/identityMembers.ts";
import { identityInvitations as invitationsEn } from "../src/i18n/locales/en-US/identityInvitations.ts";
import { identityInvitations as invitationsZh } from "../src/i18n/locales/zh-CN/identityInvitations.ts";
import { identityRoles as rolesEn } from "../src/i18n/locales/en-US/identityRoles.ts";
import { identityRoles as rolesZh } from "../src/i18n/locales/zh-CN/identityRoles.ts";
import { LocalIdentityAdministrationError } from "../src/features/local-identity/localIdentityAdministrationConsumer.ts";
import { memberFailure, memberFailureCopy, memberSuccessMessage, memberDate, type MemberSuccess } from "../src/features/local-identity/localIdentityMemberMessages.ts";
import { invitationFailureMessage, invitationDate } from "../src/features/local-identity/workspaceInvitationMessages.ts";
import { identityRoleCopy } from "../src/features/local-identity/localIdentityRoleMessages.ts";

async function messages() {
  const instance = createUiI18n();
  await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "identity", { members: membersEn, invitations: invitationsEn, roles: rolesEn }, true);
  instance.addResourceBundle("zh-CN", "identity", { members: membersZh, invitations: invitationsZh, roles: rolesZh }, true);
  return instance;
}

test("member failure state discards server prose and recovery strings and retranslates every failure class", async () => {
  const instance = await messages();
  const failures = ["workspace_membership_denied", "workspace_permission_denied", "local_identity_admin_unavailable", "local_identity_membership_conflict", "local_identity_role_assignment_conflict", "local_identity_role_catalog_mismatch", "local_identity_last_admin_removal_denied", "local_identity_self_membership_revoke_denied", "local_identity_recent_authentication_required", "local_identity_response_invalid", "unknown"].map(code => memberFailure(new LocalIdentityAdministrationError(409, code, "PRIVATE_RESPONSE", "PRIVATE_RECOVERY")));
  const before = structuredClone(failures);
  const english = failures.map(failure => memberFailureCopy(instance.getFixedT(null, "identity"), failure));
  await instance.changeLanguage("zh-CN");
  failures.forEach((failure, index) => {
    const translated = memberFailureCopy(instance.getFixedT(null, "identity"), failure);
    assert.notEqual(translated.title, english[index].title);
    assert.notEqual(translated.message, english[index].message);
    assert.doesNotMatch(JSON.stringify([failure, translated]), /PRIVATE_|\{\{|members\./);
  });
  assert.deepEqual(failures, before);
  assert.deepEqual(memberFailure(new Error("PRIVATE_EXCEPTION")), { kind: "unavailable", code: "local_identity_admin_unavailable" });
});

test("invitation guidance uses stable codes in both languages without exposing unknown diagnostics", async () => {
  const instance = await messages();
  const codes = [...Object.keys(invitationsEn.failures), "private-provider-error", "toString", "__proto__"];
  const english = codes.map(code => invitationFailureMessage(instance.getFixedT(null, "identity"), code));
  await instance.changeLanguage("zh-CN");
  codes.forEach((code, index) => {
    const translated = invitationFailureMessage(instance.getFixedT(null, "identity"), code);
    assert.notEqual(translated, english[index]);
    assert.match(translated, /[\u4e00-\u9fff]/u);
    assert.doesNotMatch(translated, /private-provider-error|\{\{|invitations\./);
  });
});

test("member success retains exact references, versions and zero or multiple counts after switching language", async () => {
  const instance = await messages();
  const outcomes: MemberSuccess[] = [
    { kind: "membershipCreated", membershipId: "mbr_exact1234567890" },
    ...[0, 1, 1234].map(count => ({ kind: "membershipRevoked" as const, count })),
    { kind: "roleAssigned", roleKey: "workspace_builder" },
    { kind: "roleRevoked", roleKey: "workspace_reviewer", version: 12 },
  ];
  const before = structuredClone(outcomes);
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale);
    const copies = outcomes.map(outcome => memberSuccessMessage(instance.getFixedT(null, "identity"), outcome, locale));
    assert.ok(copies[0].includes("mbr_exact1234567890"));
    assert.ok(copies[1].includes("0")); assert.ok(copies[2].includes("1")); assert.ok(copies[3].includes("1,234"));
    assert.ok(copies[4].includes("workspace_builder")); assert.ok(copies[5].includes("workspace_reviewer")); assert.ok(copies[5].includes("v12"));
    assert.doesNotMatch(copies.join(" "), /\{\{|success\./);
  }
  assert.deepEqual(outcomes, before);
});

test("member and invitation dates preserve UTC and role descriptions never replace the role key", async () => {
  const instance = await messages();
  const raw = "2026-10-07T23:59:00Z";
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(null, "identity");
    for (const format of [memberDate, invitationDate]) {
      assert.match(format(t, raw, locale), locale === "en-US" ? /Oct 7, 2026.*11:59 PM UTC/ : /2026年10月7日.*23:59 UTC/);
      assert.equal(format(t, "invalid", locale), locale === "en-US" ? "Unknown" : "未知");
    }
    const roleKey = "workspace_admin";
    const role = identityRoleCopy(t, roleKey);
    assert.equal(role.name, locale === "en-US" ? "Workspace administrator" : "工作区管理员");
    assert.equal(roleKey, "workspace_admin");
    assert.equal(identityRoleCopy(t, "toString").name, locale === "en-US" ? "Unrecognized role" : "未识别的角色");
  }
});
