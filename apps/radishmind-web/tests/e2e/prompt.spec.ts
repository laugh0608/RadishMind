import { randomUUID } from "node:crypto";
import { type Page, type Locator } from "@playwright/test";
import { test, expect, isEndpoint, selectApplication } from "./workflow-fixtures";

test.use({ applicationKind: "prompt_application" });

const templatePath = "/v1/user-workspace/prompt-application-templates";
const draftPath = "/v1/user-workspace/application-drafts";
const candidatePath = "/v1/user-workspace/application-publish-candidates";
const sessionPath = "/v1/user-workspace/application-sessions";
const template = (page: Page) => page.locator("#prompt-application-template-workspace");
const configuration = (page: Page) => page.locator("#application-configuration-draft");
const session = (page: Page) => page.getByRole("region", { name: "Prompt Application Session v2", exact: true });
const artifacts = (page: Page) => session(page).getByRole("region", { name: "Application result artifacts", exact: true });

async function navigate(page: Page, anchor: string) {
  await page.locator(`a[href="#${anchor}"]:visible`).first().click();
  await expect(page.locator(`#${anchor}`)).toBeVisible();
}

async function submit(page: Page, button: Locator, path: string, allowFailure = false) {
  const pending = page.waitForResponse((response) => isEndpoint(response, path, "POST"));
  await button.click();
  const response = await pending;
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  if (!allowFailure) expect(body.failure_code ?? "").toBe("");
  return body;
}

async function publishAndActivate(page: Page, applicationId: string, replace = false) {
  await navigate(page, "application-publish-review");
  const review = page.locator("#application-publish-review");
  await review.getByRole("button", { name: "Load saved drafts", exact: true }).click();
  const candidates = review.getByRole("combobox", { name: "Saved valid draft", exact: true });
  await expect(candidates.locator("option:not([value=''])")).toHaveCount(1);
  await candidates.selectOption({ index: 1 });
  const created = await submit(page, review.getByRole("button", { name: "Create immutable candidate", exact: true }), candidatePath);
  await review.getByRole("button", { name: "Read exact source", exact: true }).click();
  await expect(review.locator(".prompt-template-source")).toContainText("{{ question }}");
  await review.getByRole("textbox", { name: "Review reason", exact: true }).fill("Reviewed isolated browser regression template and exact configuration.");
  await submit(page, review.getByRole("button", { name: "Record review decision", exact: true }), `${candidatePath}/${created.candidate.candidate_id}/reviews`);
  await navigate(page, "prompt-application-runtime-assignment");
  const assignment = page.locator("#prompt-application-runtime-assignment");
  if (replace) {
    const loaded = page.waitForResponse((response) => isEndpoint(response, `/v1/user-workspace/applications/${applicationId}/prompt-runtime-assignment/events`, "GET"));
    await assignment.getByRole("button", { name: "重读 assignment 与事件", exact: true }).click();
    await loaded;
    await assignment.getByRole("combobox", { name: "CAS action", exact: true }).selectOption("replace");
  }
  return submit(page, assignment.getByRole("button", { name: "记录显式 runtime 决策", exact: true }), `/v1/user-workspace/applications/${applicationId}/prompt-runtime-assignment/decisions`);
}

