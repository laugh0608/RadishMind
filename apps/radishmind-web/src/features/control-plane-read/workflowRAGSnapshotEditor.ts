import {
  WORKFLOW_RAG_SNAPSHOT_LIMITS,
  containsWorkflowRAGSecretMaterial,
  validateWorkflowRAGFragmentInput,
  validateWorkflowRAGSnapshotWriteInput,
  type WorkflowRAGContentClassification,
  type WorkflowRAGFragmentInput,
  type WorkflowRAGSnapshotRecord,
  type WorkflowRAGSnapshotWriteInput,
  type WorkflowRAGSourceType,
} from "./workflowRAGSnapshotConsumer.ts";
import type {
  WorkflowRAGLocalMaterialFinding,
  WorkflowRAGLocalMaterialImportResult,
} from "./workflowRAGLocalMaterialImporter.ts";

const encoder = new TextEncoder();
const DYNAMIC_IMPORT_FINDINGS = new Set([
  "workflow_rag_material_source_duplicate",
  "workflow_rag_material_fragment_duplicate",
  "workflow_rag_material_budget_exceeded",
]);

export type WorkflowRAGSnapshotEditorSource = {
  sourceId: string;
  label: string;
  fileBytes: number;
  contentDigest: string;
  sourceType: WorkflowRAGSourceType;
  sourceRef: string;
  isOfficial: boolean;
};

export type WorkflowRAGSnapshotEditorFragment = {
  fragmentId: string;
  sourceId: string;
  fragmentRef: string;
  pageSlug: string;
  title: string;
  content: string;
};

export type WorkflowRAGSnapshotEditorFinding = {
  code: string;
  target: "snapshot" | "source" | "fragment";
  targetId: string;
  summary: string;
};

export type WorkflowRAGSnapshotEditor = {
  snapshotKey: string;
  displayName: string;
  contentClassification: WorkflowRAGContentClassification;
  sources: WorkflowRAGSnapshotEditorSource[];
  fragments: WorkflowRAGSnapshotEditorFragment[];
  selectedFragmentId: string;
  importFindings: WorkflowRAGSnapshotEditorFinding[];
  nextManualSequence: number;
};

export type WorkflowRAGSnapshotEditorAnalysis = {
  findings: WorkflowRAGSnapshotEditorFinding[];
  sourceCount: number;
  fragmentCount: number;
  totalContentBytes: number;
  canSubmit: boolean;
};

export type WorkflowRAGSnapshotEditorBuildResult =
  | { status: "ready"; input: WorkflowRAGSnapshotWriteInput; failureCode: "" }
  | { status: "blocked"; input: null; failureCode: string };

export function createEmptyWorkflowRAGSnapshotEditor(): WorkflowRAGSnapshotEditor {
  const source = manualSource(1);
  const fragment = manualFragment(source.sourceId, 1);
  return {
    snapshotKey: "",
    displayName: "",
    contentClassification: "workspace_internal",
    sources: [source],
    fragments: [fragment],
    selectedFragmentId: fragment.fragmentId,
    importFindings: [],
    nextManualSequence: 2,
  };
}

export function createWorkflowRAGSnapshotEditorFromRecord(record: WorkflowRAGSnapshotRecord): WorkflowRAGSnapshotEditor {
  const sources: WorkflowRAGSnapshotEditorSource[] = [];
  const sourceIds = new Map<string, string>();
  const fragments = record.fragments.map((fragment, index) => {
    const sourceKey = `${fragment.sourceRef}\u0000${fragment.sourceType}\u0000${String(fragment.isOfficial)}`;
    let sourceId = sourceIds.get(sourceKey);
    if (!sourceId) {
      sourceId = `record_source_${String(sources.length + 1).padStart(3, "0")}`;
      sourceIds.set(sourceKey, sourceId);
      sources.push({
        sourceId,
        label: fragment.sourceRef,
        fileBytes: 0,
        contentDigest: "",
        sourceType: fragment.sourceType,
        sourceRef: fragment.sourceRef,
        isOfficial: fragment.isOfficial,
      });
    }
    return {
      fragmentId: `record_fragment_${String(index + 1).padStart(3, "0")}`,
      sourceId,
      fragmentRef: fragment.fragmentRef,
      pageSlug: fragment.pageSlug,
      title: fragment.title,
      content: fragment.content,
    };
  });
  return {
    snapshotKey: record.snapshotKey,
    displayName: record.displayName,
    contentClassification: record.contentClassification,
    sources,
    fragments,
    selectedFragmentId: fragments[0]?.fragmentId ?? "",
    importFindings: [],
    nextManualSequence: 1,
  };
}

