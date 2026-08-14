import assert from "node:assert/strict";
import test from "node:test";

import {
  ADMIN_CONTROL_PLANE_RESOURCE_TASKS,
  adminControlPlaneAnchorForSurface,
  adminControlPlaneSurfaceForHash,
} from "../src/features/control-plane-read/adminControlPlaneRoute.ts";

test("Admin Control Plane exposes nine exact resource tasks without merging independent owners", () => {
  assert.deepEqual(
    ADMIN_CONTROL_PLANE_RESOURCE_TASKS.map((task) => task.surface),
    ["tenant", "user", "role", "audit", "provider", "profile", "route", "quota", "pricing"],
  );
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "tenant")?.scope, "tenant:read");
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "audit")?.scope, "audit:read");
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "user")?.scope, "Radish owner");
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "role")?.scope, "policy mapping");
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "quota")?.scope, "UTC daily CAS");
  assert.equal(ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === "pricing")?.scope, "USD / 1M CAS");
});

test("Admin Control Plane hash ownership is exact and keeps legacy evidence reachable", () => {
  assert.equal(adminControlPlaneSurfaceForHash("#admin-control-plane"), "tenant");
  for (const task of ADMIN_CONTROL_PLANE_RESOURCE_TASKS) {
    assert.equal(adminControlPlaneSurfaceForHash(`#${task.anchor}`), task.surface);
    assert.equal(adminControlPlaneAnchorForSurface(task.surface), task.anchor);
  }
  assert.equal(adminControlPlaneSurfaceForHash("#admin-operations-review"), "tenant");
  assert.equal(adminControlPlaneSurfaceForHash("#admin-provider-deployment-review"), "route");
  assert.equal(adminControlPlaneSurfaceForHash("#admin-gateway-request-quota"), "quota");
  assert.equal(adminControlPlaneSurfaceForHash("#admin-gateway-model-pricing"), "pricing");
  assert.equal(adminControlPlaneSurfaceForHash("#admin-route-config-production"), null);
  assert.equal(adminControlPlaneSurfaceForHash("admin-audit-log"), null);
});
