import assert from "node:assert/strict";
import test from "node:test";
import { createWorkflowDraftEditorRequests } from "../src/features/control-plane-read/workflowDraftEditorRequests.ts";

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

test("late save cannot mark a new draft saved or refresh its library", async () => {
  const requests = createWorkflowDraftEditorRequests();
  requests.setScope("workspace:app:draft-a");
  const isCurrent = requests.begin();
  const save = deferred<number>();
  let version = 0;
  let dirty = true;
  let refreshes = 0;
  const completion = save.promise.then((savedVersion) => {
    if (!isCurrent()) return;
    version = savedVersion;
    dirty = false;
    refreshes += 1;
  });
  requests.setScope("workspace:app:draft-b");
  save.resolve(1);
  await completion;
  assert.deepEqual({ version, dirty, refreshes }, { version: 0, dirty: true, refreshes: 0 });
});

test("leaving and returning to the same scope does not revive an old read", async () => {
  const requests = createWorkflowDraftEditorRequests();
  requests.setScope("workspace-a:app:draft");
  const isCurrent = requests.begin();
  const read = deferred<number>();
  let version = 0;
  const completion = read.promise.then((value) => { if (isCurrent()) version = value; });
  requests.setScope("workspace-b:app:draft");
  requests.setScope("workspace-a:app:draft");
  read.resolve(7);
  await completion;
  assert.equal(version, 0);
});

test("a newer editor request rejects a late failure but accepts its own result", async () => {
  const requests = createWorkflowDraftEditorRequests();
  requests.setScope("workspace:app:draft");
  const first = requests.begin();
  const validation = deferred<void>();
  let state = "reading";
  const completion = validation.promise.catch(() => { if (first()) state = "validation_failed"; });
  const latest = requests.begin();
  requests.setScope("workspace:app:draft");
  validation.reject(new Error("late validation failure"));
  await completion;
  assert.equal(state, "reading");
  assert.equal(latest(), true);
  if (latest()) state = "read_dev_record";
  assert.equal(state, "read_dev_record");
});

test("explicit selection, reset, or unmount invalidates requests within the same scope", () => {
  const requests = createWorkflowDraftEditorRequests();
  requests.setScope("workspace:app:draft");
  const prior = requests.begin();
  requests.invalidate();
  assert.equal(prior(), false);
  const next = requests.begin();
  assert.equal(next(), true);
});
