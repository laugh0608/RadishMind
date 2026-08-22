import { useEffect, useMemo, useRef, useState } from "react";

import {
  createWorkflowDefinitionHTTPToolActionPlan,
  decideWorkflowHTTPToolActionPlan,
  initialWorkflowHTTPToolActionConsumerState,
  readWorkflowHTTPToolActionConsumerConfig,
  readWorkflowHTTPToolActionPlan,
  readWorkflowHTTPToolActionPlanReference,
  rememberWorkflowHTTPToolActionPlanReference,
  workflowHTTPToolActionPermissions,
  type WorkflowHTTPToolHumanDecision,
  type WorkflowHTTPToolPublicArguments,
} from "./workflowHTTPToolActionConsumer.ts";
import { WorkflowHTTPToolActionPanel } from "./workflowHTTPToolActionPanel.tsx";
import {
  executeWorkflowHTTPToolActionPlan,
  initialWorkflowHTTPToolExecutionState,
  restoreWorkflowHTTPToolExecutionState,
} from "./workflowHTTPToolExecutionConsumer.ts";
import { WorkflowHTTPToolExecutionPanel } from "./workflowHTTPToolExecutionPanel.tsx";
import type { WorkflowDefinitionVersion } from "./workflowDefinitionPromotionConsumer.ts";

const config = readWorkflowHTTPToolActionConsumerConfig();

type Props = {
  workspaceId: string;
  applicationId: string;
  version: WorkflowDefinitionVersion;
  activationPointerVersion: number;
  onRunRecorded: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
};

