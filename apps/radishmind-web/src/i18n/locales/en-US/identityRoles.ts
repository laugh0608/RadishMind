export const identityRoles = {
  workspace_reader: { name: "Workspace reader", summary: "Read workspace applications, runs, workflows, sessions, evaluations, and usage metadata." },
  workspace_builder: { name: "Workspace builder", summary: "Create and execute workspace applications, workflows, sessions, profiles, and evaluations." },
  workspace_reviewer: { name: "Workspace reviewer", summary: "Build workspace resources and perform explicit review, activation, and confirmation actions." },
  workspace_admin: { name: "Workspace administrator", summary: "Manage the workspace, destructive lifecycle actions, policy surfaces, members, and role assignments." },
  unknown: { name: "Unrecognized role", summary: "Review the original role key and server-owned grants before proceeding." },
} as const;
