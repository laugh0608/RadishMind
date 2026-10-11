/** Retain only a bounded machine code; server prose is never UI state. */
export class EvaluationRequestFailure extends Error {
  readonly code: string;
  constructor(code: string) {
    super(`workflow evaluation route failed: ${code}`);
    this.name = "EvaluationRequestFailure";
    this.code = code;
  }
}

export function evaluationEnvelopeFailure(value: unknown): EvaluationRequestFailure {
  const code = value && typeof value === "object" && "failure_code" in value ? value.failure_code : null;
  return new EvaluationRequestFailure(typeof code === "string" && /^workflow_evaluation(?:_suite)?_[a-z_]{1,80}$/.test(code) ? code : "contract");
}
