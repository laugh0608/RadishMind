import {
  WORKFLOW_RAG_SNAPSHOT_LIMITS,
  containsWorkflowRAGSecretMaterial,
  type WorkflowRAGFragmentInput,
  type WorkflowRAGSourceType,
} from "./workflowRAGSnapshotConsumer.ts";

export const WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS = {
  maxFiles: 16,
  maxFileBytes: 256 * 1024,
  maxRawBytes: 1024 * 1024,
  targetFragmentBytes: 6 * 1024,
} as const;

const TARGET_FRAGMENT_BYTES = WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.targetFragmentBytes;
const ALLOWED_EXTENSIONS = new Set(["md", "markdown", "txt"]);
const encoder = new TextEncoder();

export type WorkflowRAGLocalMaterialFailureCode =
  | "workflow_rag_material_file_count_invalid"
  | "workflow_rag_material_file_type_unsupported"
  | "workflow_rag_material_file_too_large"
  | "workflow_rag_material_utf8_invalid"
  | "workflow_rag_material_content_invalid"
  | "workflow_rag_material_source_duplicate"
  | "workflow_rag_material_fragment_duplicate"
  | "workflow_rag_material_budget_exceeded"
  | "workflow_rag_secret_material_forbidden";

export type WorkflowRAGLocalMaterialFile = {
  fileName: string;
  bytes: Uint8Array;
  selectionIndex: number;
  sourceType?: WorkflowRAGSourceType;
  isOfficial?: boolean;
};

export type WorkflowRAGLocalMaterialFileMetadata = {
  fileName: string;
  fileBytes: number;
};

export type WorkflowRAGLocalMaterialFinding = {
  code: WorkflowRAGLocalMaterialFailureCode;
  sourceId: string;
  fragmentRef: string;
  summary: string;
};

export type WorkflowRAGLocalMaterialSource = {
  sourceId: string;
  fileName: string;
  fileBytes: number;
  contentDigest: string;
  sourceType: WorkflowRAGSourceType;
  sourceRef: string;
  isOfficial: boolean;
  fragmentRefs: string[];
};

export type WorkflowRAGLocalMaterialFragment = WorkflowRAGFragmentInput & {
  sourceId: string;
  sourceFileName: string;
  contentBytes: number;
  contentDigest: string;
};

export type WorkflowRAGLocalMaterialImportResult = {
  status: "ready" | "blocked";
  sources: WorkflowRAGLocalMaterialSource[];
  fragments: WorkflowRAGLocalMaterialFragment[];
  findings: WorkflowRAGLocalMaterialFinding[];
  sourceCount: number;
  fragmentCount: number;
  rawBytes: number;
  totalContentBytes: number;
};

type DecodedMaterial = {
  fileName: string;
  extension: "md" | "markdown" | "txt";
  fileBytes: number;
  selectionIndex: number;
  sourceType: WorkflowRAGSourceType;
  isOfficial: boolean;
  normalizedText: string;
  contentDigest: string;
};

type LogicalSection = {
  title: string;
  content: string;
};

export function preflightWorkflowRAGLocalMaterialSelection(files: readonly WorkflowRAGLocalMaterialFileMetadata[]): WorkflowRAGLocalMaterialImportResult | null {
  const rawBytes = files.reduce((total, file) => total + file.fileBytes, 0);
  if (files.length < 1 || files.length > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFiles) {
    return blockedResult(rawBytes, [finding(
      "workflow_rag_material_file_count_invalid",
      "",
      "",
      `一次必须选择 1 至 ${WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFiles} 个文件。`,
    )]);
  }
  const findings: WorkflowRAGLocalMaterialFinding[] = [];
  if (rawBytes > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxRawBytes) {
    findings.push(finding("workflow_rag_material_file_too_large", "", "", "所选文件原始内容总量超过 1 MiB。"));
  }
  for (const file of files) {
    const fileName = normalizedBaseName(file.fileName);
    if (!allowedExtension(fileName)) {
      findings.push(finding("workflow_rag_material_file_type_unsupported", "", "", `${fileName || "未命名文件"} 不是受支持的 Markdown / Text 文件。`));
    }
    if (file.fileBytes > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFileBytes) {
      findings.push(finding("workflow_rag_material_file_too_large", "", "", `${fileName || "未命名文件"} 超过 256 KiB。`));
    }
  }
  return findings.length ? blockedResult(rawBytes, findings) : null;
}

