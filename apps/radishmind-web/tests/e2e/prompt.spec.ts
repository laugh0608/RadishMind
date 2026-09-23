import { uiText, setTestLanguage } from "./ui-language";
import { randomUUID } from "node:crypto";
import { type Page, type Locator } from "@playwright/test";
import { test, expect, isEndpoint, selectApplication, type Application } from "./workflow-fixtures";

test.use({ applicationKind: "prompt_application" });

const templatePath = "/v1/user-workspace/prompt-application-templates";
const draftPath = "/v1/user-workspace/application-drafts";
const candidatePath = "/v1/user-workspace/application-publish-candidates";
const sessionPath = "/v1/user-workspace/application-sessions";
const template = (page: Page) => page.locator("#prompt-application-template-workspace");
const configuration = (page: Page) => page.locator("#application-configuration-draft");
const session = (page: Page) => page.getByRole("region", { name: uiText(page, "Prompt Application Session v2"), exact: true });
const artifacts = (page: Page) => session(page).getByRole("region", { name: uiText(page, "Application result artifacts"), exact: true });

const diagnosisSchema = {
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
const diagnosisOutput = {
  diagnosis: "Synthetic timeout diagnosis", evidence: ["Synthetic upstream timeout"],
  next_checks: ["Check synthetic upstream health"], missing_context: ["Synthetic request trace"],
  uncertainty: "Synthetic evidence is incomplete",
};

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

async function loadModelCatalog(page: Page) {
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

async function prepareApplication(page: Page, application: Application) {
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

async function execute(page: Page, sessionId: string, caseId: string, mode = "valid") {
  await session(page).getByRole("textbox", { name: uiText(page, "Turn variables"), exact: true }).fill(JSON.stringify({ question: `E2E_CASE:${caseId}:${mode}`, tone: "E2E_PROMPT_SYSTEM_V1" }));
  await session(page).getByRole("textbox", { name: uiText(page, "client_turn_key"), exact: true }).fill(`e2e-${caseId}`);
  await artifacts(page).getByRole("checkbox").check();
  return submit(page, session(page).getByRole("button", { name: uiText(page, "Execute Prompt turn"), exact: true }), `${sessionPath}/${sessionId}/turns`, true);
}

async function observedCalls(page: Page, caseId: string) {
  const url = process.env.RADISHMIND_E2E_PROVIDER_URL;
  if (!url || new URL(url).hostname !== "127.0.0.1") throw new Error("Missing isolated Prompt provider fixture");
  const response = await page.request.get(`${url}/observations/${caseId}`);
  expect(response.ok()).toBeTruthy();
  return response.json();
}

test("Prompt schema editor preserves drafts across kind switches and blocks invalid versions", async ({ page, application }, testInfo) => {
  await expect(page.getByRole("region", { name: uiText(page, "Application development context") })).toContainText(application.id);
  await navigate(page, "application-configuration-draft");
  await navigate(page, "prompt-application-template-workspace");
  const kind = template(page).getByRole("combobox", { name: uiText(page, "Kind"), exact: true });
  const editor = template(page).getByRole("textbox", { name: uiText(page, "输出 JSON Schema"), exact: true });
  const save = template(page).getByRole("button", { name: uiText(page, "Save with CAS"), exact: true });
  const version = template(page).getByRole("button", { name: uiText(page, "Create immutable version"), exact: true });
  await kind.selectOption("json_object");
  for (const source of ["{", JSON.stringify({ ...diagnosisSchema, required: ["undeclared"] }), JSON.stringify({ ...diagnosisSchema, $ref: "unsupported" })]) {
    await editor.fill(source);
    await expect(editor).toHaveAttribute("aria-invalid", "true");
    await expect(save).toBeDisabled();
    await expect(version).toBeDisabled();
    await kind.selectOption("text");
    await expect(save).toBeEnabled();
    await kind.selectOption("json_object");
    await expect(editor).toHaveValue(source);
    await expect(save).toBeDisabled();
  }
  await editor.fill(JSON.stringify(diagnosisSchema, null, 2));
  await kind.selectOption("text");
  const textDraft = await submit(page, save, templatePath);
  expect(textDraft.draft.output_contract.json_schema).toBeUndefined();
  await kind.selectOption("json_object");
  await expect(editor).toHaveValue(JSON.stringify(diagnosisSchema, null, 2));
  const saved = await submit(page, save, templatePath);
  expect(saved.draft.draft_version).toBe(2);
  await editor.fill("{");
  await template(page).getByRole("button", { name: uiText(page, "Validate"), exact: true }).click();
  await expect(version).toBeDisabled();
  await editor.fill(JSON.stringify(diagnosisSchema, null, 2));
  const corrected = await submit(page, save, templatePath);
  expect(corrected.draft.draft_version).toBe(3);
  let releaseSave!: () => void;
  let saveArrived!: () => void;
  const heldSave = new Promise<void>((resolve) => { releaseSave = resolve; });
  const saveReady = new Promise<void>((resolve) => { saveArrived = resolve; });
  await page.route(`**${templatePath}`, async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    const response = await route.fetch();
    saveArrived();
    await heldSave;
    await route.fulfill({ response });
  });
  const lateSave = submit(page, save, templatePath);
  await saveReady;
  await editor.fill("{");
  releaseSave();
  await lateSave;
  await page.unroute(`**${templatePath}`);
  await expect(save).toBeDisabled();
  await expect(editor).toHaveValue("{");
  await expect(version).toBeDisabled();
  await editor.fill(JSON.stringify(diagnosisSchema, null, 2));
  expect((await submit(page, save, templatePath)).draft.draft_version).toBe(5);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await editor.scrollIntoViewIfNeeded();
    await expect.poll(() => editor.evaluate((element) => {
      const bounds = element.getBoundingClientRect();
      const parent = element.closest("fieldset")!.getBoundingClientRect();
      const textOverflow: object[] = [];
      if (document.documentElement.scrollWidth > innerWidth) {
        const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
        for (let node = walker.nextNode(); node; node = walker.nextNode()) {
          const range = document.createRange(); range.selectNodeContents(node);
          const box = range.getBoundingClientRect();
          if (box.width && box.right + scrollX > innerWidth + 1) {
            const item = node.parentElement!;
            textOverflow.push({ tag: item.tagName, className: item.className, right: box.right + scrollX,
              parents: [item.parentElement, item.parentElement?.parentElement, item.parentElement?.parentElement?.parentElement].map(parent => parent ? `${parent.tagName}#${parent.id}.${parent.className}` : "") });
          }
        }
      }
      return {
        textOverflow: textOverflow.slice(0, 15),
        withinParent: bounds.width > 0 && bounds.left >= parent.left && bounds.right <= parent.right,
        pageWidth: document.documentElement.scrollWidth,
        viewport: window.innerWidth,
        overflowing: document.documentElement.scrollWidth <= innerWidth ? [] : [...document.querySelectorAll("body *")].flatMap(item => {
          const box = item.getBoundingClientRect();
          return box.width && box.right + scrollX > innerWidth + 1 ? [{ tag: item.tagName, className: item.className, parents: [item.parentElement, item.parentElement?.parentElement, item.parentElement?.parentElement?.parentElement].map(parent => parent ? `${parent.tagName}#${parent.id}.${parent.className}` : ""), right: box.right + scrollX }] : [];
        }).slice(0, 15),
      };
    })).toMatchObject({ withinParent: true, pageWidth: width, viewport: width, overflowing: [], textOverflow: [] });
    await template(page).screenshot({ path: testInfo.outputPath(`prompt-schema-${width}.png`) });
  }
  // A different surface must discard unsaved source, including the hidden text-mode scratch.
  await kind.selectOption("text");
  await navigate(page, "application-configuration-draft");
  await navigate(page, "prompt-application-template-workspace");
  await kind.selectOption("json_object");
  expect(JSON.parse(await editor.inputValue()).properties).toEqual({});
});

test("Prompt saves a canonical result and restores the exact artifact after refresh", async ({ page, application }, testInfo) => {
  const { sessionId } = await prepareApplication(page, application);
  const caseId = randomUUID();
  const result = await execute(page, sessionId, caseId);
  expect(result.failure_code ?? "").toBe("");
  expect(result.turn.status).toBe("succeeded");
  expect(JSON.parse(result.prompt_output)).toEqual(diagnosisOutput);
  expect(result.result_artifact_failure_code ?? "").toBe("");
  const artifactId = result.result_artifact.artifact_id;
  const runId = result.turn.run_ref.run_id;
  expect(artifactId).toMatch(/^appres_/);
  await expect(artifacts(page).getByRole("checkbox")).not.toBeChecked();
  await page.reload();
  await selectApplication(page, application);
  await expect(session(page)).toBeVisible();
  await session(page).getByLabel(uiText(page, "Active Prompt Session v2 records")).getByRole("button").filter({ hasText: sessionId }).click();
  await expect(session(page).locator(".application-publish-snapshot pre")).toHaveText(uiText(page, "(没有当前 turn output)"));
  await expect(session(page).getByRole("textbox", { name: uiText(page, "Turn variables"), exact: true })).not.toHaveValue(new RegExp(caseId));
  await artifacts(page).getByRole("button", { name: uiText(page, "Refresh artifacts"), exact: true }).click();
  await artifacts(page).getByRole("button").filter({ hasText: artifactId }).click();
  expect(JSON.parse(await artifacts(page).locator(".application-result-artifact-inspector pre").innerText())).toEqual(diagnosisOutput);
  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const openRun = artifacts(page).getByRole("button", { name: uiText(page, "Open exact run"), exact: true });
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
  await artifacts(page).getByRole("button", { name: uiText(page, "Open exact run"), exact: true }).click();
  const runs = page.getByRole("region", { name: uiText(page, "runs workflow review owner"), exact: true });
  await expect(runs).toContainText(runId);
  await expect(runs).toContainText("workflow_run_record.v6");
  await expect(runs).not.toContainText(caseId);
  expect(await observedCalls(page, caseId)).toEqual([{ caseId, accepted: true, mode: "valid" }]);
});

for (const mode of ["invalid", "missing"]) test(`Prompt rejects ${mode} output without saving or retrying`, async ({ page, application }) => {
  const { sessionId } = await prepareApplication(page, application);
  const caseId = randomUUID();
  const result = await execute(page, sessionId, caseId, mode);
  expect(result.failure_code).toBe("prompt_invocation_output_contract_failed");
  expect(result.turn.status).toBe("failed");
  expect(result.result_artifact).toBeUndefined();
  expect(result.result_artifact_failure_code).toBeUndefined();
  expect(result.prompt_output ?? "").toBe("");
  await expect(session(page)).toContainText("prompt_invocation_output_contract_failed");
  await expect(session(page)).not.toContainText("E2E_INVALID_OUTPUT");
  await artifacts(page).getByRole("button", { name: uiText(page, "Refresh artifacts"), exact: true }).click();
  await expect(artifacts(page).getByLabel(uiText(page, "active result artifacts")).getByRole("button")).toHaveCount(0);
  // Explicit repeat of the same turn key must remain metadata-only and not call the fixture again.
  const replay = await submit(page, session(page).getByRole("button", { name: uiText(page, "Execute Prompt turn"), exact: true }), `${sessionPath}/${sessionId}/turns`, true);
  expect(replay.idempotent_replay).toBe(true);
  expect(await observedCalls(page, caseId)).toEqual([{ caseId, accepted: true, mode }]);
});

test("Prompt configuration drift blocks calls until explicit review and replacement", async ({ page, application }) => {
  const { sessionId } = await prepareApplication(page, application);
  // The draft recovery reload deliberately cleared the in-memory credential and catalog.
  await loadModelCatalog(page);
  await navigate(page, "application-configuration-draft");
  await configuration(page).getByRole("button", { name: uiText(page, "Refresh"), exact: true }).click();
  await configuration(page).locator(".application-draft-summary").click();
  await expect(configuration(page).getByRole("combobox", { name: uiText(page, "Default model"), exact: true })).toHaveValue("profile:prompt-e2e");
  await configuration(page).getByRole("textbox", { name: uiText(page, "Description"), exact: true }).fill("Changed reviewed configuration for authority drift regression.");
  await submit(page, configuration(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }), draftPath);
  await navigate(page, "prompt-application-session");
  const blockedId = randomUUID();
  const blocked = await execute(page, sessionId, blockedId);
  expect(blocked.failure_code).toMatch(/authority_changed|candidate_ineligible/);
  await expect(session(page)).toContainText(/authority_changed|candidate_ineligible/);
  expect(await observedCalls(page, blockedId)).toEqual([]);
  await publishAndActivate(page, application.id, true);
  await navigate(page, "prompt-application-session");
  const created = await submit(page, session(page).getByRole("button", { name: uiText(page, "Create Session v2"), exact: true }), sessionPath);
  const recoveredId = randomUUID();
  const recovered = await execute(page, created.session.session_id, recoveredId);
  expect(recovered.failure_code ?? "").toBe("");
  expect(recovered.turn.status).toBe("succeeded");
  expect(await observedCalls(page, recoveredId)).toEqual([{ caseId: recoveredId, accepted: true, mode: "valid" }]);
});

