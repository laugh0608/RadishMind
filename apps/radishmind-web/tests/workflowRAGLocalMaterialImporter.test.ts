import assert from "node:assert/strict";
import test from "node:test";

import { importWorkflowRAGLocalMaterials, preflightWorkflowRAGLocalMaterialSelection, type WorkflowRAGLocalMaterialFile } from "../src/features/control-plane-read/workflowRAGLocalMaterialImporter.ts";
import { validateWorkflowRAGSnapshotWriteInput } from "../src/features/control-plane-read/workflowRAGSnapshotConsumer.ts";

const textEncoder = new TextEncoder();

test("imports Markdown and text in deterministic source order without treating fenced headings as sections", async () => {
  const markdown = "# 概览\r\n正文。\r\n\r\n```text\r\n# 代码内标题\r\n```\r\n\r\n## 约束\r\n只读建议。";
  const files = [file("z-notes.txt", "第一段。\n\n第二段。", 0), file("a-guide.md", markdown, 1)];
  const result = await importWorkflowRAGLocalMaterials(files);

  assert.equal(result.status, "ready");
  assert.deepEqual(result.sources.map((source) => source.fileName), ["a-guide.md", "z-notes.txt"]);
  assert.equal(result.sourceCount, 2);
  assert.equal(result.fragmentCount, 3);
  assert.deepEqual(result.fragments.map((fragment) => fragment.title), ["概览", "约束", "z-notes"]);
  assert.match(result.fragments[0]!.content, /# 代码内标题/u);
  assert.equal(result.fragments[0]!.content.includes("\r"), false);
  assert.match(result.fragments[0]!.sourceRef, /^local_material\/a-guide-[a-f0-9]{12}$/u);
  assert.match(result.fragments[0]!.fragmentRef, /^m_[a-f0-9]{10}_001$/u);
  assert.match(result.fragments[0]!.pageSlug, /^local-material\/a-guide-[a-f0-9]{8}\/001$/u);

  const reordered = await importWorkflowRAGLocalMaterials([file("a-guide.md", markdown, 0), file("z-notes.txt", "第一段。\n\n第二段。", 1)]);
  assert.deepEqual(reordered.fragments.map(stableProjection), result.fragments.map(stableProjection));
});

test("accepts one UTF-8 BOM and rejects invalid UTF-8, NUL, unsupported files, and secret material", async () => {
  const withBOM = new Uint8Array([0xef, 0xbb, 0xbf, ...textEncoder.encode("# 标题\n正文")]);
  const accepted = await importWorkflowRAGLocalMaterials([{ fileName: "guide.md", bytes: withBOM, selectionIndex: 0 }]);
  assert.equal(accepted.status, "ready");
  assert.equal(accepted.fragments[0]?.content.startsWith("# 标题"), true);

  const invalidUTF8 = await importWorkflowRAGLocalMaterials([{ fileName: "invalid.txt", bytes: new Uint8Array([0xc3, 0x28]), selectionIndex: 0 }]);
  assert.deepEqual(invalidUTF8.findings.map((finding) => finding.code), ["workflow_rag_material_utf8_invalid"]);

  const invalidInputs = await importWorkflowRAGLocalMaterials([
    file("binary.pdf", "not accepted", 0),
    file("nul.txt", "before\u0000after", 1),
    file("secret.md", "# Credential\napi_key=hidden", 2),
  ]);
  assert.equal(invalidInputs.status, "blocked");
  assert.deepEqual(new Set(invalidInputs.findings.map((finding) => finding.code)), new Set([
    "workflow_rag_material_file_type_unsupported",
    "workflow_rag_material_content_invalid",
    "workflow_rag_secret_material_forbidden",
    "workflow_rag_material_content_invalid",
  ]));
  assert.equal(JSON.stringify(invalidInputs).includes("hidden"), false);
});

test("splits long Unicode content below the target without breaking code points", async () => {
  const repeated = Array.from({ length: 900 }, (_, index) => `知识${index}🙂证据`).join(" ");
  const result = await importWorkflowRAGLocalMaterials([file("unicode.txt", repeated, 0)]);

  assert.equal(result.status, "ready");
  assert.ok(result.fragmentCount > 1);
  for (const fragment of result.fragments) {
    assert.ok(fragment.contentBytes <= 6 * 1024);
    assert.equal(fragment.content.includes("�"), false);
    assert.equal(fragment.contentDigest.startsWith("sha256:"), true);
  }
  assert.equal(result.fragments.reduce((count, fragment) => count + (fragment.content.match(/🙂/gu)?.length ?? 0), 0), 900);
});

test("blocks duplicate sources and duplicate generated fragment content", async () => {
  const result = await importWorkflowRAGLocalMaterials([
    file("first.md", "# Same\n相同正文。", 0),
    file("second.md", "# Same\n相同正文。", 1),
  ]);

  assert.equal(result.status, "blocked");
  assert.equal(result.sourceCount, 2);
  assert.equal(result.fragmentCount, 2);
  assert.ok(result.findings.some((finding) => finding.code === "workflow_rag_material_source_duplicate"));
  assert.ok(result.findings.some((finding) => finding.code === "workflow_rag_material_fragment_duplicate"));
});

test("enforces file count and raw byte budgets before snapshot submission", async () => {
  assert.deepEqual((await importWorkflowRAGLocalMaterials([])).findings.map((finding) => finding.code), ["workflow_rag_material_file_count_invalid"]);
  const tooMany = Array.from({ length: 17 }, (_, index) => file(`file-${index}.txt`, "content", index));
  assert.deepEqual((await importWorkflowRAGLocalMaterials(tooMany)).findings.map((finding) => finding.code), ["workflow_rag_material_file_count_invalid"]);

  const tooLarge = await importWorkflowRAGLocalMaterials([{ fileName: "large.txt", bytes: new Uint8Array(256 * 1024 + 1), selectionIndex: 0 }]);
  assert.ok(tooLarge.findings.some((finding) => finding.code === "workflow_rag_material_file_too_large"));

  const totalTooLarge = await importWorkflowRAGLocalMaterials(Array.from({ length: 5 }, (_, index) => ({
    fileName: `part-${index}.txt`,
    bytes: textEncoder.encode(`${String(index)}${"x".repeat(220 * 1024)}`),
    selectionIndex: index,
  })));
  assert.ok(totalTooLarge.findings.some((finding) => finding.code === "workflow_rag_material_file_too_large"));
});

test("preflights file metadata before the browser reads unsupported or oversized files", () => {
  const tooMany = preflightWorkflowRAGLocalMaterialSelection(Array.from({ length: 17 }, (_, index) => ({ fileName: `file-${index}.txt`, fileBytes: 8 })));
  assert.deepEqual(tooMany?.findings.map((finding) => finding.code), ["workflow_rag_material_file_count_invalid"]);

  const invalid = preflightWorkflowRAGLocalMaterialSelection([
    { fileName: "oversized.md", fileBytes: 256 * 1024 + 1 },
    { fileName: "archive.zip", fileBytes: 900 * 1024 },
  ]);
  assert.ok(invalid?.findings.some((finding) => finding.code === "workflow_rag_material_file_type_unsupported"));
  assert.ok(invalid?.findings.filter((finding) => finding.code === "workflow_rag_material_file_too_large").length >= 2);
  assert.equal(preflightWorkflowRAGLocalMaterialSelection([{ fileName: "guide.md", fileBytes: 128 }]), null);
});

test("produces fragments accepted by the existing snapshot write contract", async () => {
  const result = await importWorkflowRAGLocalMaterials([
    { ...file("product-guide.md", "# Overview\nRadishMind 提供可审查建议。\n\n## Boundary\n不自动写回业务真相源。", 0), sourceType: "document", isOfficial: true },
  ]);
  assert.equal(result.status, "ready");
  assert.equal(validateWorkflowRAGSnapshotWriteInput({
    snapshotKey: "product_guide",
    displayName: "产品指南",
    contentClassification: "workspace_internal",
    fragments: result.fragments,
  }), "");
});

function file(fileName: string, content: string, selectionIndex: number): WorkflowRAGLocalMaterialFile {
  return { fileName, bytes: textEncoder.encode(content), selectionIndex };
}

function stableProjection(fragment: { fragmentRef: string; sourceRef: string; pageSlug: string; title: string; content: string; contentDigest: string }) {
  return { fragmentRef: fragment.fragmentRef, sourceRef: fragment.sourceRef, pageSlug: fragment.pageSlug, title: fragment.title, content: fragment.content, contentDigest: fragment.contentDigest };
}
