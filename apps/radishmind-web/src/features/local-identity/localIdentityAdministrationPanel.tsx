import { useTranslation } from "react-i18next";
import { useLocalePreference } from "../../i18n/LocaleProvider.tsx";
import { formatDisplayNumber } from "../../i18n/formatters.ts";
import "../../i18n/identityMemberResources.ts";
import { identityRoleCopy } from "./localIdentityRoleMessages.ts";
import { memberFailure, memberFailureCopy, memberSuccessMessage, memberDate, type MemberFailure, type MemberSuccess } from "./localIdentityMemberMessages.ts";
import type { TFunction } from "i18next";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  LocalIdentityAdministrationError,
  assignLocalIdentityWorkspaceRole,
  createLocalIdentityWorkspaceMembership,
  isValidLocalIdentityAdministrationUserId,
  listLocalIdentityWorkspaceMembers,
  normalizeLocalIdentityAdministrationExpiration,
  readLocalIdentityRoleCatalog,
  readLocalIdentityWorkspaceMember,
  revokeLocalIdentityWorkspaceMembership,
  revokeLocalIdentityWorkspaceRole,
  type LocalIdentityAdministrationConfig,
  type LocalIdentityMembershipState,
  type LocalIdentityRoleCatalog,
  type LocalIdentityRoleDefinition,
  type LocalIdentityWorkspaceMemberDetail,
  type LocalIdentityWorkspaceMemberSummary,
  type LocalIdentityWorkspaceMembershipView,
  type LocalIdentityWorkspaceRoleAssignmentView,
} from "./localIdentityAdministrationConsumer.ts";
import type { LocalIdentityAccountProfile } from "./localIdentityConsumer.ts";
import { useLocalIdentity } from "./localIdentityGateway.tsx";

type LoadStatus = "idle" | "loading" | "ready" | "empty" | "denied" | "unavailable" | "failed";

type LoadFailure = MemberFailure;

type MutationCandidate =
  | { kind: "membership_create"; userId: string; expiresAt?: string }
  | {
      kind: "membership_revoke";
      userId: string;
      membershipId: string;
      expectedRecordVersion: number;
      activeAssignmentCount: number;
    }
  | {
      kind: "role_assign";
      userId: string;
      role: LocalIdentityRoleDefinition;
      expiresAt?: string;
    }
  | {
      kind: "role_revoke";
      userId: string;
      assignmentId: string;
      roleKey: string;
      expectedRecordVersion: number;
    };

type OperationNotice =
  | { kind: "idle"; code: "" }
  | { kind: "success"; code: "mutation_succeeded"; success: MemberSuccess }
  | (MemberFailure & { input?: "membership" | "role" });

const EMPTY_OPERATION: OperationNotice = { kind: "idle", code: "" };

export function LocalIdentityAdministrationPanel({
  surface,
  tenantRef,
  workspaceId,
}: {
  surface: "user" | "role";
  tenantRef: string;
  workspaceId: string;
}) {
  const identity = useLocalIdentity();
  if (!identity) return <DisabledLocalIdentityAdministration surface={surface} />;

  const scopeKey = localIdentityAdministrationScopeKey(identity.profile, tenantRef, workspaceId);
  return (
    <LocalIdentityAdministrationWorkspace
      key={scopeKey}
      surface={surface}
      tenantRef={tenantRef}
      workspaceId={workspaceId}
      profile={identity.profile}
      baseConfig={identity.config}
      refreshIdentity={identity.refresh}
    />
  );
}

