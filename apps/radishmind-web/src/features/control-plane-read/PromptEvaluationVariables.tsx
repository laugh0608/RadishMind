import { useTranslation } from "react-i18next";
import PromptVariableFields from "./PromptVariableFields.tsx";
import type { PromptPlanIssue, PromptPlanSample } from "./promptEvaluationPlanDraft.ts";
import type { PromptTemplateVariable } from "./promptApplicationTemplateConsumer.ts";

export default function PromptEvaluationVariables({ sample, declarations, disabled, issues, onChange }: {
  sample: PromptPlanSample; declarations: PromptTemplateVariable[]; disabled: boolean;
  issues: PromptPlanIssue[]; onChange: (variables: PromptPlanSample["variables"]) => void;
}) {
  const { t } = useTranslation("evaluation");
  return <PromptVariableFields values={sample.variables} declarations={declarations} disabled={disabled}
    issues={issues.filter(issue => issue.itemKey === sample.itemKey).map(issue => ({ field: issue.field, message: t($ => $.promptPlan.issues[issue.code]) }))}
    onChange={onChange} />;
}
