import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

export type WorkflowEvaluationSuiteCaseRef = { caseId: string; version: number };
export type WorkflowEvaluationReleaseDecisionKind = "approved" | "rejected" | "needs_review";
export type WorkflowEvaluationSuite = { suiteId: string; name: string; caseRefs: WorkflowEvaluationSuiteCaseRef[]; currentDecisionVersion: number; currentDecision: WorkflowEvaluationReleaseDecisionKind | ""; createdAt: string; actorRef: string; requestId: string; auditRef: string };
export type WorkflowEvaluationReleaseDecision = { decisionId: string; suiteId: string; version: number; decision: WorkflowEvaluationReleaseDecisionKind; reviewDigest: string; reviewOutcome: "passed" | "mismatch" | "inconclusive"; passed: number; mismatch: number; inconclusive: number; unavailable: number; createdAt: string; actorRef: string; requestId: string; auditRef: string };
export type WorkflowEvaluationSuiteReviewItem = { caseId: string; version: number; name: string; outcome: "passed" | "mismatch" | "inconclusive" | "unavailable"; matched: number; mismatched: number; inconclusive: number; unavailable: number; auditRef: string; runProfile: "workflow_standard.v1" | "workflow_rag_retrieval.v1" | "unavailable" };
export type WorkflowEvaluationSuiteReview = { suiteId: string; outcome: "passed" | "mismatch" | "inconclusive"; reviewDigest: string; passed: number; mismatch: number; inconclusive: number; unavailable: number; items: WorkflowEvaluationSuiteReviewItem[] };

type SuiteDocument = { schema_version: "workflow_evaluation_suite.v1"; suite_id: string; name: string; workspace_id: string; application_id: string; case_refs: Array<{case_id:string;version:number}>; current_decision_version: number; current_decision: WorkflowEvaluationSuite["currentDecision"]; created_at: string; actor_ref: string; request_id: string; audit_ref: string };
type DecisionDocument = { schema_version: "workflow_evaluation_release_decision.v1"; decision_id: string; suite_id: string; version: number; decision: WorkflowEvaluationReleaseDecisionKind; review_digest: string; review_outcome: WorkflowEvaluationReleaseDecision["reviewOutcome"]; passed: number; mismatch: number; inconclusive: number; unavailable: number; created_at: string; actor_ref: string; request_id: string; audit_ref: string };
type ReviewDocument = { suite_id: string; outcome: WorkflowEvaluationSuiteReview["outcome"]; review_digest: string; passed: number; mismatch: number; inconclusive: number; unavailable: number; items: Array<{case_id:string;version:number;name:string;outcome:WorkflowEvaluationSuiteReviewItem["outcome"];matched:number;mismatched:number;inconclusive:number;unavailable:number;audit_ref:string;run_profile:WorkflowEvaluationSuiteReviewItem["runProfile"]}> };
type Envelope = { request_id:string; workspace_id:string; application_id:string; suite:SuiteDocument|null; decision:DecisionDocument|null; review:ReviewDocument|null; failure_code:string|null; failure_summary:string; audit_ref:string };
type SuiteListEnvelope = { request_id:string; suites:SuiteDocument[]; next_cursor:string; has_more:boolean; failure_code:string|null; failure_summary:string };
type DecisionListEnvelope = { request_id:string; decisions:DecisionDocument[]; next_cursor:string; has_more:boolean; failure_code:string|null; failure_summary:string };

export class WorkflowEvaluationDecisionConflict extends Error {
  readonly currentSuite: WorkflowEvaluationSuite;
  constructor(currentSuite: WorkflowEvaluationSuite) { super("Release decision changed; refresh the suite before deciding."); this.name = "WorkflowEvaluationDecisionConflict"; this.currentSuite = currentSuite; }
}

export async function createWorkflowEvaluationSuite(applicationId:string,name:string,caseRefs:WorkflowEvaluationSuiteCaseRef[],config:WorkflowExecutorConsumerConfig):Promise<WorkflowEvaluationSuite> {
  assertLive(config);
  const body = await writeJSON(`${config.baseUrl}/v1/user-workspace/workflow-evaluation-suites`, applicationId, {workspace_id:config.workspaceId,application_id:applicationId,name,case_refs:caseRefs.map(item=>({case_id:item.caseId,version:item.version}))}, config);
  if (!isEnvelope(body) || body.failure_code || !body.suite) throw new Error(envelopeFailure(body, 200));
  return mapSuite(body.suite);
}

