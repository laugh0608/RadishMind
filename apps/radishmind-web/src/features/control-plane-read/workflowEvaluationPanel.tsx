import "../../i18n/evaluationCaseResources.ts";
import { useTranslation } from "react-i18next";
import { evaluationState, evaluationFailure, evaluationFailureCode } from "./evaluationMessages.ts";

import { useEffect, useRef, useState } from "react";

import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";
import { isWorkflowRunComparisonCompatible, type WorkflowRunHistorySummary } from "./workflowRunHistoryConsumer.ts";
import {
  WorkflowEvaluationVersionConflict,
  createWorkflowEvaluationCase,
  listWorkflowEvaluationCases,
  listWorkflowEvaluationRevisions,
  reviseWorkflowEvaluationCase,
  reviewWorkflowEvaluationCase,
  type WorkflowEvaluationCase,
  type WorkflowEvaluationExpectation,
  type WorkflowEvaluationReview,
} from "./workflowEvaluationConsumer.ts";

export default function WorkflowEvaluationPanel({
  applicationId,
  runs,
  config,
  readOnly = false,
}: {
  applicationId: string;
  runs: WorkflowRunHistorySummary[];
  config: WorkflowExecutorConsumerConfig;
  readOnly?: boolean;
}) {
  const { t } = useTranslation("evaluation");
  const [cases, setCases] = useState<WorkflowEvaluationCase[]>([]);
  const [name, setName] = useState("");
  const [baseline, setBaseline] = useState("");
  const [candidate, setCandidate] = useState("");
  const [expected, setExpected] = useState<WorkflowEvaluationExpectation["expectedClassification"]>("unchanged");
  const [expectations, setExpectations] = useState<WorkflowEvaluationExpectation[]>([]);
  const [selected, setSelected] = useState<WorkflowEvaluationCase | null>(null);
  const [history, setHistory] = useState<WorkflowEvaluationCase[]>([]);
  const [reviewTarget, setReviewTarget] = useState<WorkflowEvaluationCase | null>(null);
  const [review, setReview] = useState<WorkflowEvaluationReview | null>(null);
  const [failure, setFailure] = useState("");
  const [revisionName, setRevisionName] = useState("");
  const [revisionBaseline, setRevisionBaseline] = useState("");
  const [revisionExpectations, setRevisionExpectations] = useState<WorkflowEvaluationExpectation[]>([]);
  const [revisionCandidate, setRevisionCandidate] = useState("");
  const generationRef = useRef(0);
  const selectionRef = useRef(0);
  const listRef = useRef(0);
  const writeRef = useRef(false);
  const [pending, setPending] = useState(false);
  const [loading, setLoading] = useState(false);
  const [reviewState, setReviewState] = useState<"idle" | "loading" | "ready" | "failed">("idle");
  const offline = config.mode !== "dev_workflow_executor_http";
  const disabled = readOnly || offline || pending;
  const [conflict, setConflict] = useState<WorkflowEvaluationCase | null>(null);

  const baselineRun = runs.find((run) => run.runId === baseline);
  const candidateRuns = runs.filter((run) => isWorkflowRunComparisonCompatible(baselineRun, run));
  const revisionBaselineRun = runs.find((run) => run.runId === revisionBaseline);
  const revisionCandidateRuns = runs.filter((run) => isWorkflowRunComparisonCompatible(revisionBaselineRun, run));

  useEffect(() => {
    const generation = ++generationRef.current;
    setCases([]);
    setName("");
    setBaseline("");
    setCandidate("");
    setExpectations([]);
    setSelected(null);
    setHistory([]);
    setReview(null);
    setReviewTarget(null);
    setFailure("");
    setRevisionName("");
    setRevisionBaseline("");
    setRevisionExpectations([]);
    setRevisionCandidate("");
    selectionRef.current++;
    writeRef.current = false;
    setPending(false);
    setReviewState("idle");
    setConflict(null);
    setExpected("unchanged");
    void refresh(generation);
    return () => { generationRef.current++; selectionRef.current++; };
  }, [applicationId, config]);

  async function refresh(expectedGeneration = generationRef.current) {
    if (generationRef.current !== expectedGeneration) return;
    const request = ++listRef.current;
    setLoading(true);
    setFailure("");
    try {
      const page = await listWorkflowEvaluationCases(applicationId, config);
      if (generationRef.current === expectedGeneration && listRef.current === request) setCases(page.cases);
    } catch (error) {
      if (generationRef.current === expectedGeneration && listRef.current === request) setFailure(evaluationFailureCode(error));
    } finally {
      if (generationRef.current === expectedGeneration && listRef.current === request) setLoading(false);
    }
  }

  function addCandidate() {
    if (disabled || !candidate || candidate === baseline || expectations.some((item) => item.candidateRunId === candidate)) return;
    setExpectations([...expectations, { candidateRunId: candidate, expectedClassification: expected }]);
    setCandidate("");
  }

  async function create() {
    if (disabled || writeRef.current) return;
    const generation = generationRef.current;
    writeRef.current = true;
    setPending(true);
    setFailure("");
    try {
      await createWorkflowEvaluationCase(applicationId, name, baseline, expectations, config);
      if (generationRef.current !== generation) return;
      setName("");
      setExpectations([]);
      await refresh(generation);
    } catch (error) {
      if (generationRef.current === generation) setFailure(evaluationFailureCode(error));
    } finally {
      if (generationRef.current === generation) { writeRef.current = false; setPending(false); }
    }
  }

  async function inspect(value: WorkflowEvaluationCase) {
    const generation = generationRef.current;
    const selection = ++selectionRef.current;
    setReviewTarget(value);
    setReviewState("loading");
    setFailure("");
    setConflict(null);
    setRevisionCandidate("");
    setSelected(value);
    setRevisionName(value.name);
    setRevisionBaseline(value.baselineRunId);
    setRevisionExpectations(value.expectations);
    setHistory([]);
    setReview(null);
    try {
      const [revisions, result] = await Promise.all([
        listWorkflowEvaluationRevisions(applicationId, value.caseId, config),
        reviewWorkflowEvaluationCase(applicationId, value.caseId, config, value.version),
      ]);
      if (generationRef.current !== generation || selectionRef.current !== selection) return;
      setHistory(revisions.revisions);
      setReview(result);
      setReviewState("ready");
    } catch (error) {
      if (generationRef.current === generation && selectionRef.current === selection) { setFailure(evaluationFailureCode(error)); setReviewState("failed"); }
    }
  }

  function addRevisionCandidate() {
    if (disabled || !revisionCandidate || revisionCandidate === revisionBaseline || revisionExpectations.some((item) => item.candidateRunId === revisionCandidate)) return;
    setRevisionExpectations([...revisionExpectations, { candidateRunId: revisionCandidate, expectedClassification: "unchanged" }]);
    setRevisionCandidate("");
  }

  async function revise() {
    if (disabled || writeRef.current || !selected || conflict) return;
    writeRef.current = true;
    setPending(true);
    const generation = generationRef.current;
    setFailure("");
    try {
      const value = await reviseWorkflowEvaluationCase(applicationId, selected.caseId, {
        expectedVersion: selected.version,
        revisionKind: revisionBaseline === selected.baselineRunId ? "case_revision" : "baseline_promotion",
        name: revisionName,
        baselineRunId: revisionBaseline,
        expectations: revisionExpectations,
      }, config);
      if (generationRef.current !== generation) return;
      await refresh(generation);
      if (generationRef.current === generation) await inspect(value);
    } catch (error) {
      if (generationRef.current !== generation) return;
      if (error instanceof WorkflowEvaluationVersionConflict) {
        setConflict(error.currentCase);
        await refresh(generation);
      } else {
        setFailure(evaluationFailureCode(error));
      }
    } finally {
      if (generationRef.current === generation) { writeRef.current = false; setPending(false); }
    }
  }

  async function inspectVersion(value: WorkflowEvaluationCase) {
    const generation = generationRef.current;
    const selection = ++selectionRef.current;
    setReviewTarget(value);
    setFailure("");
    setReview(null);
    setReviewState("loading");
    try {
      const next = await reviewWorkflowEvaluationCase(applicationId, value.caseId, config, value.version);
      if (generationRef.current !== generation || selectionRef.current !== selection) return;
      setReview(next);
      setReviewState("ready");
    } catch (error) {
      if (generationRef.current === generation && selectionRef.current === selection) {
        setFailure(evaluationFailureCode(error));
        setReviewState("failed");
      }
    }
  }

  return (
    <section className="workflow-evaluation-panel workflow-review-cases" aria-label={t($ => $.cases.caseLabel)}>
      <div className="card-title-row">
        <div><p className="eyebrow">{t($ => $.cases.cases)}</p><h4>{t($ => $.cases.caseTitle)}</h4></div>
        <span className="status-badge neutral">{t($ => $.cases.refsOnly)}</span>
      </div>
      <p>{t($ => $.cases.caseDescription)}</p>
      {readOnly ? <p className="workflow-review-read-only">{t($ => $.cases.caseReadOnly)}</p> : null}
      <details className="workflow-review-disclosure" open={!readOnly}>
        <summary><span>{t($ => $.cases.createCaseTitle)}</span><small>{t($ => $.cases.exactRuns)}</small></summary>
        <fieldset className="workflow-evaluation-builder" disabled={disabled}>
          <label>{t($ => $.cases.caseName)}<input value={name} onChange={(event) => setName(event.target.value)} placeholder={t($ => $.cases.namePlaceholder)} /></label>
          <label>{t($ => $.cases.baseline)}<RunSelect value={baseline} onChange={(value) => { setBaseline(value); setCandidate(""); setExpectations([]); }} runs={runs} /></label>
          <label>{t($ => $.cases.candidate)}<RunSelect value={candidate} onChange={setCandidate} runs={candidateRuns} /></label>
          <label>{t($ => $.cases.expected)}<ClassificationSelect value={expected} onChange={setExpected} /></label>
          <button type="button" onClick={addCandidate} disabled={!candidate || candidate === baseline}>{t($ => $.cases.addCandidate)}</button>
          <button type="button" onClick={() => void create()} disabled={!name.trim() || !baseline || expectations.length === 0}>{t($ => $.cases.createCase)}</button>
        </fieldset>
        <ExpectationBadges values={expectations} readOnly={disabled} onRemove={(id) => setExpectations(expectations.filter((value) => value.candidateRunId !== id))} />
      </details>
      {offline ? <p>{t($ => $.cases.offline)}</p> : null}
      <button type="button" disabled={offline || pending || loading} onClick={() => void refresh()}>{t($ => $.cases.refresh)}</button>
      {loading || pending ? <p role="status">{pending ? t($ => $.cases.saving) : t($ => $.cases.loading)}</p> : null}
      {!loading && !offline && !failure && cases.length === 0 ? <p>{t($ => $.cases.emptyCases)}</p> : null}
      {reviewState === "loading" ? <p role="status">{t($ => $.cases.loading)}</p> : null}
      {reviewState === "failed" && selected ? <p>{t($ => $.cases.reviewFailed)} <button type="button" disabled={pending} onClick={() => void inspectVersion(reviewTarget ?? selected)}>{t($ => $.cases.retryReview)}</button></p> : null}
      {failure ? <p role="alert" className="failure-summary">{evaluationFailure(t, failure)}{failure.startsWith("workflow_evaluation_") ? <> <code>{failure}</code></> : null}</p> : null}
      <div className="workflow-evaluation-case-list">
        {cases.map((value) => (
          <article className={selected?.caseId === value.caseId ? "is-selected" : ""} key={value.caseId}>
            <span><strong>{value.name}</strong><small>{value.caseId} · v{value.version} · {t($ => $.cases.candidateCount, { count: value.expectations.length })}</small></span>
            <button type="button" disabled={pending} onClick={() => void inspect(value)}>{t($ => $.cases.reviewCase)}</button>
          </article>
        ))}
      </div>
      {conflict ? <div role="alert"><p>{t($ => $.cases.conflict)}</p><button type="button" disabled={pending} onClick={() => void inspect(conflict)}>{t($ => $.cases.loadCurrent)}</button></div> : null}
      {selected ? (
        <details className="workflow-review-disclosure" open>
          <summary><span>{selected.name} · {t($ => $.cases.currentVersion, { version: selected.version })}</span><small>{t($ => $.cases.revisionHistory)}</small></summary>
          <fieldset className="workflow-evaluation-builder" disabled={disabled}>
            <label>{t($ => $.cases.name)}<input value={revisionName} onChange={(event) => setRevisionName(event.target.value)} /></label>
            <label>{t($ => $.cases.baseline)}<RunSelect value={revisionBaseline} onChange={(value) => { setRevisionBaseline(value); setRevisionCandidate(""); setRevisionExpectations([]); }} runs={runs} /></label>
            <label>{t($ => $.cases.addCandidate)}<RunSelect value={revisionCandidate} onChange={setRevisionCandidate} runs={revisionCandidateRuns} /></label>
            <button type="button" onClick={addRevisionCandidate}>{t($ => $.cases.addCandidate)}</button>
            <button type="button" onClick={() => void revise()} disabled={Boolean(conflict) || !revisionName.trim() || !revisionBaseline || revisionExpectations.length === 0}>{t($ => $.cases.createRevision)}</button>
          </fieldset>
          <div className="workflow-evaluation-expectations">
            {revisionExpectations.map((item) => (
              <span className="status-badge neutral" key={item.candidateRunId}>
                {item.candidateRunId}
                <ClassificationSelect value={item.expectedClassification} disabled={disabled} label={`${t($ => $.cases.expected)} · ${item.candidateRunId}`} onChange={(value) => setRevisionExpectations(revisionExpectations.map((current) => current.candidateRunId === item.candidateRunId ? { ...current, expectedClassification: value } : current))} />
                <button type="button" disabled={disabled} onClick={() => setRevisionExpectations(revisionExpectations.filter((value) => value.candidateRunId !== item.candidateRunId))}>{t($ => $.cases.remove)}</button>
              </span>
            ))}
          </div>
          <div className="workflow-evaluation-case-list">
            {history.map((value) => <article key={value.version}><span><strong>v{value.version} · {evaluationState(t, value.revisionKind)} <code>{value.revisionKind}</code></strong><small>{value.changeCodes.map(code => `${evaluationState(t, code)} (${code})`).join(", ")} · {value.actorRef} · {value.auditRef}</small></span><button type="button" disabled={pending} onClick={() => void inspectVersion(value)}>{t($ => $.cases.reviewVersion, { version: value.version })}</button></article>)}
          </div>
        </details>
      ) : null}
      {review ? <Review value={review} /> : null}
    </section>
  );
}

