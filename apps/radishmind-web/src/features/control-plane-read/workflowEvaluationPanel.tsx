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
  const [cases, setCases] = useState<WorkflowEvaluationCase[]>([]);
  const [name, setName] = useState("");
  const [baseline, setBaseline] = useState("");
  const [candidate, setCandidate] = useState("");
  const [expected, setExpected] = useState<WorkflowEvaluationExpectation["expectedClassification"]>("unchanged");
  const [expectations, setExpectations] = useState<WorkflowEvaluationExpectation[]>([]);
  const [selected, setSelected] = useState<WorkflowEvaluationCase | null>(null);
  const [history, setHistory] = useState<WorkflowEvaluationCase[]>([]);
  const [review, setReview] = useState<WorkflowEvaluationReview | null>(null);
  const [failure, setFailure] = useState("");
  const [revisionName, setRevisionName] = useState("");
  const [revisionBaseline, setRevisionBaseline] = useState("");
  const [revisionExpectations, setRevisionExpectations] = useState<WorkflowEvaluationExpectation[]>([]);
  const [revisionCandidate, setRevisionCandidate] = useState("");
  const generationRef = useRef(0);

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
    setFailure("");
    setRevisionName("");
    setRevisionBaseline("");
    setRevisionExpectations([]);
    setRevisionCandidate("");
    void refresh(generation);
  }, [applicationId, config]);

  async function refresh(expectedGeneration = generationRef.current) {
    try {
      const page = await listWorkflowEvaluationCases(applicationId, config);
      if (generationRef.current === expectedGeneration) setCases(page.cases);
    } catch (error) {
      if (generationRef.current === expectedGeneration) setFailure(message(error, "Evaluation cases unavailable."));
    }
  }

  function addCandidate() {
    if (readOnly || !candidate || candidate === baseline || expectations.some((item) => item.candidateRunId === candidate)) return;
    setExpectations([...expectations, { candidateRunId: candidate, expectedClassification: expected }]);
    setCandidate("");
  }

  async function create() {
    if (readOnly) return;
    const generation = generationRef.current;
    setFailure("");
    try {
      await createWorkflowEvaluationCase(applicationId, name, baseline, expectations, config);
      if (generationRef.current !== generation) return;
      setName("");
      setExpectations([]);
      await refresh(generation);
    } catch (error) {
      if (generationRef.current === generation) setFailure(message(error, "Evaluation case creation failed."));
    }
  }

  async function inspect(value: WorkflowEvaluationCase) {
    const generation = generationRef.current;
    setFailure("");
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
      if (generationRef.current !== generation) return;
      setHistory(revisions.revisions);
      setReview(result);
    } catch (error) {
      if (generationRef.current === generation) setFailure(message(error, "Evaluation review failed."));
    }
  }

  function addRevisionCandidate() {
    if (readOnly || !revisionCandidate || revisionCandidate === revisionBaseline || revisionExpectations.some((item) => item.candidateRunId === revisionCandidate)) return;
    setRevisionExpectations([...revisionExpectations, { candidateRunId: revisionCandidate, expectedClassification: "unchanged" }]);
    setRevisionCandidate("");
  }

  async function revise() {
    if (readOnly || !selected) return;
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
      await inspect(value);
    } catch (error) {
      if (generationRef.current !== generation) return;
      if (error instanceof WorkflowEvaluationVersionConflict) {
        await refresh(generation);
        await inspect(error.currentCase);
      }
      setFailure(message(error, "Evaluation revision failed."));
    }
  }

  async function inspectVersion(value: WorkflowEvaluationCase) {
    const generation = generationRef.current;
    setFailure("");
    try {
      const next = await reviewWorkflowEvaluationCase(applicationId, value.caseId, config, value.version);
      if (generationRef.current === generation) setReview(next);
    } catch (error) {
      if (generationRef.current === generation) setFailure(message(error, "Historical review failed."));
    }
  }

  return (
    <section className="workflow-evaluation-panel workflow-review-cases" aria-label="Workflow evaluation cases">
      <div className="card-title-row">
        <div><p className="eyebrow">Evaluation cases</p><h4>Versioned regression evidence</h4></div>
        <span className="status-badge neutral">durable refs only</span>
      </div>
      <p>Cases bind one exact baseline and 1–20 compatible candidates. They do not store run contents or comparison snapshots.</p>
      {readOnly ? <p className="workflow-review-read-only">Archived application: existing cases and reviews remain readable; create and revise actions are disabled.</p> : null}
      <details className="workflow-review-disclosure" open={!readOnly}>
        <summary><span>Create evaluation case</span><small>Exact run references</small></summary>
        <fieldset className="workflow-evaluation-builder" disabled={readOnly}>
          <label>Case name<input value={name} onChange={(event) => setName(event.target.value)} placeholder="release candidate review" /></label>
          <label>Baseline<RunSelect value={baseline} onChange={(value) => { setBaseline(value); setCandidate(""); setExpectations([]); }} runs={runs} /></label>
          <label>Candidate<RunSelect value={candidate} onChange={setCandidate} runs={candidateRuns} /></label>
          <label>Expected<ClassificationSelect value={expected} onChange={setExpected} /></label>
          <button type="button" onClick={addCandidate} disabled={!candidate || candidate === baseline}>Add candidate</button>
          <button type="button" onClick={() => void create()} disabled={!name.trim() || !baseline || expectations.length === 0}>Create case</button>
        </fieldset>
        <ExpectationBadges values={expectations} readOnly={readOnly} onRemove={(id) => setExpectations(expectations.filter((value) => value.candidateRunId !== id))} />
      </details>
      {failure ? <p className="failure-summary">{failure}</p> : null}
      <div className="workflow-evaluation-case-list">
        {cases.map((value) => (
          <article className={selected?.caseId === value.caseId ? "is-selected" : ""} key={value.caseId}>
            <span><strong>{value.name}</strong><small>{value.caseId} · v{value.version} · {value.expectations.length} candidates</small></span>
            <button type="button" onClick={() => void inspect(value)}>Review case</button>
          </article>
        ))}
      </div>
      {selected ? (
        <details className="workflow-review-disclosure" open>
          <summary><span>{selected.name} · current v{selected.version}</span><small>Revision history and exact review</small></summary>
          <fieldset className="workflow-evaluation-builder" disabled={readOnly}>
            <label>Name<input value={revisionName} onChange={(event) => setRevisionName(event.target.value)} /></label>
            <label>Baseline<RunSelect value={revisionBaseline} onChange={(value) => { setRevisionBaseline(value); setRevisionCandidate(""); setRevisionExpectations([]); }} runs={runs} /></label>
            <label>Add candidate<RunSelect value={revisionCandidate} onChange={setRevisionCandidate} runs={revisionCandidateRuns} /></label>
            <button type="button" onClick={addRevisionCandidate}>Add candidate</button>
            <button type="button" onClick={() => void revise()} disabled={!revisionName.trim() || !revisionBaseline || revisionExpectations.length === 0}>Create revision</button>
          </fieldset>
          <div className="workflow-evaluation-expectations">
            {revisionExpectations.map((item) => (
              <span className="status-badge neutral" key={item.candidateRunId}>
                {item.candidateRunId}
                <ClassificationSelect value={item.expectedClassification} onChange={(value) => setRevisionExpectations(revisionExpectations.map((current) => current.candidateRunId === item.candidateRunId ? { ...current, expectedClassification: value } : current))} />
                <button type="button" disabled={readOnly} onClick={() => setRevisionExpectations(revisionExpectations.filter((value) => value.candidateRunId !== item.candidateRunId))}>Remove</button>
              </span>
            ))}
          </div>
          <div className="workflow-evaluation-case-list">
            {history.map((value) => <article key={value.version}><span><strong>v{value.version} · {value.revisionKind}</strong><small>{value.changeCodes.join(", ")} · {value.actorRef} · {value.auditRef}</small></span><button type="button" onClick={() => void inspectVersion(value)}>Review v{value.version}</button></article>)}
          </div>
        </details>
      ) : null}
      {review ? <Review value={review} /> : null}
    </section>
  );
}

