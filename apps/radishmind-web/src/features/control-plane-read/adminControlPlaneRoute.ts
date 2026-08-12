export type AdminControlPlaneSurface =
  | "tenant"
  | "user"
  | "role"
  | "audit"
  | "provider"
  | "profile"
  | "route"
  | "quota"
  | "pricing";

export type AdminControlPlaneResourceTask = {
  surface: AdminControlPlaneSurface;
  anchor: string;
  label: string;
  scope: string;
  number: string;
};

export const ADMIN_CONTROL_PLANE_RESOURCE_TASKS: ReadonlyArray<AdminControlPlaneResourceTask> = [
  { surface: "tenant", anchor: "admin-tenant-overview", label: "Tenant", scope: "tenant:read", number: "01" },
  { surface: "user", anchor: "admin-user-directory", label: "User", scope: "Radish owner", number: "02" },
  { surface: "role", anchor: "admin-role-policy", label: "Role", scope: "policy mapping", number: "03" },
  { surface: "audit", anchor: "admin-audit-log", label: "Audit", scope: "audit:read", number: "04" },
  { surface: "provider", anchor: "admin-provider-config", label: "Provider", scope: "inventory ref", number: "05" },
  { surface: "profile", anchor: "admin-profile-config", label: "Profile", scope: "assignment", number: "06" },
  { surface: "route", anchor: "admin-route-config", label: "Route", scope: "generation CAS", number: "07" },
  { surface: "quota", anchor: "admin-gateway-request-quota", label: "Quota", scope: "UTC daily CAS", number: "08" },
  { surface: "pricing", anchor: "admin-gateway-model-pricing", label: "Pricing", scope: "USD / 1M CAS", number: "09" },
];

const SURFACE_BY_HASH = new Map<string, AdminControlPlaneSurface>([
  ["#admin-control-plane", "tenant"],
  ...ADMIN_CONTROL_PLANE_RESOURCE_TASKS.map((task) => [`#${task.anchor}`, task.surface] as const),
  ["#admin-operations-review", "tenant"],
  ["#admin-provider-deployment-review", "route"],
]);

export function adminControlPlaneSurfaceForHash(hash: string): AdminControlPlaneSurface | null {
  return SURFACE_BY_HASH.get(hash) ?? null;
}

export function adminControlPlaneAnchorForSurface(surface: AdminControlPlaneSurface): string {
  return ADMIN_CONTROL_PLANE_RESOURCE_TASKS.find((task) => task.surface === surface)?.anchor ??
    "admin-tenant-overview";
}
