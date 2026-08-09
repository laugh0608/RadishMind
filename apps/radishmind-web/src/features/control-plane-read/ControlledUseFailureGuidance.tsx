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
  const guidance = controlledUseFailureGuidance(owner, failureCode);
  if (!guidance) return null;

  return (
    <article className="controlled-use-failure-guidance" aria-label="Controlled use eligibility guidance">
      <div className="application-api-card-heading">
        <div><p className="eyebrow">Eligibility blocked</p><h5>{guidance.title}</h5></div>
        <span className="status-badge bad">no provider call</span>
      </div>
      <p>{guidance.summary}</p>
      <p className="boundary-note">{guidance.sideEffectSummary}</p>
      <a href={`#${guidance.assignmentAnchor}`}>{guidance.assignmentLabel} <span aria-hidden="true">→</span></a>
    </article>
  );
}