async function prepareApplication(page: Page, applicationId: string) {
  await navigate(page, "workspace-api-keys");
  const keys = page.locator("#workspace-api-keys");
  await keys.locator("summary").filter({ hasText: "Issue new credential" }).click();
  await keys.getByRole("textbox", { name: "Display name", exact: true }).fill("Isolated Prompt catalog access");
  await keys.getByRole("button", { name: "Issue API key", exact: true }).click();
  await keys.getByRole("button", { name: "Use in Playground", exact: true }).click();
  const playground = page.locator("#model-gateway-playground");
  await playground.getByRole("button", { name: "Load models", exact: true }).click();
  await expect(playground.getByRole("combobox", { name: "Validated model", exact: true })).toBeVisible();
  await navigate(page, "application-configuration-draft");
  await configuration(page).getByRole("combobox", { name: "Default protocol", exact: true }).selectOption("chat_completions");
  await configuration(page).getByRole("combobox", { name: "Default model", exact: true }).selectOption("profile:prompt-e2e");
  const draft = await submit(page, configuration(page).getByRole("button", { name: "Save draft", exact: true }), draftPath);
  await navigate(page, "prompt-application-template-workspace");
  await template(page).locator(".prompt-template-message textarea").nth(0).fill("{{ tone }}");
  await template(page).locator(".prompt-template-message textarea").nth(1).fill("{{ question }}");
  await template(page).getByRole("combobox", { name: "Kind", exact: true }).selectOption("json_object");
  const saved = await submit(page, template(page).getByRole("button", { name: "Save with CAS", exact: true }), templatePath);
  const version = await submit(page, template(page).getByRole("button", { name: "Create immutable version", exact: true }), `${templatePath}/${saved.draft.template_id}/versions`);
  expect(version.version.output_contract.kind).toBe("json_object");
  await template(page).getByRole("button", { name: "Load drafts", exact: true }).click();
  await template(page).getByRole("combobox", { name: "Valid Prompt Application draft", exact: true }).selectOption(draft.draft.draft_id);
  await template(page).getByRole("combobox", { name: "Immutable template version", exact: true }).selectOption("1");
  await submit(page, template(page).getByRole("button", { name: "Bind and open Publish Review", exact: true }), `/v1/user-workspace/application-configuration-drafts/${draft.draft.draft_id}/prompt-template-binding`);
  await publishAndActivate(page, applicationId);
  await navigate(page, "prompt-application-session");
  const created = await submit(page, session(page).getByRole("button", { name: "Create Session v2", exact: true }), sessionPath);
  return { draftId: draft.draft.draft_id as string, sessionId: created.session.session_id as string };
}

async function execute(page: Page, sessionId: string, caseId: string, mode = "valid") {
  await session(page).getByRole("textbox", { name: "Turn variables", exact: true }).fill(JSON.stringify({ question: `E2E_CASE:${caseId}:${mode}`, tone: "E2E_PROMPT_SYSTEM_V1" }));
  await session(page).getByRole("textbox", { name: "client_turn_key", exact: true }).fill(`e2e-${caseId}`);
  await artifacts(page).getByRole("checkbox").check();
  return submit(page, session(page).getByRole("button", { name: "Execute Prompt turn", exact: true }), `${sessionPath}/${sessionId}/turns`, true);
}

async function observedCalls(page: Page, caseId: string) {
  const url = process.env.RADISHMIND_E2E_PROVIDER_URL;
  if (!url || new URL(url).hostname !== "127.0.0.1") throw new Error("Missing isolated Prompt provider fixture");
  const response = await page.request.get(`${url}/observations/${caseId}`);
  expect(response.ok()).toBeTruthy();
  return response.json();
}

test("Prompt saves a canonical result and restores the exact artifact after refresh", async ({ page, application }, testInfo) => {
  const { sessionId } = await prepareApplication(page, application.id);
  const caseId = randomUUID();
  const result = await execute(page, sessionId, caseId);
  expect(result.failure_code ?? "").toBe("");
  expect(result.turn.status).toBe("succeeded");
  expect(result.prompt_output).toBe("{}");
  expect(result.result_artifact_failure_code ?? "").toBe("");
  const artifactId = result.result_artifact.artifact_id;
  const runId = result.turn.run_ref.run_id;
  expect(artifactId).toMatch(/^appres_/);
  await expect(artifacts(page).getByRole("checkbox")).not.toBeChecked();
  await page.reload();
  await selectApplication(page, application);
  await expect(session(page)).toBeVisible();
  await session(page).getByLabel("Active Prompt Session v2 records").getByRole("button").filter({ hasText: sessionId }).click();
  await expect(session(page).locator(".application-publish-snapshot pre")).toHaveText("(没有当前 turn output)");
  await expect(session(page).getByRole("textbox", { name: "Turn variables", exact: true })).not.toHaveValue(new RegExp(caseId));
  await artifacts(page).getByRole("button", { name: "Refresh artifacts", exact: true }).click();
  await artifacts(page).getByRole("button").filter({ hasText: artifactId }).click();
  await expect(artifacts(page).locator(".application-result-artifact-inspector pre")).toHaveText("{}");
  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const openRun = artifacts(page).getByRole("button", { name: "Open exact run", exact: true });
    await openRun.scrollIntoViewIfNeeded();
    const geometry = await openRun.evaluate((element) => {
      const bounds = element.getBoundingClientRect();
      const panel = element.closest(".application-result-artifact-inspector")!.getBoundingClientRect();
      return {
        withinPanel: bounds.left >= panel.left && bounds.right <= panel.right,
        pointerHits: [0.2, 0.5, 0.8].every((point) => element.contains(document.elementFromPoint(bounds.x + bounds.width * point, bounds.y + bounds.height / 2))),
        pageWidth: document.documentElement.scrollWidth,
      };
    });
    expect(geometry).toEqual({ withinPanel: true, pointerHits: true, pageWidth: width });
    const labelsFit = await artifacts(page).locator(".application-result-artifact-list button span > *").evaluateAll((elements) =>
      elements.every((element) => element.scrollWidth <= element.clientWidth)
    );
    expect(labelsFit, "Artifact identifiers and metadata must not overlap inside the list row").toBe(true);
    const inspectorFits = await artifacts(page).locator(".application-result-artifact-inspector").evaluate((element) =>
      [element, ...element.querySelectorAll("strong, dd, pre")].every((item) => item.scrollWidth <= item.clientWidth)
    );
    expect(inspectorFits, "Exact artifact content must fit the available workbench column").toBe(true);
    await artifacts(page).screenshot({ path: testInfo.outputPath(`prompt-result-${width}.png`) });
  }
  await artifacts(page).getByRole("button", { name: "Open exact run", exact: true }).click();
  const runs = page.getByRole("region", { name: "runs workflow review owner", exact: true });
  await expect(runs).toContainText(runId);
  await expect(runs).toContainText("workflow_run_record.v6");
  await expect(runs).not.toContainText(caseId);
  expect(await observedCalls(page, caseId)).toEqual([{ caseId, accepted: true, mode: "valid" }]);
});

