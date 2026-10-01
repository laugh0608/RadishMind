import { uiText, setTestLanguage, draftVersionText } from "./ui-language";
import {
  test, expect, designer, draftField, createDraft, saveDraft, promptLabel, selectApplication,
  openDraft, isEndpoint, draftRoute, holdDraftResponse,
} from "./workflow-fixtures";

test("saved draft revisions and two-tab conflict preserve explicit recovery", async ({ page, context, application }, testInfo) => {
  const id = await createDraft(page);
  const originalLabel = await (await promptLabel(page)).inputValue();
  await saveDraft(page, 1);
  await (await promptLabel(page)).fill("Revision two prompt");
  await saveDraft(page, 2);

  const history = page.getByRole("region", { name: uiText(page, "Draft revision history and restore"), exact: true });
  await history.getByRole("button", { name: uiText(page, "Refresh history"), exact: true }).click();
  await history.getByRole("button", { name: /^v1 · / }).click();
  await history.getByRole("button", { name: uiText(page, "Prepare restore from this version"), exact: true }).click();
  await history.getByRole("button", { name: uiText(page, "Confirm new revision"), exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 3, 1));
  await expect(await promptLabel(page)).toHaveValue(originalLabel);
  await history.getByRole("button", { name: uiText(page, "Refresh history"), exact: true }).click();
  await expect(history.getByRole("button", { name: /^v1 · / })).toBeVisible();
  await expect(history.getByRole("button", { name: /^v2 · / })).toBeVisible();

  const other = await context.newPage();
  setTestLanguage(other, testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  await other.goto("/#workspace-applications");
  await selectApplication(other, application);
  await openDraft(other, id);
  await (await promptLabel(other)).fill("Other tab saved prompt");
  await saveDraft(other, 4);

  await (await promptLabel(page)).fill("Unsaved first tab prompt");
  const conflict = page.waitForResponse((response) => isEndpoint(response, draftRoute, "POST"));
  await designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }).click();
  const conflictEnvelope = await (await conflict).json();
  expect(conflictEnvelope.failure_code).toBe("draft_version_conflict");
  expect(conflictEnvelope.current_draft_version).toBe(4);
  await expect(designer(page)).toContainText(uiText(page, "Version conflict review"));
  await expect(await promptLabel(page)).toHaveValue("Unsaved first tab prompt");
  await designer(page).getByRole("button", { name: uiText(page, "Open saved draft"), exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 4, 1));
  await expect(await promptLabel(page)).toHaveValue("Other tab saved prompt");

  await page.reload();
  await selectApplication(page, application);
  await openDraft(page, id);
  await expect(await promptLabel(page)).toHaveValue("Other tab saved prompt");
  const freshId = await createDraft(page);
  expect(freshId).not.toBe(id);
  await saveDraft(page, 1);
});

test("late save and validation responses cannot cross draft or workspace scope", async ({ page, application }) => {
  const original = await createDraft(page);
  await saveDraft(page, 1);
  await (await promptLabel(page)).fill("Old draft write reaches server");
  const save = await holdDraftResponse(page, draftRoute, "POST");
  let freshId: string;
  try {
    await designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }).click();
    await save.arrived;
    freshId = await createDraft(page);
    expect(freshId).not.toBe(original);
    await save.deliver();
    await expect(draftField(page, "Draft")).toHaveText(freshId);
    await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 0, 0));
  } finally {
    await save.dispose();
  }
  await saveDraft(page, 1);
  await openDraft(page, original);
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 2, 1));
  await expect(await promptLabel(page)).toHaveValue("Old draft write reaches server");

  // Establish that the injected transport failure is observable before testing its stale counterpart.
  const currentFailure = await holdDraftResponse(page, `${draftRoute}/validate`, "POST", true);
  try {
    await designer(page).getByRole("button", { name: uiText(page, "Validate"), exact: true }).click();
    await currentFailure.arrived;
    await currentFailure.deliver();
    await expect(designer(page)).toContainText(uiText(page, "Validation failed"));
  } finally {
    await currentFailure.dispose();
  }
  await openDraft(page, original);
  const validation = await holdDraftResponse(page, `${draftRoute}/validate`, "POST", true);
  try {
    await (await promptLabel(page)).fill("Transient workspace marker");
    await designer(page).getByRole("button", { name: uiText(page, "Validate"), exact: true }).click();
    await validation.arrived;
    await page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true }).fill("workspace_e2e_other");
    await page.getByRole("button", { name: uiText(page, "Switch workspace"), exact: true }).click();
    await expect(page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true })).toHaveValue("workspace_e2e_other");
    await validation.deliver();
    await expect(designer(page)).not.toContainText(uiText(page, "Validation failed"));
    await expect(draftField(page, "Draft")).not.toHaveText(original);
  } finally {
    await validation.dispose();
  }
  await page.getByRole("combobox", { name: uiText(page, "Workspace"), exact: true }).fill("workspace_demo");
  await page.getByRole("button", { name: uiText(page, "Switch workspace"), exact: true }).click();
  await selectApplication(page, application);
  await expect(draftField(page, "Draft")).not.toHaveText(original);
  await openDraft(page, original);
  await expect(await promptLabel(page)).toHaveValue("Old draft write reaches server");
});

