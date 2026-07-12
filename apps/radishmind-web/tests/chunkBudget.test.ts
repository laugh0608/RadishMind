import assert from "node:assert/strict";
import test from "node:test";

import {
  WEB_CHUNK_BUDGETS,
  enforceWebChunkBudgets,
  type BuildOutput,
} from "../config/chunkBudget.ts";

function chunk(
  name: string,
  size: number,
  options: { isEntry?: boolean } = {},
): BuildOutput {
  return {
    type: "chunk",
    fileName: `${name}-hash.js`,
    name,
    isEntry: options.isEntry ?? false,
    code: "x".repeat(size),
  };
}

function validBundle(): Record<string, BuildOutput> {
  return {
    "index.js": chunk("index", WEB_CHUNK_BUDGETS.entryBytes, { isEntry: true }),
    "react-vendor.js": chunk("react-vendor", WEB_CHUNK_BUDGETS.namedBytes["react-vendor"]),
    "workflow-node-designer.js": chunk(
      "workflowNodeDesigner",
      WEB_CHUNK_BUDGETS.namedBytes.workflowNodeDesigner,
    ),
    "feature.js": chunk("featurePanel", WEB_CHUNK_BUDGETS.lazyBytes),
    "styles.css": { type: "asset", fileName: "styles.css" },
  };
}

test("web chunk budget accepts each limit and ignores assets", () => {
  const reports = enforceWebChunkBudgets(validBundle());
  assert.equal(reports.length, 4);
  assert.equal(reports.find((report) => report.chunkName === "index")?.budgetKind, "entry");
  assert.equal(reports.find((report) => report.chunkName === "react-vendor")?.budgetKind, "named");
  assert.equal(reports.find((report) => report.chunkName === "featurePanel")?.budgetKind, "lazy");
});

test("web chunk budget rejects an oversized entry with actionable values", () => {
  const bundle = validBundle();
  bundle["index.js"] = chunk("index", WEB_CHUNK_BUDGETS.entryBytes + 1, { isEntry: true });

  assert.throws(
    () => enforceWebChunkBudgets(bundle),
    /entry chunk index .* is 500\.00 KiB \(512001 bytes\), budget 500\.00 KiB \(512000 bytes\)/,
  );
});

test("web chunk budget rejects oversized lazy chunks and missing required chunks", () => {
  const bundle = validBundle();
  delete bundle["react-vendor.js"];
  bundle["feature.js"] = chunk("featurePanel", WEB_CHUNK_BUDGETS.lazyBytes + 1024);

  assert.throws(
    () => enforceWebChunkBudgets(bundle),
    (error: unknown) => {
      assert.ok(error instanceof Error);
      assert.match(error.message, /required named chunk react-vendor was not emitted/);
      assert.match(
        error.message,
        /lazy chunk featurePanel .* is 65\.00 KiB \(66560 bytes\), budget 64\.00 KiB \(65536 bytes\)/,
      );
      return true;
    },
  );
});