function LocalIdentityAdministrationWorkspace({
  surface,
  tenantRef,
  workspaceId,
  profile,
  baseConfig,
  refreshIdentity,
}: {
  surface: "user" | "role";
  tenantRef: string;
  workspaceId: string;
  profile: LocalIdentityAccountProfile;
  baseConfig: { mode: "disabled" | "local_identity_dev"; baseUrl: string };
  refreshIdentity: () => Promise<void>;
}) {
  const { t } = useTranslation("identity");
  const config = useMemo<LocalIdentityAdministrationConfig>(() => ({
    ...baseConfig,
    tenantRef: tenantRef.trim(),
    workspaceId: workspaceId.trim(),
  }), [baseConfig.baseUrl, baseConfig.mode, tenantRef, workspaceId]);
  const [membershipState, setMembershipState] = useState<LocalIdentityMembershipState>("active");
  const [directoryStatus, setDirectoryStatus] = useState<LoadStatus>("idle");
  const [directoryFailure, setDirectoryFailure] = useState<LoadFailure | null>(null);
  const [members, setMembers] = useState<LocalIdentityWorkspaceMemberSummary[]>([]);
  const [nextCursor, setNextCursor] = useState("");
  const [paginationPending, setPaginationPending] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState("");
  const [selectedMembershipId, setSelectedMembershipId] = useState("");
  const [detailStatus, setDetailStatus] = useState<LoadStatus>("idle");
  const [detailFailure, setDetailFailure] = useState<LoadFailure | null>(null);
  const [detail, setDetail] = useState<LocalIdentityWorkspaceMemberDetail | null>(null);
  const [catalogStatus, setCatalogStatus] = useState<LoadStatus>("idle");
  const [catalogFailure, setCatalogFailure] = useState<LoadFailure | null>(null);
  const [catalog, setCatalog] = useState<LocalIdentityRoleCatalog | null>(null);
  const [directoryRefresh, setDirectoryRefresh] = useState(0);
  const [detailRefresh, setDetailRefresh] = useState(0);
  const [catalogRefresh, setCatalogRefresh] = useState(0);
  const [membershipCreateOpen, setMembershipCreateOpen] = useState(false);
  const [membershipUserId, setMembershipUserId] = useState("");
  const [membershipExpiresAt, setMembershipExpiresAt] = useState("");
  const [roleExpiresAt, setRoleExpiresAt] = useState("");
  const [selectedRoleKey, setSelectedRoleKey] = useState("");
  const [candidate, setCandidate] = useState<MutationCandidate | null>(null);
  const [operation, setOperation] = useState<OperationNotice>(EMPTY_OPERATION);
  const [mutationPending, setMutationPending] = useState(false);
  const directoryGeneration = useRef(0);
  const detailGeneration = useRef(0);
  const catalogGeneration = useRef(0);
  const mutationGeneration = useRef(0);
  const preferredSelection = useRef({ userId: "", membershipId: "" });
  const mounted = useRef(true);
  const previousSurface = useRef(surface);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      directoryGeneration.current += 1;
      detailGeneration.current += 1;
      catalogGeneration.current += 1;
      mutationGeneration.current += 1;
    };
  }, []);

  useEffect(() => {
    if (previousSurface.current === surface) return;
    previousSurface.current = surface;
    mutationGeneration.current += 1;
    setOperation(EMPTY_OPERATION);
    clearMutationDrafts();
  }, [surface]);

  useEffect(() => {
    const controller = new AbortController();
    const generation = directoryGeneration.current + 1;
    const preferred = preferredSelection.current;
    preferredSelection.current = { userId: "", membershipId: "" };
    directoryGeneration.current = generation;
    detailGeneration.current += 1;
    setDirectoryStatus("loading");
    setDirectoryFailure(null);
    setMembers([]);
    setNextCursor("");
    setPaginationPending(false);
    setSelectedUserId("");
    setSelectedMembershipId("");
    setDetailStatus("idle");
    setDetailFailure(null);
    setDetail(null);
    clearMutationDrafts();
    void listLocalIdentityWorkspaceMembers(config, { membershipState, limit: 100 }, controller.signal)
      .then((page) => {
        if (!mounted.current || directoryGeneration.current !== generation) return;
        setMembers(page.members);
        setNextCursor(page.nextCursor);
        setDirectoryStatus(page.members.length === 0 ? "empty" : "ready");
        const selected = page.members.find((member) => member.membershipId === preferred.membershipId) ??
          page.members.find((member) => member.userId === preferred.userId) ?? page.members[0];
        setSelectedUserId(selected?.userId ?? "");
        setSelectedMembershipId(selected?.membershipId ?? "");
      })
      .catch((error: unknown) => {
        if (isAbort(error) || !mounted.current || directoryGeneration.current !== generation) return;
        const failure = memberFailure(error);
        setDirectoryStatus(loadStatusForFailure(failure));
        setDirectoryFailure(failure);
      });
    return () => controller.abort();
  }, [config, directoryRefresh, membershipState]);

  useEffect(() => {
    if (!selectedUserId) return;
    const controller = new AbortController();
    const generation = detailGeneration.current + 1;
    detailGeneration.current = generation;
    setDetailStatus("loading");
    setDetailFailure(null);
    setDetail(null);
    setCandidate(null);
    setSelectedRoleKey("");
    setRoleExpiresAt("");
    void readLocalIdentityWorkspaceMember(config, selectedUserId, controller.signal)
      .then((result) => {
        if (!mounted.current || detailGeneration.current !== generation) return;
        setDetail(result.member);
        setDetailStatus("ready");
      })
      .catch((error: unknown) => {
        if (isAbort(error) || !mounted.current || detailGeneration.current !== generation) return;
        const failure = memberFailure(error);
        setDetailStatus(loadStatusForFailure(failure));
        setDetailFailure(failure);
      });
    return () => controller.abort();
  }, [config, detailRefresh, selectedUserId]);

  useEffect(() => {
    const controller = new AbortController();
    const generation = catalogGeneration.current + 1;
    catalogGeneration.current = generation;
    setCatalogStatus("loading");
    setCatalogFailure(null);
    setCatalog(null);
    setSelectedRoleKey("");
    setRoleExpiresAt("");
    setCandidate((current) => current?.kind === "membership_create" || current?.kind === "membership_revoke"
      ? current
      : null);
    void readLocalIdentityRoleCatalog(config, controller.signal)
      .then((result) => {
        if (!mounted.current || catalogGeneration.current !== generation) return;
        setCatalog(result.catalog);
        setCatalogStatus("ready");
      })
      .catch((error: unknown) => {
        if (isAbort(error) || !mounted.current || catalogGeneration.current !== generation) return;
        const failure = memberFailure(error);
        setCatalogStatus(loadStatusForFailure(failure));
        setCatalogFailure(failure);
      });
    return () => controller.abort();
  }, [catalogRefresh, config]);

  function clearMutationDrafts() {
    setMembershipCreateOpen(false);
    setMembershipUserId("");
    setMembershipExpiresAt("");
    setRoleExpiresAt("");
    setSelectedRoleKey("");
    setCandidate(null);
  }

  function selectMember(userId: string, membershipId: string) {
    if (membershipId === selectedMembershipId) return;
    detailGeneration.current += 1;
    setSelectedUserId(userId);
    setSelectedMembershipId(membershipId);
    setOperation(EMPTY_OPERATION);
    clearMutationDrafts();
  }

  function switchMembershipState(state: LocalIdentityMembershipState) {
    if (state === membershipState) return;
    preferredSelection.current = { userId: "", membershipId: "" };
    setMembershipState(state);
    setOperation(EMPTY_OPERATION);
  }

  function retryDirectory() {
    preferredSelection.current = { userId: selectedUserId, membershipId: selectedMembershipId };
    setOperation(EMPTY_OPERATION);
    setDirectoryRefresh((value) => value + 1);
  }

  function retryAll() {
    preferredSelection.current = { userId: selectedUserId, membershipId: selectedMembershipId };
    setOperation(EMPTY_OPERATION);
    setDirectoryRefresh((value) => value + 1);
    setCatalogRefresh((value) => value + 1);
    void refreshIdentity();
  }

  function retryDetail() {
    setCandidate(null);
    setOperation(EMPTY_OPERATION);
    setDetailRefresh((value) => value + 1);
  }

  function retryCatalog() {
    setCandidate(null);
    setSelectedRoleKey("");
    setOperation(EMPTY_OPERATION);
    setCatalogRefresh((value) => value + 1);
  }

  async function loadMoreMembers() {
    if (!nextCursor || paginationPending) return;
    const generation = directoryGeneration.current + 1;
    directoryGeneration.current = generation;
    setPaginationPending(true);
    setDirectoryFailure(null);
    try {
      const page = await listLocalIdentityWorkspaceMembers(config, {
        membershipState,
        limit: 100,
        cursor: nextCursor,
      });
      if (!mounted.current || directoryGeneration.current !== generation) return;
      const known = new Set(members.map((member) => member.membershipId));
      if (page.members.some((member) => known.has(member.membershipId))) {
        throw new LocalIdentityAdministrationError(
          200,
          "local_identity_response_invalid",
          "The workspace member cursor returned duplicate records.",
        );
      }
      setMembers((current) => [...current, ...page.members]);
      setNextCursor(page.nextCursor);
    } catch (error) {
      if (!mounted.current || directoryGeneration.current !== generation) return;
      const failure = memberFailure(error);
      setMembers([]);
      setNextCursor("");
      setSelectedUserId("");
      setSelectedMembershipId("");
      setDetail(null);
      setDetailStatus("idle");
      setDirectoryStatus(loadStatusForFailure(failure));
      setDirectoryFailure(failure);
    } finally {
      if (mounted.current && directoryGeneration.current === generation) setPaginationPending(false);
    }
  }

  function reviewMembershipCreate() {
    const userId = membershipUserId.trim();
    const expiresAt = normalizeLocalIdentityAdministrationExpiration(membershipExpiresAt);
    if (!isValidLocalIdentityAdministrationUserId(userId) || expiresAt === null) {
      setOperation({
        kind: "failed",
        code: "local_identity_input_invalid",
        input: "membership",
      });
      setCandidate(null);
      return;
    }
    setOperation(EMPTY_OPERATION);
    setCandidate({ kind: "membership_create", userId, ...(expiresAt ? { expiresAt } : {}) });
  }

  function reviewRoleAssignment() {
    const role = catalog?.roles.find((entry) => entry.roleKey === selectedRoleKey);
    const expiresAt = normalizeLocalIdentityAdministrationExpiration(roleExpiresAt);
    if (!detail || !role || expiresAt === null || !effectiveMembership(detail, selectedSummary)?.effective) {
      setOperation({
        kind: "failed",
        code: "local_identity_input_invalid",
        input: "role",
      });
      setCandidate(null);
      return;
    }
    setOperation(EMPTY_OPERATION);
    setCandidate({ kind: "role_assign", userId: detail.userId, role, ...(expiresAt ? { expiresAt } : {}) });
  }

  function reviewMembershipRevoke(membership: LocalIdentityWorkspaceMembershipView) {
    if (!detail) return;
    setOperation(EMPTY_OPERATION);
    setCandidate({
      kind: "membership_revoke",
      userId: detail.userId,
      membershipId: membership.membershipId,
      expectedRecordVersion: membership.recordVersion,
      activeAssignmentCount: detail.roleAssignments.filter((assignment) =>
        assignment.scope === "workspace" && assignment.effective
      ).length,
    });
  }

  function reviewRoleRevoke(assignment: LocalIdentityWorkspaceRoleAssignmentView) {
    if (!detail) return;
    setOperation(EMPTY_OPERATION);
    setCandidate({
      kind: "role_revoke",
      userId: detail.userId,
      assignmentId: assignment.assignmentId,
      roleKey: assignment.roleKey,
      expectedRecordVersion: assignment.recordVersion,
    });
  }

  async function confirmCandidate() {
    if (!candidate || mutationPending) return;
    const generation = mutationGeneration.current + 1;
    mutationGeneration.current = generation;
    setMutationPending(true);
    setOperation(EMPTY_OPERATION);
    try {
      let success: MemberSuccess;
      let preferredUserId = candidate.userId;
      let preferredMembershipId = selectedMembershipId;
      if (candidate.kind === "membership_create") {
        const result = await createLocalIdentityWorkspaceMembership(config, candidate);
        success = { kind: "membershipCreated", membershipId: result.membership.membershipId };
        preferredMembershipId = "";
      } else if (candidate.kind === "membership_revoke") {
        const result = await revokeLocalIdentityWorkspaceMembership(config, candidate);
        success = { kind: "membershipRevoked", count: result.revokedRoleAssignments.length };
        preferredUserId = "";
        preferredMembershipId = "";
      } else if (candidate.kind === "role_assign") {
        const result = await assignLocalIdentityWorkspaceRole(config, {
          userId: candidate.userId,
          roleKey: candidate.role.roleKey,
          expectedCatalogVersion: candidate.role.catalogVersion,
          expectedRoleDefinitionDigest: candidate.role.definitionDigest,
          ...(candidate.expiresAt ? { expiresAt: candidate.expiresAt } : {}),
        });
        success = { kind: "roleAssigned", roleKey: result.roleAssignment.roleKey };
      } else {
        const result = await revokeLocalIdentityWorkspaceRole(config, candidate);
        success = { kind: "roleRevoked", roleKey: result.roleAssignment.roleKey, version: result.roleAssignment.recordVersion };
      }
      if (!mounted.current || mutationGeneration.current !== generation) return;
      setOperation({ kind: "success", code: "mutation_succeeded", success });
      clearMutationDrafts();
      setMutationPending(false);
      preferredSelection.current = { userId: preferredUserId, membershipId: preferredMembershipId };
      directoryGeneration.current += 1;
      detailGeneration.current += 1;
      catalogGeneration.current += 1;
      setDirectoryRefresh((value) => value + 1);
      setCatalogRefresh((value) => value + 1);
      void refreshIdentity();
    } catch (error) {
      if (!mounted.current || mutationGeneration.current !== generation) return;
      const failure = memberFailure(error);
      const kind = failure.kind;
      setCandidate(null);
      setOperation(failure);
      if (kind === "denied" || kind === "unavailable") {
        directoryGeneration.current += 1;
        detailGeneration.current += 1;
        catalogGeneration.current += 1;
        setMembers([]);
        setNextCursor("");
        setSelectedUserId("");
        setSelectedMembershipId("");
        setDetail(null);
        setCatalog(null);
        setDirectoryStatus(kind);
        setDetailStatus("idle");
        setCatalogStatus(kind);
      }
    } finally {
      if (mounted.current && mutationGeneration.current === generation) setMutationPending(false);
    }
  }

  const selectedSummary = members.find((member) => member.membershipId === selectedMembershipId) ??
    members.find((member) => member.userId === selectedUserId) ?? null;
  const membership = detail ? effectiveMembership(detail, selectedSummary) : null;
  const activeAssignments = detail?.roleAssignments.filter((assignment) =>
    assignment.effective && assignment.scope === "workspace" && assignment.workspaceId === config.workspaceId
  ) ?? [];
  const selfMembership = detail?.userId === profile.account.userId;
  const roleAlreadyAssigned = activeAssignments.some((assignment) => assignment.roleKey === selectedRoleKey);
  const mutationsEnabled = profile.capabilities.recentAuthentication && operation.kind !== "denied" &&
    operation.kind !== "unavailable";

  return (
    <section
      className="admin-control-owner-surface local-identity-administration"
      aria-labelledby="local-identity-administration-title"
      data-surface={surface}
    >
      <header>
        <div>
          <p className="eyebrow">{surface === "user" ? t($ => $.members.user) : t($ => $.members.role)}{t($ => $.members.administrationSuffix)}</p>
          <h4 id="local-identity-administration-title">
            {surface === "user" ? t($ => $.members.members) : t($ => $.members.memberRoles)}
          </h4>
          <p>
            {surface === "user"
              ? t($ => $.members.membersHelp)
              : t($ => $.members.rolesHelp)}
          </p>
        </div>
        <span className={`admin-control-status is-${directoryStatus === "ready" ? "ready" : directoryStatus === "denied" || directoryStatus === "unavailable" ? "blocked" : "neutral"}`}>
          {directoryStatus === "ready" ? t($ => $.members.localSession) : t($ => $.members.states[directoryStatus])}
        </span>
      </header>

      <dl className="local-identity-administration-scope">
        <div><dt>{t($ => $.members.tenant)}</dt><dd>{config.tenantRef}</dd><small>{t($ => $.members.verifiedScope)}</small></div>
        <div><dt>{t($ => $.members.workspace)}</dt><dd>{config.workspaceId}</dd><small>{t($ => $.members.activeMembership)}</small></div>
        <div><dt>{t($ => $.members.permission)}</dt><dd>{surface === "user" ? "members:read" : "roles:read + assign"}</dd><small>{t($ => $.members.perRequest)}</small></div>
        <div><dt>{surface === "user" ? t($ => $.members.cursor) : t($ => $.members.catalog)}</dt><dd>{surface === "user" ? t($ => $.members.filterBound) : catalog?.catalogVersion ?? t($ => $.members.states[catalogStatus])}</dd><small>{t($ => $.members.volatileState)}</small></div>
      </dl>

      {operation.kind !== "idle" ? (
        <OperationBanner
          operation={operation}
          onReloadDetail={retryDetail}
          onReloadCatalog={retryCatalog}
          onRetryAll={retryAll}
        />
      ) : null}

      {surface === "user" ? (
        <div className="local-identity-member-workspace">
          <MemberDirectory
            status={directoryStatus}
            failure={directoryFailure}
            membershipState={membershipState}
            members={members}
            nextCursor={nextCursor}
            paginationPending={paginationPending}
            selectedMembershipId={selectedMembershipId}
            createOpen={membershipCreateOpen}
            createUserId={membershipUserId}
            createExpiresAt={membershipExpiresAt}
            candidate={candidate}
            mutationPending={mutationPending}
            mutationsEnabled={mutationsEnabled}
            onMembershipState={switchMembershipState}
            onSelect={selectMember}
            onLoadMore={() => void loadMoreMembers()}
            onRetry={retryDirectory}
            onToggleCreate={() => {
              setMembershipCreateOpen((open) => !open);
              setCandidate(null);
              setOperation(EMPTY_OPERATION);
            }}
            onCreateUserId={setMembershipUserId}
            onCreateExpiresAt={setMembershipExpiresAt}
            onReviewCreate={reviewMembershipCreate}
          />
          <MemberInspector
            actorUserId={profile.account.userId}
            status={detailStatus}
            failure={detailFailure}
            summary={selectedSummary}
            detail={detail}
            membership={membership}
            candidate={candidate}
            mutationPending={mutationPending}
            mutationsEnabled={mutationsEnabled}
            onRetry={retryDetail}
            onReviewRevoke={reviewMembershipRevoke}
          />
        </div>
      ) : (
        <RoleAssignmentWorkspace
          directoryStatus={directoryStatus}
          selectedSummary={selectedSummary}
          detailStatus={detailStatus}
          detailFailure={detailFailure}
          detail={detail}
          membership={membership}
          catalogStatus={catalogStatus}
          catalogFailure={catalogFailure}
          catalog={catalog}
          selectedRoleKey={selectedRoleKey}
          roleExpiresAt={roleExpiresAt}
          roleAlreadyAssigned={roleAlreadyAssigned}
          candidate={candidate}
          mutationPending={mutationPending}
          mutationsEnabled={mutationsEnabled}
          onSelectRole={(roleKey) => {
            setSelectedRoleKey(roleKey);
            setCandidate(null);
            setOperation(EMPTY_OPERATION);
          }}
          onRoleExpiresAt={setRoleExpiresAt}
          onReviewAssign={reviewRoleAssignment}
          onReviewRevoke={reviewRoleRevoke}
          onRetryDetail={retryDetail}
          onRetryCatalog={retryCatalog}
        />
      )}

      {candidate ? (
        <MutationConfirmation
          candidate={candidate}
          tenantRef={config.tenantRef}
          workspaceId={config.workspaceId}
          pending={mutationPending}
          recentAuthentication={profile.capabilities.recentAuthentication}
          onCancel={() => setCandidate(null)}
          onConfirm={() => void confirmCandidate()}
        />
      ) : null}

      <p className="local-identity-administration-boundary">
        <span aria-hidden="true">!</span>{t($ => $.members.boundary)}</p>
    </section>
  );
}

