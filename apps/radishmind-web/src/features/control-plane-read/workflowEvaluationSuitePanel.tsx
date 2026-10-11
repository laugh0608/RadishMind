import "../../i18n/evaluationCaseResources.ts";
import { useTranslation } from "react-i18next";
import { evaluationState, evaluationFailure, evaluationFailureCode } from "./evaluationMessages.ts";

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
  const { t } = useTranslation("evaluation");
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
  const selectionRef = useRef(0);
  const listRef = useRef(0);
  const writeRef = useRef(false);
  const [pending, setPending] = useState(false);
  const [loading, setLoading] = useState(false);
  const [reviewState, setReviewState] = useState<"idle" | "loading" | "ready" | "failed">("idle");
  const offline = config.mode !== "dev_workflow_executor_http";
  const disabled = readOnly || offline || pending;
  const [decisionConflict, setDecisionConflict] = useState(false);

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
    selectionRef.current++;
    writeRef.current = false;
    setPending(false);
    setReviewState("idle");
    setDecisionConflict(false);
    void refresh(generation);
    return () => { generationRef.current++; selectionRef.current++; };
  }, [applicationId, config]);

  async function refresh(expectedGeneration = generationRef.current) {
    if (generationRef.current !== expectedGeneration) return;
    const request = ++listRef.current;
    setLoading(true);
    setFailure("");
    try {
      const [casePage, suitePage] = await Promise.all([
        listWorkflowEvaluationCases(applicationId, config),
        listWorkflowEvaluationSuites(applicationId, config),
      ]);
      if (generationRef.current !== expectedGeneration || listRef.current !== request) return;
      setCases(casePage.cases);
      setSuites(suitePage.suites);
    } catch (error) {
      if (generationRef.current === expectedGeneration && listRef.current === request) setFailure(evaluationFailureCode(error));
    } finally {
      if (generationRef.current === expectedGeneration && listRef.current === request) setLoading(false);
    }
  }

  function chooseCase(value: string) {
    setCaseId(value);
    const found = cases.find((item) => item.caseId === value);
    if (found) setVersion(found.version);
  }

  function addRef() {
    if (disabled || !caseId || version < 1 || refs.some((item) => item.caseId === caseId && item.version === version)) return;
    setRefs([...refs, { caseId, version }]);
  }

  async function create() {
    if (disabled || writeRef.current) return;
    const generation = generationRef.current;
    writeRef.current = true;
    setPending(true);
    setFailure("");
    try {
      await createWorkflowEvaluationSuite(applicationId, name, refs, config);
      if (generationRef.current !== generation) return;
      setName("");
      setRefs([]);
      await refresh(generation);
    } catch (error) {
      if (generationRef.current === generation) setFailure(evaluationFailureCode(error));
    } finally {
      if (generationRef.current === generation) { writeRef.current = false; setPending(false); }
    }
  }

  async function inspect(value: WorkflowEvaluationSuite) {
    const generation = generationRef.current;
    const selection = ++selectionRef.current;
    setReviewState("loading");
    setFailure("");
    setDecisionConflict(false);
    setSelected(value);
    setReview(null);
    setDecisions([]);
    try {
      const [nextReview, history] = await Promise.all([
        reviewWorkflowEvaluationSuite(applicationId, value.suiteId, config),
        listWorkflowEvaluationDecisions(applicationId, value.suiteId, config),
      ]);
      if (generationRef.current !== generation || selectionRef.current !== selection) return;
      setReview(nextReview);
      setReviewState("ready");
      setDecisions(history.decisions);
    } catch (error) {
      if (generationRef.current === generation && selectionRef.current === selection) { setFailure(evaluationFailureCode(error)); setReviewState("failed"); }
    }
  }

  async function recordDecision() {
    if (disabled || writeRef.current || !selected || !review) return;
    writeRef.current = true;
    setPending(true);
    const generation = generationRef.current;
    setFailure("");
    try {
      const result = await decideWorkflowEvaluationSuite(applicationId, selected.suiteId, selected.currentDecisionVersion, decision, review.reviewDigest, config);
      if (generationRef.current !== generation) return;
      setSelected(result.suite);
      await refresh(generation);
      if (generationRef.current === generation) await inspect(result.suite);
    } catch (error) {
      if (generationRef.current !== generation) return;
      if (error instanceof WorkflowEvaluationDecisionConflict) {
        setSelected(error.currentSuite);
        await refresh(generation);
        if (generationRef.current !== generation) return;
        await inspect(error.currentSuite);
        if (generationRef.current === generation) setDecisionConflict(true);
      } else {
        setFailure(evaluationFailureCode(error));
      }
    } finally {
      if (generationRef.current === generation) { writeRef.current = false; setPending(false); }
    }
  }

  return (
    <section className="workflow-evaluation-panel workflow-review-release" aria-labelledby="workflow-evaluation-suite-title">
      <div className="card-title-row">
        <div><p className="eyebrow">{t($ => $.cases.suiteEyebrow)}</p><h4 id="workflow-evaluation-suite-title">{t($ => $.cases.suites)}</h4></div>
        <span className="status-badge neutral">{t($ => $.cases.humanOnly)}</span>
      </div>
      <p>{t($ => $.cases.suiteDescription)}</p>
      {readOnly ? <p className="workflow-review-read-only">{t($ => $.cases.suiteReadOnly)}</p> : null}
      <details className="workflow-review-disclosure" open={!readOnly}>
        <summary><span>{t($ => $.cases.createSuiteTitle)}</span><small>{t($ => $.cases.exactCases)}</small></summary>
        <fieldset className="workflow-evaluation-builder" disabled={disabled}>
          <label>{t($ => $.cases.suiteName)}<input value={name} onChange={(event) => setName(event.target.value)} placeholder={t($ => $.cases.namePlaceholder)} /></label>
          <label>{t($ => $.cases.case)}<select value={caseId} onChange={(event) => chooseCase(event.target.value)}><option value="">{t($ => $.cases.chooseCase)}</option>{cases.map((item) => <option key={item.caseId} value={item.caseId}>{item.name} · {t($ => $.cases.currentVersion, { version: item.version })}</option>)}</select></label>
          <label>{t($ => $.cases.version)}<input type="number" min="1" value={version} onChange={(event) => setVersion(Number(event.target.value))} /></label>
          <button type="button" disabled={!caseId || version < 1} onClick={addRef}>{t($ => $.cases.addVersion)}</button>
          <button type="button" disabled={!name.trim() || refs.length === 0} onClick={() => void create()}>{t($ => $.cases.createSuite)}</button>
        </fieldset>
        <div className="workflow-evaluation-expectations">{refs.map((item) => <span className="status-badge neutral" key={`${item.caseId}-${item.version}`}>{item.caseId} · v{item.version} <button type="button" disabled={disabled} onClick={() => setRefs(refs.filter((ref) => ref !== item))}>{t($ => $.cases.remove)}</button></span>)}</div>
      </details>
      {offline ? <p>{t($ => $.cases.offline)}</p> : null}
      <button type="button" disabled={offline || pending || loading} onClick={() => void refresh()}>{t($ => $.cases.refresh)}</button>
      {loading || pending ? <p role="status">{pending ? t($ => $.cases.saving) : t($ => $.cases.loading)}</p> : null}
      {!loading && !offline && !failure && suites.length === 0 ? <p>{t($ => $.cases.emptySuites)}</p> : null}
      {reviewState === "loading" ? <p role="status">{t($ => $.cases.loading)}</p> : null}
      {reviewState === "failed" && selected ? <p>{t($ => $.cases.reviewFailed)} <button type="button" disabled={pending} onClick={() => void inspect(selected)}>{t($ => $.cases.retryReview)}</button></p> : null}
      {failure ? <p role="alert" className="failure-summary">{evaluationFailure(t, failure)}{failure.startsWith("workflow_evaluation_") ? <> <code>{failure}</code></> : null}</p> : null}
      {decisionConflict ? <p role="alert">{t($ => $.cases.decisionConflict)}</p> : null}
      <div className="workflow-review-suite-layout">
        <div className="workflow-run-history-live-list" aria-label={t($ => $.cases.suiteHistory)}>
          {suites.map((item) => <button type="button" className={`workflow-run-history-live-row ${selected?.suiteId === item.suiteId ? "is-selected" : ""}`} key={item.suiteId} disabled={pending} onClick={() => void inspect(item)}><span className="workflow-run-history-live-identity"><strong>{item.name}</strong><small>{item.suiteId}</small></span><span><small>{t($ => $.cases.caseCount)}</small><strong>{item.caseRefs.length}</strong></span><span><small>{t($ => $.cases.decision)}</small><strong>{evaluationState(t, item.currentDecision || "unreviewed")}</strong><small>{t($ => $.cases.decisionVersion, { version: item.currentDecisionVersion })}</small></span></button>)}
        </div>
        {selected && review ? (
          <article className="workflow-evaluation-review">
            <div className="card-title-row"><div><p className="eyebrow">{t($ => $.cases.aggregate)}</p><h5>{evaluationState(t, review.outcome)}</h5></div><span className={`status-badge ${review.outcome === "passed" ? "good" : "neutral"}`}>{t($ => $.cases.passedCount, { count: review.passed })}</span></div>
            <dl className="tenant-meta"><div><dt>{t($ => $.cases.mismatch)}</dt><dd>{review.mismatch}</dd></div><div><dt>{t($ => $.cases.inconclusive)}</dt><dd>{review.inconclusive}</dd></div><div><dt>{t($ => $.cases.unavailable)}</dt><dd>{review.unavailable}</dd></div><div><dt>{t($ => $.cases.digest)}</dt><dd>{review.reviewDigest.slice(0, 12)}…</dd></div></dl>
            {review.items.map((item) => <div className="workflow-evaluation-review-item" key={`${item.caseId}-${item.version}`}><strong>{item.name || item.caseId}</strong><span>{item.caseId} · v{item.version}</span><span>{evaluationState(t, item.outcome)} · {evaluationState(t, item.runProfile)}</span></div>)}
            <fieldset className="workflow-review-decision-form" disabled={disabled}>
              <label>{t($ => $.cases.humanDecision)}<select value={decision} onChange={(event) => setDecision(event.target.value as WorkflowEvaluationReleaseDecisionKind)}><option value="needs_review">{t($ => $.cases.states.needs_review)}</option><option value="approved">{t($ => $.cases.states.approved)}</option><option value="rejected">{t($ => $.cases.states.rejected)}</option></select></label>
              <button type="button" onClick={() => void recordDecision()}>{t($ => $.cases.recordDecision, { version: selected.currentDecisionVersion + 1 })}</button>
            </fieldset>
            <div className="workflow-review-decision-history" aria-label={t($ => $.cases.decisionHistory)}>{decisions.map((item) => <div key={item.decisionId}><span><strong>v{item.version} · {evaluationState(t, item.decision)}</strong><small>{item.createdAt}</small></span><span><small>{t($ => $.cases.reviewed)}</small><strong>{evaluationState(t, item.reviewOutcome)}</strong></span><span><small>{t($ => $.cases.digest)}</small><code>{item.reviewDigest.slice(0, 12)}…</code></span></div>)}</div>
            <p className="workflow-review-stop-line">{t($ => $.cases.suiteStop)}</p>
          </article>
        ) : !selected ? <div className="workflow-review-empty-detail"><span aria-hidden="true">◇</span><h5>{t($ => $.cases.selectSuite)}</h5><p>{t($ => $.cases.selectSuiteHelp)}</p></div> : null}
      </div>
    </section>
  );
}
