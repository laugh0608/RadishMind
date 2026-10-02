import { useTranslation } from "react-i18next";
import "../../i18n/actionSafetyResources.ts";
import type { ActionSafetyReadProjection } from "./actionSafetyConsumer.ts";

export default function ActionSafetyReadPanel({
  projection,
  title,
  transient = false,
}: {
  projection: ActionSafetyReadProjection | null;
  title?: string;
  transient?: boolean;
}) {
  const { t } = useTranslation("workflow");
  if (!projection) return null;
  const observed = projection.observedSideEffects;
  const heading = title ?? t($ => $.actionSafety.title);
  return (
    <article className="workflow-executor-state-card" aria-label={heading}>
      <div className="action-safety-read-heading">
        <span>{heading}</span>
        <strong>{t($ => $.actionSafety.status[projection.status])}</strong>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div><dt>{t($ => $.actionSafety.owner)}</dt><dd>{t($ => $.actionSafety.ownerKind[projection.owner.kind])} · {projection.owner.id} · v{projection.owner.version}</dd></div>
        <div><dt>{t($ => $.actionSafety.projection)}</dt><dd>{projection.projectionVersion || t($ => $.actionSafety.notRecorded)}</dd></div>
        <div><dt>{t($ => $.actionSafety.decisions)}</dt><dd>{projection.decisions.length}</dd></div>
        <div><dt>{t($ => $.actionSafety.observed)}</dt><dd>{observed ? t($ => $.actionSafety.calls, { provider: observed.providerCalls, tool: observed.toolCalls, confirmation: observed.confirmationCalls }) : t($ => $.actionSafety.notApplicable)}</dd></div>
      </dl>
      {projection.status === "not_recorded_legacy" ? (
        <p className="boundary-note">{t($ => $.actionSafety.legacy)}</p>
      ) : projection.decisions.map((decision) => (
        <div className="prompt-template-summary" key={decision.decisionId}>
          <strong>{decision.actionKind} · {t($ => $.actionSafety.level[decision.effectiveLevel])}</strong>
          <span>{t($ => $.actionSafety.levels, {
            requested: t($ => $.actionSafety.level[decision.requestedLevel]),
            maximum: t($ => $.actionSafety.level[decision.maximumAllowedLevel]),
            target: t($ => $.actionSafety.target[decision.targetKind]), method: decision.method,
          })}</span>
          {decision.blockers.length ? <div><small>{t($ => $.actionSafety.blockersLabel)}</small>{decision.blockers.map(code => <p key={code}>{t($ => $.actionSafety.blocker[code])} · <code>{code}</code></p>)}</div>
            : <small>{t($ => $.actionSafety.confirmation, { state: t($ => $.actionSafety.confirmationState[decision.confirmationState]), policy: decision.policyVersion })}</small>}
        </div>
      ))}
      <p className="boundary-note">{t($ => transient ? $.actionSafety.transient : $.actionSafety.readOnly)}</p>
    </article>
  );
}
