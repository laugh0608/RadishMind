import { createContext, useContext } from "react";
import type { LocalIdentityAccountProfile, LocalIdentityConsumerConfig } from "./localIdentityConsumer.ts";

export type LocalIdentityContextValue = {
  config: LocalIdentityConsumerConfig;
  profile: LocalIdentityAccountProfile;
  authorityRevision: number;
  readAuthorityRevision: () => number;
  onWorkspaceScopeChange: (tenantRef: string, workspaceId: string) => void;
  refresh: () => Promise<void>;
  linkOIDC: () => Promise<void>;
  revokeExternalIdentity: (bindingId: string, expectedRecordVersion: number) => Promise<void>;
};

export const LocalIdentityContext = createContext<LocalIdentityContextValue | null>(null);

export function useLocalIdentity(): LocalIdentityContextValue | null {
  return useContext(LocalIdentityContext);
}