export async function listWorkflowEvaluationSuites(applicationId:string,config:WorkflowExecutorConsumerConfig,cursor=""):Promise<{suites:WorkflowEvaluationSuite[];nextCursor:string;hasMore:boolean}> {
  if (!isLive(config)) return {suites:[],nextCursor:"",hasMore:false};
  const query = scopedQuery(applicationId, config); query.set("limit", "25"); if (cursor) query.set("cursor", cursor);
  const body = await readJSON(`${config.baseUrl}/v1/user-workspace/workflow-evaluation-suites?${query}`, applicationId, config);
  if (!isSuiteListEnvelope(body) || body.failure_code) throw new Error(envelopeFailure(body, 200));
  return {suites:body.suites.map(mapSuite),nextCursor:body.next_cursor,hasMore:body.has_more};
}

export async function reviewWorkflowEvaluationSuite(applicationId:string,suiteId:string,config:WorkflowExecutorConsumerConfig):Promise<WorkflowEvaluationSuiteReview> {
  assertLive(config);
  const body = await readJSON(`${config.baseUrl}/v1/user-workspace/workflow-evaluation-suites/${encodeURIComponent(suiteId)}/review?${scopedQuery(applicationId,config)}`, applicationId, config);
  if (!isEnvelope(body) || body.failure_code || !body.review) throw new Error(envelopeFailure(body, 200));
  return mapReview(body.review);
}

export async function decideWorkflowEvaluationSuite(applicationId:string,suiteId:string,expectedDecisionVersion:number,decision:WorkflowEvaluationReleaseDecisionKind,reviewDigest:string,config:WorkflowExecutorConsumerConfig):Promise<{suite:WorkflowEvaluationSuite;decision:WorkflowEvaluationReleaseDecision}> {
  assertLive(config);
  const body = await writeJSON(`${config.baseUrl}/v1/user-workspace/workflow-evaluation-suites/${encodeURIComponent(suiteId)}/decisions`, applicationId, {workspace_id:config.workspaceId,application_id:applicationId,expected_decision_version:expectedDecisionVersion,decision,review_digest:reviewDigest}, config);
  if (!isEnvelope(body)) throw new Error(envelopeFailure(body, 200));
  if (body.failure_code === "workflow_evaluation_suite_decision_conflict" && body.suite) throw new WorkflowEvaluationDecisionConflict(mapSuite(body.suite));
  if (body.failure_code || !body.suite || !body.decision) throw new Error(envelopeFailure(body, 200));
  return {suite:mapSuite(body.suite),decision:mapDecision(body.decision)};
}

export async function listWorkflowEvaluationDecisions(applicationId:string,suiteId:string,config:WorkflowExecutorConsumerConfig,cursor=""):Promise<{decisions:WorkflowEvaluationReleaseDecision[];nextCursor:string;hasMore:boolean}> {
  if (!isLive(config)) return {decisions:[],nextCursor:"",hasMore:false};
  const query = scopedQuery(applicationId, config); query.set("limit", "25"); if (cursor) query.set("cursor", cursor);
  const body = await readJSON(`${config.baseUrl}/v1/user-workspace/workflow-evaluation-suites/${encodeURIComponent(suiteId)}/decisions?${query}`, applicationId, config);
  if (!isDecisionListEnvelope(body) || body.failure_code) throw new Error(envelopeFailure(body, 200));
  return {decisions:body.decisions.map(mapDecision),nextCursor:body.next_cursor,hasMore:body.has_more};
}

