import assert from "node:assert/strict";
import test from "node:test";

import { importWorkflowRAGLocalMaterials } from "../src/features/control-plane-read/workflowRAGLocalMaterialImporter.ts";
import {
  addWorkflowRAGSnapshotManualFragment,
  analyzeWorkflowRAGSnapshotEditor,
  buildWorkflowRAGSnapshotWriteInput,
  createEmptyWorkflowRAGSnapshotEditor,
  createWorkflowRAGSnapshotEditorFromRecord,
  removeWorkflowRAGSnapshotEditorFragment,
  removeWorkflowRAGSnapshotEditorSource,
  replaceWorkflowRAGSnapshotEditorWithImport,
  updateWorkflowRAGSnapshotEditorFragment,
  updateWorkflowRAGSnapshotEditorSource,
} from "../src/features/control-plane-read/workflowRAGSnapshotEditor.ts";
import type { WorkflowRAGSnapshotRecord } from "../src/features/control-plane-read/workflowRAGSnapshotConsumer.ts";

const encoder = new TextEncoder();

test("projects imported Markdown and text through one structured snapshot owner", async () => {
  const imported = await importWorkflowRAGLocalMaterials([
    { fileName: "guide.md", bytes: encoder.encode("# Overview\nProduct guidance.\n\n## Boundary\nNo automatic writeback."), selectionIndex: 0, sourceType: "document", isOfficial: true },
    { fileName: "faq.txt", bytes: encoder.encode("Question and answer."), selectionIndex: 1, sourceType: "faq", isOfficial: false },
  ]);
  let editor = replaceWorkflowRAGSnapshotEditorWithImport(createEmptyWorkflowRAGSnapshotEditor(), imported);
  editor = { ...editor, snapshotKey: "product_docs", displayName: "Product docs" };

  const analysis = analyzeWorkflowRAGSnapshotEditor(editor);
  const built = buildWorkflowRAGSnapshotWriteInput(editor);
  assert.equal(analysis.canSubmit, true);
  assert.equal(analysis.sourceCount, 2);
  assert.equal(analysis.fragmentCount, 3);
  assert.equal(built.status, "ready");
  assert.deepEqual(built.input?.fragments.map((fragment) => fragment.sourceType), ["faq", "document", "document"]);
  assert.equal(built.input?.fragments.some((fragment) => "sourceFileName" in fragment || "contentDigest" in fragment), false);
});

test("keeps source properties in one owner and projects edits into every linked fragment", async () => {
  const imported = await importWorkflowRAGLocalMaterials([
    { fileName: "guide.md", bytes: encoder.encode("# One\nFirst.\n\n## Two\nSecond."), selectionIndex: 0 },
  ]);
  let editor = replaceWorkflowRAGSnapshotEditorWithImport(createEmptyWorkflowRAGSnapshotEditor(), imported);
  editor = { ...editor, snapshotKey: "guide_docs", displayName: "Guide docs" };
  const sourceId = editor.sources[0]!.sourceId;
  editor = updateWorkflowRAGSnapshotEditorSource(editor, sourceId, { sourceType: "wiki", sourceRef: "knowledge/guide", isOfficial: true });

  const built = buildWorkflowRAGSnapshotWriteInput(editor);
  assert.equal(built.status, "ready");
  assert.ok(built.input?.fragments.every((fragment) => fragment.sourceType === "wiki" && fragment.sourceRef === "knowledge/guide" && fragment.isOfficial));
});