function MemberDirectory({
  status,
  failure,
  membershipState,
  members,
  nextCursor,
  paginationPending,
  selectedMembershipId,
  createOpen,
  createUserId,
  createExpiresAt,
  candidate,
  mutationPending,
  mutationsEnabled,
  onMembershipState,
  onSelect,
  onLoadMore,
  onRetry,
  onToggleCreate,
  onCreateUserId,
  onCreateExpiresAt,
  onReviewCreate,
}: {
  status: LoadStatus;
  failure: LoadFailure | null;
  membershipState: LocalIdentityMembershipState;
  members: LocalIdentityWorkspaceMemberSummary[];
  nextCursor: string;
  paginationPending: boolean;
  selectedMembershipId: string;
  createOpen: boolean;
  createUserId: string;
  createExpiresAt: string;
  candidate: MutationCandidate | null;
  mutationPending: boolean;
  mutationsEnabled: boolean;
  onMembershipState: (state: LocalIdentityMembershipState) => void;
  onSelect: (userId: string, membershipId: string) => void;
  onLoadMore: () => void;
  onRetry: () => void;
  onToggleCreate: () => void;
  onCreateUserId: (value: string) => void;
  onCreateExpiresAt: (value: string) => void;
  onReviewCreate: () => void;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  return (
    <section className="local-identity-member-directory" aria-labelledby="local-identity-directory-title">
      <header>
        <div><span>{t($ => $.members.directoryTitle)}</span><h5 id="local-identity-directory-title">{membershipState === "active" ? t($ => $.members.activeMembers) : t($ => $.members.revokedHistory)}</h5></div>
        <button type="button" className="secondary-action" onClick={onToggleCreate} disabled={!mutationsEnabled || mutationPending}>
          {createOpen ? t($ => $.members.closeUserForm) : t($ => $.members.addExactUser)}
        </button>
      </header>
      <div className="local-identity-directory-filters" aria-label={t($ => $.members.lifecycleFilter)}>
        <button type="button" aria-pressed={membershipState === "active"} onClick={() => onMembershipState("active")}>{t($ => $.members.activeFilter)}</button>
        <button type="button" aria-pressed={membershipState === "revoked"} onClick={() => onMembershipState("revoked")}>{t($ => $.members.revokedFilter)}</button>
        <span>{t($ => $.members.pageLimit)}</span>
      </div>
      {createOpen ? (
        <div className="local-identity-membership-create">
          <label>{t($ => $.members.exactUserId)}<input value={createUserId} onChange={(event) => onCreateUserId(event.target.value)} placeholder="usr_…" autoComplete="off" /></label>
          <label>{t($ => $.members.expiresInput)}<input value={createExpiresAt} onChange={(event) => onCreateExpiresAt(event.target.value)} placeholder="2026-09-01T00:00:00Z" autoComplete="off" /></label>
          <button type="button" onClick={onReviewCreate} disabled={!mutationsEnabled || mutationPending}>{t($ => $.members.reviewMembership)}</button>
          <small>{t($ => $.members.exactUserHelp)}</small>
        </div>
      ) : null}
      {status === "loading" ? <StateNotice state="loading" title={t($ => $.members.loadingMembers)} /> : null}
      {status === "empty" ? <StateNotice state="empty" title={t($ => $.members.emptyMemberships, { state: t($ => $.members.states[membershipState]) })} /> : null}
      {status === "denied" || status === "unavailable" || status === "failed" ? (
        <StateFailure failure={failure} onRetry={onRetry} />
      ) : null}
      {status === "ready" ? (
        <div className="local-identity-member-rows">
          {members.map((member) => (
            <button
              type="button"
              key={member.membershipId}
              className={member.membershipId === selectedMembershipId ? "is-selected" : ""}
              aria-pressed={member.membershipId === selectedMembershipId}
              onClick={() => onSelect(member.userId, member.membershipId)}
            >
              <i aria-hidden="true" />
              <span className="local-identity-member-avatar" aria-hidden="true">{initials(member.displayName)}</span>
              <span><strong>{member.displayName}</strong><small>{member.userId}</small></span>
              <span className="local-identity-member-role">
                <strong>{member.roleKeys[0] ?? t($ => $.members.noAssignment)}</strong>
                <small>{member.roleCatalogDrift ? t($ => $.members.catalogDrift) : member.membershipEffective ? t($ => $.members.effective) : t($ => $.members.states[member.membershipLifecycleState])}</small>
              </span>
            </button>
          ))}
        </div>
      ) : null}
      {status === "ready" ? (
        <footer>
          <span>{t($ => $.members.showingMembers, { countText: (formatDisplayNumber(members.length, locale) ?? t($ => $.members.unknown)), state: t($ => $.members.states[membershipState]) })}</span>
          <button type="button" className="secondary-action" disabled={!nextCursor || paginationPending} onClick={onLoadMore}>
            {paginationPending ? t($ => $.members.loading) : nextCursor ? t($ => $.members.nextPage) : t($ => $.members.endDirectory)}
          </button>
        </footer>
      ) : null}
      {candidate?.kind === "membership_create" ? <small className="local-identity-candidate-marker">{t($ => $.members.membershipCandidate)}</small> : null}
    </section>
  );
}

function MemberInspector({
  actorUserId,
  status,
  failure,
  summary,
  detail,
  membership,
  candidate,
  mutationPending,
  mutationsEnabled,
  onRetry,
  onReviewRevoke,
}: {
  actorUserId: string;
  status: LoadStatus;
  failure: LoadFailure | null;
  summary: LocalIdentityWorkspaceMemberSummary | null;
  detail: LocalIdentityWorkspaceMemberDetail | null;
  membership: LocalIdentityWorkspaceMembershipView | null;
  candidate: MutationCandidate | null;
  mutationPending: boolean;
  mutationsEnabled: boolean;
  onRetry: () => void;
  onReviewRevoke: (membership: LocalIdentityWorkspaceMembershipView) => void;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  if (!summary && status === "idle") {
    return <section className="local-identity-member-inspector"><StateNotice state="empty" title={t($ => $.members.selectMember)} /></section>;
  }
  return (
    <section className="local-identity-member-inspector" aria-labelledby="local-identity-inspector-title">
      {status === "loading" ? <StateNotice state="loading" title={t($ => $.members.loadingMember)} /> : null}
      {status === "denied" || status === "unavailable" || status === "failed" ? <StateFailure failure={failure} onRetry={onRetry} /> : null}
      {detail ? (
        <>
          <header>
            <span className="local-identity-member-avatar" aria-hidden="true">{initials(detail.displayName)}</span>
            <div><small>{t($ => $.members.selectedMember)}</small><h5 id="local-identity-inspector-title">{detail.displayName}</h5><p>{detail.userId}</p></div>
            <em className={membership?.effective ? "is-ready" : "is-revoked"}>{membership?.effective ? t($ => $.members.states.active) : t($ => $.members.states.revoked)}</em>
          </header>
          {membership ? (
            <dl className="local-identity-member-facts">
              <div><dt>{t($ => $.members.membership)}</dt><dd>{membership.membershipId} · v{membership.recordVersion}</dd></div>
              <div><dt>{t($ => $.members.window)}</dt><dd>{membership.effective ? t($ => $.members.states.active) : t($ => $.members.states[membership.lifecycleState])} · {membership.expiresAt ? <time dateTime={membership.expiresAt} title={membership.expiresAt}>{memberDate(t, membership.expiresAt, locale)}</time> : t($ => $.members.noExpiry)}</dd></div>
              <div><dt>{t($ => $.members.updated)}</dt><dd><time dateTime={membership.updatedAt} title={membership.updatedAt}>{memberDate(t, membership.updatedAt, locale)}</time></dd></div>
              <div><dt>{t($ => $.members.account)}</dt><dd>{t($ => $.members.states[detail.accountLifecycleState])} · v{detail.accountRecordVersion}</dd></div>
            </dl>
          ) : <StateNotice state="empty" title={t($ => $.members.noMatchingMembership)} />}
          <div className="local-identity-assignment-evidence">
            <header><strong>{t($ => $.members.roleAssignments)}</strong><small>{t($ => $.members.lifecycleRecords, { countText: (formatDisplayNumber(detail.roleAssignments.length, locale) ?? t($ => $.members.unknown)) })}</small></header>
            {detail.roleAssignments.length === 0 ? <p>{t($ => $.members.noRoles)}</p> : detail.roleAssignments.map((assignment) => (
              <article key={assignment.assignmentId}>
                <span><strong>{assignment.roleKey}</strong><small>{assignment.assignmentId} · v{assignment.recordVersion}</small></span>
                <em>{assignment.catalogDrift ? t($ => $.members.drift) : t($ => $.members.states[assignment.lifecycleState])}</em>
              </article>
            ))}
          </div>
          {detail.userId === actorUserId && membership?.effective ? (
            <p className="local-identity-protection"><span aria-hidden="true">△</span><strong>{t($ => $.members.currentAdminProtected)}</strong><small>{t($ => $.members.selfRevokeHelp)}</small></p>
          ) : detail.canManageLocalIdentity && membership?.effective ? (
            <p className="local-identity-protection"><span aria-hidden="true">△</span><strong>{t($ => $.members.serverAdminProtection)}</strong><small>{t($ => $.members.lastAdminHelp)}</small></p>
          ) : null}
          {membership?.effective ? (
            <div className="local-identity-inspector-action">
              <span><strong>{t($ => $.members.mutationBoundary)}</strong><small>{t($ => $.members.mutationRequirements, { version: membership.recordVersion })}</small></span>
              <button
                type="button"
                className="danger-action"
                disabled={!mutationsEnabled || mutationPending || detail.userId === actorUserId}
                onClick={() => onReviewRevoke(membership)}
              >{t($ => $.members.reviewMembershipRevoke)}</button>
            </div>
          ) : null}
          {candidate?.kind === "membership_revoke" ? <small className="local-identity-candidate-marker">{t($ => $.members.revokeCandidate)}</small> : null}
        </>
      ) : null}
    </section>
  );
}

function RoleAssignmentWorkspace({
  directoryStatus,
  selectedSummary,
  detailStatus,
  detailFailure,
  detail,
  membership,
  catalogStatus,
  catalogFailure,
  catalog,
  selectedRoleKey,
  roleExpiresAt,
  roleAlreadyAssigned,
  candidate,
  mutationPending,
  mutationsEnabled,
  onSelectRole,
  onRoleExpiresAt,
  onReviewAssign,
  onReviewRevoke,
  onRetryDetail,
  onRetryCatalog,
}: {
  directoryStatus: LoadStatus;
  selectedSummary: LocalIdentityWorkspaceMemberSummary | null;
  detailStatus: LoadStatus;
  detailFailure: LoadFailure | null;
  detail: LocalIdentityWorkspaceMemberDetail | null;
  membership: LocalIdentityWorkspaceMembershipView | null;
  catalogStatus: LoadStatus;
  catalogFailure: LoadFailure | null;
  catalog: LocalIdentityRoleCatalog | null;
  selectedRoleKey: string;
  roleExpiresAt: string;
  roleAlreadyAssigned: boolean;
  candidate: MutationCandidate | null;
  mutationPending: boolean;
  mutationsEnabled: boolean;
  onSelectRole: (roleKey: string) => void;
  onRoleExpiresAt: (value: string) => void;
  onReviewAssign: () => void;
  onReviewRevoke: (assignment: LocalIdentityWorkspaceRoleAssignmentView) => void;
  onRetryDetail: () => void;
  onRetryCatalog: () => void;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  return (
    <div className="local-identity-role-workspace">
      <section className="local-identity-role-member">
        <header><span>{t($ => $.members.selectedMember)}</span><a href="#admin-user-directory">{t($ => $.members.openUserDirectory)}</a></header>
        {directoryStatus === "loading" || detailStatus === "loading" ? <StateNotice state="loading" title={t($ => $.members.loadingContext)} /> : null}
        {!selectedSummary && directoryStatus === "empty" ? <StateNotice state="empty" title={t($ => $.members.noMemberForRole)} /> : null}
        {detailStatus === "denied" || detailStatus === "unavailable" || detailStatus === "failed" ? <StateFailure failure={detailFailure} onRetry={onRetryDetail} /> : null}
        {detail ? (
          <>
            <div className="local-identity-role-member-heading">
              <span className="local-identity-member-avatar" aria-hidden="true">{initials(detail.displayName)}</span>
              <div><strong>{detail.displayName}</strong><small>{detail.userId}</small></div>
              <em>{membership?.effective ? t($ => $.members.states.active) : t($ => $.members.states.revoked)}</em>
            </div>
            <dl className="local-identity-role-member-facts">
              <div><dt>{t($ => $.members.membership)}</dt><dd>{membership ? `${membership.membershipId} · v${membership.recordVersion}` : t($ => $.members.unavailable)}</dd></div>
              <div><dt>{t($ => $.members.assignments)}</dt><dd>{t($ => $.members.assignmentCounts, { activeText: (formatDisplayNumber(detail.roleAssignments.filter((item) => item.effective).length, locale) ?? t($ => $.members.unknown)), retainedText: (formatDisplayNumber(detail.roleAssignments.filter((item) => !item.effective).length, locale) ?? t($ => $.members.unknown)) })}</dd></div>
            </dl>
            <div className="local-identity-current-assignments">
              <header><strong>{t($ => $.members.currentAssignments)}</strong><small>{t($ => $.members.frozenGrants)}</small></header>
              {detail.roleAssignments.length === 0 ? <p>{t($ => $.members.noAssignmentEvidence)}</p> : detail.roleAssignments.map((assignment) => (
                <article key={assignment.assignmentId} className={assignment.catalogDrift ? "has-drift" : ""}>
                  <span><strong>{assignment.roleKey}</strong><small>{assignment.assignmentId} · v{assignment.recordVersion} · {assignment.roleCatalogVersion ?? t($ => $.members.legacyCatalog)}</small></span>
                  <em>{assignment.catalogDrift ? t($ => $.members.catalogDrift) : t($ => $.members.states[assignment.lifecycleState])}</em>
                  {assignment.effective && assignment.scope === "workspace" ? (
                    <button type="button" className="secondary-action" disabled={!mutationsEnabled || mutationPending} onClick={() => onReviewRevoke(assignment)}>{t($ => $.members.reviewRevoke)}</button>
                  ) : null}
                </article>
              ))}
            </div>
          </>
        ) : null}
      </section>

      <section className="local-identity-role-catalog" aria-labelledby="local-identity-role-catalog-title">
        <header><div><span>{t($ => $.members.roleCatalog)}</span><h5 id="local-identity-role-catalog-title">{t($ => $.members.chooseDefinition)}</h5></div><em>{catalog ? t($ => $.members.roleCount, { countText: (formatDisplayNumber(catalog.roles.length, locale) ?? t($ => $.members.unknown)) }) : t($ => $.members.states[catalogStatus])}</em></header>
        {catalogStatus === "loading" ? <StateNotice state="loading" title={t($ => $.members.loadingCatalog)} /> : null}
        {catalogStatus === "denied" || catalogStatus === "unavailable" || catalogStatus === "failed" ? <StateFailure failure={catalogFailure} onRetry={onRetryCatalog} /> : null}
        {catalog ? (
          <>
            <div className="local-identity-role-definitions">
              {catalog.roles.map((role, index) => (
                <button
                  type="button"
                  key={role.roleKey}
                  className={selectedRoleKey === role.roleKey ? "is-selected" : ""}
                  aria-pressed={selectedRoleKey === role.roleKey}
                  onClick={() => onSelectRole(role.roleKey)}
                >
                  <b>{String(index + 1).padStart(2, "0")}</b>
                  <span><strong>{role.roleKey}</strong><small>{identityRoleCopy(t, role.roleKey).summary}</small></span>
                  <em>{role.canManageLocalIdentity ? t($ => $.members.identity) : roleAlreadyAssigned && selectedRoleKey === role.roleKey ? t($ => $.members.assigned) : t($ => $.members.available)}</em>
                </button>
              ))}
            </div>
            <footer><span>{catalog.catalogVersion}</span><small>{shortDigest(catalog.definitionDigest)}{t($ => $.members.noClientGrants)}</small></footer>
            <div className="local-identity-role-candidate-editor">
              <label>{t($ => $.members.assignmentExpiresInput)}<input value={roleExpiresAt} onChange={(event) => onRoleExpiresAt(event.target.value)} placeholder={t($ => $.members.noExpiryPlaceholder)} autoComplete="off" /></label>
              <button type="button" disabled={!selectedRoleKey || roleAlreadyAssigned || !detail || !membership?.effective || !mutationsEnabled || mutationPending} onClick={onReviewAssign}>
                {roleAlreadyAssigned ? t($ => $.members.alreadyActive) : t($ => $.members.reviewRole)}
              </button>
            </div>
            {detail?.roleAssignments.some((assignment) => assignment.catalogDrift) ? (
              <p className="local-identity-catalog-drift"><span aria-hidden="true">!</span><strong>{t($ => $.members.driftEvidence)}</strong><small>{t($ => $.members.driftHelp)}</small></p>
            ) : null}
          </>
        ) : null}
        {candidate?.kind === "role_assign" || candidate?.kind === "role_revoke" ? <small className="local-identity-candidate-marker">{t($ => $.members.assignmentCandidate)}</small> : null}
      </section>
    </div>
  );
}

function MutationConfirmation({
  candidate,
  tenantRef,
  workspaceId,
  pending,
  recentAuthentication,
  onCancel,
  onConfirm,
}: {
  candidate: MutationCandidate;
  tenantRef: string;
  workspaceId: string;
  pending: boolean;
  recentAuthentication: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  const destructive = candidate.kind === "membership_revoke" || candidate.kind === "role_revoke";
  return (
    <aside className={`local-identity-mutation-confirmation ${destructive ? "is-destructive" : ""}`} aria-label={t($ => $.members.mutationConfirmation)}>
      <header><div><span>{t($ => $.members.explicitConfirmation)}</span><h5>{candidateTitle(t, candidate)}</h5></div><em>{recentAuthentication ? t($ => $.members.recentAuth) : t($ => $.members.reauthRequired)}</em></header>
      <dl>
        <div><dt>{t($ => $.members.tenantWorkspace)}</dt><dd>{tenantRef} / {workspaceId}</dd></div>
        <div><dt>{t($ => $.members.exactUser)}</dt><dd>{candidate.userId}</dd></div>
        {candidate.kind === "membership_create" ? <div><dt>{t($ => $.members.expiration)}</dt><dd>{candidate.expiresAt ?? t($ => $.members.none)}</dd></div> : null}
        {candidate.kind === "membership_revoke" ? <><div><dt>{t($ => $.members.expectedMembership)}</dt><dd>{candidate.membershipId} · v{candidate.expectedRecordVersion}</dd></div><div><dt>{t($ => $.members.atomicImpact)}</dt><dd>{t($ => $.members.atomicAssignments, { countText: (formatDisplayNumber(candidate.activeAssignmentCount, locale) ?? t($ => $.members.unknown)) })}</dd></div></> : null}
        {candidate.kind === "role_assign" ? <><div><dt>{t($ => $.members.expectedCatalog)}</dt><dd>{candidate.role.catalogVersion}</dd></div><div><dt>{t($ => $.members.definitionDigest)}</dt><dd>{shortDigest(candidate.role.definitionDigest)}</dd></div></> : null}
        {candidate.kind === "role_revoke" ? <div><dt>{t($ => $.members.expectedAssignment)}</dt><dd>{candidate.assignmentId} · v{candidate.expectedRecordVersion}</dd></div> : null}
      </dl>
      <p>{t($ => $.members.serverRecheck)}</p>
      <div><button type="button" className="secondary-action" onClick={onCancel} disabled={pending}>{t($ => $.members.cancel)}</button><button type="button" className={destructive ? "danger-action" : ""} onClick={onConfirm} disabled={pending || !recentAuthentication}>{pending ? t($ => $.members.applying) : t($ => $.members.confirmMutation)}</button></div>
    </aside>
  );
}

function OperationBanner({
  operation,
  onReloadDetail,
  onReloadCatalog,
  onRetryAll,
}: {
  operation: OperationNotice;
  onReloadDetail: () => void;
  onReloadCatalog: () => void;
  onRetryAll: () => void;
}) {
  const { t } = useTranslation("identity");
  const { locale } = useLocalePreference();
  let message = "";
  if (operation.kind === "success") message = memberSuccessMessage(t, operation.success, locale);
  else if (operation.kind !== "idle") {
    const input = operation.input;
    message = input ? t($ => $.members.inputErrors[input]) : memberFailureCopy(t, operation).message;
  }
  const action = operation.kind === "catalog_drift"
    ? { label: t($ => $.members.reloadCatalog), run: onReloadCatalog }
    : operation.kind === "stale_conflict" || operation.kind === "last_admin"
    ? { label: t($ => $.members.reloadMember), run: onReloadDetail }
    : operation.kind === "recent_authentication" || operation.kind === "denied"
    ? { label: t($ => $.members.refreshSession), run: onRetryAll }
    : operation.kind === "unavailable" || operation.kind === "invalid_response"
    ? { label: t($ => $.members.reloadOwners), run: onRetryAll }
    : null;
  return (
    <div className={`local-identity-operation is-${operation.kind}`} role={operation.kind === "success" ? "status" : "alert"}>
      <span aria-hidden="true">{operation.kind === "success" ? "✓" : operation.kind === "last_admin" ? "△" : "!"}</span>
      <div><strong>{operationLabel(t, operation.kind)}</strong><p>{message}</p><small>{operation.code}</small></div>
      {action ? <button type="button" className="secondary-action" onClick={action.run}>{action.label}</button> : null}
    </div>
  );
}

function StateNotice({ state, title }: { state: "loading" | "empty"; title: string }) {
  const { t } = useTranslation("identity");
  return <div className={`local-identity-state-notice is-${state}`}><span aria-hidden="true">{state === "loading" ? "…" : "∅"}</span><div><strong>{title}</strong><p>{state === "loading" ? t($ => $.members.readingHelp) : t($ => $.members.emptyHelp)}</p></div></div>;
}

function StateFailure({ failure, onRetry }: { failure: LoadFailure | null; onRetry: () => void }) {
  const { t } = useTranslation("identity");
  return (
    <div className="local-identity-state-failure" role="alert">
      <span aria-hidden="true">!</span>
      <div><strong>{failure ? operationLabel(t, failure.kind) : t($ => $.members.failedClosed)}</strong><p>{failure ? memberFailureCopy(t, failure).message : t($ => $.members.verifyFailed)}</p><small>{failure?.code ?? "local_identity_admin_unavailable"}</small></div>
      <button type="button" className="secondary-action" onClick={onRetry}>{t($ => $.members.retryRead)}</button>
    </div>
  );
}

function DisabledLocalIdentityAdministration({ surface }: { surface: "user" | "role" }) {
  const { t } = useTranslation("identity");
  return (
    <section className="admin-control-owner-surface local-identity-administration" aria-labelledby="local-identity-administration-title">
      <header><div><p className="eyebrow">{surface === "user" ? t($ => $.members.user) : t($ => $.members.role)}{t($ => $.members.boundarySuffix)}</p><h4 id="local-identity-administration-title">{t($ => $.members.disabled)}</h4></div><span className="admin-control-status is-blocked">{t($ => $.members.notConnected)}</span></header>
      <StateNotice state="empty" title={t($ => $.members.noIdentityFacts)} />
      <p className="local-identity-administration-boundary"><span aria-hidden="true">!</span>{t($ => $.members.offlineBoundary)}</p>
    </section>
  );
}

function localIdentityAdministrationScopeKey(
  profile: LocalIdentityAccountProfile,
  tenantRef: string,
  workspaceId: string,
): string {
  const memberships = profile.workspaceMemberships
    .filter((membership) => membership.tenantRef === tenantRef && membership.workspaceId === workspaceId)
    .map((membership) => `${membership.membershipId}:${membership.recordVersion}:${membership.lifecycleState}`)
    .sort()
    .join(",");
  const assignments = profile.roleAssignments
    .filter((assignment) => assignment.tenantRef === tenantRef && (!assignment.workspaceId || assignment.workspaceId === workspaceId))
    .map((assignment) => `${assignment.assignmentId}:${assignment.recordVersion}:${assignment.lifecycleState}`)
    .sort()
    .join(",");
  return `${profile.session.sessionId}:${profile.capabilities.recentAuthentication}:${tenantRef}:${workspaceId}:${memberships}:${assignments}`;
}

function effectiveMembership(
  detail: LocalIdentityWorkspaceMemberDetail,
  summary: LocalIdentityWorkspaceMemberSummary | null,
): LocalIdentityWorkspaceMembershipView | null {
  return detail.memberships.find((membership) => membership.membershipId === summary?.membershipId) ??
    detail.memberships.find((membership) => membership.effective) ?? detail.memberships[0] ?? null;
}

function loadStatusForFailure(failure: LoadFailure): LoadStatus {
  if (failure.kind === "denied") return "denied";
  if (failure.kind === "unavailable") return "unavailable";
  return "failed";
}

function candidateTitle(t: TFunction<"identity">, candidate: MutationCandidate): string {
  if (candidate.kind === "membership_create") return t($ => $.members.addUserTitle, { userId: candidate.userId });
  if (candidate.kind === "membership_revoke") return t($ => $.members.revokeTitle, { reference: candidate.membershipId });
  if (candidate.kind === "role_assign") return `${candidate.role.roleKey} → ${candidate.userId}`;
  return t($ => $.members.revokeTitle, { reference: candidate.roleKey });
}

function operationLabel(t: TFunction<"identity">, kind: OperationNotice["kind"]): string {
  if (kind === "success") return t($ => $.members.mutationSucceeded);
  if (kind === "idle") return "";
  return t($ => $.members.failures[kind].title);
}

function initials(displayName: string): string {
  return displayName.trim().split(/\s+/u).slice(0, 2).map((part) => part.slice(0, 1).toUpperCase()).join("") || "?";
}

function shortDigest(value: string): string {
  return value.length > 20 ? `${value.slice(0, 12)}…${value.slice(-6)}` : value;
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}
