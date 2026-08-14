import assert from "node:assert/strict";
import test from "node:test";

import {
  promptAgentTypeWorkspaceCanonicalHashForTypeSwitch,
  promptAgentTypeWorkspaceOwnsHash,
  promptAgentTypeWorkspaceSurfaceForHash,
  promptAgentTypeWorkspaceTasks,
} from "../src/features/control-plane-read/promptAgentTypeWorkspaceModel.ts";

test("Prompt type path preserves every existing owner in task order", () => {
  const tasks = promptAgentTypeWorkspaceTasks("prompt_application", "active");
  assert.deepEqual(tasks.map((task) => task.surface), [
    "source",
    "configuration",
    "candidate",
    "assignment",
    "access",
    "controlled_use",
    "session",
    "evaluation",
  ]);
  assert.equal(tasks.every((task) => task.availability === "available"), true);
  assert.equal(tasks.find((task) => task.surface === "source")?.anchor, "prompt-application-template-workspace");
  assert.equal(tasks.find((task) => task.surface === "evaluation")?.anchor, "workspace-run-history");
});

test("Agent type path uses Profile and one Session v3 suggestion owner", () => {
  const tasks = promptAgentTypeWorkspaceTasks("agent_copilot", "active");
  assert.deepEqual(tasks.map((task) => task.surface), [
    "source",
    "configuration",
    "candidate",
    "assignment",
    "access",
    "controlled_use",
    "evaluation",
  ]);
  assert.equal(tasks.find((task) => task.surface === "source")?.anchor, "agent-copilot-profile-workspace");
  assert.equal(tasks.find((task) => task.surface === "controlled_use")?.anchor, "agent-copilot-invocation");
});

test("archived type paths keep evidence readable and block new controlled use", () => {
  const prompt = promptAgentTypeWorkspaceTasks("prompt_application", "archived");
  assert.equal(prompt.find((task) => task.surface === "source")?.availability, "read_only");
  assert.equal(prompt.find((task) => task.surface === "assignment")?.availability, "read_only");
  assert.equal(prompt.find((task) => task.surface === "access")?.availability, "read_only");
  assert.equal(prompt.find((task) => task.surface === "controlled_use")?.availability, "blocked");
  assert.equal(prompt.find((task) => task.surface === "session")?.availability, "blocked");
  assert.equal(prompt.find((task) => task.surface === "evaluation")?.availability, "read_only");
});

test("type workspace resolves exact owner anchors and shared S6 review anchors", () => {
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#prompt-application-template-workspace", "prompt_application", "configure_build"), "source");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#prompt-application-runtime-assignment", "prompt_application", "human_promotion"), "assignment");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#prompt-application-session", "prompt_application", "controlled_test"), "session");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#agent-copilot-profile-workspace", "agent_copilot", "configure_build"), "source");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#agent-copilot-invocation", "agent_copilot", "controlled_test"), "controlled_use");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#workflow-evaluation-release-review", "agent_copilot", "evidence_review"), "evaluation");
});

test("type workspace uses conservative stage fallbacks and rejects unrelated kinds", () => {
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#application-development-workspace", "prompt_application", "configure_build"), "source");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#application-development-workspace", "agent_copilot", "human_promotion"), "candidate");
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#application-development-workspace", "prompt_application", "release_readiness"), null);
  assert.equal(promptAgentTypeWorkspaceSurfaceForHash("#prompt-application-invocation", "workflow_rag", "controlled_test"), null);
  assert.deepEqual(promptAgentTypeWorkspaceTasks("unsupported", "unavailable"), []);
});

test("only exact S8 hashes keep the type owner surface open", () => {
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#application-publish-review", "prompt_application"), true);
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#workspace-api-keys", "agent_copilot"), true);
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#workspace-run-history", "agent_copilot"), true);
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#admin-control-plane", "agent_copilot"), false);
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#prompt-application-invocation-preview", "prompt_application"), false);
  assert.equal(promptAgentTypeWorkspaceOwnsHash("#prompt-application-invocation", "workflow_rag"), false);
});

test("application kind switches replace only the previous type's exact S8 anchor", () => {
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#agent-copilot-invocation",
      "agent_copilot",
      "prompt_application",
    ),
    "#prompt-application-invocation",
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#prompt-application-runtime-assignment",
      "prompt_application",
      "agent_copilot",
    ),
    "#agent-copilot-runtime-assignment",
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#prompt-application-session",
      "prompt_application",
      "agent_copilot",
    ),
    "#agent-copilot-invocation",
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#application-api-integration",
      "prompt_application",
      "agent_copilot",
    ),
    "#workspace-api-keys",
  );
});

test("type switch canonicalization preserves shared, review, unrelated, and same-type hashes", () => {
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#application-configuration-draft",
      "prompt_application",
      "agent_copilot",
    ),
    null,
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#workflow-run-comparison",
      "prompt_application",
      "agent_copilot",
    ),
    null,
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#application-development-workspace",
      "prompt_application",
      "agent_copilot",
    ),
    null,
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#prompt-application-invocation",
      "prompt_application",
      "prompt_application",
    ),
    null,
  );
  assert.equal(
    promptAgentTypeWorkspaceCanonicalHashForTypeSwitch(
      "#prompt-application-session",
      "unsupported",
      "agent_copilot",
    ),
    "#agent-copilot-invocation",
  );
});
