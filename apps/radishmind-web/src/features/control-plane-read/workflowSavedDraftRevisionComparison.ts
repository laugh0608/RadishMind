import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";

export type WorkflowSavedDraftRevisionChange = {
  kind:
    | "metadata"
    | "layout"
    | "validation"
    | "provenance"
    | "node_added"
    | "node_removed"
    | "node_changed"
    | "edge_added"
    | "edge_removed"
    | "edge_changed";
  subject: string;
  summary: string;
};

export type WorkflowSavedDraftRevisionComparison = {
  changed: boolean;
  metadataChangeCount: number;
  nodeChangeCount: number;
  edgeChangeCount: number;
  reviewContextChangeCount: number;
  changes: WorkflowSavedDraftRevisionChange[];
};

export function compareWorkflowSavedDraftRevision(
  historical: WorkflowDraftDesignerDraft,
  active: WorkflowDraftDesignerDraft,
): WorkflowSavedDraftRevisionComparison {
  const changes: WorkflowSavedDraftRevisionChange[] = [];
  if (historical.label !== active.label) {
    changes.push({ kind: "metadata", subject: "name", summary: `${historical.label} → ${active.label}` });
  }
  if (historical.summary !== active.summary) {
    changes.push({ kind: "metadata", subject: "summary", summary: "草案说明已变更。" });
  }
  if (historical.baseDefinitionVersion !== active.baseDefinitionVersion) {
    changes.push({
      kind: "metadata",
      subject: "base definition",
      summary: `${historical.baseDefinitionVersion ?? 0} → ${active.baseDefinitionVersion ?? 0}`,
    });
  }
  compareMetadataValue(
    "workflow definition",
    historical.workflowDefinitionId,
    active.workflowDefinitionId,
    changes,
  );
  compareMetadataValue(
    "provider profile",
    historical.providerProfileRef,
    active.providerProfileRef,
    changes,
  );
  compareMetadataValue(
    "execution profile",
    historical.executionProfile ?? "review_only",
    active.executionProfile ?? "review_only",
    changes,
  );

  compareByStableId(
    historical.nodes,
    active.nodes,
    (node) => node.nodeId,
    (node) => JSON.stringify({
      label: node.label,
      nodeType: node.nodeType,
      providerRef: node.providerRef,
      toolRef: node.toolRef,
      ragRef: node.ragRef,
      inputContractFields: node.inputContractFields,
      outputContractFields: node.outputContractFields,
      outputMappingSummary: node.outputMappingSummary,
      riskLevel: node.riskLevel,
      requiresConfirmation: node.requiresConfirmation,
    }),
    "node",
    changes,
  );
  compareByStableId(
    historical.edges,
    active.edges,
    (edge) => edge.edgeId,
    (edge) => JSON.stringify({
      fromNodeId: edge.fromNodeId,
      toNodeId: edge.toNodeId,
      conditionSummary: edge.conditionSummary,
    }),
    "edge",
    changes,
  );
  if (stableFingerprint(normalizedLayout(historical)) !== stableFingerprint(normalizedLayout(active))) {
    changes.push({
      kind: "layout",
      subject: "designer layout",
      summary: "节点画布布局已变更。",
    });
  }
  if (stableFingerprint(normalizedReviewContext(historical)) !== stableFingerprint(normalizedReviewContext(active))) {
    changes.push({
      kind: "validation",
      subject: "review context",
      summary: "校验、风险或阻塞能力摘要已变更。",
    });
  }
  if (stableFingerprint(historical.derivation ?? null) !== stableFingerprint(active.derivation ?? null)) {
    changes.push({
      kind: "provenance",
      subject: "derivation",
      summary: "草案直接来源已变更。",
    });
  }
  return {
    changed: changes.length > 0,
    metadataChangeCount: changes.filter((change) => change.kind === "metadata").length,
    nodeChangeCount: changes.filter((change) => change.kind.startsWith("node_")).length,
    edgeChangeCount: changes.filter((change) => change.kind.startsWith("edge_")).length,
    reviewContextChangeCount: changes.filter((change) =>
      change.kind === "layout" || change.kind === "validation" || change.kind === "provenance"
    ).length,
    changes,
  };
}

function compareMetadataValue(
  subject: string,
  historical: string,
  active: string,
  changes: WorkflowSavedDraftRevisionChange[],
) {
  if (historical !== active) {
    changes.push({
      kind: "metadata",
      subject,
      summary: `${historical || "未设置"} → ${active || "未设置"}`,
    });
  }
}

function normalizedLayout(draft: WorkflowDraftDesignerDraft) {
  return {
    source: draft.designerLayout.source,
    persistence: draft.designerLayout.persistence,
    nodePositions: [...draft.designerLayout.nodePositions]
      .sort((left, right) => left.nodeId.localeCompare(right.nodeId)),
  };
}

function normalizedReviewContext(draft: WorkflowDraftDesignerDraft) {
  return {
    readiness: normalizedByStableId(draft.readiness, (item) => item.checkId),
    risks: normalizedByStableId(draft.risks, (item) => item.riskId),
    blockedCapabilities: normalizedByStableId(
      draft.blockedCapabilities,
      (item) => item.capabilityId,
    ),
  };
}

function normalizedByStableId<T>(values: T[], id: (value: T) => string): T[] {
  return [...values].sort((left, right) => id(left).localeCompare(id(right)));
}

function stableFingerprint(value: unknown): string {
  return JSON.stringify(value);
}

function compareByStableId<T>(
  historical: T[],
  active: T[],
  id: (value: T) => string,
  fingerprint: (value: T) => string,
  subjectKind: "node" | "edge",
  changes: WorkflowSavedDraftRevisionChange[],
) {
  const historicalById = new Map(historical.map((value) => [id(value), value]));
  const activeById = new Map(active.map((value) => [id(value), value]));
  for (const [stableId, historicalValue] of historicalById) {
    const activeValue = activeById.get(stableId);
    if (!activeValue) {
      changes.push({
        kind: subjectKind === "node" ? "node_removed" : "edge_removed",
        subject: stableId,
        summary: `${subjectKind} 已从当前草案移除。`,
      });
    } else if (fingerprint(historicalValue) !== fingerprint(activeValue)) {
      changes.push({
        kind: subjectKind === "node" ? "node_changed" : "edge_changed",
        subject: stableId,
        summary: `${subjectKind} 配置已变更。`,
      });
    }
  }
  for (const stableId of activeById.keys()) {
    if (!historicalById.has(stableId)) {
      changes.push({
        kind: subjectKind === "node" ? "node_added" : "edge_added",
        subject: stableId,
        summary: `${subjectKind} 已加入当前草案。`,
      });
    }
  }
}
