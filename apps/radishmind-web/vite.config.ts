import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

import { enforceWebChunkBudgets } from "./config/chunkBudget";

function webChunkBudgetPlugin(): Plugin {
  return {
    name: "radishmind-web-chunk-budget",
    generateBundle(_options, bundle) {
      enforceWebChunkBudgets(bundle);
    },
  };
}

export default defineConfig(({ mode }) => ({
  // Workbench scope and addresses arrive at runtime; do not embed ambient VITE_* values.
  envPrefix: mode === "workbench" ? [] : "VITE_",
  define: mode === "workbench" ? { "import.meta.env.VITE_RADISHMIND_RUNTIME_PROFILE": JSON.stringify("local-workbench") } : {},
  plugins: [react(), webChunkBudgetPlugin(), {
    name: "radishmind-workbench-build-marker",
    generateBundle() {
      if (mode === "workbench") this.emitFile({ type: "asset", fileName: "workbench-build.json", source: JSON.stringify({ runtime_profile: "local-workbench" }) });
    },
  }],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.endsWith("/runtime/workbenchConfiguration.ts") || id.endsWith("/contracts/typescript/workbench-public-config.ts")) return "workbench-runtime";
          if (id.includes("node_modules/i18next/") || id.includes("node_modules/react-i18next/")) return "i18n-vendor";
          if (/\/i18n\/locales\/(?:zh-CN|en-US)\/(?:common|shell|identity|appShell)\.ts$/.test(id)) return "i18n-core-resources";
          if (id.includes("node_modules/react/") || id.includes("node_modules/react-dom/") || id.includes("node_modules/scheduler/")) return "react-vendor";
          if (id.endsWith("/workflowRunRecordConsumer.ts")) return "workflow-run-record-consumer";
          if (id.endsWith("/savedWorkflowDraftConsumer.ts")) return "workflow-saved-draft-consumer";
        },
      },
    },
  },
}));
