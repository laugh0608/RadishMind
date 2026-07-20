import type {
  ApplicationDevelopmentHandoff,
  ApplicationDevelopmentHandoffInput,
} from "./applicationDevelopmentHandoff.ts";
import type {
  ApplicationDevelopmentEvidenceInput,
  ApplicationDevelopmentReadinessViewModel,
} from "./applicationDevelopmentReadiness.ts";
import type { ApplicationDevelopmentStageId } from "./applicationDevelopmentWorkspace.ts";

export type ApplicationDevelopmentEvidenceReport = ApplicationDevelopmentEvidenceInput & {
  surfaceKey: string;
};

export type ApplicationDevelopmentWorkspaceControls = {
  readiness: ApplicationDevelopmentReadinessViewModel;
  pendingHandoff: ApplicationDevelopmentHandoff | null;
  reportEvidence: (input: ApplicationDevelopmentEvidenceReport) => void;
  issueHandoff: (input: ApplicationDevelopmentHandoffInput) => void;
  consumeHandoff: (targetStage: ApplicationDevelopmentStageId, handoffId: string) => void;
};
