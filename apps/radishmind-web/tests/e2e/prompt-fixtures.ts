import { uiText } from "./ui-language";
import type { Page, Locator } from "@playwright/test";
import { expect, isEndpoint, selectApplication, type Application } from "./workflow-fixtures";

export const templatePath = "/v1/user-workspace/prompt-application-templates";
export const draftPath = "/v1/user-workspace/application-drafts";
export const candidatePath = "/v1/user-workspace/application-publish-candidates";
export const sessionPath = "/v1/user-workspace/application-sessions";
export const template = (page: Page) => page.locator("#prompt-application-template-workspace");
export const configuration = (page: Page) => page.locator("#application-configuration-draft");
export const session = (page: Page) => page.getByRole("region", { name: uiText(page, "Prompt Application Session v2"), exact: true });
export const artifacts = (page: Page) => session(page).getByRole("region", { name: uiText(page, "Application result artifacts"), exact: true });

export const diagnosisSchema = {
  type: "object", additionalProperties: false,
  properties: {
    diagnosis: { type: "string", additionalProperties: false },
    evidence: { type: "array", additionalProperties: false, items: { type: "string", additionalProperties: false } },
    next_checks: { type: "array", additionalProperties: false, items: { type: "string", additionalProperties: false } },
    missing_context: { type: "array", additionalProperties: false, items: { type: "string", additionalProperties: false } },
    uncertainty: { type: "string", additionalProperties: false },
  },
  required: ["diagnosis", "evidence", "missing_context", "next_checks", "uncertainty"],
};
export const diagnosisOutput = {
  diagnosis: "Synthetic timeout diagnosis", evidence: ["Synthetic upstream timeout"],
  next_checks: ["Check synthetic upstream health"], missing_context: ["Synthetic request trace"],
  uncertainty: "Synthetic evidence is incomplete",
};

export async function navigate(page: Page, anchor: string) {
  await page.locator(`a[href="#${anchor}"]:visible`).first().click();
  await expect(page.locator(`#${anchor}`)).toBeVisible();
}

export async function submit(page: Page, button: Locator, path: string, allowFailure = false) {
  const pending = page.waitForResponse((response) => isEndpoint(response, path, "POST"));
  await button.click();
  const response = await pending;
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  if (!allowFailure) expect(body.failure_code ?? "").toBe("");
  return body;
}

export async function publishAndActivate(page: Page, applicationId: string, replace = false) {
  await navigate(page, "application-publish-review");
  const review = page.locator("#application-publish-review");
  await review.getByRole("button", { name: uiText(page, "Load saved drafts"), exact: true }).click();
  const candidates = review.getByRole("combobox", { name: uiText(page, "Saved valid draft"), exact: true });
  await expect(candidates.locator("option:not([value=''])")).toHaveCount(1);
  await candidates.selectOption({ index: 1 });
  const created = await submit(page, review.getByRole("button", { name: uiText(page, "Create immutable candidate"), exact: true }), candidatePath);
  await review.getByRole("button", { name: uiText(page, "Read exact source"), exact: true }).click();
  await expect(review.locator(".prompt-template-source")).toContainText("{{ question }}");
  const reviewedContract = JSON.parse(await review.getByLabel(uiText(page, "审查输出契约")).innerText());
  expect(reviewedContract.jsonSchema).toEqual(diagnosisSchema);
  await review.getByRole("textbox", { name: uiText(page, "Review reason"), exact: true }).fill("Reviewed isolated browser regression template and exact configuration.");
  await submit(page, review.getByRole("button", { name: uiText(page, "Record review decision"), exact: true }), `${candidatePath}/${created.candidate.candidate_id}/reviews`);
  await navigate(page, "prompt-application-runtime-assignment");
  const assignment = page.locator("#prompt-application-runtime-assignment");
  if (replace) {
    const loaded = page.waitForResponse((response) => isEndpoint(response, `/v1/user-workspace/applications/${applicationId}/prompt-runtime-assignment/events`, "GET"));
    await assignment.getByRole("button", { name: uiText(page, "重读 assignment 与事件"), exact: true }).click();
    await loaded;
    await assignment.getByRole("combobox", { name: uiText(page, "CAS action"), exact: true }).selectOption("replace");
  }
  return submit(page, assignment.getByRole("button", { name: uiText(page, "记录显式 runtime 决策"), exact: true }), `/v1/user-workspace/applications/${applicationId}/prompt-runtime-assignment/decisions`);
}