test("recomputes duplicate findings after source and fragment removal", async () => {
  const imported = await importWorkflowRAGLocalMaterials([
    { fileName: "first.md", bytes: encoder.encode("# Same\nRepeated body."), selectionIndex: 0 },
    { fileName: "second.md", bytes: encoder.encode("# Same\nRepeated body."), selectionIndex: 1 },
  ]);
  let editor = replaceWorkflowRAGSnapshotEditorWithImport(createEmptyWorkflowRAGSnapshotEditor(), imported);
  editor = { ...editor, snapshotKey: "duplicate_docs", displayName: "Duplicate docs" };
  assert.ok(analyzeWorkflowRAGSnapshotEditor(editor).findings.some((finding) => finding.code === "workflow_rag_material_source_duplicate"));
  assert.ok(analyzeWorkflowRAGSnapshotEditor(editor).findings.some((finding) => finding.code === "workflow_rag_material_fragment_duplicate"));

  editor = removeWorkflowRAGSnapshotEditorSource(editor, editor.sources[1]!.sourceId);
  assert.equal(analyzeWorkflowRAGSnapshotEditor(editor).findings.some((finding) => finding.code === "workflow_rag_material_source_duplicate"), false);
  assert.equal(analyzeWorkflowRAGSnapshotEditor(editor).findings.some((finding) => finding.code === "workflow_rag_material_fragment_duplicate"), false);
  assert.equal(buildWorkflowRAGSnapshotWriteInput(editor).status, "ready");
});

test("maps historical records into the same editor and supports explicit manual fragments", () => {
  let editor = createWorkflowRAGSnapshotEditorFromRecord(sampleRecord());
  assert.equal(editor.sources.length, 1);
  assert.equal(editor.fragments.length, 2);
  assert.equal(editor.selectedFragmentId, "record_fragment_001");

  editor = removeWorkflowRAGSnapshotEditorFragment(editor, "record_fragment_001");
  assert.equal(editor.selectedFragmentId, "record_fragment_002");
  editor = addWorkflowRAGSnapshotManualFragment(editor);
  assert.equal(editor.fragments.length, 2);
  assert.match(editor.fragments.at(-1)!.fragmentRef, /^manual_\d{3}$/u);
  assert.equal(editor.selectedFragmentId, editor.fragments.at(-1)!.fragmentId);
});

test("fails closed for secret material and duplicate edited references", () => {
  let editor = createWorkflowRAGSnapshotEditorFromRecord(sampleRecord());
  const first = editor.fragments[0]!;
  const second = editor.fragments[1]!;
  editor = updateWorkflowRAGSnapshotEditorFragment(editor, second.fragmentId, { fragmentRef: first.fragmentRef, content: "api_key=hidden" });
  const analysis = analyzeWorkflowRAGSnapshotEditor(editor);
  assert.equal(analysis.canSubmit, false);
  assert.ok(analysis.findings.some((finding) => finding.code === "workflow_rag_secret_material_forbidden"));
  assert.ok(analysis.findings.some((finding) => finding.code === "workflow_rag_material_fragment_duplicate"));
  assert.equal(buildWorkflowRAGSnapshotWriteInput(editor).status, "blocked");
});

function sampleRecord(): WorkflowRAGSnapshotRecord {
  return {
    schemaVersion: "workflow_rag_snapshot.v1",
    snapshotId: "rags_aaaaaaaaaaaaaaaa",
    tenantRef: "tenant_demo",
    workspaceId: "workspace_demo",
    applicationId: "application_demo",
    snapshotKey: "product_docs",
    ragRef: "workflow.rag.product_docs.v1",
    snapshotVersion: 1,
    displayName: "Product docs",
    lifecycleState: "active",
    contentClassification: "workspace_internal",
    profileRef: "workflow.rag.lexical-ngram-dev.v1",
    fragmentCount: 2,
    totalContentBytes: 25,
    snapshotDigest: `sha256:${"a".repeat(64)}`,
    createdAt: "2026-08-12T00:00:00Z",
    createdByActorRef: "subject_demo_user",
    requestId: "request_demo",
    auditRef: "audit_demo",
    fragments: [
      { schemaVersion: "workflow_rag_fragment.v1", fragmentRef: "overview", sourceType: "document", sourceRef: "knowledge/product", pageSlug: "product/overview", title: "Overview", isOfficial: true, content: "Product overview.", contentClassification: "workspace_internal", contentBytes: 17, contentDigest: `sha256:${"b".repeat(64)}` },
      { schemaVersion: "workflow_rag_fragment.v1", fragmentRef: "boundary", sourceType: "document", sourceRef: "knowledge/product", pageSlug: "product/boundary", title: "Boundary", isOfficial: true, content: "No writeback.", contentClassification: "workspace_internal", contentBytes: 13, contentDigest: `sha256:${"c".repeat(64)}` },
    ],
  };
}
