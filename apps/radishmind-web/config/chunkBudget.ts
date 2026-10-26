export type BuildOutput = {
  type: string;
  fileName: string;
  name?: string;
  isEntry?: boolean;
  facadeModuleId?: string | null;
  moduleIds?: string[];
  code?: string;
};

export type ChunkBudgetReport = {
  chunkName: string;
  fileName: string;
  actualBytes: number;
  budgetBytes: number;
  budgetKind: "entry" | "named" | "lazy";
};

const KIB = 1024;

export const WEB_CHUNK_BUDGETS = Object.freeze({
  entryBytes: 500 * KIB,
  lazyBytes: 64 * KIB,
  namedBytes: Object.freeze({
    "react-vendor": 220 * KIB,
    workflowNodeDesigner: 220 * KIB,
  }),
});

export function enforceWebChunkBudgets(bundle: Record<string, BuildOutput>): ChunkBudgetReport[] {
  const chunks = Object.values(bundle).filter(
    (output): output is BuildOutput & { type: "chunk"; code: string } =>
      output.type === "chunk" && typeof output.code === "string",
  );
  const chunkNames = new Set(chunks.map((chunk) => chunk.name ?? ""));
  const missingNamedChunks = Object.keys(WEB_CHUNK_BUDGETS.namedBytes).filter(
    (chunkName) => !chunkNames.has(chunkName),
  );
  const reports = chunks.map(buildChunkBudgetReport);
  const bootstrapBytes = chunks.filter(isBootstrapChunk).reduce((total, chunk) => total + new TextEncoder().encode(chunk.code).byteLength, 0);
  const bootstrapExceeded = chunks.some(chunk => !chunk.isEntry && isBootstrapChunk(chunk)) && bootstrapBytes > WEB_CHUNK_BUDGETS.entryBytes;
  const violations = reports.filter((report) => report.actualBytes > report.budgetBytes);

  if (missingNamedChunks.length > 0 || violations.length > 0 || bootstrapExceeded) {
    const details = [
      ...(bootstrapExceeded ? [`- combined bootstrap chunks are ${formatKiB(bootstrapBytes)}, budget ${formatKiB(WEB_CHUNK_BUDGETS.entryBytes)}`] : []),
      ...missingNamedChunks.map((chunkName) => `- required named chunk ${chunkName} was not emitted`),
      ...violations.map(
        (report) =>
          `- ${report.budgetKind} chunk ${report.chunkName} (${report.fileName}) is ${formatKiB(report.actualBytes)} (${report.actualBytes} bytes), budget ${formatKiB(report.budgetBytes)} (${report.budgetBytes} bytes)`,
      ),
    ];
    throw new Error(`RadishMind Web chunk budget failed:\n${details.join("\n")}`);
  }

  return reports;
}

function buildChunkBudgetReport(chunk: BuildOutput & { type: "chunk"; code: string }): ChunkBudgetReport {
  const chunkName = chunk.name?.trim() || chunk.fileName;
  const namedBudget = WEB_CHUNK_BUDGETS.namedBytes[chunkName as keyof typeof WEB_CHUNK_BUDGETS.namedBytes];
  const budgetKind = isBootstrapChunk(chunk) ? "entry" : namedBudget !== undefined ? "named" : "lazy";
  const budgetBytes = isBootstrapChunk(chunk)
    ? WEB_CHUNK_BUDGETS.entryBytes
    : namedBudget ?? WEB_CHUNK_BUDGETS.lazyBytes;

  return {
    chunkName,
    fileName: chunk.fileName,
    actualBytes: new TextEncoder().encode(chunk.code).byteLength,
    budgetBytes,
    budgetKind,
  };
}

function formatKiB(bytes: number): string {
  return `${(bytes / KIB).toFixed(2)} KiB`;
}

function isBootstrapChunk(chunk: BuildOutput): boolean {
  return Boolean(chunk.isEntry || [chunk.facadeModuleId, ...(chunk.moduleIds ?? [])].some(id => id?.replaceAll("\\", "/").endsWith("/src/app/App.tsx")));
}