export function replaceWorkflowRAGSnapshotEditorWithImport(
  editor: WorkflowRAGSnapshotEditor,
  result: WorkflowRAGLocalMaterialImportResult,
): WorkflowRAGSnapshotEditor {
  const sources = result.sources.map((source) => ({
    sourceId: source.sourceId,
    label: source.fileName,
    fileBytes: source.fileBytes,
    contentDigest: source.contentDigest,
    sourceType: source.sourceType,
    sourceRef: source.sourceRef,
    isOfficial: source.isOfficial,
  }));
  const fragments = result.fragments.map((fragment, index) => ({
    fragmentId: `import_fragment_${String(index + 1).padStart(3, "0")}`,
    sourceId: fragment.sourceId,
    fragmentRef: fragment.fragmentRef,
    pageSlug: fragment.pageSlug,
    title: fragment.title,
    content: fragment.content,
  }));
  return {
    ...editor,
    sources,
    fragments,
    selectedFragmentId: fragments[0]?.fragmentId ?? "",
    importFindings: result.findings.filter((finding) => !DYNAMIC_IMPORT_FINDINGS.has(finding.code)).map(mapImportFinding),
    nextManualSequence: 1,
  };
}

export function updateWorkflowRAGSnapshotEditorSource(
  editor: WorkflowRAGSnapshotEditor,
  sourceId: string,
  patch: Partial<Pick<WorkflowRAGSnapshotEditorSource, "sourceType" | "sourceRef" | "isOfficial">>,
): WorkflowRAGSnapshotEditor {
  return { ...editor, sources: editor.sources.map((source) => source.sourceId === sourceId ? { ...source, ...patch } : source) };
}

export function updateWorkflowRAGSnapshotEditorFragment(
  editor: WorkflowRAGSnapshotEditor,
  fragmentId: string,
  patch: Partial<Pick<WorkflowRAGSnapshotEditorFragment, "fragmentRef" | "pageSlug" | "title" | "content">>,
): WorkflowRAGSnapshotEditor {
  return { ...editor, fragments: editor.fragments.map((fragment) => fragment.fragmentId === fragmentId ? { ...fragment, ...patch } : fragment) };
}

export function selectWorkflowRAGSnapshotEditorFragment(editor: WorkflowRAGSnapshotEditor, fragmentId: string): WorkflowRAGSnapshotEditor {
  return editor.fragments.some((fragment) => fragment.fragmentId === fragmentId) ? { ...editor, selectedFragmentId: fragmentId } : editor;
}

export function removeWorkflowRAGSnapshotEditorSource(editor: WorkflowRAGSnapshotEditor, sourceId: string): WorkflowRAGSnapshotEditor {
  const fragments = editor.fragments.filter((fragment) => fragment.sourceId !== sourceId);
  return withFragments({ ...editor, sources: editor.sources.filter((source) => source.sourceId !== sourceId), importFindings: [] }, fragments);
}

export function removeWorkflowRAGSnapshotEditorFragment(editor: WorkflowRAGSnapshotEditor, fragmentId: string): WorkflowRAGSnapshotEditor {
  return withFragments(editor, editor.fragments.filter((fragment) => fragment.fragmentId !== fragmentId));
}

export function addWorkflowRAGSnapshotManualFragment(editor: WorkflowRAGSnapshotEditor): WorkflowRAGSnapshotEditor {
  const sequence = editor.nextManualSequence;
  let source = editor.sources.find((candidate) => candidate.sourceType === "manual" && candidate.sourceRef === "application_manual");
  const sources = [...editor.sources];
  if (!source) {
    source = manualSource(sequence);
    sources.push(source);
  }
  const fragment = manualFragment(source.sourceId, sequence);
  return {
    ...editor,
    sources,
    fragments: [...editor.fragments, fragment],
    selectedFragmentId: fragment.fragmentId,
    importFindings: [],
    nextManualSequence: sequence + 1,
  };
}