test("reviewed definition runs once and remains readable after refresh across layouts", async ({ page, application }) => {
  await createDraft(page);
  await saveDraft(page, 1);
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true }).click();
  await expect(page.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true })).toBeEnabled();

  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const link = page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true });
    await link.scrollIntoViewIfNeeded();
    const geometry = await link.evaluate((element) => {
      const workbench = element.closest(".application-development-workbench")!;
      const rail = workbench.querySelector(".application-development-readiness-pane")!;
      const area = workbench.getBoundingClientRect();
      const evidence = rail.getBoundingClientRect();
      const target = element.getBoundingClientRect();
      return {
        railWithin: evidence.left >= area.left && evidence.right <= area.right,
        pointerHits: [0.2, 0.5, 0.8].every((point) => element.contains(document.elementFromPoint(target.x + target.width * point, target.y + target.height / 2))),
        pageWidth: document.documentElement.scrollWidth,
      };
    });
    expect(geometry).toEqual({ railWithin: true, pointerHits: true, pageWidth: width });
    await page.getByRole("link", { name: uiText(page, "Open Application candidate owner"), exact: true }).click();
    await link.click();
    await expect(page).toHaveURL(/#workflow-definition-promotion$/);
    await expect(page.locator("#application-development-owner-surface")).toBeVisible();
  }

  await page.setViewportSize({ width: 1440, height: 900 });
  const executions: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/v1/user-workspace/workflow-definition-runs") executions.push(request.url());
  });
  await page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true }).fill("Synthetic E2E advisory input; no business action is authorized.");
  await page.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true }).click();
  await expect(page.getByRole("button", { name: uiText(page, "Open Run History"), exact: true })).toBeVisible();
  await expect(page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true })).toHaveValue("");
  const runId = await page.locator(".workflow-definition-run-result strong").innerText();
  expect(runId).toMatch(/^run_[a-z0-9]+$/);
  await page.getByRole("button", { name: uiText(page, "Open Run History"), exact: true }).click();
  const runOwner = page.getByRole("region", { name: uiText(page, "runs workflow review owner"), exact: true });
  await expect(runOwner).toContainText(runId);
  await expect(runOwner).toContainText("workflow_run_record.v5");
  await expect(runOwner).toContainText(uiText(page, "Succeeded"));
  await page.reload();
  await selectApplication(page, application);
  // The catalog projection is refreshed on reload, after the definition was published.
  await expect(page.locator("#workspace-workflow-definitions .workflow-definition-row").first()).toBeVisible();
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const definitionsFit = await page.locator("#workspace-workflow-definitions .workflow-definition-row").evaluateAll(rows => rows.length > 0 && rows.every(row => {
      const card = row.getBoundingClientRect();
      return [...row.querySelectorAll(".workflow-definition-row-main, .workflow-definition-row-meta, .workflow-definition-row-actions, h4")].every(item => {
        const box = item.getBoundingClientRect();
        return box.left >= card.left && box.right <= card.right && item.scrollWidth <= item.clientWidth + 1;
      });
    }));
    expect(definitionsFit, "Long definition IDs and row content must fit each catalog card").toBe(true);
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.getByRole("button", { name: new RegExp(`^${runId} `) }).click();
  await expect(runOwner).toContainText("workflow_run_record.v5");
  await expect(runOwner).toContainText(uiText(page, "Succeeded"));
  await expect(runOwner).not.toContainText("Synthetic E2E advisory input");
  expect(executions).toHaveLength(1);
});

