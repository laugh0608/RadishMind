import type { TFunction } from "i18next";

export function identityRoleCopy(t: TFunction<"identity">, roleKey: string) {
  const key = roleKey === "workspace_reader" || roleKey === "workspace_builder" ||
    roleKey === "workspace_reviewer" || roleKey === "workspace_admin" ? roleKey : "unknown";
  return { name: t($ => $.roles[key].name), summary: t($ => $.roles[key].summary) };
}