export function analyzeWorkflowRAGSnapshotEditor(editor: WorkflowRAGSnapshotEditor): WorkflowRAGSnapshotEditorAnalysis {
  const findings = [...editor.importFindings];
  if (!/^[a-z][a-z0-9_]{2,47}$/u.test(editor.snapshotKey.trim()) || editor.displayName.trim().length < 2 || editor.displayName.trim().length > 120 || containsWorkflowRAGSecretMaterial(editor.displayName)) {
    findings.push(snapshotFinding("workflow_rag_snapshot_payload_invalid", "Snapshot key、display name 或 classification 不符合既有快照合同。"));
  }
  if (editor.fragments.length < 1 || editor.fragments.length > WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragments) {
    findings.push(snapshotFinding("workflow_rag_snapshot_payload_invalid", `最终 fragment 数必须为 1 至 ${WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragments}。`));
  }

  const sourceDigestOwners = new Map<string, WorkflowRAGSnapshotEditorSource>();
  for (const source of editor.sources) {
    if (!source.contentDigest) continue;
    const previous = sourceDigestOwners.get(source.contentDigest);
    if (previous) {
      findings.push({ code: "workflow_rag_material_source_duplicate", target: "source", targetId: source.sourceId, summary: `${source.label} 与 ${previous.label} 内容完全重复。` });
    } else {
      sourceDigestOwners.set(source.contentDigest, source);
    }
  }

  const fragmentRefOwners = new Map<string, WorkflowRAGSnapshotEditorFragment>();
  const contentOwners = new Map<string, WorkflowRAGSnapshotEditorFragment>();
  let totalContentBytes = 0;
  const input = projectWorkflowRAGSnapshotWriteInput(editor);
  for (const fragment of editor.fragments) {
    const source = editor.sources.find((candidate) => candidate.sourceId === fragment.sourceId);
    if (!source) {
      findings.push(fragmentFinding(fragment, "workflow_rag_fragment_invalid", "Fragment 已失去来源 owner。"));
      continue;
    }
    const projected = projectFragment(source, fragment);
    const validationFailure = validateWorkflowRAGFragmentInput(projected);
    if (validationFailure) findings.push(fragmentFinding(fragment, validationFailure, fragmentFailureSummary(validationFailure)));
    const normalizedRef = projected.fragmentRef.trim();
    const previousRef = fragmentRefOwners.get(normalizedRef);
    if (previousRef) {
      findings.push(fragmentFinding(fragment, "workflow_rag_material_fragment_duplicate", `${projected.fragmentRef || "未命名 fragment"} 与 ${previousRef.fragmentRef || "另一 fragment"} 使用重复引用。`));
    } else if (normalizedRef) {
      fragmentRefOwners.set(normalizedRef, fragment);
    }
    const normalizedContent = projected.content.trim();
    const previousContent = contentOwners.get(normalizedContent);
    if (previousContent && normalizedContent) {
      findings.push(fragmentFinding(fragment, "workflow_rag_material_fragment_duplicate", `${projected.fragmentRef || "未命名 fragment"} 与 ${previousContent.fragmentRef || "另一 fragment"} 正文完全重复。`));
    } else if (normalizedContent) {
      contentOwners.set(normalizedContent, fragment);
    }
    totalContentBytes += encoder.encode(normalizedContent).byteLength;
  }
  if (totalContentBytes > WORKFLOW_RAG_SNAPSHOT_LIMITS.maxTotalContentBytes) {
    findings.push(snapshotFinding("workflow_rag_budget_exceeded", "最终正文总量超过 1 MiB。"));
  }
  if (input) {
    const finalFailure = validateWorkflowRAGSnapshotWriteInput(input);
    if (finalFailure && !findings.some((finding) => finding.code === finalFailure)) {
      findings.push(snapshotFinding(finalFailure, "最终 replacement 未通过既有快照写入校验。"));
    }
  }
  return {
    findings,
    sourceCount: editor.sources.length,
    fragmentCount: editor.fragments.length,
    totalContentBytes,
    canSubmit: findings.length === 0 && input !== null,
  };
}

