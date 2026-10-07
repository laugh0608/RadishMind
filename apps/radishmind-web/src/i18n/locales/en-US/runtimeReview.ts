export const runtimeReview = {
  "eyebrow": "Application runtime review",
  "title": "Run, review, and observe",
  "intro": "Move from one controlled request to its sanitized record, then inspect the current application windows without implying correlation.",
  "application": "Application",
  "workspace": "Workspace",
  "lifecycle": "Lifecycle",
  "tasks": "Application runtime review tasks",
  "path": "Runtime path",
  "oneTask": "One task at a time",
  "boundary": "Request input and output are temporary. History is sanitized. Operations combines independent current windows only. Saved results are explicit owner-scoped artifacts.",
  "loading": "Loading the current runtime owner…",
  "archived": "archived · read only",
  "unavailable": "unavailable",
  "tasksBySurface": {
    "run": {
      "label": "Run request",
      "summary": "Temporary input and output"
    },
    "request": {
      "label": "Review request",
      "summary": "Exact sanitized record"
    },
    "evidence": {
      "label": "Application evidence",
      "summary": "Current loaded windows"
    },
    "results": {
      "label": "Saved results",
      "summary": "Cross-Session exact artifacts"
    }
  }
} as const;