export async function loadModelCatalog(page: Page) {
  await navigate(page, "workspace-api-keys");
  const keys = page.locator("#workspace-api-keys");
  await keys.locator("summary").filter({ hasText: uiText(page, "Issue new credential") }).click();
  await keys.getByRole("textbox", { name: uiText(page, "Display name"), exact: true }).fill("Isolated Prompt catalog access");
  await keys.getByRole("button", { name: uiText(page, "Issue API key"), exact: true }).click();
  await keys.getByRole("button", { name: uiText(page, "Use in Playground"), exact: true }).click();
  const playground = page.locator("#model-gateway-playground");
  await playground.getByRole("button", { name: uiText(page, "Load models"), exact: true }).click();
  await expect(playground.getByRole("combobox", { name: uiText(page, "Validated model"), exact: true })).toBeVisible();
}

export async function prepareApplication(page: Page, application: Application) {
  const applicationId = application.id;
  await loadModelCatalog(page);
  await navigate(page, "application-configuration-draft");
  await configuration(page).getByRole("combobox", { name: uiText(page, "Default protocol"), exact: true }).selectOption("chat_completions");
  await configuration(page).getByRole("combobox", { name: uiText(page, "Default model"), exact: true }).selectOption("profile:prompt-e2e");
  const draft = await submit(page, configuration(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }), draftPath);
  await navigate(page, "prompt-application-template-workspace");
  await template(page).locator(".prompt-template-message textarea").nth(0).fill("{{ tone }}");
  await template(page).locator(".prompt-template-message textarea").nth(1).fill("{{ question }}");
  await template(page).getByRole("combobox", { name: uiText(page, "Kind"), exact: true }).selectOption("json_object");
  const schemaEditor = template(page).getByRole("textbox", { name: uiText(page, "输出 JSON Schema"), exact: true });
  await schemaEditor.fill(JSON.stringify(diagnosisSchema, null, 2));
  const saved = await submit(page, template(page).getByRole("button", { name: uiText(page, "Save with CAS"), exact: true }), templatePath);
  expect(saved.draft.output_contract.json_schema).toEqual(diagnosisSchema);
  // Reload from the persisted draft before creating a version, not from editor memory.
  await page.reload();
  await selectApplication(page, application);
  await navigate(page, "prompt-application-template-workspace");
  await template(page).getByRole("button", { name: uiText(page, "Refresh"), exact: true }).first().click();
  await template(page).locator(".prompt-template-summary").filter({ hasText: saved.draft.template_id }).click();
  await expect(schemaEditor).toHaveValue(/diagnosis/);
  expect(JSON.parse(await schemaEditor.inputValue())).toEqual(diagnosisSchema);
  const version = await submit(page, template(page).getByRole("button", { name: uiText(page, "Create immutable version"), exact: true }), `${templatePath}/${saved.draft.template_id}/versions`);
  expect(version.version.output_contract.kind).toBe("json_object");
  expect(version.version.output_contract.json_schema).toEqual(diagnosisSchema);
  expect(JSON.parse(await template(page).getByLabel(uiText(page, "不可变版本输出契约")).innerText()).jsonSchema).toEqual(diagnosisSchema);
  await template(page).getByRole("button", { name: uiText(page, "Load drafts"), exact: true }).click();
  await template(page).getByRole("combobox", { name: uiText(page, "Valid Prompt Application draft"), exact: true }).selectOption(draft.draft.draft_id);
  await template(page).getByRole("combobox", { name: uiText(page, "Immutable template version"), exact: true }).selectOption("1");
  await submit(page, template(page).getByRole("button", { name: uiText(page, "Bind and open Publish Review"), exact: true }), `/v1/user-workspace/application-configuration-drafts/${draft.draft.draft_id}/prompt-template-binding`);
  await publishAndActivate(page, applicationId);
  await navigate(page, "prompt-application-session");
  const created = await submit(page, session(page).getByRole("button", { name: uiText(page, "Create Session v2"), exact: true }), sessionPath);
  return { draftId: draft.draft.draft_id as string, sessionId: created.session.session_id as string };
}

