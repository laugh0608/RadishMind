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

export default defineConfig({
  plugins: [react(), webChunkBudgetPlugin()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules/i18next/") || id.includes("node_modules/react-i18next/")) return "i18n-vendor";
          if (/\/i18n\/locales\/(?:zh-CN|en-US)\/(?:common|shell|identity|appShell)\.ts$/.test(id)) return "i18n-core-resources";
          if (id.includes("node_modules/react/") || id.includes("node_modules/react-dom/") || id.includes("node_modules/scheduler/")) return "react-vendor";
          if (id.endsWith("/workflowRunRecordConsumer.ts")) return "workflow-run-record-consumer";
          if (id.endsWith("/savedWorkflowDraftConsumer.ts")) return "workflow-saved-draft-consumer";
        },
      },
    },
  },
});
