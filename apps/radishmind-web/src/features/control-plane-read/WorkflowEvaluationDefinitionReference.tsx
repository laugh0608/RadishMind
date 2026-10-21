import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { ApplicationEvaluationConfig, ApplicationEvaluationDefinitionTarget } from "./applicationEvaluationCampaignConsumer.ts";
import { listWorkflowDefinitionCandidates, readWorkflowDefinitionPromotionAuthority, type WorkflowDefinitionPromotionAuthority, type WorkflowDefinitionPromotionConfig } from "./workflowDefinitionPromotionConsumer.ts";
import { useLocalIdentity } from "../local-identity/localIdentityContext.ts";

type Props = {
  config: ApplicationEvaluationConfig;
  disabled: boolean;
  onAdopt: (target: ApplicationEvaluationDefinitionTarget) => void;
};

export default function WorkflowEvaluationDefinitionReference({ config, disabled, onAdopt }: Props) {
  const { t } = useTranslation("evaluation");
  const identity = useLocalIdentity();
  const [definitions, setDefinitions] = useState<string[]>([]);
  const [selected, setSelected] = useState("");
  const [authority, setAuthority] = useState<WorkflowDefinitionPromotionAuthority | null>(null);
  const [state, setState] = useState<"idle" | "loading" | "ready" | "failed">("idle");
  const generation = useRef(0);
  const promotionConfig = useMemo<WorkflowDefinitionPromotionConfig>(() => ({ ...config,
    mode: config.mode === "offline" ? "offline" : "dev_workflow_definition_promotion_http",
  }), [config]);
  useEffect(() => () => { generation.current += 1; }, []);

  async function read() {
    const request = ++generation.current;
    const revision = identity?.readAuthorityRevision();
    const current = () => request === generation.current && revision === identity?.readAuthorityRevision();
    setAuthority(null);
    setState("loading");
    try {
      const candidates = await listWorkflowDefinitionCandidates(promotionConfig, config.applicationId);
      if (!current()) return;
      const ids = [...new Set(candidates.filter(candidate => candidate.snapshot.executionProfile !== "workflow_definition_http_tool_v1").map(candidate => candidate.definitionId))];
      setDefinitions(ids);
      if (!selected || !ids.includes(selected)) {
        setSelected("");
        setState("idle");
        return;
      }
      const result = await readWorkflowDefinitionPromotionAuthority(promotionConfig, config.applicationId, selected);
      if (!current()) return;
      setAuthority(result);
      setState("ready");
    } catch {
      if (current()) { setAuthority(null); setState("failed"); }
    }
  }
  const activation = authority?.activation;
  const version = activation?.state === "active" ? authority?.versions.find(value => value.version === activation.activeVersion) : null;
  const supported = version && ["workflow_definition_executor_v1", "workflow_definition_executor_v2"].includes(version.snapshot.executionProfile);
  function adopt() {
    if (!supported || !version || !activation || state !== "ready" || disabled) return;
    const contract = version.snapshot.inputContract;
    onAdopt({ definitionId: authority!.definitionId, expectedPointerVersion: activation.pointerVersion,
      expectedDefinitionVersion: version.version, expectedDefinitionDigest: version.definitionDigest,
      inputContract: "fields" in contract ? contract : null });
  }
  return <section className="workflow-evaluation-reference" aria-label={t($ => $.campaign.referenceTitle)}>
    <strong>{t($ => $.campaign.referenceTitle)}</strong>
    <p>{t($ => $.campaign.referenceBoundary)}</p>
    <label><span>{t($ => $.campaign.referenceFamily)}</span><select value={selected} disabled={disabled} onChange={event => {
      generation.current += 1; setSelected(event.target.value); setAuthority(null); setState("idle");
    }}><option value="">{t($ => $.campaign.referenceSelect)}</option>{definitions.map(id => <option key={id} value={id}>{id}</option>)}</select></label>
    <button type="button" className="secondary-action" onClick={() => void read()} disabled={disabled || state === "loading" || config.mode === "offline"}>{t($ => $.campaign.referenceRead)}</button>
    {state === "loading" ? <p role="status">{t($ => $.campaign.referenceLoading)}</p> : null}
    {state === "failed" ? <p role="alert">{t($ => $.campaign.referenceFailed)}</p> : null}
    {state === "ready" && (!version || !supported) ? <p role="status">{t($ => $.campaign.referenceInactive)}</p> : null}
    {version && activation ? <dl>
      <dt>{t($ => $.campaign.referenceVersion)}</dt><dd>{version.definitionId} · v{version.version} · {activation.pointerVersion}</dd>
      <dt>{t($ => $.campaign.digest)}</dt><dd><code>{version.definitionDigest}</code></dd>
      {"fields" in version.snapshot.inputContract ? <><dt>{t($ => $.campaign.referenceContract)}</dt><dd>{version.snapshot.inputContract.contractId}<br /><code>{version.snapshot.inputContract.contractDigest}</code><ul>{version.snapshot.inputContract.fields.map(field => <li key={field.name}>{field.name} · {field.valueType} · {String(field.required)}</li>)}</ul></dd></> : null}
    </dl> : null}
    <button type="button" onClick={adopt} disabled={disabled || state !== "ready" || !supported}>{t($ => $.campaign.referenceAdopt)}</button>
  </section>;
}
