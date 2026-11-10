import { useCallback, useEffect, useRef } from "react";
import { PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT, createPromptApplicationCredentialHandoffDetail, type PromptApplicationCredentialHandoffDetail } from "./promptApplicationInvocationEvents.ts";

// The two surfaces mount exclusively. Hold the one-time handoff only until the
// invocation surface consumes it; never put credentials in workspace state.
export function usePromptCredentialHandoff(applicationId: string, generationKey: string) {
  const pending = useRef<{ generationKey: string; detail: PromptApplicationCredentialHandoffDetail } | null>(null);
  useEffect(() => {
    pending.current = null;
    function receive(event: Event) {
      const detail = (event as CustomEvent<PromptApplicationCredentialHandoffDetail>).detail;
      if (window.location.hash !== "#workspace-api-keys" || detail?.applicationId !== applicationId) return;
      try {
        pending.current = { generationKey, detail: createPromptApplicationCredentialHandoffDetail(detail.applicationId, detail.apiKeyId, detail.token) };
      } catch {
        pending.current = null;
      }
    }
    function clearOnLeave() {
      if (window.location.hash !== "#prompt-application-invocation") pending.current = null;
    }
    window.addEventListener(PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receive);
    window.addEventListener("hashchange", clearOnLeave);
    return () => {
      pending.current = null;
      window.removeEventListener(PROMPT_APPLICATION_CREDENTIAL_HANDOFF_EVENT, receive);
      window.removeEventListener("hashchange", clearOnLeave);
    };
  }, [applicationId, generationKey]);
  return useCallback(() => {
    const handoff = pending.current;
    pending.current = null;
    return window.location.hash === "#prompt-application-invocation" && handoff?.generationKey === generationKey
      ? handoff.detail : null;
  }, [generationKey]);
}
