package httpapi

import (
	"context"
	"crypto/sha256"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"
)

type localIdentityRepository interface {
	CreateAccount(context.Context, UserAccount, LocalCredential) error
	CreateAccountAndWebSession(context.Context, UserAccount, LocalCredential, WebSession) error
	CreateOIDCAccountAndWebSession(context.Context, UserAccount, ExternalIdentityBinding, WebSession) error
	ReadAccount(context.Context, string) (UserAccount, error)
	ReadAccountAccessProfile(context.Context, string) (LocalIdentityAccountAccessProfile, error)
	FindAccountByLoginIdentifier(context.Context, string) (UserAccount, error)
	DisableAccount(context.Context, string, int, time.Time, string) (UserAccount, error)
	ReadActiveCredential(context.Context, string) (LocalCredential, error)
	ReplaceCredential(context.Context, string, string, int, LocalCredential) error
	BindExternalIdentity(context.Context, ExternalIdentityBinding) error
	ResolveExternalIdentity(context.Context, string, string) (ExternalIdentityBinding, error)
	RevokeExternalIdentity(context.Context, string, int, time.Time, string) (ExternalIdentityBinding, error)
	CreateOIDCAuthorizationTransaction(context.Context, LocalIdentityOIDCAuthorizationTransaction) error
	ConsumeOIDCAuthorizationTransaction(context.Context, [sha256.Size]byte, time.Time) (LocalIdentityOIDCAuthorizationTransaction, error)
	CreateWebSession(context.Context, WebSession) error
	ReadWebSession(context.Context, string, time.Time) (WebSession, UserAccount, error)
	ResolveWebSession(context.Context, [sha256.Size]byte, time.Time) (WebSession, UserAccount, error)
	RevokeWebSession(context.Context, string, int, time.Time, string) (WebSession, error)
	CreateRoleAssignment(context.Context, LocalRoleAssignment) error
	RevokeRoleAssignment(context.Context, string, int, time.Time, string) (LocalRoleAssignment, error)
	CreateWorkspaceMembership(context.Context, WorkspaceMembership) error
	RevokeWorkspaceMembership(context.Context, string, int, time.Time, string) (WorkspaceMembership, error)
	AuthorizeWorkspace(context.Context, string, string, string, []string, time.Time) (LocalWorkspaceAuthorization, error)
}

func validLocalAccountAndSessionRegistration(account UserAccount, credential LocalCredential, session WebSession) bool {
	return validUserAccount(account) && validLocalCredential(credential) && validWebSession(session) &&
		account.LifecycleState == localIdentityStateActive && credential.LifecycleState == localIdentityStateActive &&
		session.LifecycleState == localIdentityStateActive && account.UserID == credential.UserID &&
		account.UserID == session.UserID && account.CreatedAt.Equal(credential.CreatedAt) &&
		account.CreatedAt.Equal(session.CreatedAt)
}

type memoryLocalIdentityRepository struct {
	mu sync.RWMutex

	accounts                      map[string]UserAccount
	accountByLoginIdentifier      map[string]string
	credentials                   map[string]LocalCredential
	activeCredentialByUser        map[string]string
	externalBindings              map[string]ExternalIdentityBinding
	activeBindingByIssuerSubject  map[string]string
	oidcAuthorizationTransactions map[string]LocalIdentityOIDCAuthorizationTransaction
	oidcTransactionByStateDigest  map[string]string
	sessions                      map[string]WebSession
	sessionByCredentialDigest     map[string]string
	roleAssignments               map[string]LocalRoleAssignment
	activeRoleByScope             map[string]string
	memberships                   map[string]WorkspaceMembership
	activeMembershipByScope       map[string]string
}

func newMemoryLocalIdentityRepository() *memoryLocalIdentityRepository {
	return &memoryLocalIdentityRepository{
		accounts: make(map[string]UserAccount), accountByLoginIdentifier: make(map[string]string),
		credentials: make(map[string]LocalCredential), activeCredentialByUser: make(map[string]string),
		externalBindings: make(map[string]ExternalIdentityBinding), activeBindingByIssuerSubject: make(map[string]string),
		oidcAuthorizationTransactions: make(map[string]LocalIdentityOIDCAuthorizationTransaction),
		oidcTransactionByStateDigest:  make(map[string]string),
		sessions:                      make(map[string]WebSession), sessionByCredentialDigest: make(map[string]string),
		roleAssignments: make(map[string]LocalRoleAssignment), activeRoleByScope: make(map[string]string),
		memberships: make(map[string]WorkspaceMembership), activeMembershipByScope: make(map[string]string),
	}
}