test("Workflow language switching preserves edits, pending saves and revision confirmation", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  const locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next = locale === "zh-CN" ? "en-US" : "zh-CN";
  const id = await createDraft(page);
  await saveDraft(page, 1);
  await designer(page).getByRole("combobox", { name: uiText(page, "Inspect node"), exact: true }).selectOption("node_executor_model");
  const nodeEditor = () => designer(page).getByRole("textbox", { name: uiText(page, "Label"), exact: true });
  const source = "未保存 Node <literal>";
  await nodeEditor().fill(source);
  await nodeEditor().focus();
  const requests: string[] = [];
  page.on("request", request => {
    if (new URL(request.url()).port === "17000") requests.push(request.method() + " " + new URL(request.url()).pathname);
  });
  async function switchLanguage(value: "en-US" | "zh-CN") {
    await page.evaluate(locale => {
      localStorage.setItem("radishmind.uiLocale.v1", locale);
      window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: locale, storageArea: localStorage }));
    }, value);
    setTestLanguage(page, value);
    await expect(page.locator("html")).toHaveAttribute("lang", value);
  }
  await switchLanguage(next);
  await expect(nodeEditor()).toHaveValue(source);
  await expect(nodeEditor()).toBeFocused();
  await expect(designer(page).getByRole("combobox", { name: uiText(page, "Inspect node"), exact: true })).toHaveValue("node_executor_model");
  await expect(draftField(page, "Draft")).toHaveText(id);
  await expect(draftField(page, "Edit state")).toHaveText(uiText(page, "Unsaved local"));
  expect(requests).toEqual([]);

  const held = await holdDraftResponse(page, draftRoute, "POST");
  try {
    await designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }).click();
    await held.arrived;
    const before = [...requests];
    await switchLanguage(locale);
    expect(requests).toEqual(before);
    await expect(designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true })).toBeDisabled();
    await held.deliver();
    await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 2, 1));
    await expect(nodeEditor()).toHaveValue(source);
    expect(requests.filter(request => request === `POST ${draftRoute}`)).toHaveLength(1);
  } finally {
    await held.dispose();
  }

  const history = () => page.getByRole("region", { name: uiText(page, "Draft revision history and restore"), exact: true });
  await history().getByRole("button", { name: uiText(page, "Refresh history"), exact: true }).click();
  await history().getByRole("button", { name: /^v1 · / }).click();
  await history().getByRole("button", { name: uiText(page, "Prepare restore from this version"), exact: true }).click();
  const beforeConfirmationSwitch = [...requests];
  await switchLanguage(next);
  expect(requests).toEqual(beforeConfirmationSwitch);
  await expect(history().getByRole("button", { name: uiText(page, "Confirm new revision"), exact: true })).toBeEnabled();
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 2, 1));
  await history().getByRole("button", { name: uiText(page, "Cancel"), exact: true }).click();
  await expect(nodeEditor()).toHaveValue(source);

  for (const language of [locale, next] as const) {
    await switchLanguage(language);
    for (const width of [1440, 720, 390]) {
      await page.setViewportSize({ width, height: 900 });
      await designer(page).scrollIntoViewIfNeeded();
      await expect(draftField(page, "Draft")).toHaveText(id);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
      const fits = await designer(page).locator(".workflow-designer-context, .workflow-designer-actions, .workflow-node-designer-toolbar").evaluateAll(elements => elements.every(element => element.scrollWidth <= element.clientWidth + 1));
      expect(fits, "Draft context, actions and toolbar must fit in both languages").toBe(true);
      await designer(page).screenshot({ path: testInfo.outputPath(`workflow-${language}-${width}.png`) });
    }
  }
});

test("Workflow library keeps archive confirmation across languages and requires explicit reopen", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  const locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next = locale === "zh-CN" ? "en-US" : "zh-CN";
  const id = await createDraft(page);
  await saveDraft(page, 1);
  const library = () => page.getByLabel(uiText(page, "User workspace saved draft list"), { exact: true });
  const row = () => page.getByLabel(uiText(page, "Saved draft summaries"), { exact: true }).locator("article").filter({ hasText: id });
  await row().getByRole("button", { name: uiText(page, "Archive"), exact: true }).click();
  const requests: string[] = [];
  page.on("request", request => {
    if (new URL(request.url()).port === "17000") requests.push(request.method() + " " + new URL(request.url()).pathname);
  });
  await page.locator(".ui-language-selector select:visible").first().selectOption(next);
  setTestLanguage(page, next);
  await expect(page.locator("html")).toHaveAttribute("lang", next);
  await expect(row().getByRole("button", { name: uiText(page, "Confirm archive"), exact: true })).toBeVisible();
  expect(requests).toEqual([]);
  await row().getByRole("button", { name: uiText(page, "Confirm archive"), exact: true }).click();
  await expect(row()).toHaveCount(0);
  await library().getByRole("tab", { name: uiText(page, "Archived drafts"), exact: true }).click();
  await row().getByRole("button", { name: uiText(page, "Read-only review"), exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 1, 2));
  await expect(designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true })).toBeDisabled();
  await row().getByRole("button", { name: uiText(page, "Unarchive"), exact: true }).click();
  await expect(row()).toHaveCount(0);
  await expect(designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true })).toBeDisabled();
  await library().getByRole("tab", { name: uiText(page, "Active drafts"), exact: true }).click();
  await openDraft(page, id);
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 1, 3));
  await library().getByRole("textbox", { name: uiText(page, "Name prefix"), exact: true }).fill("missing-language-test");
  await library().getByRole("button", { name: uiText(page, "Apply filters"), exact: true }).click();
  await expect(library()).toContainText(next === "zh-CN" ? "当前生命周期和筛选条件下没有草案" : "No drafts match this lifecycle and filter selection");
});
