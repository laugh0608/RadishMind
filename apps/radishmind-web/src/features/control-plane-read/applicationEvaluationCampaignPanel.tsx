import { lazy, useMemo } from "react";
import { readApplicationEvaluationConfig } from "./applicationEvaluationCampaignConsumer.ts";
export type { LoadState, PendingPlanOperation } from "./ApplicationEvaluationCampaignWorkspace.tsx";

import { useLocalIdentity } from "../local-identity/localIdentityContext.ts";

const Workspace = lazy(() => import("./ApplicationEvaluationCampaignWorkspace.tsx"));

export type EvaluationPanelProps = {
  applicationId: string;
  applicationName: string;
  applicationKind: string;
  workspaceId: string;
  applicationActive: boolean;
};
// Scope changes discard all private draft/source state before a new workspace
// can render. Keep this boundary separate from the lazily loaded task owners.
export default function ApplicationEvaluationCampaignPanel(props: EvaluationPanelProps) {
  const identity = useLocalIdentity();
  const config = useMemo(() => readApplicationEvaluationConfig(props), [props.applicationId, props.workspaceId]);
  const scope = JSON.stringify([identity?.profile.account.userId, identity?.authorityRevision, config.tenantRef, config.workspaceId, config.applicationId, config.subjectRef,
    config.authMode, config.baseUrl, config.environment, config.mode, props.applicationKind, props.applicationActive]);
  return <Workspace key={scope} {...props} config={config} />;
}