async function writeJSON(url:string,applicationId:string,payload:unknown,config:WorkflowExecutorConsumerConfig):Promise<unknown> { const response=await fetch(url,{method:"POST",headers:{...headers(config,applicationId,true),"Content-Type":"application/json"},body:JSON.stringify(payload)});const body:unknown=await response.json();if(!response.ok)throw new Error(envelopeFailure(body,response.status));return body; }
async function readJSON(url:string,applicationId:string,config:WorkflowExecutorConsumerConfig):Promise<unknown> { const response=await fetch(url,{headers:headers(config,applicationId,false)});const body:unknown=await response.json();if(!response.ok)throw new Error(envelopeFailure(body,response.status));return body; }
function scopedQuery(applicationId:string,config:WorkflowExecutorConsumerConfig){return new URLSearchParams({workspace_id:config.workspaceId,application_id:applicationId})}
function isLive(config:WorkflowExecutorConsumerConfig){return config.mode==="dev_workflow_executor_http"}
function assertLive(config:WorkflowExecutorConsumerConfig){if(!isLive(config))throw new Error("Workflow evaluation suites are unavailable in offline mode.")}
function headers(config:WorkflowExecutorConsumerConfig,applicationId:string,write:boolean):HeadersInit{return{Accept:"application/json","X-Request-Id":`dev-workflow-evaluation-suite-${applicationId}`,"X-RadishMind-Dev-Read-Identity":"dev-workflow-evaluation-suite-consumer","X-RadishMind-Dev-Read-Tenant":config.tenantRef,"X-RadishMind-Dev-Read-Subject":config.subjectRef,"X-RadishMind-Dev-Read-Scopes":write?"workflow_evaluations:write,workflow_evaluations:read,workflow_runs:read":"workflow_evaluations:read,workflow_runs:read","X-RadishMind-Dev-Read-Audit":"audit_dev_workflow_evaluation_suite_consumer","X-RadishMind-Dev-Workflow-Workspace":config.workspaceId,"X-RadishMind-Dev-Workflow-Application":applicationId}}
function mapSuite(value:SuiteDocument):WorkflowEvaluationSuite{return{suiteId:value.suite_id,name:value.name,caseRefs:value.case_refs.map(item=>({caseId:item.case_id,version:item.version})),currentDecisionVersion:value.current_decision_version,currentDecision:value.current_decision,createdAt:value.created_at,actorRef:value.actor_ref,requestId:value.request_id,auditRef:value.audit_ref}}
function mapDecision(value:DecisionDocument):WorkflowEvaluationReleaseDecision{return{decisionId:value.decision_id,suiteId:value.suite_id,version:value.version,decision:value.decision,reviewDigest:value.review_digest,reviewOutcome:value.review_outcome,passed:value.passed,mismatch:value.mismatch,inconclusive:value.inconclusive,unavailable:value.unavailable,createdAt:value.created_at,actorRef:value.actor_ref,requestId:value.request_id,auditRef:value.audit_ref}}
function mapReview(value:ReviewDocument):WorkflowEvaluationSuiteReview{return{suiteId:value.suite_id,outcome:value.outcome,reviewDigest:value.review_digest,passed:value.passed,mismatch:value.mismatch,inconclusive:value.inconclusive,unavailable:value.unavailable,items:value.items.map(item=>({caseId:item.case_id,version:item.version,name:item.name,outcome:item.outcome,matched:item.matched,mismatched:item.mismatched,inconclusive:item.inconclusive,unavailable:item.unavailable,auditRef:item.audit_ref,runProfile:item.run_profile}))}}
function hasForbiddenKey(value:unknown):boolean{if(Array.isArray(value))return value.some(hasForbiddenKey);if(!value||typeof value!=="object")return false;const forbidden=new Set(["input_text","input_bytes","condition_values","condition_node_ids","query","raw_query","fragment_content","content","prompt","prompt_packet","answer","limitations","confidence","output","output_preview","credential","authorization","cookie","endpoint","raw_response","provider_raw_envelope","comparison","change_reason"]);return Object.entries(value as Record<string,unknown>).some(([key,child])=>forbidden.has(key.toLowerCase())||hasForbiddenKey(child))}
function isEnvelope(value:unknown):value is Envelope{if(!value||typeof value!=="object"||hasForbiddenKey(value))return false;const item=value as Partial<Envelope>;return typeof item.request_id==="string"&&(item.failure_code===null||typeof item.failure_code==="string")&&(item.suite===null||isSuite(item.suite))&&(item.decision===null||isDecision(item.decision))&&(item.review===null||isReview(item.review))}
function isSuiteListEnvelope(value:unknown):value is SuiteListEnvelope{if(!value||typeof value!=="object"||hasForbiddenKey(value))return false;const item=value as Partial<SuiteListEnvelope>;return typeof item.request_id==="string"&&Array.isArray(item.suites)&&item.suites.every(isSuite)&&typeof item.next_cursor==="string"&&typeof item.has_more==="boolean"&&(item.failure_code===null||typeof item.failure_code==="string")}
function isDecisionListEnvelope(value:unknown):value is DecisionListEnvelope{if(!value||typeof value!=="object"||hasForbiddenKey(value))return false;const item=value as Partial<DecisionListEnvelope>;return typeof item.request_id==="string"&&Array.isArray(item.decisions)&&item.decisions.every(isDecision)&&typeof item.next_cursor==="string"&&typeof item.has_more==="boolean"&&(item.failure_code===null||typeof item.failure_code==="string")}
function integer(value:unknown){return Number.isInteger(value)&&Number(value)>=0}
function isSuite(value:unknown):value is SuiteDocument{if(!value||typeof value!=="object")return false;const item=value as Partial<SuiteDocument>;return item.schema_version==="workflow_evaluation_suite.v1"&&typeof item.suite_id==="string"&&typeof item.name==="string"&&Array.isArray(item.case_refs)&&item.case_refs.every(ref=>typeof ref.case_id==="string"&&Number.isInteger(ref.version)&&ref.version>=1)&&integer(item.current_decision_version)&&["","approved","rejected","needs_review"].includes(item.current_decision??"")&&typeof item.created_at==="string"&&typeof item.actor_ref==="string"&&typeof item.request_id==="string"&&typeof item.audit_ref==="string"}
function isDecision(value:unknown):value is DecisionDocument{if(!value||typeof value!=="object")return false;const item=value as Partial<DecisionDocument>;return item.schema_version==="workflow_evaluation_release_decision.v1"&&typeof item.decision_id==="string"&&typeof item.suite_id==="string"&&Number.isInteger(item.version)&&Number(item.version)>=1&&["approved","rejected","needs_review"].includes(item.decision??"")&&typeof item.review_digest==="string"&&/^[0-9a-f]{64}$/.test(item.review_digest)&&["passed","mismatch","inconclusive"].includes(item.review_outcome??"")&&integer(item.passed)&&integer(item.mismatch)&&integer(item.inconclusive)&&integer(item.unavailable)&&typeof item.created_at==="string"}
function isReview(value:unknown):value is ReviewDocument{if(!value||typeof value!=="object")return false;const item=value as Partial<ReviewDocument>;return typeof item.suite_id==="string"&&["passed","mismatch","inconclusive"].includes(item.outcome??"")&&typeof item.review_digest==="string"&&/^[0-9a-f]{64}$/.test(item.review_digest)&&integer(item.passed)&&integer(item.mismatch)&&integer(item.inconclusive)&&integer(item.unavailable)&&Array.isArray(item.items)&&item.items.every(isReviewItem)}
function isReviewItem(value:unknown):value is ReviewDocument["items"][number]{if(!value||typeof value!=="object")return false;const item=value as Record<string,unknown>;return typeof item.case_id==="string"&&Number.isInteger(item.version)&&Number(item.version)>=1&&typeof item.name==="string"&&["passed","mismatch","inconclusive","unavailable"].includes(String(item.outcome))&&integer(item.matched)&&integer(item.mismatched)&&integer(item.inconclusive)&&integer(item.unavailable)&&typeof item.audit_ref==="string"&&["workflow_standard.v1","workflow_rag_retrieval.v1","unavailable"].includes(String(item.run_profile))}
function envelopeFailure(value:unknown,status:number):string{if(value&&typeof value==="object"){const item=value as Partial<Envelope>;if(typeof item.failure_code==="string")return`${item.failure_code}: ${item.failure_summary??""}`};return`workflow evaluation suite route failed with HTTP ${status}`}
