import { mergeConfig } from "vite";
import webConfig from "../../vite.config";

// Browser regressions use only the launcher environment, never a developer's .env files.
export default mergeConfig(webConfig, { envDir: false, server: { strictPort: true } });