export async function importWorkflowRAGLocalMaterials(files: readonly WorkflowRAGLocalMaterialFile[]): Promise<WorkflowRAGLocalMaterialImportResult> {
  if (files.length < 1 || files.length > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFiles) {
    return blockedResult(files.reduce((total, file) => total + file.bytes.byteLength, 0), [finding(
      "workflow_rag_material_file_count_invalid",
      "",
      "",
      `一次必须选择 1 至 ${WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFiles} 个文件。`,
    )]);
  }

  const rawBytes = files.reduce((total, file) => total + file.bytes.byteLength, 0);
  const findings: WorkflowRAGLocalMaterialFinding[] = [];
  if (rawBytes > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxRawBytes) {
    findings.push(finding("workflow_rag_material_file_too_large", "", "", "所选文件原始内容总量超过 1 MiB。"));
  }

  const decoded: DecodedMaterial[] = [];
  for (const file of files) {
    const fileName = normalizedBaseName(file.fileName);
    const extension = allowedExtension(fileName);
    if (!extension) {
      findings.push(finding("workflow_rag_material_file_type_unsupported", "", "", `${fileName || "未命名文件"} 不是受支持的 Markdown / Text 文件。`));
      continue;
    }
    if (file.bytes.byteLength > WORKFLOW_RAG_LOCAL_MATERIAL_LIMITS.maxFileBytes) {
      findings.push(finding("workflow_rag_material_file_too_large", "", "", `${fileName} 超过 256 KiB。`));
      continue;
    }
    const normalizedText = decodeStrictUTF8(file.bytes);
    if (normalizedText === null) {
      findings.push(finding("workflow_rag_material_utf8_invalid", "", "", `${fileName} 不是严格 UTF-8。`));
      continue;
    }
    if (!normalizedText.trim() || normalizedText.includes("\u0000")) {
      findings.push(finding("workflow_rag_material_content_invalid", "", "", `${fileName} 为空或包含 NUL。`));
      continue;
    }
    if (containsWorkflowRAGSecretMaterial(normalizedText)) {
      findings.push(finding("workflow_rag_secret_material_forbidden", "", "", `${fileName} 命中敏感材料规则。`));
      continue;
    }
    decoded.push({
      fileName,
      extension,
      fileBytes: file.bytes.byteLength,
      selectionIndex: file.selectionIndex,
      sourceType: file.sourceType ?? "document",
      isOfficial: file.isOfficial ?? false,
      normalizedText,
      contentDigest: await sha256Digest(encoder.encode(normalizedText)),
    });
  }

  decoded.sort(compareDecodedMaterials);
  const sourceDigestOwner = new Map<string, string>();
  const fragmentDigestOwner = new Map<string, string>();
  const sources: WorkflowRAGLocalMaterialSource[] = [];
  const fragments: WorkflowRAGLocalMaterialFragment[] = [];

  for (const material of decoded) {
    const sourceIdentityDigest = await sha256Digest(encoder.encode(`${material.fileName}\u0000${material.contentDigest}`));
    const sourceId = `source_${sourceIdentityDigest.slice(7, 19)}`;
    const sourceSlug = slugifySourceName(material.fileName);
    const sourceRef = `local_material/${sourceSlug}-${material.contentDigest.slice(7, 19)}`;
    const previousSource = sourceDigestOwner.get(material.contentDigest);
    if (previousSource) {
      findings.push(finding("workflow_rag_material_source_duplicate", sourceId, "", `${material.fileName} 与 ${previousSource} 内容完全重复。`));
    } else {
      sourceDigestOwner.set(material.contentDigest, material.fileName);
    }

    const sections = material.extension === "txt"
      ? plainTextSections(material.normalizedText, displayStem(material.fileName))
      : markdownSections(material.normalizedText, displayStem(material.fileName));
    if (!sections.length) {
      findings.push(finding("workflow_rag_material_content_invalid", sourceId, "", `${material.fileName} 没有可导入的正文。`));
      continue;
    }

    const sourceFragments: WorkflowRAGLocalMaterialFragment[] = [];
    for (let index = 0; index < sections.length; index += 1) {
      const section = sections[index]!;
      const fragmentRef = `m_${sourceIdentityDigest.slice(7, 17)}_${String(index + 1).padStart(3, "0")}`;
      const content = section.content.trim();
      const contentBytes = encoder.encode(content).byteLength;
      const contentDigest = await sha256Digest(encoder.encode(content));
      const fragment: WorkflowRAGLocalMaterialFragment = {
        sourceId,
        sourceFileName: material.fileName,
        fragmentRef,
        sourceType: material.sourceType,
        sourceRef,
        pageSlug: `local-material/${sourceSlug}-${material.contentDigest.slice(7, 15)}/${String(index + 1).padStart(3, "0")}`,
        title: boundedTitle(section.title, index + 1),
        isOfficial: material.isOfficial,
        content,
        contentBytes,
        contentDigest,
      };
      if (contentBytes < 1 || contentBytes > WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragmentBytes) {
        findings.push(finding("workflow_rag_material_budget_exceeded", sourceId, fragmentRef, `${fragmentRef} 超过单片段预算。`));
      }
      const previousFragment = fragmentDigestOwner.get(contentDigest);
      if (previousFragment) {
        findings.push(finding("workflow_rag_material_fragment_duplicate", sourceId, fragmentRef, `${fragmentRef} 与 ${previousFragment} 正文完全重复。`));
      } else {
        fragmentDigestOwner.set(contentDigest, fragmentRef);
      }
      sourceFragments.push(fragment);
      fragments.push(fragment);
    }
    sources.push({
      sourceId,
      fileName: material.fileName,
      fileBytes: material.fileBytes,
      contentDigest: material.contentDigest,
      sourceType: material.sourceType,
      sourceRef,
      isOfficial: material.isOfficial,
      fragmentRefs: sourceFragments.map((fragment) => fragment.fragmentRef),
    });
  }

  const totalContentBytes = fragments.reduce((total, fragment) => total + fragment.contentBytes, 0);
  if (fragments.length > WORKFLOW_RAG_SNAPSHOT_LIMITS.maxFragments || totalContentBytes > WORKFLOW_RAG_SNAPSHOT_LIMITS.maxTotalContentBytes) {
    findings.push(finding("workflow_rag_material_budget_exceeded", "", "", "最终片段数量或正文总量超过知识快照预算。"));
  }
  if (!fragments.length && !findings.length) {
    findings.push(finding("workflow_rag_material_content_invalid", "", "", "没有生成可提交的知识片段。"));
  }

  return {
    status: findings.length ? "blocked" : "ready",
    sources,
    fragments,
    findings,
    sourceCount: sources.length,
    fragmentCount: fragments.length,
    rawBytes,
    totalContentBytes,
  };
}

