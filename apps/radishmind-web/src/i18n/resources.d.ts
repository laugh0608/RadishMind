import "i18next";
import type { common } from "./locales/en-US/common.ts";
import type { shell } from "./locales/en-US/shell.ts";
import type { identity } from "./locales/en-US/identity.ts";
import type { prompt } from "./locales/en-US/prompt.ts";
import type { catalog } from "./locales/en-US/catalog.ts";
import type { configuration } from "./locales/en-US/configuration.ts";
import type { publish } from "./locales/en-US/publish.ts";
import type { workspacePanel } from "./locales/en-US/workspacePanel.ts";
import type { workspaceSurface } from "./locales/en-US/workspaceSurface.ts";
import type { promptWorkspace } from "./locales/en-US/promptWorkspace.ts";
import type { playground } from "./locales/en-US/playground.ts";
import type { artifact } from "./locales/en-US/artifact.ts";
import type { artifactLibrary } from "./locales/en-US/artifactLibrary.ts";
import type { apiIntegration } from "./locales/en-US/apiIntegration.ts";
import type { apiKey } from "./locales/en-US/apiKey.ts";
import type { runReview } from "./locales/en-US/runReview.ts";
import type { appShell } from "./locales/en-US/appShell.ts";

declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "common";
    enableSelector: true;
    returnNull: false;
    resources: { common: typeof common; shell: typeof shell & { appShell: typeof appShell }; identity: typeof identity; prompt: typeof prompt; evaluation: { runReview: typeof runReview }; gateway: { playground: typeof playground; apiIntegration: typeof apiIntegration; apiKey: typeof apiKey }; applications: { catalog: typeof catalog; configuration: typeof configuration; publish: typeof publish; workspacePanel: typeof workspacePanel; workspaceSurface: typeof workspaceSurface; promptWorkspace: typeof promptWorkspace; artifact: typeof artifact; artifactLibrary: typeof artifactLibrary } };
  }
}
