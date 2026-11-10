// A rejection of a later request cannot resolve an earlier unknown outcome.
export function promptExecutionNeedsReview(previous: boolean, failureCode: string, runStatus?: string): boolean {
  if ([
    "prompt_invocation_transport_failed", "prompt_invocation_response_invalid", "prompt_invocation_canceled",
    "prompt_invocation_duplicate_running", "prompt_invocation_outcome_unknown",
    "application_session_store_unavailable", "application_session_response_invalid", "application_session_request_canceled",
    "application_session_run_outcome_unknown", "application_session_turn_interrupted", "application_session_turn_canceled",
  ].includes(failureCode) || ["running", "canceled", "outcome_unknown"].includes(runStatus ?? "")) return true;
  if (runStatus === "succeeded" || runStatus === "failed") return false;
  return previous || Boolean(runStatus);
}

export function isPromptClientKey(value: string): boolean {
  return /^[A-Za-z0-9][A-Za-z0-9._:-]{7,159}$/u.test(value);
}
