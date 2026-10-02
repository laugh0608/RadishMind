import type { identitySecurity } from "./locales/en-US/identitySecurity.ts";
import type { workflowInspectionProjection } from "./locales/en-US/workflowInspectionProjection.ts";
import type { workflowHandoffProjection } from "./locales/en-US/workflowHandoffProjection.ts";
import type { actionSafety } from "./locales/en-US/actionSafety.ts";
import type { workflowHTTPTool } from "./locales/en-US/workflowHTTPTool.ts";
import type { workflowRAGApplication } from "./locales/en-US/workflowRAGApplication.ts";
import type { workflowRAGPromotion } from "./locales/en-US/workflowRAGPromotion.ts";
import type { workflowRAGExecution } from "./locales/en-US/workflowRAGExecution.ts";
import type { workflowRAGSnapshot } from "./locales/en-US/workflowRAGSnapshot.ts";
import type { workflowInput } from "./locales/en-US/workflowInput.ts";
import type { workflowHistory } from "./locales/en-US/workflowHistory.ts";
import type { workflowExecutor } from "./locales/en-US/workflowExecutor.ts";
import type { workflowTemplate } from "./locales/en-US/workflowTemplate.ts";
import type { workflowPromotion } from "./locales/en-US/workflowPromotion.ts";
import type { workflowHandoff } from "./locales/en-US/workflowHandoff.ts";
import type { workflowInspection } from "./locales/en-US/workflowInspection.ts";
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

import type { workflowDraft } from "./locales/en-US/workflowDraft.ts";

import type { workflowCanvas } from "./locales/en-US/workflowCanvas.ts";

import type { workflowRevision } from "./locales/en-US/workflowRevision.ts";

import type { workflowLibrary } from "./locales/en-US/workflowLibrary.ts";

declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "common";
    enableSelector: true;
    returnNull: false;
    resources: { workflow: { projection: typeof workflowInspectionProjection & typeof workflowHandoffProjection; actionSafety: typeof actionSafety; httpTool: typeof workflowHTTPTool; ragApplication: typeof workflowRAGApplication; ragPromotion: typeof workflowRAGPromotion; ragExecution: typeof workflowRAGExecution; ragSnapshot: typeof workflowRAGSnapshot; input: typeof workflowInput; history: typeof workflowHistory; executor: typeof workflowExecutor; template: typeof workflowTemplate; promotion: typeof workflowPromotion; handoff: typeof workflowHandoff; inspection: typeof workflowInspection; library: typeof workflowLibrary; revision: typeof workflowRevision; canvas: typeof workflowCanvas; draft: typeof workflowDraft }; common: typeof common; shell: typeof shell & { appShell: typeof appShell }; identity: typeof identity & { security: typeof identitySecurity }; prompt: typeof prompt; evaluation: { runReview: typeof runReview }; gateway: { playground: typeof playground; apiIntegration: typeof apiIntegration; apiKey: typeof apiKey }; applications: { catalog: typeof catalog; configuration: typeof configuration; publish: typeof publish; workspacePanel: typeof workspacePanel; workspaceSurface: typeof workspaceSurface; promptWorkspace: typeof promptWorkspace; artifact: typeof artifact; artifactLibrary: typeof artifactLibrary } };
  }
}