func (repository *memoryLocalIdentityRepository) CreateOIDCAccountAndWebSession(
	_ context.Context,
	account UserAccount,
	binding ExternalIdentityBinding,
	session WebSession,
) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalOIDCAccountAndSessionRegistration(account, binding, session) {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	digestKey := string(session.credentialDigest)
	bindingKey := externalIdentityKey(binding.Issuer, binding.Subject)
	if _, exists := repository.accounts[account.UserID]; exists ||
		repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] != "" ||
		repository.externalBindings[binding.BindingID].BindingID != "" || repository.activeBindingByIssuerSubject[bindingKey] != "" ||
		repository.sessions[session.SessionID].SessionID != "" || repository.sessionByCredentialDigest[digestKey] != "" {
		return errLocalIdentityExternalConflict
	}
	repository.accounts[account.UserID] = cloneUserAccount(account)
	repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = account.UserID
	repository.externalBindings[binding.BindingID] = cloneExternalIdentityBinding(binding)
	repository.activeBindingByIssuerSubject[bindingKey] = binding.BindingID
	repository.sessions[session.SessionID] = cloneWebSession(session)
	repository.sessionByCredentialDigest[digestKey] = session.SessionID
	return nil
}

func (repository *memoryLocalIdentityRepository) CreateAccount(_ context.Context, account UserAccount, credential LocalCredential) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validUserAccount(account) || !validLocalCredential(credential) || account.LifecycleState != localIdentityStateActive ||
		credential.LifecycleState != localIdentityStateActive || account.UserID != credential.UserID ||
		!account.CreatedAt.Equal(credential.CreatedAt) {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, exists := repository.accounts[account.UserID]; exists || repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] != "" ||
		repository.credentials[credential.CredentialID].CredentialID != "" || repository.activeCredentialByUser[account.UserID] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.accounts[account.UserID] = cloneUserAccount(account)
	repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = account.UserID
	repository.credentials[credential.CredentialID] = cloneLocalCredential(credential)
	repository.activeCredentialByUser[account.UserID] = credential.CredentialID
	return nil
}

func (repository *memoryLocalIdentityRepository) CreateAccountAndWebSession(
	_ context.Context,
	account UserAccount,
	credential LocalCredential,
	session WebSession,
) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalAccountAndSessionRegistration(account, credential, session) {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	digestKey := string(session.credentialDigest)
	if _, exists := repository.accounts[account.UserID]; exists ||
		repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] != "" ||
		repository.credentials[credential.CredentialID].CredentialID != "" ||
		repository.activeCredentialByUser[account.UserID] != "" ||
		repository.sessions[session.SessionID].SessionID != "" || repository.sessionByCredentialDigest[digestKey] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.accounts[account.UserID] = cloneUserAccount(account)
	repository.accountByLoginIdentifier[account.NormalizedLoginIdentifier] = account.UserID
	repository.credentials[credential.CredentialID] = cloneLocalCredential(credential)
	repository.activeCredentialByUser[account.UserID] = credential.CredentialID
	repository.sessions[session.SessionID] = cloneWebSession(session)
	repository.sessionByCredentialDigest[digestKey] = session.SessionID
	return nil
}

func (repository *memoryLocalIdentityRepository) ReadAccount(_ context.Context, userID string) (UserAccount, error) {
	if repository == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	account, exists := repository.accounts[strings.TrimSpace(userID)]
	if !exists {
		return UserAccount{}, errLocalIdentityNotFound
	}
	return cloneUserAccount(account), nil
}

