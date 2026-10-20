export const artifactFeedback = {
  "generic": "The request failed. Reload the current result and review it before trying again.",
  "offline": "Result access is disabled in this mode; no result request was sent.",
  "permission": "Result access was denied. Check the current application, workspace and permissions.",
  "invalid": "The result request is invalid. Check the selected scope and filters.",
  "unavailable": "The result store is unavailable. Retry the read explicitly after it recovers.",
  "notFound": "This exact result is unavailable in the current scope. Refresh the list and select again.",
  "contract": "The response failed scope or content verification. No content or download was accepted.",
  "conflict": "The lifecycle changed. Refresh the list, read the current result, then explicitly review the operation again.",
  "canceled": "The request was canceled. No result was accepted.",
  "content": "The result could not be saved because its content exceeds the allowed size or its source is unavailable."
} as const;
