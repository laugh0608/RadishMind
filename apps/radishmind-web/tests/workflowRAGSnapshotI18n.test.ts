import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { workflowRAGSnapshot as en } from "../src/i18n/locales/en-US/workflowRAGSnapshot.ts";
import { workflowRAGSnapshot as zh } from "../src/i18n/locales/zh-CN/workflowRAGSnapshot.ts";
import { workflowRAGSnapshotFindingMessage } from "../src/features/control-plane-read/workflowRAGSnapshotMessages.ts";
import type { WorkflowRAGSnapshotFindingMessage } from "../src/features/control-plane-read/workflowRAGSnapshotFeedback.ts";
import { importWorkflowRAGLocalMaterials, preflightWorkflowRAGLocalMaterialSelection } from "../src/features/control-plane-read/workflowRAGLocalMaterialImporter.ts";
import { analyzeWorkflowRAGSnapshotEditor, buildWorkflowRAGSnapshotWriteInput, createEmptyWorkflowRAGSnapshotEditor, replaceWorkflowRAGSnapshotEditorWithImport } from "../src/features/control-plane-read/workflowRAGSnapshotEditor.ts";

test("RAG findings translate retained descriptors and preserve literal source names and references", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "workflow", { ragSnapshot: en });
  instance.addResourceBundle("zh-CN", "workflow", { ragSnapshot: zh });
  const file = "材料<img>.pdf";
  const preflight = preflightWorkflowRAGLocalMaterialSelection([{ fileName: file, fileBytes: 300_000 }])!;
  assert.deepEqual(preflight.findings.map(value => value.message), [
    { key: "unsupportedFile", values: { file } }, { key: "fileBudget", values: { file } },
  ]);
  const messages: WorkflowRAGSnapshotFindingMessage[] = [
    ...preflight.findings.map(value => value.message),
    { key: "fileCount", values: { maximum: 16 } }, { key: "rawBudget" },
    { key: "utf8", values: { file } }, { key: "invalidContent", values: { file } },
    { key: "fileSecret", values: { file } }, { key: "noContent", values: { file } },
    { key: "duplicateSource", values: { source: "原始资料.md", previous: "copy.md" } },
    { key: "fragmentBudget", values: { fragment: "fragment_001" } },
    { key: "duplicateContent", values: { fragment: "fragment_001", previous: "fragment_002" } },
    { key: "snapshotBudget" }, { key: "noFragments" }, { key: "metadata" },
    { key: "fragmentCount", values: { maximum: 256 } }, { key: "missingSource" },
    { key: "duplicateRef", values: { fragment: "fragment_001", previous: "fragment_001" } },
    { key: "contentBudget" }, { key: "replacementInvalid" }, { key: "fragmentSecret" }, { key: "fragmentInvalid" },
  ];
  const before = JSON.stringify(messages);
  const render = () => messages.map(message => workflowRAGSnapshotFindingMessage(instance.getFixedT(null, "workflow"), message));
  const english = render(); await instance.changeLanguage("zh-CN"); const chinese = render();
  assert.equal(new Set(messages.map(value => value.key)).size, Object.keys(en.findings).length);
  english.forEach((value, index) => {
    assert.notEqual(value, chinese[index]);
    assert.doesNotMatch(value, /\{\{/u);
    assert.doesNotMatch(chinese[index]!, /\{\{/u);
  });
  for (const output of [english, chinese]) { assert.ok(output[0]!.includes(file)); assert.match(output[9]!, /fragment_001/); }
  assert.equal(JSON.stringify(messages), before);
});

test("language changes leave RAG imported material, stable refs and the exact write payload untouched", async () => {
  const content = "# 原始标题\nLiteral <b> content 保持原文。";
  const imported = await importWorkflowRAGLocalMaterials([{ fileName: "原始材料.md", bytes: new TextEncoder().encode(content), selectionIndex: 0 }]);
  const editor = { ...replaceWorkflowRAGSnapshotEditorWithImport(createEmptyWorkflowRAGSnapshotEditor(), imported), snapshotKey: "bilingual_docs", displayName: "知识资料 Original" };
  assert.equal(analyzeWorkflowRAGSnapshotEditor(editor).canSubmit, true);
  const before = JSON.stringify({ imported, editor, built: buildWorkflowRAGSnapshotWriteInput(editor) });
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "workflow", { ragSnapshot: en });
  instance.addResourceBundle("zh-CN", "workflow", { ragSnapshot: zh });
  for (const locale of ["zh-CN", "en-US"]) {
    await instance.changeLanguage(locale);
    assert.equal(JSON.stringify({ imported, editor, built: buildWorkflowRAGSnapshotWriteInput(editor) }), before);
    assert.equal(editor.sources[0]!.label, "原始材料.md");
    const built = buildWorkflowRAGSnapshotWriteInput(editor);
    assert.equal(built.input!.fragments[0]!.content, content);
    assert.equal(Object.keys(built.input!.fragments[0]!).length, 7);
    assert.ok(!JSON.stringify(built.input).includes("原始材料.md"));
  }
});
