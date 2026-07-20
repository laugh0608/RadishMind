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
          if (id.includes("node_modules/react/") || id.includes("node_modules/react-dom/") || id.includes("node_modules/scheduler/")) return "react-vendor";
          if (id.endsWith("/workflowRunRecordConsumer.ts")) return "workflow-run-record-consumer";
        },
      },
    },
  },
});
