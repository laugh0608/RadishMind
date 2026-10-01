import { useTranslation } from "react-i18next";
import type { prompt } from "../../i18n/locales/en-US/prompt.ts";
import {
  controlledUseFailureGuidance,
  type ControlledUseOwner,
} from "./controlledUseFailureGuidance.ts";

export default function ControlledUseFailureGuidance({
  owner,
  failureCode,
}: {
  owner: ControlledUseOwner;
  failureCode: string;
}) {
  const { t } = useTranslation("prompt");
  const guidance = controlledUseFailureGuidance(owner, failureCode);
  if (!guidance) return null;
  const code = failureCode as keyof typeof prompt.eligibilityFailures;

  return (
    <article className="controlled-use-failure-guidance" aria-label={t($ => $.guidanceRegion)}>
      <div className="application-api-card-heading">
        <div><p className="eyebrow">{t($ => $.eligibilityBlocked)}</p><h5>{t($ => $.eligibilityFailures[code].title)}</h5></div>
        <span className="status-badge bad">{t($ => $.noProviderCall)}</span>
      </div>
      <p>{t($ => $.eligibilityFailures[code].summary)}</p>
      <p className="boundary-note">{t($ => $.sideEffectBoundary)}</p>
      <a href={`#${guidance.assignmentAnchor}`}>{owner === "agent_session" ? t($ => $.openAgentAssignment) : t($ => $.openPromptAssignment)} <span aria-hidden="true">→</span></a>
    </article>
  );
}
