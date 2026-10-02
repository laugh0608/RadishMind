import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { identitySecurity as en } from "../src/i18n/locales/en-US/identitySecurity.ts";
import { identitySecurity as zh } from "../src/i18n/locales/zh-CN/identitySecurity.ts";
import { identityFailureMessage } from "../src/features/local-identity/localIdentityMessages.ts";
import { securityDate, securityFailure, securityFailureCopy, securitySuccessMessage, type SecurityFailure, type SecuritySuccess } from "../src/features/local-identity/localIdentitySecurityMessages.ts";
import { LocalIdentitySelfServiceSecurityError } from "../src/features/local-identity/localIdentitySelfServiceSecurityConsumer.ts";

async function messages() {
  const instance = createUiI18n();
  await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "identity", { security: en }, true);
  instance.addResourceBundle("zh-CN", "identity", { security: zh }, true);
  return instance;
}

test("stored security failures retranslate without retaining server diagnostics or changing their code", async () => {
  const instance = await messages();
  const cases = [
    "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED", "LOCAL_IDENTITY_CSRF_INVALID",
    "local_identity_session_recent_authentication_required", "local_identity_session_version_conflict",
    "local_identity_credential_unavailable", "local_identity_credential_current_invalid",
    "local_identity_credential_policy_rejected", "local_identity_service_unavailable",
    "local_identity_response_invalid", "unknown_failure",
  ];
  const failures: SecurityFailure[] = cases.map(code => securityFailure(new LocalIdentitySelfServiceSecurityError(409, code, "PRIVATE_DIAGNOSTIC")));
  failures.push({ kind: "conflict", code: "local_identity_session_selection_stale" }, { kind: "conflict", code: "local_identity_credential_review_stale" });
  const before = structuredClone(failures);
  const english = failures.map(failure => securityFailureCopy(instance.getFixedT(null, "identity"), failure));
  await instance.changeLanguage("zh-CN");
  failures.forEach((failure, index) => {
    const copy = securityFailureCopy(instance.getFixedT(null, "identity"), failure);
    assert.notEqual(copy.title, english[index].title);
    assert.notEqual(copy.message, english[index].message);
    assert.doesNotMatch(JSON.stringify([copy, failure]), /PRIVATE_DIAGNOSTIC/);
    assert.match(copy.message, /[\u4e00-\u9fff]/u);
  });
  assert.deepEqual(failures, before);
  assert.deepEqual(securityFailure(new Error("PRIVATE_DIAGNOSTIC")), { kind: "failed", code: "local_identity_request_failed" });
});

test("security success feedback keeps counts and policy identifiers across a locale switch", async () => {
  const instance = await messages();
  const feedback: SecuritySuccess[] = [
    { kind: "currentRevoked" }, { kind: "exactRevoked" },
    { kind: "othersRevoked", count: 0 }, { kind: "othersRevoked", count: 1 }, { kind: "othersRevoked", count: 1234 },
    { kind: "credentialRevoked", count: 1 }, { kind: "credentialRevoked", count: 2 },
    { kind: "credentialClosed", policyVersion: "local_credential_policy.v1" },
  ];
  const before = structuredClone(feedback);
  const english = feedback.map(value => securitySuccessMessage(instance.getFixedT(null, "identity"), value, "en-US"));
  assert.match(english[2], /0 other active sessions/);
  assert.match(english[3], /1 other active session revoked/);
  assert.match(english[4], /1,234 other active sessions/);
  await instance.changeLanguage("zh-CN");
  feedback.forEach((value, index) => {
    const copy = securitySuccessMessage(instance.getFixedT(null, "identity"), value, "zh-CN");
    assert.notEqual(copy, english[index]);
    assert.doesNotMatch(copy, /security\.|\{\{/);
    if (value.kind === "credentialClosed") assert.ok(copy.includes(value.policyVersion));
  });
  assert.deepEqual(feedback, before);
});

test("security dates remain UTC and invalid input never becomes a current date", async () => {
  const instance = await messages();
  const raw = "2026-10-02T23:59:00Z";
  const english = securityDate(instance.getFixedT(null, "identity"), raw, "en-US");
  assert.match(english, /Oct 2, 2026.*11:59 PM UTC/);
  assert.equal(securityDate(instance.getFixedT(null, "identity"), "invalid", "en-US"), "Unknown");
  await instance.changeLanguage("zh-CN");
  assert.match(securityDate(instance.getFixedT(null, "identity"), raw, "zh-CN"), /2026年10月2日.*23:59 UTC/);
  assert.equal(securityDate(instance.getFixedT(null, "identity"), "", "zh-CN"), "未知");
  assert.equal(raw, "2026-10-02T23:59:00Z");
});

test("authentication feedback localizes known codes and uses a safe unknown-code fallback", async () => {
  const instance = await messages();
  const codes = ["LOCAL_IDENTITY_AUTHENTICATION_FAILED", "LOCAL_IDENTITY_ACCOUNT_CHANGE_REQUIRES_RECENT_AUTHENTICATION", "LOCAL_IDENTITY_LAST_LOGIN_METHOD_REMOVAL_DENIED", "LOCAL_IDENTITY_AUTHENTICATION_REQUIRED", "LOCAL_IDENTITY_CSRF_INVALID", "LOCAL_IDENTITY_ORIGIN_FORBIDDEN", "LOCAL_IDENTITY_ACCOUNT_LINK_REQUIRES_RECENT_AUTHENTICATION", "LOCAL_IDENTITY_PAYLOAD_INVALID", "LOCAL_IDENTITY_RETURN_TARGET_INVALID", "LOCAL_IDENTITY_ACCOUNT_CONFLICT", "LOCAL_IDENTITY_ALREADY_AUTHENTICATED", "LOCAL_IDENTITY_OIDC_DISABLED", "LOCAL_IDENTITY_EXTERNAL_IDENTITY_UNBOUND", "LOCAL_IDENTITY_EXTERNAL_IDENTITY_CONFLICT", "LOCAL_IDENTITY_EXTERNAL_IDENTITY_VERSION_CONFLICT", "LOCAL_IDENTITY_EXTERNAL_IDENTITY_OWNERSHIP_DENIED", "LOCAL_IDENTITY_OIDC_ADMISSION_DENIED", "LOCAL_IDENTITY_OIDC_STATE_INVALID", "LOCAL_IDENTITY_OIDC_CALLBACK_MISMATCH", "LOCAL_IDENTITY_OIDC_TOKEN_EXCHANGE_FAILED", "LOCAL_IDENTITY_OIDC_IDENTITY_MISMATCH", "LOCAL_IDENTITY_OIDC_PROVIDER_UNAVAILABLE", "unknown_code"];
  const english = codes.map(code => identityFailureMessage(instance.getFixedT(null, "identity"), code));
  await instance.changeLanguage("zh-CN");
  codes.forEach((code, index) => {
    const copy = identityFailureMessage(instance.getFixedT(null, "identity"), code);
    assert.notEqual(copy, english[index]);
    assert.doesNotMatch(copy, /unknown_code/);
  });
});