test("Prompt rejects output outside the template contract without saving or retrying", async ({ page, application }) => {
  const { sessionId } = await prepareApplication(page, application.id);
  const caseId = randomUUID();
  const result = await execute(page, sessionId, caseId, "invalid");
  expect(result.failure_code).toBe("prompt_invocation_output_contract_failed");
  expect(result.turn.status).toBe("failed");
  expect(result.result_artifact).toBeUndefined();
  expect(result.result_artifact_failure_code).toBeUndefined();
  expect(result.prompt_output ?? "").toBe("");
  await expect(session(page)).toContainText("prompt_invocation_output_contract_failed");
  await expect(session(page)).not.toContainText("E2E_INVALID_OUTPUT");
  await artifacts(page).getByRole("button", { name: "Refresh artifacts", exact: true }).click();
  await expect(artifacts(page).getByLabel("active result artifacts").getByRole("button")).toHaveCount(0);
  // Explicit repeat of the same turn key must remain metadata-only and not call the fixture again.
  const replay = await submit(page, session(page).getByRole("button", { name: "Execute Prompt turn", exact: true }), `${sessionPath}/${sessionId}/turns`, true);
  expect(replay.idempotent_replay).toBe(true);
  expect(await observedCalls(page, caseId)).toEqual([{ caseId, accepted: true, mode: "invalid" }]);
});

test("Prompt configuration drift blocks calls until explicit review and replacement", async ({ page, application }) => {
  const { sessionId } = await prepareApplication(page, application.id);
  await navigate(page, "application-configuration-draft");
  await configuration(page).getByRole("button", { name: "Refresh", exact: true }).click();
  await configuration(page).locator(".application-draft-summary").click();
  await expect(configuration(page).getByRole("combobox", { name: "Default model", exact: true })).toHaveValue("profile:prompt-e2e");
  await configuration(page).getByRole("textbox", { name: "Description", exact: true }).fill("Changed reviewed configuration for authority drift regression.");
  await submit(page, configuration(page).getByRole("button", { name: "Save draft", exact: true }), draftPath);
  await navigate(page, "prompt-application-session");
  const blockedId = randomUUID();
  const blocked = await execute(page, sessionId, blockedId);
  expect(blocked.failure_code).toMatch(/authority_changed|candidate_ineligible/);
  await expect(session(page)).toContainText(/authority_changed|candidate_ineligible/);
  expect(await observedCalls(page, blockedId)).toEqual([]);
  await publishAndActivate(page, application.id, true);
  await navigate(page, "prompt-application-session");
  const created = await submit(page, session(page).getByRole("button", { name: "Create Session v2", exact: true }), sessionPath);
  const recoveredId = randomUUID();
  const recovered = await execute(page, created.session.session_id, recoveredId);
  expect(recovered.failure_code ?? "").toBe("");
  expect(recovered.turn.status).toBe("succeeded");
  expect(await observedCalls(page, recoveredId)).toEqual([{ caseId: recoveredId, accepted: true, mode: "valid" }]);
});
