export type WorkflowRAGSnapshotFindingMessage =
  | { key: "fileCount"; values: { maximum: number } }
  | { key: "rawBudget" }
  | { key: "unsupportedFile"; values: { file: string } }
  | { key: "fileBudget"; values: { file: string } }
  | { key: "utf8"; values: { file: string } }
  | { key: "invalidContent"; values: { file: string } }
  | { key: "fileSecret"; values: { file: string } }
  | { key: "duplicateSource"; values: { source: string; previous: string } }
  | { key: "noContent"; values: { file: string } }
  | { key: "fragmentBudget"; values: { fragment: string } }
  | { key: "duplicateContent"; values: { fragment: string; previous: string } }
  | { key: "snapshotBudget" }
  | { key: "noFragments" }
  | { key: "metadata" }
  | { key: "fragmentCount"; values: { maximum: number } }
  | { key: "missingSource" }
  | { key: "duplicateRef"; values: { fragment: string; previous: string } }
  | { key: "contentBudget" }
  | { key: "replacementInvalid" }
  | { key: "fragmentSecret" }
  | { key: "fragmentInvalid" };
