import { useEffect, useRef, useState } from "react";

import { listWorkflowEvaluationCases, type WorkflowEvaluationCase } from "./workflowEvaluationConsumer.ts";
import {
  WorkflowEvaluationDecisionConflict,
  createWorkflowEvaluationSuite,
  decideWorkflowEvaluationSuite,
  listWorkflowEvaluationDecisions,
  listWorkflowEvaluationSuites,
  reviewWorkflowEvaluationSuite,
  type WorkflowEvaluationReleaseDecision,
  type WorkflowEvaluationReleaseDecisionKind,
  type WorkflowEvaluationSuite,
  type WorkflowEvaluationSuiteCaseRef,
  type WorkflowEvaluationSuiteReview,
} from "./workflowEvaluationSuiteConsumer.ts";
import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

export default function WorkflowEvaluationSuitePanel({ applicationId, config, readOnly = false }: { applicationId: string; config: WorkflowExecutorConsumerConfig; readOnly?: boolean }) {
  const [cases, setCases] = useState<WorkflowEvaluationCase[]>([]);
  const [suites, setSuites] = useState<WorkflowEvaluationSuite[]>([]);
  const [name, setName] = useState("");
  const [caseId, setCaseId] = useState("");
  const [version, setVersion] = useState(1);
  const [refs, setRefs] = useState<WorkflowEvaluationSuiteCaseRef[]>([]);
  const [selected, setSelected] = useState<WorkflowEvaluationSuite | null>(null);
  const [review, setReview] = useState<WorkflowEvaluationSuiteReview | null>(null);
  const [decisions, setDecisions] = useState<WorkflowEvaluationReleaseDecision[]>([]);
  const [decision, setDecision] = useState<WorkflowEvaluationReleaseDecisionKind>("needs_review");
  const [failure, setFailure] = useState("");
  const generationRef = useRef(0);

  useEffect(() => {
    const generation = ++generationRef.current;
    setCases([]);
    setSuites([]);
    setName("");
    setCaseId("");
    setVersion(1);
    setRefs([]);
    setSelected(null);
    setReview(null);
    setDecisions([]);
    setDecision("needs_review");
    setFailure("");
    void refresh(generation);
  }, [applicationId, config]);

  async function refresh(expectedGeneration = generationRef.current) {
    try {
      const [casePage, suitePage] = await Promise.all([
        listWorkflowEvaluationCases(applicationId, config),
        listWorkflowEvaluationSuites(applicationId, config),
      ]);
      if (generationRef.current !== expectedGeneration) return;
      setCases(casePage.cases);
      setSuites(suitePage.suites);
    } catch (error) {
      if (generationRef.current === expectedGeneration) setFailure(message(error, "Evaluation suites unavailable."));
    }
  }

  function chooseCase(value: string) {
    setCaseId(value);
    const found = cases.find((item) => item.caseId === value);
    if (found) setVersion(found.version);
  }

  function addRef() {
    if (readOnly || !caseId || version < 1 || refs.some((item) => item.caseId === caseId && item.version === version)) return;
    setRefs([...refs, { caseId, version }]);
  }

  async function create() {
    if (readOnly) return;
    const generation = generationRef.current;
    setFailure("");
    try {
      await createWorkflowEvaluationSuite(applicationId, name, refs, config);
      if (generationRef.current !== generation) return;
      setName("");
      setRefs([]);
      await refresh(generation);
    } catch (error) {
      if (generationRef.current === generation) setFailure(message(error, "Evaluation suite creation failed."));
    }
  }

  async function inspect(value: WorkflowEvaluationSuite) {
    const generation = generationRef.current;
    setFailure("");
    setSelected(value);
    setReview(null);
    setDecisions([]);
    try {
      const [nextReview, history] = await Promise.all([
        reviewWorkflowEvaluationSuite(applicationId, value.suiteId, config),
        listWorkflowEvaluationDecisions(applicationId, value.suiteId, config),
      ]);
      if (generationRef.current !== generation) return;
      setReview(nextReview);
      setDecisions(history.decisions);
    } catch (error) {
      if (generationRef.current === generation) setFailure(message(error, "Release review failed."));
    }
  }

  async function recordDecision() {
    if (readOnly || !selected || !review) return;
    const generation = generationRef.current;
    setFailure("");
    try {
      const result = await decideWorkflowEvaluationSuite(applicationId, selected.suiteId, selected.currentDecisionVersion, decision, review.reviewDigest, config);
      if (generationRef.current !== generation) return;
      setSelected(result.suite);
      await refresh(generation);
      await inspect(result.suite);
    } catch (error) {
      if (generationRef.current !== generation) return;
      if (error instanceof WorkflowEvaluationDecisionConflict) {
        setSelected(error.currentSuite);
        await refresh(generation);
        await inspect(error.currentSuite);
      }
      setFailure(message(error, "Release decision failed."));
    }
  }

  return (
    <section className="workflow-evaluation-panel workflow-review-release" aria-labelledby="workflow-evaluation-suite-title">
      <div className="card-title-row">
        <div><p className="eyebrow">Durable release evidence</p><h4 id="workflow-evaluation-suite-title">Evaluation suites</h4></div>
        <span className="status-badge neutral">human review only</span>
      </div>
      <p>Suites bind 1–50 exact case versions. Decisions retain the reviewed digest and never authorize deployment or release.</p>
      {readOnly ? <p className="workflow-review-read-only">Archived application: suites, reviews, and decision history remain readable; create and decision actions are disabled.</p> : null}
      <details className="workflow-review-disclosure" open={!readOnly}>
        <summary><span>Create evaluation suite</span><small>Exact case versions</small></summary>
        <fieldset className="workflow-evaluation-builder" disabled={readOnly}>
          <label>Suite name<input value={name} onChange={(event) => setName(event.target.value)} placeholder="release candidate review" /></label>
          <label>Evaluation case<select value={caseId} onChange={(event) => chooseCase(event.target.value)}><option value="">Choose exact case</option>{cases.map((item) => <option key={item.caseId} value={item.caseId}>{item.name} · current v{item.version}</option>)}</select></label>
          <label>Version<input type="number" min="1" value={version} onChange={(event) => setVersion(Number(event.target.value))} /></label>
          <button type="button" disabled={!caseId || version < 1} onClick={addRef}>Add exact version</button>
          <button type="button" disabled={!name.trim() || refs.length === 0} onClick={() => void create()}>Create suite</button>
        </fieldset>
        <div className="workflow-evaluation-expectations">{refs.map((item) => <span className="status-badge neutral" key={`${item.caseId}-${item.version}`}>{item.caseId} · v{item.version} <button type="button" disabled={readOnly} onClick={() => setRefs(refs.filter((ref) => ref !== item))}>Remove</button></span>)}</div>
      </details>
      {failure ? <p className="failure-summary">{failure}</p> : null}
      <div className="workflow-review-suite-layout">
        <div className="workflow-run-history-live-list" aria-label="Workflow evaluation suite history">
          {suites.map((item) => <button type="button" className={`workflow-run-history-live-row ${selected?.suiteId === item.suiteId ? "is-selected" : ""}`} key={item.suiteId} onClick={() => void inspect(item)}><span className="workflow-run-history-live-identity"><strong>{item.name}</strong><small>{item.suiteId}</small></span><span><small>Cases</small><strong>{item.caseRefs.length}</strong></span><span><small>Decision</small><strong>{item.currentDecision || "unreviewed"}</strong><small>version {item.currentDecisionVersion}</small></span></button>)}
        </div>
        {selected && review ? (
          <article className="workflow-evaluation-review">
            <div className="card-title-row"><div><p className="eyebrow">Aggregate release review</p><h5>{review.outcome}</h5></div><span className={`status-badge ${review.outcome === "passed" ? "good" : "neutral"}`}>{review.passed} passed</span></div>
            <dl className="tenant-meta"><div><dt>Mismatch</dt><dd>{review.mismatch}</dd></div><div><dt>Inconclusive</dt><dd>{review.inconclusive}</dd></div><div><dt>Unavailable</dt><dd>{review.unavailable}</dd></div><div><dt>Digest</dt><dd>{review.reviewDigest.slice(0, 12)}…</dd></div></dl>
            {review.items.map((item) => <div className="workflow-evaluation-review-item" key={`${item.caseId}-${item.version}`}><strong>{item.name || item.caseId}</strong><span>{item.caseId} · v{item.version}</span><span>{item.outcome} · {item.runProfile}</span></div>)}
            <fieldset className="workflow-review-decision-form" disabled={readOnly}>
              <label>Human decision<select value={decision} onChange={(event) => setDecision(event.target.value as WorkflowEvaluationReleaseDecisionKind)}><option value="needs_review">Needs review</option><option value="approved">Approved</option><option value="rejected">Rejected</option></select></label>
              <button type="button" onClick={() => void recordDecision()}>Record decision v{selected.currentDecisionVersion + 1}</button>
            </fieldset>
            <div className="workflow-review-decision-history" aria-label="Release decision evidence">{decisions.map((item) => <div key={item.decisionId}><span><strong>v{item.version} · {item.decision}</strong><small>{item.createdAt}</small></span><span><small>Reviewed</small><strong>{item.reviewOutcome}</strong></span><span><small>Digest</small><code>{item.reviewDigest.slice(0, 12)}…</code></span></div>)}</div>
            <p className="workflow-review-stop-line">Approval is append-only evidence only. It does not publish, deploy, execute, retry, replay, assign, or authorize production.</p>
          </article>
        ) : <div className="workflow-review-empty-detail"><span aria-hidden="true">◇</span><h5>Select a suite</h5><p>Review exact case versions, digest, aggregate outcome, and append-only human decisions.</p></div>}
      </div>
    </section>
  );
}

function message(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