export function buildWorkflowRAGSnapshotWriteInput(editor: WorkflowRAGSnapshotEditor): WorkflowRAGSnapshotEditorBuildResult {
  const analysis = analyzeWorkflowRAGSnapshotEditor(editor);
  const input = projectWorkflowRAGSnapshotWriteInput(editor);
  if (!analysis.canSubmit || !input) return { status: "blocked", input: null, failureCode: analysis.findings[0]?.code || "workflow_rag_snapshot_payload_invalid" };
  return { status: "ready", input, failureCode: "" };
}

function projectWorkflowRAGSnapshotWriteInput(editor: WorkflowRAGSnapshotEditor): WorkflowRAGSnapshotWriteInput | null {
  const fragments: WorkflowRAGFragmentInput[] = [];
  for (const fragment of editor.fragments) {
    const source = editor.sources.find((candidate) => candidate.sourceId === fragment.sourceId);
    if (!source) return null;
    fragments.push(projectFragment(source, fragment));
  }
  return {
    snapshotKey: editor.snapshotKey.trim(),
    displayName: editor.displayName.trim(),
    contentClassification: editor.contentClassification,
    fragments,
  };
}

function projectFragment(source: WorkflowRAGSnapshotEditorSource, fragment: WorkflowRAGSnapshotEditorFragment): WorkflowRAGFragmentInput {
  return {
    fragmentRef: fragment.fragmentRef.trim(),
    sourceType: source.sourceType,
    sourceRef: source.sourceRef.trim(),
    pageSlug: fragment.pageSlug.trim(),
    title: fragment.title.trim(),
    isOfficial: source.isOfficial,
    content: fragment.content.trim(),
  };
}

function withFragments(editor: WorkflowRAGSnapshotEditor, fragments: WorkflowRAGSnapshotEditorFragment[]): WorkflowRAGSnapshotEditor {
  const retainedSourceIds = new Set(fragments.map((fragment) => fragment.sourceId));
  const selectedFragmentId = fragments.some((fragment) => fragment.fragmentId === editor.selectedFragmentId)
    ? editor.selectedFragmentId
    : fragments[0]?.fragmentId ?? "";
  return { ...editor, sources: editor.sources.filter((source) => retainedSourceIds.has(source.sourceId)), fragments, selectedFragmentId, importFindings: [] };
}

function manualSource(sequence: number): WorkflowRAGSnapshotEditorSource {
  return { sourceId: `manual_source_${String(sequence).padStart(3, "0")}`, label: "手工片段", fileBytes: 0, contentDigest: "", sourceType: "manual", sourceRef: "application_manual", isOfficial: true };
}

function manualFragment(sourceId: string, sequence: number): WorkflowRAGSnapshotEditorFragment {
  const suffix = String(sequence).padStart(3, "0");
  return { fragmentId: `manual_fragment_${suffix}`, sourceId, fragmentRef: `manual_${suffix}`, pageSlug: `manual/${suffix}`, title: "", content: "" };
}

function mapImportFinding(finding: WorkflowRAGLocalMaterialFinding): WorkflowRAGSnapshotEditorFinding {
  if (finding.fragmentRef) return { code: finding.code, target: "fragment", targetId: finding.fragmentRef, summary: finding.summary };
  if (finding.sourceId) return { code: finding.code, target: "source", targetId: finding.sourceId, summary: finding.summary };
  return snapshotFinding(finding.code, finding.summary);
}

function snapshotFinding(code: string, summary: string): WorkflowRAGSnapshotEditorFinding {
  return { code, target: "snapshot", targetId: "", summary };
}

function fragmentFinding(fragment: WorkflowRAGSnapshotEditorFragment, code: string, summary: string): WorkflowRAGSnapshotEditorFinding {
  return { code, target: "fragment", targetId: fragment.fragmentId, summary };
}

function fragmentFailureSummary(code: string): string {
  return code === "workflow_rag_secret_material_forbidden"
    ? "Fragment 的来源引用、页面引用、标题或正文命中敏感材料规则。"
    : "Fragment 引用、来源、页面、标题、正文或单片段预算不符合既有合同。";
}