function markdownSections(text: string, fallbackTitle: string): LogicalSection[] {
  const lines = text.split("\n");
  const sections: LogicalSection[] = [];
  let currentTitle = fallbackTitle;
  let currentLines: string[] = [];
  let fenceCharacter = "";
  let fenceLength = 0;

  const flush = () => {
    const content = currentLines.join("\n").trim();
    if (content) sections.push(...splitLogicalSection({ title: currentTitle, content }));
    currentLines = [];
  };

  for (const line of lines) {
    const fence = line.match(/^ {0,3}(`{3,}|~{3,})/u)?.[1] ?? "";
    if (fence) {
      const character = fence[0]!;
      if (!fenceCharacter) {
        fenceCharacter = character;
        fenceLength = fence.length;
      } else if (character === fenceCharacter && fence.length >= fenceLength) {
        fenceCharacter = "";
        fenceLength = 0;
      }
      currentLines.push(line);
      continue;
    }
    const heading = fenceCharacter ? null : line.match(/^ {0,3}#{1,6}[ \t]+(.+?)[ \t]*#*[ \t]*$/u);
    if (heading) {
      flush();
      currentTitle = heading[1]!.trim() || fallbackTitle;
    }
    currentLines.push(line);
  }
  flush();
  return sections;
}

function plainTextSections(text: string, fallbackTitle: string): LogicalSection[] {
  const paragraphs = text.split(/\n[ \t]*\n+/u).map((paragraph) => paragraph.trim()).filter(Boolean);
  const sections: LogicalSection[] = [];
  let buffer = "";
  let sectionIndex = 1;
  const flush = () => {
    if (!buffer) return;
    sections.push(...splitLogicalSection({ title: numberedTitle(fallbackTitle, sectionIndex), content: buffer }));
    sectionIndex += 1;
    buffer = "";
  };
  for (const paragraph of paragraphs) {
    const candidate = buffer ? `${buffer}\n\n${paragraph}` : paragraph;
    if (encoder.encode(candidate).byteLength > TARGET_FRAGMENT_BYTES && buffer) flush();
    if (encoder.encode(paragraph).byteLength > TARGET_FRAGMENT_BYTES) {
      flush();
      const pieces = splitByUTF8Budget(paragraph, TARGET_FRAGMENT_BYTES);
      for (const piece of pieces) {
        sections.push({ title: numberedTitle(fallbackTitle, sectionIndex), content: piece });
        sectionIndex += 1;
      }
      continue;
    }
    buffer = buffer ? `${buffer}\n\n${paragraph}` : paragraph;
  }
  flush();
  return sections;
}

function splitLogicalSection(section: LogicalSection): LogicalSection[] {
  const pieces = splitByUTF8Budget(section.content, TARGET_FRAGMENT_BYTES);
  return pieces.map((content, index) => ({
    title: index === 0 ? section.title : `${section.title}（${index + 1}）`,
    content,
  }));
}

function splitByUTF8Budget(text: string, budget: number): string[] {
  const pieces: string[] = [];
  let remaining = text.trim();
  while (encoder.encode(remaining).byteLength > budget) {
    const cut = safeUTF8Cut(remaining, budget);
    const piece = remaining.slice(0, cut).trim();
    if (!piece) break;
    pieces.push(piece);
    remaining = remaining.slice(cut).trimStart();
  }
  if (remaining.trim()) pieces.push(remaining.trim());
  return pieces;
}

function safeUTF8Cut(value: string, budget: number): number {
  let bytes = 0;
  let maximumCut = 0;
  let preferredCut = 0;
  let offset = 0;
  for (const character of value) {
    const characterBytes = encoder.encode(character).byteLength;
    if (bytes + characterBytes > budget) break;
    bytes += characterBytes;
    offset += character.length;
    maximumCut = offset;
    if (/\s/u.test(character) && bytes >= Math.floor(budget / 2)) preferredCut = offset;
  }
  return preferredCut || maximumCut;
}

function decodeStrictUTF8(bytes: Uint8Array): string | null {
  try {
    const decoded = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
    return decoded.replace(/^\uFEFF/u, "").replace(/\r\n?/gu, "\n");
  } catch {
    return null;
  }
}

async function sha256Digest(bytes: Uint8Array): Promise<string> {
  const copy = new Uint8Array(bytes.byteLength);
  copy.set(bytes);
  const digest = await globalThis.crypto.subtle.digest("SHA-256", copy.buffer);
  return `sha256:${Array.from(new Uint8Array(digest), (value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function normalizedBaseName(fileName: string): string {
  return fileName.normalize("NFC").trim();
}

function allowedExtension(fileName: string): DecodedMaterial["extension"] | null {
  const separator = fileName.lastIndexOf(".");
  const extension = separator >= 0 ? fileName.slice(separator + 1).toLowerCase() : "";
  return ALLOWED_EXTENSIONS.has(extension) ? extension as DecodedMaterial["extension"] : null;
}

function displayStem(fileName: string): string {
  const separator = fileName.lastIndexOf(".");
  return (separator > 0 ? fileName.slice(0, separator) : fileName).trim() || "本地材料";
}

function slugifySourceName(fileName: string): string {
  const normalized = displayStem(fileName).normalize("NFKD").toLowerCase().replace(/[\u0300-\u036f]/gu, "");
  const slug = normalized.replace(/[^a-z0-9]+/gu, "-").replace(/^-+|-+$/gu, "").slice(0, 32);
  return slug || "material";
}

function boundedTitle(value: string, index: number): string {
  const title = value.trim() || `本地材料 ${index}`;
  return Array.from(title).slice(0, 160).join("");
}

function numberedTitle(value: string, index: number): string {
  return index === 1 ? value : `${value}（${index}）`;
}

function compareDecodedMaterials(left: DecodedMaterial, right: DecodedMaterial): number {
  const leftName = left.fileName.toLowerCase();
  const rightName = right.fileName.toLowerCase();
  if (leftName !== rightName) return leftName < rightName ? -1 : 1;
  if (left.contentDigest !== right.contentDigest) return left.contentDigest < right.contentDigest ? -1 : 1;
  return left.selectionIndex - right.selectionIndex;
}

function finding(code: WorkflowRAGLocalMaterialFailureCode, sourceId: string, fragmentRef: string, summary: string): WorkflowRAGLocalMaterialFinding {
  return { code, sourceId, fragmentRef, summary };
}

function blockedResult(rawBytes: number, findings: WorkflowRAGLocalMaterialFinding[]): WorkflowRAGLocalMaterialImportResult {
  return { status: "blocked", sources: [], fragments: [], findings, sourceCount: 0, fragmentCount: 0, rawBytes, totalContentBytes: 0 };
}