func (repository *memoryLocalIdentityRepository) FindAccountByLoginIdentifier(_ context.Context, rawIdentifier string) (UserAccount, error) {
	if repository == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	identifier, err := NormalizeLocalLoginIdentifier(rawIdentifier)
	if err != nil {
		return UserAccount{}, err
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	userID := repository.accountByLoginIdentifier[identifier]
	account, exists := repository.accounts[userID]
	if !exists {
		return UserAccount{}, errLocalIdentityNotFound
	}
	return cloneUserAccount(account), nil
}

func (repository *memoryLocalIdentityRepository) DisableAccount(_ context.Context, userID string, expectedVersion int, disabledAt time.Time, auditRef string) (UserAccount, error) {
	if repository == nil {
		return UserAccount{}, errLocalIdentityStoreUnavailable
	}
	disabledAt = disabledAt.UTC()
	if !localUserIDPattern.MatchString(userID) || expectedVersion < 1 || disabledAt.IsZero() || !validAuditRef(auditRef) {
		return UserAccount{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[userID]
	if !exists {
		return UserAccount{}, errLocalIdentityNotFound
	}
	if account.RecordVersion != expectedVersion || account.LifecycleState != localIdentityStateActive {
		return UserAccount{}, errLocalIdentityVersionConflict
	}
	account.LifecycleState = localIdentityStateDisabled
	account.RecordVersion++
	account.UpdatedAt = disabledAt
	account.DisabledAt = timePointer(disabledAt)
	account.AuditRef = strings.TrimSpace(auditRef)
	if !validUserAccount(account) {
		return UserAccount{}, errLocalIdentityContractMismatch
	}
	repository.accounts[userID] = account
	for sessionID, session := range repository.sessions {
		if session.UserID != userID || session.LifecycleState != localIdentityStateActive {
			continue
		}
		session.LifecycleState = localIdentityStateRevoked
		session.RecordVersion++
		session.UpdatedAt = disabledAt
		session.RevokedAt = timePointer(disabledAt)
		session.AuditRef = strings.TrimSpace(auditRef)
		repository.sessions[sessionID] = session
	}
	return cloneUserAccount(account), nil
}

func (repository *memoryLocalIdentityRepository) ReadActiveCredential(_ context.Context, userID string) (LocalCredential, error) {
	if repository == nil {
		return LocalCredential{}, errLocalIdentityStoreUnavailable
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	credential, exists := repository.credentials[repository.activeCredentialByUser[userID]]
	if !exists || credential.LifecycleState != localIdentityStateActive {
		return LocalCredential{}, errLocalIdentityNotFound
	}
	return cloneLocalCredential(credential), nil
}

func (repository *memoryLocalIdentityRepository) ReplaceCredential(_ context.Context, userID string, expectedCredentialID string, expectedVersion int, replacement LocalCredential) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalCredential(replacement) || replacement.UserID != userID || replacement.LifecycleState != localIdentityStateActive ||
		!localCredentialIDPattern.MatchString(expectedCredentialID) || expectedVersion < 1 {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, accountExists := repository.accounts[userID]
	currentID := repository.activeCredentialByUser[userID]
	current, credentialExists := repository.credentials[currentID]
	if !accountExists || !credentialExists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	if currentID != expectedCredentialID || current.RecordVersion != expectedVersion || current.LifecycleState != localIdentityStateActive {
		return errLocalIdentityVersionConflict
	}
	if _, exists := repository.credentials[replacement.CredentialID]; exists || !replacement.CreatedAt.After(current.CreatedAt) {
		return errLocalIdentityIdentifierConflict
	}
	current.LifecycleState = localIdentityStateSuperseded
	current.RecordVersion++
	current.UpdatedAt = replacement.CreatedAt
	current.AuditRef = replacement.AuditRef
	repository.credentials[currentID] = current
	repository.credentials[replacement.CredentialID] = cloneLocalCredential(replacement)
	repository.activeCredentialByUser[userID] = replacement.CredentialID
	return nil
}

func (repository *memoryLocalIdentityRepository) BindExternalIdentity(_ context.Context, binding ExternalIdentityBinding) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validExternalIdentityBinding(binding) || binding.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[binding.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	key := externalIdentityKey(binding.Issuer, binding.Subject)
	if repository.activeBindingByIssuerSubject[key] != "" || repository.externalBindings[binding.BindingID].BindingID != "" {
		return errLocalIdentityExternalConflict
	}
	repository.externalBindings[binding.BindingID] = cloneExternalIdentityBinding(binding)
	repository.activeBindingByIssuerSubject[key] = binding.BindingID
	return nil
}

func (repository *memoryLocalIdentityRepository) ResolveExternalIdentity(_ context.Context, rawIssuer string, rawSubject string) (ExternalIdentityBinding, error) {
	if repository == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	issuer, err := NormalizeExternalIssuer(rawIssuer)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	subject, err := NormalizeExternalSubject(rawSubject)
	if err != nil {
		return ExternalIdentityBinding{}, err
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	binding, exists := repository.externalBindings[repository.activeBindingByIssuerSubject[externalIdentityKey(issuer, subject)]]
	if !exists || binding.LifecycleState != localIdentityStateActive {
		return ExternalIdentityBinding{}, errLocalIdentityNotFound
	}
	return cloneExternalIdentityBinding(binding), nil
}

func (repository *memoryLocalIdentityRepository) RevokeExternalIdentity(_ context.Context, bindingID string, expectedVersion int, revokedAt time.Time, auditRef string) (ExternalIdentityBinding, error) {
	if repository == nil {
		return ExternalIdentityBinding{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localBindingIDPattern.MatchString(bindingID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return ExternalIdentityBinding{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	binding, exists := repository.externalBindings[bindingID]
	if !exists {
		return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
	}
	if binding.RecordVersion != expectedVersion || binding.LifecycleState != localIdentityStateActive {
		return ExternalIdentityBinding{}, errLocalIdentityVersionConflict
	}
	if _, hasCredential := repository.activeCredentialByUser[binding.UserID]; !hasCredential {
		activeBindings := 0
		for _, candidate := range repository.externalBindings {
			if candidate.UserID == binding.UserID && candidate.LifecycleState == localIdentityStateActive {
				activeBindings++
			}
		}
		if activeBindings <= 1 {
			return ExternalIdentityBinding{}, errLocalIdentityLastLoginMethodRemoval
		}
	}
	binding.LifecycleState = localIdentityStateRevoked
	binding.RecordVersion++
	binding.UpdatedAt = revokedAt
	binding.RevokedAt = timePointer(revokedAt)
	binding.AuditRef = strings.TrimSpace(auditRef)
	delete(repository.activeBindingByIssuerSubject, externalIdentityKey(binding.Issuer, binding.Subject))
	repository.externalBindings[bindingID] = binding
	return cloneExternalIdentityBinding(binding), nil
}

func (repository *memoryLocalIdentityRepository) CreateOIDCAuthorizationTransaction(
	_ context.Context,
	transaction LocalIdentityOIDCAuthorizationTransaction,
) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validLocalIdentityOIDCAuthorizationTransaction(transaction) || transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	stateKey := string(transaction.stateDigest)
	if _, exists := repository.oidcAuthorizationTransactions[transaction.TransactionID]; exists ||
		repository.oidcTransactionByStateDigest[stateKey] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.oidcAuthorizationTransactions[transaction.TransactionID] = cloneLocalIdentityOIDCAuthorizationTransaction(transaction)
	repository.oidcTransactionByStateDigest[stateKey] = transaction.TransactionID
	return nil
}

func (repository *memoryLocalIdentityRepository) ConsumeOIDCAuthorizationTransaction(
	_ context.Context,
	stateDigest [sha256.Size]byte,
	now time.Time,
) (LocalIdentityOIDCAuthorizationTransaction, error) {
	if repository == nil {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	if now.IsZero() {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	transactionID := repository.oidcTransactionByStateDigest[string(stateDigest[:])]
	transaction, exists := repository.oidcAuthorizationTransactions[transactionID]
	if !exists || transaction.LifecycleState != localIdentityOIDCTransactionPending {
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateInvalid
	}
	consumed := cloneLocalIdentityOIDCAuthorizationTransaction(transaction)
	transaction.RecordVersion++
	transaction.UpdatedAt = now
	transaction.ConsumedAt = timePointer(now)
	transaction.codeVerifier = ""
	delete(repository.oidcTransactionByStateDigest, string(stateDigest[:]))
	if !transaction.ExpiresAt.After(now) {
		transaction.LifecycleState = localIdentityOIDCTransactionExpired
		repository.oidcAuthorizationTransactions[transactionID] = transaction
		return LocalIdentityOIDCAuthorizationTransaction{}, errLocalIdentityOIDCStateExpired
	}
	transaction.LifecycleState = localIdentityOIDCTransactionConsumed
	repository.oidcAuthorizationTransactions[transactionID] = transaction
	consumed.LifecycleState = localIdentityOIDCTransactionConsumed
	consumed.RecordVersion = transaction.RecordVersion
	consumed.UpdatedAt = now
	consumed.ConsumedAt = timePointer(now)
	return consumed, nil
}

func (repository *memoryLocalIdentityRepository) CreateWebSession(_ context.Context, session WebSession) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWebSession(session) || session.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[session.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	digestKey := string(session.credentialDigest)
	if repository.sessions[session.SessionID].SessionID != "" || repository.sessionByCredentialDigest[digestKey] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.sessions[session.SessionID] = cloneWebSession(session)
	repository.sessionByCredentialDigest[digestKey] = session.SessionID
	return nil
}

func (repository *memoryLocalIdentityRepository) ResolveWebSession(_ context.Context, digest [sha256.Size]byte, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	session, exists := repository.sessions[repository.sessionByCredentialDigest[string(digest[:])]]
	if !exists || session.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionInvalid
	}
	if !session.ExpiresAt.After(now) {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionExpired
	}
	account, exists := repository.accounts[session.UserID]
	if !exists || account.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentityAccountInactive
	}
	return cloneWebSession(session), cloneUserAccount(account), nil
}

func (repository *memoryLocalIdentityRepository) ReadWebSession(_ context.Context, sessionID string, now time.Time) (WebSession, UserAccount, error) {
	if repository == nil {
		return WebSession{}, UserAccount{}, errLocalIdentityStoreUnavailable
	}
	now = now.UTC()
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	session, exists := repository.sessions[strings.TrimSpace(sessionID)]
	if !exists || session.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionInvalid
	}
	if !session.ExpiresAt.After(now) {
		return WebSession{}, UserAccount{}, errLocalIdentitySessionExpired
	}
	account, exists := repository.accounts[session.UserID]
	if !exists || account.LifecycleState != localIdentityStateActive {
		return WebSession{}, UserAccount{}, errLocalIdentityAccountInactive
	}
	return cloneWebSession(session), cloneUserAccount(account), nil
}

func (repository *memoryLocalIdentityRepository) RevokeWebSession(_ context.Context, sessionID string, expectedVersion int, revokedAt time.Time, auditRef string) (WebSession, error) {
	if repository == nil {
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localSessionIDPattern.MatchString(sessionID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WebSession{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	session, exists := repository.sessions[sessionID]
	if !exists {
		return WebSession{}, errLocalIdentityVersionConflict
	}
	if session.RecordVersion != expectedVersion || session.LifecycleState != localIdentityStateActive {
		return WebSession{}, errLocalIdentityVersionConflict
	}
	session.LifecycleState = localIdentityStateRevoked
	session.RecordVersion++
	session.UpdatedAt = revokedAt
	session.RevokedAt = timePointer(revokedAt)
	session.AuditRef = strings.TrimSpace(auditRef)
	repository.sessions[sessionID] = session
	return cloneWebSession(session), nil
}

func (repository *memoryLocalIdentityRepository) CreateRoleAssignment(_ context.Context, assignment LocalRoleAssignment) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	grants, ok := normalizedPermissionGrants(assignment.PermissionGrants)
	assignment.PermissionGrants = grants
	if !ok || assignment.RoleCatalogVersion != "" || assignment.RoleDefinitionDigest != "" ||
		localIdentityContainsManagementPermission(grants) ||
		!validLocalRoleAssignment(assignment) || assignment.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[assignment.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	key := localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey)
	if repository.roleAssignments[assignment.AssignmentID].AssignmentID != "" || repository.activeRoleByScope[key] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.roleAssignments[assignment.AssignmentID] = cloneLocalRoleAssignment(assignment)
	repository.activeRoleByScope[key] = assignment.AssignmentID
	return nil
}

func (repository *memoryLocalIdentityRepository) RevokeRoleAssignment(_ context.Context, assignmentID string, expectedVersion int, revokedAt time.Time, auditRef string) (LocalRoleAssignment, error) {
	if repository == nil {
		return LocalRoleAssignment{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localRoleAssignmentIDPattern.MatchString(assignmentID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	assignment, exists := repository.roleAssignments[assignmentID]
	if !exists {
		return LocalRoleAssignment{}, errLocalIdentityVersionConflict
	}
	if assignment.RecordVersion != expectedVersion || assignment.LifecycleState != localIdentityStateActive {
		return LocalRoleAssignment{}, errLocalIdentityVersionConflict
	}
	if assignment.RoleCatalogVersion != "" || assignment.RoleDefinitionDigest != "" {
		return LocalRoleAssignment{}, errLocalIdentityContractMismatch
	}
	assignment.LifecycleState = localIdentityStateRevoked
	assignment.RecordVersion++
	assignment.UpdatedAt = revokedAt
	assignment.RevokedAt = timePointer(revokedAt)
	assignment.AuditRef = strings.TrimSpace(auditRef)
	delete(repository.activeRoleByScope, localRoleScopeKey(assignment.UserID, assignment.TenantRef, assignment.WorkspaceID, assignment.RoleKey))
	repository.roleAssignments[assignmentID] = assignment
	return cloneLocalRoleAssignment(assignment), nil
}

func (repository *memoryLocalIdentityRepository) CreateWorkspaceMembership(_ context.Context, membership WorkspaceMembership) error {
	if repository == nil {
		return errLocalIdentityStoreUnavailable
	}
	if !validWorkspaceMembership(membership) || membership.LifecycleState != localIdentityStateActive {
		return errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	account, exists := repository.accounts[membership.UserID]
	if !exists {
		return errLocalIdentityNotFound
	}
	if account.LifecycleState != localIdentityStateActive {
		return errLocalIdentityAccountInactive
	}
	key := localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID)
	if repository.memberships[membership.MembershipID].MembershipID != "" || repository.activeMembershipByScope[key] != "" {
		return errLocalIdentityIdentifierConflict
	}
	repository.memberships[membership.MembershipID] = cloneWorkspaceMembership(membership)
	repository.activeMembershipByScope[key] = membership.MembershipID
	return nil
}

func (repository *memoryLocalIdentityRepository) RevokeWorkspaceMembership(_ context.Context, membershipID string, expectedVersion int, revokedAt time.Time, auditRef string) (WorkspaceMembership, error) {
	if repository == nil {
		return WorkspaceMembership{}, errLocalIdentityStoreUnavailable
	}
	revokedAt = revokedAt.UTC()
	if !localMembershipIDPattern.MatchString(membershipID) || expectedVersion < 1 || revokedAt.IsZero() || !validAuditRef(auditRef) {
		return WorkspaceMembership{}, errLocalIdentityContractMismatch
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	membership, exists := repository.memberships[membershipID]
	if !exists {
		return WorkspaceMembership{}, errLocalIdentityVersionConflict
	}
	if membership.RecordVersion != expectedVersion || membership.LifecycleState != localIdentityStateActive {
		return WorkspaceMembership{}, errLocalIdentityVersionConflict
	}
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID == membership.UserID && assignment.TenantRef == membership.TenantRef &&
			assignment.WorkspaceID == membership.WorkspaceID && assignment.LifecycleState == localIdentityStateActive {
			return WorkspaceMembership{}, errLocalIdentityContractMismatch
		}
	}
	membership.LifecycleState = localIdentityStateRevoked
	membership.RecordVersion++
	membership.UpdatedAt = revokedAt
	membership.RevokedAt = timePointer(revokedAt)
	membership.AuditRef = strings.TrimSpace(auditRef)
	delete(repository.activeMembershipByScope, localMembershipScopeKey(membership.UserID, membership.TenantRef, membership.WorkspaceID))
	repository.memberships[membershipID] = membership
	return cloneWorkspaceMembership(membership), nil
}

func (repository *memoryLocalIdentityRepository) AuthorizeWorkspace(_ context.Context, userID string, tenantRef string, workspaceID string, required []string, now time.Time) (LocalWorkspaceAuthorization, error) {
	if repository == nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	required, ok := normalizedPermissionGrants(required)
	if !ok || !localUserIDPattern.MatchString(userID) || !validControlPlaneReadAuthReference(tenantRef, false) ||
		!validControlPlaneReadAuthReference(workspaceID, false) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityContractMismatch
	}
	now = now.UTC()
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	account, exists := repository.accounts[userID]
	if !exists || account.LifecycleState != localIdentityStateActive {
		return LocalWorkspaceAuthorization{}, errLocalIdentityAccountInactive
	}
	membership, exists := repository.memberships[repository.activeMembershipByScope[localMembershipScopeKey(userID, tenantRef, workspaceID)]]
	if !exists || membership.LifecycleState != localIdentityStateActive || membership.ExpiresAt != nil && !membership.ExpiresAt.After(now) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
	}
	assignments := make([]LocalRoleAssignment, 0)
	grantSet := make(map[string]struct{})
	for _, assignment := range repository.roleAssignments {
		if assignment.UserID != userID || assignment.TenantRef != tenantRef || assignment.LifecycleState != localIdentityStateActive ||
			assignment.WorkspaceID != "" && assignment.WorkspaceID != workspaceID || assignment.ExpiresAt != nil && !assignment.ExpiresAt.After(now) {
			continue
		}
		assignments = append(assignments, cloneLocalRoleAssignment(assignment))
		for _, grant := range assignment.PermissionGrants {
			grantSet[grant] = struct{}{}
		}
	}
	for _, permission := range required {
		if _, granted := grantSet[permission]; !granted {
			return LocalWorkspaceAuthorization{}, errLocalIdentityPermissionDenied
		}
	}
	grants := make([]string, 0, len(grantSet))
	for grant := range grantSet {
		grants = append(grants, grant)
	}
	slices.Sort(grants)
	slices.SortFunc(assignments, func(left, right LocalRoleAssignment) int {
		return strings.Compare(left.AssignmentID, right.AssignmentID)
	})
	return LocalWorkspaceAuthorization{
		Account: cloneUserAccount(account), Membership: cloneWorkspaceMembership(membership),
		RoleAssignments: assignments, PermissionGrants: grants,
	}, nil
}

func externalIdentityKey(issuer string, subject string) string {
	return issuer + "\x00" + subject
}

func localRoleScopeKey(userID string, tenantRef string, workspaceID string, roleKey string) string {
	return strings.Join([]string{userID, tenantRef, workspaceID, roleKey}, "\x00")
}

func localMembershipScopeKey(userID string, tenantRef string, workspaceID string) string {
	return strings.Join([]string{userID, tenantRef, workspaceID}, "\x00")
}

func cloneUserAccount(account UserAccount) UserAccount {
	account.DisabledAt = cloneTimePointer(account.DisabledAt)
	return account
}

func cloneLocalCredential(credential LocalCredential) LocalCredential {
	credential.salt = append([]byte(nil), credential.salt...)
	credential.derivedKey = append([]byte(nil), credential.derivedKey...)
	return credential
}

func cloneExternalIdentityBinding(binding ExternalIdentityBinding) ExternalIdentityBinding {
	binding.RevokedAt = cloneTimePointer(binding.RevokedAt)
	return binding
}

func cloneWebSession(session WebSession) WebSession {
	session.credentialDigest = append([]byte(nil), session.credentialDigest...)
	session.RevokedAt = cloneTimePointer(session.RevokedAt)
	return session
}

func cloneLocalRoleAssignment(assignment LocalRoleAssignment) LocalRoleAssignment {
	assignment.PermissionGrants = append([]string(nil), assignment.PermissionGrants...)
	assignment.ExpiresAt = cloneTimePointer(assignment.ExpiresAt)
	assignment.RevokedAt = cloneTimePointer(assignment.RevokedAt)
	return assignment
}

func cloneWorkspaceMembership(membership WorkspaceMembership) WorkspaceMembership {
	membership.ExpiresAt = cloneTimePointer(membership.ExpiresAt)
	membership.RevokedAt = cloneTimePointer(membership.RevokedAt)
	return membership
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func timePointer(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

func localIdentityRepositoryError(err error) string {
	switch {
	case errors.Is(err, errLocalIdentityStoreUnavailable):
		return LocalIdentityFailureStoreUnavailable
	case errors.Is(err, errLocalIdentityContractMismatch):
		return LocalIdentityFailureContractMismatch
	case errors.Is(err, errLocalIdentityIdentifierConflict):
		return LocalIdentityFailureIdentifierConflict
	case errors.Is(err, errLocalIdentityExternalConflict):
		return LocalIdentityFailureExternalConflict
	case errors.Is(err, errLocalIdentityVersionConflict):
		return LocalIdentityFailureVersionConflict
	case errors.Is(err, errLocalIdentityAccountInactive):
		return LocalIdentityFailureAccountInactive
	case errors.Is(err, errLocalIdentityCredentialInvalid):
		return LocalIdentityFailureCredentialInvalid
	case errors.Is(err, errLocalIdentitySessionExpired):
		return LocalIdentityFailureSessionExpired
	case errors.Is(err, errLocalIdentitySessionInvalid):
		return LocalIdentityFailureSessionInvalid
	case errors.Is(err, errLocalIdentityOIDCStateInvalid):
		return LocalIdentityFailureOIDCStateInvalid
	case errors.Is(err, errLocalIdentityOIDCStateExpired):
		return LocalIdentityFailureOIDCStateExpired
	case errors.Is(err, errLocalIdentityLastLoginMethodRemoval):
		return LocalIdentityFailureLastLoginMethodRemoval
	case errors.Is(err, errLocalIdentityMembershipDenied):
		return LocalIdentityFailureMembershipDenied
	case errors.Is(err, errLocalIdentityPermissionDenied):
		return LocalIdentityFailurePermissionDenied
	case errors.Is(err, errLocalIdentityAdminUnavailable):
		return LocalIdentityFailureAdminUnavailable
	case errors.Is(err, errLocalIdentityAdminScopeMismatch):
		return LocalIdentityFailureAdminScopeMismatch
	case errors.Is(err, errLocalIdentityMemberUnavailable):
		return LocalIdentityFailureMemberUnavailable
	case errors.Is(err, errLocalIdentityMemberCursorInvalid):
		return LocalIdentityFailureMemberCursorInvalid
	case errors.Is(err, errLocalIdentityRoleCatalogMismatch):
		return LocalIdentityFailureRoleCatalogMismatch
	case errors.Is(err, errLocalIdentityMembershipConflict):
		return LocalIdentityFailureMembershipConflict
	case errors.Is(err, errLocalIdentityRoleAssignmentConflict):
		return LocalIdentityFailureRoleAssignmentConflict
	case errors.Is(err, errLocalIdentitySelfMembershipRevoke):
		return LocalIdentityFailureSelfMembershipRevoke
	case errors.Is(err, errLocalIdentityLastAdminRemoval):
		return LocalIdentityFailureLastAdminRemoval
	case errors.Is(err, errLocalIdentityRecentAuthentication):
		return LocalIdentityFailureRecentAuthentication
	case errors.Is(err, errLocalIdentityAdminBootstrapDenied):
		return LocalIdentityFailureAdminBootstrapDenied
	case errors.Is(err, errLocalIdentitySessionCursorInvalid):
		return LocalIdentityFailureSessionCursorInvalid
	case errors.Is(err, errLocalIdentitySessionScopeDenied):
		return LocalIdentityFailureSessionScopeDenied
	case errors.Is(err, errLocalIdentitySessionVersionConflict):
		return LocalIdentityFailureSessionVersionConflict
	case errors.Is(err, errLocalIdentitySessionRecentAuthentication):
		return LocalIdentityFailureSessionRecentAuthentication
	case errors.Is(err, errLocalIdentitySessionBulkRevokeConflict):
		return LocalIdentityFailureSessionBulkRevokeConflict
	case errors.Is(err, errLocalIdentityCredentialUnavailable):
		return LocalIdentityFailureCredentialUnavailable
	case errors.Is(err, errLocalIdentityCredentialCurrentInvalid):
		return LocalIdentityFailureCredentialCurrentInvalid
	case errors.Is(err, errLocalIdentityCredentialPolicyRejected):
		return LocalIdentityFailureCredentialPolicyRejected
	case errors.Is(err, errLocalIdentityCredentialReuseDenied):
		return LocalIdentityFailureCredentialReuseDenied
	case errors.Is(err, errLocalIdentityCredentialRotationConflict):
		return LocalIdentityFailureCredentialRotationConflict
	case errors.Is(err, errLocalIdentitySelfServiceUnavailable):
		return LocalIdentityFailureSelfServiceUnavailable
	case errors.Is(err, errLocalIdentityNotFound):
		return LocalIdentityFailureNotFound
	default:
		return LocalIdentityFailureStoreUnavailable
	}
}