function RunSelect({ value, onChange, runs }: { value: string; onChange: (value: string) => void; runs: WorkflowRunHistorySummary[] }) {
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="">Choose exact run</option>{runs.map((run) => <option value={run.runId} key={run.runId}>{run.runId} · {run.status} · {run.schemaVersion}</option>)}</select>;
}

function ClassificationSelect({ value, onChange }: { value: WorkflowEvaluationExpectation["expectedClassification"]; onChange: (value: WorkflowEvaluationExpectation["expectedClassification"]) => void }) {
  return <select value={value} onChange={(event) => onChange(event.target.value as WorkflowEvaluationExpectation["expectedClassification"])}><option value="unchanged">Unchanged</option><option value="regression">Regression</option><option value="improvement">Improvement</option><option value="changed">Changed</option></select>;
}

function ExpectationBadges({ values, readOnly, onRemove }: { values: WorkflowEvaluationExpectation[]; readOnly: boolean; onRemove: (id: string) => void }) {
  return <div className="workflow-evaluation-expectations">{values.map((item) => <span className="status-badge neutral" key={item.candidateRunId}>{item.candidateRunId} · {item.expectedClassification} <button type="button" disabled={readOnly} onClick={() => onRemove(item.candidateRunId)}>Remove</button></span>)}</div>;
}

function Review({ value }: { value: WorkflowEvaluationReview }) {
  return <article className="workflow-evaluation-review"><div className="card-title-row"><div><p className="eyebrow">Batch result · version {value.version}</p><h5>{value.outcome}</h5></div><span className="status-badge good">{value.matched} matched</span></div><dl className="tenant-meta"><div><dt>Run profile</dt><dd>{value.runProfile}</dd></div><div><dt>Mismatch</dt><dd>{value.mismatched}</dd></div><div><dt>Inconclusive</dt><dd>{value.inconclusive}</dd></div><div><dt>Unavailable</dt><dd>{value.unavailable}</dd></div></dl>{value.items.map((item) => <div className="workflow-evaluation-review-item" key={item.candidateRunId}><strong>{item.candidateRunId}</strong><span>{item.expectedClassification} → {item.actualClassification || item.outcome}</span><span>{item.outcome} · {item.runProfile} · {item.comparisonSchemaVersion || "unavailable"}</span></div>)}<p className="workflow-review-stop-line">References only; no query, fragment content, prompt, answer, evaluation execution, replay, retry, tool, confirmation, or business write is available.</p></article>;
}

function message(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