test("Prompt language switching preserves invalid schema, focus and pending saves without business requests", async ({ page, application }, testInfo) => {
  await navigate(page, "application-configuration-draft");
  await navigate(page, "prompt-application-template-workspace");
  const locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next = locale === "zh-CN" ? "en-US" : "zh-CN";
  await template(page).getByRole("combobox", { name: uiText(page, "Kind"), exact: true }).selectOption("json_object");
  const editor = template(page).locator("#prompt-output-schema");
  const source = '{"uncommitted":';
  await editor.fill(source);
  await editor.focus();
  const requests: string[] = [];
  const observe = (request: import("@playwright/test").Request) => {
    if (new URL(request.url()).port === "17000") requests.push(request.method() + " " + new URL(request.url()).pathname);
  };
  page.on("request", observe);
  // A storage event from another tab must update text without moving focus or touching the draft.
  await page.evaluate(value => {
    localStorage.setItem("radishmind.uiLocale.v1", value);
    window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
  }, next);
  setTestLanguage(page, next);
  await expect(page.locator("html")).toHaveAttribute("lang", next);
  await expect(editor).toBeFocused();
  await expect(editor).toHaveValue(source);
  await expect(editor).toHaveAttribute("aria-invalid", "true");
  await expect(page.locator("#prompt-output-schema-error")).toContainText(next === "zh-CN" ? "完整有效的 JSON" : "valid, complete JSON");
  await expect(template(page).getByRole("button", { name: uiText(page, "Save with CAS"), exact: true })).toBeDisabled();
  expect(requests).toEqual([]);
  await editor.fill(JSON.stringify(diagnosisSchema));
  let release!: () => void, arrived!: () => void;
  const hold = new Promise<void>(resolve => { release = resolve; });
  const ready = new Promise<void>(resolve => { arrived = resolve; });
  await page.route(`**${templatePath}`, async route => {
    if (route.request().method() !== "POST") return route.continue();
    const response = await route.fetch(); arrived(); await hold; await route.fulfill({ response });
  });
  const saved = submit(page, template(page).getByRole("button", { name: uiText(page, "Save with CAS"), exact: true }), templatePath);
  await ready;
  const before = [...requests];
  await page.locator(".ui-language-selector select:visible").first().selectOption(locale);
  setTestLanguage(page, locale);
  await expect(page.locator("html")).toHaveAttribute("lang", locale);
  await expect(editor).toHaveValue(JSON.stringify(diagnosisSchema));
  expect(requests).toEqual(before);
  release();
  const result = await saved;
  expect(result.draft.output_contract.json_schema).toEqual(diagnosisSchema);
  await expect(template(page).getByRole("button", { name: uiText(page, "Create immutable version"), exact: true })).toBeEnabled();
  await page.unroute(`**${templatePath}`);
  const savedSource = await editor.inputValue();
  expect(JSON.parse(savedSource)).toEqual(diagnosisSchema);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const menu = page.locator(".product-nav-mobile-menu");
    if (await menu.isVisible()) await menu.locator(":scope > summary").click();
    const selector = page.locator(".ui-language-selector select:visible").first();
    await expect(selector).toHaveValue(locale);
    await selector.focus();
    await expect(selector).toBeFocused();
    await expect(selector).toHaveAccessibleName(locale === "zh-CN" ? "界面语言" : "Interface language");
    if (await menu.isVisible()) await menu.locator(":scope > summary").click();
    await expect(editor).toHaveValue(savedSource);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
    await editor.scrollIntoViewIfNeeded();
    await page.screenshot({ path: testInfo.outputPath(`prompt-${locale}-${width}.png`) });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  const persisted = await page.evaluate(() => localStorage.getItem("radishmind.uiLocale.v1"));
  await page.evaluate(() => {
    const original = Storage.prototype.setItem;
    Storage.prototype.setItem = function (key, value) {
      if (key === "radishmind.uiLocale.v1") throw new DOMException("Storage disabled", "QuotaExceededError");
      return original.call(this, key, value);
    };
  });
  const beforeDeniedWrite = [...requests];
  await page.locator(".ui-language-selector select:visible").first().selectOption(next);
  setTestLanguage(page, next);
  await expect(page.locator("html")).toHaveAttribute("lang", next);
  await expect(page.locator(".ui-language-selector:visible").first()).toContainText(next === "zh-CN" ? "未能保存" : "could not be saved");
  await expect(editor).toHaveValue(savedSource);
  expect(requests).toEqual(beforeDeniedWrite);
  expect(await page.evaluate(() => localStorage.getItem("radishmind.uiLocale.v1"))).toBe(persisted);
  page.off("request", observe);
  expect(application.kind).toBe("prompt_application");
});
