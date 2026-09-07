import type { ActionSafetyReadProjection } from "./actionSafetyConsumer.ts";

export default function ActionSafetyReadPanel({
  projection,
  title = "Action Safety evidence",
  transient = false,
}: {
  projection: ActionSafetyReadProjection | null;
  title?: string;
  transient?: boolean;
}) {
  if (!projection) return null;
  const observed = projection.observedSideEffects;
  return (
    <article className="workflow-executor-state-card" aria-label={title}>
      <div className="action-safety-read-heading">
        <span>{title}</span>
        <strong>{projection.status}</strong>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div><dt>Owner</dt><dd>{projection.owner.kind} · {projection.owner.id} · v{projection.owner.version}</dd></div>
        <div><dt>Projection</dt><dd>{projection.projectionVersion || "not recorded"}</dd></div>
        <div><dt>Decisions</dt><dd>{projection.decisions.length}</dd></div>
        <div><dt>Observed</dt><dd>{observed ? `provider ${observed.providerCalls} · tool ${observed.toolCalls} · confirmation ${observed.confirmationCalls}` : "not applicable"}</dd></div>
      </dl>
      {projection.status === "not_recorded_legacy" ? (
        <p className="boundary-note">旧 owner 未记录 Action Safety 快照；界面不会用当前策略回填历史结论。</p>
      ) : projection.decisions.map((decision) => (
        <div className="prompt-template-summary" key={decision.decisionId}>
          <strong>{decision.actionKind} · {decision.effectiveLevel}</strong>
          <span>{decision.requestedLevel} → max {decision.maximumAllowedLevel} · {decision.targetKind} · {decision.method}</span>
          <small>{decision.blockers.length ? `blockers: ${decision.blockers.join(", ")}` : `confirmation: ${decision.confirmationState} · policy ${decision.policyVersion}`}</small>
        </div>
      ))}
      <p className="boundary-note">
        {transient
          ? "该投影仅随当前响应展示；候选动作仍由既有 review owner 承接，不创建新的 mutation owner。"
          : "只读投影仅解释既有 owner；它不授予执行、确认或业务写权限。"}
      </p>
    </article>
  );
}