function RunSelect({ value, onChange, runs }: { value: string; onChange: (value: string) => void; runs: WorkflowRunHistorySummary[] }) {
  const { t } = useTranslation("evaluation");
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="">{t($ => $.cases.chooseRun)}</option>{runs.map((run) => <option value={run.runId} key={run.runId}>{run.runId} · {evaluationState(t, run.status)} · {run.schemaVersion}</option>)}</select>;
}

function ClassificationSelect({ value, onChange, disabled = false, label }: { disabled?: boolean; label?: string; value: WorkflowEvaluationExpectation["expectedClassification"]; onChange: (value: WorkflowEvaluationExpectation["expectedClassification"]) => void }) {
  const { t } = useTranslation("evaluation");
  return <select disabled={disabled} aria-label={label} value={value} onChange={(event) => onChange(event.target.value as WorkflowEvaluationExpectation["expectedClassification"])}><option value="unchanged">{t($ => $.cases.states.unchanged)}</option><option value="regression">{t($ => $.cases.states.regression)}</option><option value="improvement">{t($ => $.cases.states.improvement)}</option><option value="changed">{t($ => $.cases.states.changed)}</option></select>;
}

function ExpectationBadges({ values, readOnly, onRemove }: { values: WorkflowEvaluationExpectation[]; readOnly: boolean; onRemove: (id: string) => void }) {
  const { t } = useTranslation("evaluation");
  return <div className="workflow-evaluation-expectations">{values.map((item) => <span className="status-badge neutral" key={item.candidateRunId}>{item.candidateRunId} · {evaluationState(t, item.expectedClassification)} <button type="button" disabled={readOnly} onClick={() => onRemove(item.candidateRunId)}>{t($ => $.cases.remove)}</button></span>)}</div>;
}

