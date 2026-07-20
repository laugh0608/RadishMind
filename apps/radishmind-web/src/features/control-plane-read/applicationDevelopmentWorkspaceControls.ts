import type {
  ApplicationDevelopmentHandoff,
  ApplicationDevelopmentHandoffInput,
} from "./applicationDevelopmentHandoff.ts";
import type {
  ApplicationDevelopmentEvidenceInput,
  ApplicationDevelopmentReadinessViewModel,
} from "./applicationDevelopmentReadiness.ts";
import type { ApplicationDevelopmentStageId } from "./applicationDevelopmentWorkspace.ts";

export type ApplicationDevelopmentWorkspaceControls = {
  readiness: ApplicationDevelopmentReadinessViewModel;
  pendingHandoff: ApplicationDevelopmentHandoff | null;
  reportEvidence: (input: ApplicationDevelopmentEvidenceInput) => void;
  issueHandoff: (input: ApplicationDevelopmentHandoffInput) => void;
  consumeHandoff: (targetStage: ApplicationDevelopmentStageId, handoffId: string) => void;
};