export default function WorkflowDefinitionHTTPToolRuntimePanel({
  applicationId,
  workspaceId,
  version,
  activationPointerVersion,
  onRunRecorded,
  onOpenRun,
}: Props) {
  const epochRef = useRef(0);
  const liveConfig = useMemo(() => ({ ...config, workspaceId }), [workspaceId]);
  const [actionState, setActionState] = useState(() => initialWorkflowHTTPToolActionConsumerState(liveConfig));
  const [executionState, setExecutionState] = useState(() => initialWorkflowHTTPToolExecutionState(liveConfig));
  const [resourceKey, setResourceKey] = useState("");
  const [locale, setLocale] = useState("");
  const [inputText, setInputText] = useState("");
  const [model, setModel] = useState("");
  const permissions = useMemo(() => workflowHTTPToolActionPermissions(liveConfig), [liveConfig]);
  const toolNode = version.snapshot.nodes.find((node) => node.nodeType === "http_tool") ?? null;
  const eligible = version.schemaVersion === "workflow_definition_version.v3" &&
    version.snapshot.executionProfile === "workflow_definition_http_tool_v1" && Boolean(toolNode) && activationPointerVersion > 0;
  const eligibility = {
    eligible,
    draftVersion: 0,
    nodeId: toolNode?.nodeId ?? "",
    toolId: toolNode?.toolRef ?? "",
    reasons: eligible ? [] : [{
      code: "workflow_definition_http_tool_authority_invalid",
      summary: "The exact active Definition v3 HTTP Tool authority is unavailable.",
    }],
  };

  useEffect(() => {
    epochRef.current += 1;
    const epoch = epochRef.current;
    setActionState(initialWorkflowHTTPToolActionConsumerState(liveConfig));
    setExecutionState(initialWorkflowHTTPToolExecutionState(liveConfig));
    setResourceKey("");
    setLocale("");
    setInputText("");
    setModel("");
    if (liveConfig.mode !== "dev_workflow_http_tool_http") return;
    const reference = readWorkflowHTTPToolActionPlanReference();
    if (!reference || reference.workspaceId !== liveConfig.workspaceId || reference.applicationId !== applicationId ||
      reference.sourceKind !== "workflow_definition" || reference.sourceId !== version.definitionId) return;
    setActionState((state) => ({ ...state, status: "reading", summary: "Reloading the durable Definition action plan.", failureCode: "" }));
    readWorkflowHTTPToolActionPlan(liveConfig, {
      planId: reference.planId,
      applicationId,
      sourceKind: "workflow_definition",
      workflowDefinitionId: version.definitionId,
    }).then(async (state) => {
      if (epochRef.current !== epoch) return;
      setActionState(state);
      if (state.actionPlan?.status !== "consumed" || !reference.runId) return;
      setExecutionState((current) => ({ ...current, summary: "Reloading the exact durable v9 run; no execution request is sent." }));
      const restored = await restoreWorkflowHTTPToolExecutionState(liveConfig, state.actionPlan, reference.runId);
      if (epochRef.current === epoch) setExecutionState(restored);
    });
    return () => { epochRef.current += 1; };
  }, [applicationId, liveConfig, version.definitionId, version.version, version.definitionDigest, activationPointerVersion]);

  async function createPlan(publicArguments: WorkflowHTTPToolPublicArguments) {
    if (!toolNode) return;
    const epoch = epochRef.current;
    setActionState((state) => ({ ...state, status: "creating", summary: "Creating a plan from the exact active Definition authority.", failureCode: "" }));
    const state = await createWorkflowDefinitionHTTPToolActionPlan(liveConfig, {
      definitionId: version.definitionId,
      applicationId,
      nodeId: toolNode.nodeId,
      expectedDefinitionVersion: version.version,
      expectedDefinitionDigest: version.definitionDigest,
      expectedPointerVersion: activationPointerVersion,
      publicArguments,
    });
    if (epochRef.current !== epoch) return;
    if (state.actionPlan) rememberWorkflowHTTPToolActionPlanReference(state.actionPlan);
    setActionState(state);
  }

  async function reloadPlan() {
    const plan = actionState.actionPlan;
    if (!plan) return;
    const epoch = epochRef.current;
    setActionState((state) => ({ ...state, status: "reading", summary: "Reloading durable plan detail.", failureCode: "" }));
    const state = await readWorkflowHTTPToolActionPlan(liveConfig, plan);
    if (epochRef.current === epoch) setActionState(state);
  }

  async function decide(decision: WorkflowHTTPToolHumanDecision) {
    const plan = actionState.actionPlan;
    if (!plan) return;
    const epoch = epochRef.current;
    setActionState((state) => ({ ...state, status: "deciding", summary: `Recording ${decision} against plan version ${plan.recordVersion}.`, failureCode: "" }));
    const state = await decideWorkflowHTTPToolActionPlan(liveConfig, plan, decision);
    if (epochRef.current !== epoch) return;
    if (state.actionPlan) rememberWorkflowHTTPToolActionPlanReference(state.actionPlan);
    setActionState(state);
  }

  async function execute() {
    const plan = actionState.actionPlan;
    if (!plan) return;
    const epoch = epochRef.current;
    const boundedInput = inputText;
    const requestedModel = model;
    setInputText("");
    setModel("");
    setExecutionState((state) => ({ ...state, status: "executing", summary: "Starting the single confirmed Definition-bound attempt.", failureCode: "", run: null }));
    const state = await executeWorkflowHTTPToolActionPlan(liveConfig, plan, { inputText: boundedInput, model: requestedModel });
    if (epochRef.current !== epoch) return;
    setExecutionState(state);
    if (state.actionPlan) {
      rememberWorkflowHTTPToolActionPlanReference(state.actionPlan, state.run?.runId ?? "");
      setActionState((current) => ({ ...current, actionPlan: state.actionPlan }));
    }
    if (state.run) onRunRecorded(state.run.runId);
  }

  return <div className="workflow-definition-http-tool-runtime" aria-label="Workflow Definition HTTP Tool controlled runtime">
    <WorkflowHTTPToolActionPanel
      sectionId="workflow-definition-http-tool-action-review"
      source={{ kind: "workflow_definition", id: version.definitionId, version: version.version,
        nodeId: toolNode?.nodeId ?? "", toolId: toolNode?.toolRef ?? "", label: "Exact active Definition" }}
      consumerState={actionState}
      eligibility={eligibility}
      permissions={permissions}
      resourceKey={resourceKey}
      locale={locale}
      onResourceKeyChange={setResourceKey}
      onLocaleChange={setLocale}
      onCreatePlan={(argumentsValue) => void createPlan(argumentsValue)}
      onReloadPlan={() => void reloadPlan()}
      onDecision={(decision) => void decide(decision)}
    />
    <WorkflowHTTPToolExecutionPanel
      sectionId="workflow-definition-http-tool-execution"
      plan={actionState.actionPlan}
      state={executionState}
      permissions={permissions}
      inputText={inputText}
      model={model}
      onInputTextChange={setInputText}
      onModelChange={setModel}
      onExecute={() => void execute()}
    />
    {executionState.run ? <div className="workflow-executor-action-row"><button type="button" onClick={() => onOpenRun?.(executionState.run!.runId)}>Open v9 evidence in Run History</button></div> : null}
  </div>;
}