function Review({ value }: { value: WorkflowEvaluationReview }) {
  const { t } = useTranslation("evaluation");
  return <article className="workflow-evaluation-review"><div className="card-title-row"><div><p className="eyebrow">{t($ => $.cases.batchResult, { version: value.version })}</p><h5>{evaluationState(t, value.outcome)}</h5></div><span className={`status-badge ${value.outcome === "passed" ? "good" : "neutral"}`}>{t($ => $.cases.matchedCount, { count: value.matched })}</span></div><dl className="tenant-meta"><div><dt>{t($ => $.cases.runProfile)}</dt><dd>{evaluationState(t, value.runProfile)}</dd></div><div><dt>{t($ => $.cases.mismatch)}</dt><dd>{value.mismatched}</dd></div><div><dt>{t($ => $.cases.inconclusive)}</dt><dd>{value.inconclusive}</dd></div><div><dt>{t($ => $.cases.unavailable)}</dt><dd>{value.unavailable}</dd></div></dl>{value.items.map((item) => <div className="workflow-evaluation-review-item" key={item.candidateRunId}><strong>{item.candidateRunId}</strong><span>{evaluationState(t, item.expectedClassification)} → {evaluationState(t, item.actualClassification || item.outcome)}</span><span>{evaluationState(t, item.outcome)} · {evaluationState(t, item.runProfile)} · {item.comparisonSchemaVersion || t($ => $.cases.unavailable)}</span></div>)}<p className="workflow-review-stop-line">{t($ => $.cases.caseStop)}</p></article>;
}
