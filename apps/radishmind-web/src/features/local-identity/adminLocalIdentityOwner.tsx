import { LocalIdentityAdministrationPanel } from "./localIdentityAdministrationPanel.tsx";

export function AdminLocalIdentityOwner({
  surface,
  tenantRef,
  workspaceId,
}: {
  surface: "user" | "role";
  tenantRef: string;
  workspaceId: string;
}) {
  return (
    <LocalIdentityAdministrationPanel
      surface={surface}
      tenantRef={tenantRef}
      workspaceId={workspaceId}
    />
  );
}
