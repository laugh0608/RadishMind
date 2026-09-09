import {
  test, expect, designer, draftField, createDraft, saveDraft, promptLabel, selectApplication,
  openDraft, isEndpoint, draftRoute, holdDraftResponse,
} from "./workflow-fixtures";

test("saved draft revisions and two-tab conflict preserve explicit recovery", async ({ page, context, application }) => {
  const id = await createDraft(page);
  const originalLabel = await (await promptLabel(page)).inputValue();
  await saveDraft(page, 1);
  await (await promptLabel(page)).fill("Revision two prompt");
  await saveDraft(page, 2);

  const history = page.getByRole("region", { name: "草案修订历史与恢复", exact: true });
  await history.getByRole("button", { name: "刷新历史", exact: true }).click();
  await history.getByRole("button", { name: /^v1 · / }).click();
  await history.getByRole("button", { name: "准备从此版本恢复", exact: true }).click();
  await history.getByRole("button", { name: "确认创建新修订", exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText("content 3 / lifecycle 1");
  await expect(await promptLabel(page)).toHaveValue(originalLabel);
  await history.getByRole("button", { name: "刷新历史", exact: true }).click();
  await expect(history.getByRole("button", { name: /^v1 · / })).toBeVisible();
  await expect(history.getByRole("button", { name: /^v2 · / })).toBeVisible();

  const other = await context.newPage();
  await other.goto("/#workspace-applications");
  await selectApplication(other, application);
  await openDraft(other, id);
  await (await promptLabel(other)).fill("Other tab saved prompt");
  await saveDraft(other, 4);

  await (await promptLabel(page)).fill("Unsaved first tab prompt");
  const conflict = page.waitForResponse((response) => isEndpoint(response, draftRoute, "POST"));
  await designer(page).getByRole("button", { name: "Save draft", exact: true }).click();
  const conflictEnvelope = await (await conflict).json();
  expect(conflictEnvelope.failure_code).toBe("draft_version_conflict");
  expect(conflictEnvelope.current_draft_version).toBe(4);
  await expect(designer(page)).toContainText("Version conflict review");
  await expect(await promptLabel(page)).toHaveValue("Unsaved first tab prompt");
  await designer(page).getByRole("button", { name: "Open saved draft", exact: true }).click();
  await expect(draftField(page, "Version")).toHaveText("content 4 / lifecycle 1");
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
    await designer(page).getByRole("button", { name: "Save draft", exact: true }).click();
    await save.arrived;
    freshId = await createDraft(page);
    expect(freshId).not.toBe(original);
    await save.deliver();
    await expect(draftField(page, "Draft")).toHaveText(freshId);
    await expect(draftField(page, "Version")).toHaveText("content 0 / lifecycle 0");
  } finally {
    await save.dispose();
  }
  await saveDraft(page, 1);
  await openDraft(page, original);
  await expect(draftField(page, "Version")).toHaveText("content 2 / lifecycle 1");
  await expect(await promptLabel(page)).toHaveValue("Old draft write reaches server");

  // Establish that the injected transport failure is observable before testing its stale counterpart.
  const currentFailure = await holdDraftResponse(page, `${draftRoute}/validate`, "POST", true);
  try {
    await designer(page).getByRole("button", { name: "Validate", exact: true }).click();
    await currentFailure.arrived;
    await currentFailure.deliver();
    await expect(designer(page)).toContainText("validation_failed");
  } finally {
    await currentFailure.dispose();
  }
  await openDraft(page, original);
  const validation = await holdDraftResponse(page, `${draftRoute}/validate`, "POST", true);
  try {
    await (await promptLabel(page)).fill("Transient workspace marker");
    await designer(page).getByRole("button", { name: "Validate", exact: true }).click();
    await validation.arrived;
    await page.getByRole("combobox", { name: "Workspace", exact: true }).fill("workspace_e2e_other");
    await page.getByRole("button", { name: "Switch workspace", exact: true }).click();
    await expect(page.getByRole("combobox", { name: "Workspace", exact: true })).toHaveValue("workspace_e2e_other");
    await validation.deliver();
    await expect(designer(page)).not.toContainText("validation_failed");
    await expect(draftField(page, "Draft")).not.toHaveText(original);
  } finally {
    await validation.dispose();
  }
  await page.getByRole("combobox", { name: "Workspace", exact: true }).fill("workspace_demo");
  await page.getByRole("button", { name: "Switch workspace", exact: true }).click();
  await selectApplication(page, application);
  await expect(draftField(page, "Draft")).not.toHaveText(original);
  await openDraft(page, original);
  await expect(await promptLabel(page)).toHaveValue("Old draft write reaches server");
});

test("reviewed definition runs once and remains readable after refresh across layouts", async ({ page, application }) => {
  await createDraft(page);
  await saveDraft(page, 1);
  await page.getByRole("link", { name: "Open Workflow definition owner", exact: true }).click();
  await page.getByRole("button", { name: "创建晋级候选", exact: true }).click();
  await page.getByRole("button", { name: "追加 review v1", exact: true }).click();
  await page.getByRole("button", { name: "activate · expected pointer v0", exact: true }).click();
  await expect(page.getByRole("button", { name: "启动精确版本运行", exact: true })).toBeEnabled();

  for (const width of [1440, 1200, 390]) {
    await page.setViewportSize({ width, height: 900 });
    const link = page.getByRole("link", { name: "Open Workflow definition owner", exact: true });
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
    await page.getByRole("link", { name: "Open Application candidate owner", exact: true }).click();
    await link.click();
    await expect(page).toHaveURL(/#workflow-definition-promotion$/);
    await expect(page.locator("#application-development-owner-surface")).toBeVisible();
  }

  await page.setViewportSize({ width: 1440, height: 900 });
  const executions: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/v1/user-workspace/workflow-definition-runs") executions.push(request.url());
  });
  await page.getByRole("textbox", { name: "一次性输入", exact: true }).fill("Synthetic E2E advisory input; no business action is authorized.");
  await page.getByRole("button", { name: "启动精确版本运行", exact: true }).click();
  await expect(page.getByRole("button", { name: "打开 Run History", exact: true })).toBeVisible();
  await expect(page.getByRole("textbox", { name: "一次性输入", exact: true })).toHaveValue("");
  const runId = await page.locator(".workflow-definition-run-result strong").innerText();
  expect(runId).toMatch(/^run_[a-z0-9]+$/);
  await page.getByRole("button", { name: "打开 Run History", exact: true }).click();
  const runOwner = page.getByRole("region", { name: "runs workflow review owner", exact: true });
  await expect(runOwner).toContainText(runId);
  await expect(runOwner).toContainText("workflow_run_record.v5");
  await expect(runOwner).toContainText(/succeeded/i);
  await page.reload();
  await selectApplication(page, application);
  await page.getByRole("button", { name: new RegExp(`^${runId} `) }).click();
  await expect(runOwner).toContainText("workflow_run_record.v5");
  await expect(runOwner).toContainText(/succeeded/i);
  await expect(runOwner).not.toContainText("Synthetic E2E advisory input");
  expect(executions).toHaveLength(1);
});
