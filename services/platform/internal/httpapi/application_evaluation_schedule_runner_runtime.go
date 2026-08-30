package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (server *Server) revalidateApplicationEvaluationScheduleOccurrence(
	requestContext context.Context,
	ctx ApplicationEvaluationContext,
	version ApplicationEvaluationScheduleVersion,
) string {
	if server == nil || server.localIdentityRepository == nil || ctx.ScheduleExecution == nil ||
		ctx.ActorRef != version.Authorization.DelegatedByUserRef ||
		ctx.ScheduleExecution.SystemActorRef != version.Authorization.SystemActorRef ||
		ctx.ScheduleExecution.DelegatedByUserRef != version.Authorization.DelegatedByUserRef ||
		ctx.ScheduleExecution.ScheduleID != version.ScheduleID || ctx.ScheduleExecution.ScheduleVersion != version.ScheduleVersion ||
		ctx.ScheduleExecution.ScheduleDigest != version.ScheduleDigest {
		return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
	}
	schedule, found, err := server.applicationEvaluationScheduleRepository.ReadSchedule(ctx, version.ScheduleID)
	if err != nil {
		return applicationEvaluationScheduleRunnerStoreFailure(err)
	}
	if !found || schedule.LifecycleState != applicationEvaluationScheduleStateActive ||
		schedule.LatestScheduleVersion != version.ScheduleVersion || schedule.LatestScheduleDigest != version.ScheduleDigest ||
		schedule.SystemActorRef != version.Authorization.SystemActorRef || schedule.DelegatedByUserRef != version.Authorization.DelegatedByUserRef {
		return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
	}
	delegatedUserID, validDelegatedUser := applicationEvaluationScheduleDelegatedUserID(version.Authorization.DelegatedByUserRef)
	if !validDelegatedUser {
		return ApplicationEvaluationScheduleFailureMembershipDenied
	}
	_, err = server.localIdentityRepository.AuthorizeWorkspace(
		requestContext,
		delegatedUserID,
		version.TenantRef,
		version.WorkspaceID,
		version.Authorization.RequiredPermissions,
		time.Now().UTC(),
	)
	if err != nil {
		switch {
		case errors.Is(err, errLocalIdentityAccountInactive), errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityPermissionDenied):
			return ApplicationEvaluationScheduleFailureMembershipDenied
		case errors.Is(err, errLocalIdentityContractMismatch):
			return ApplicationEvaluationScheduleFailureStoreContract
		default:
			return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
		}
	}
	return server.applicationEvaluationScheduleService().revalidateActivation(ctx, version)
}

func applicationEvaluationScheduleDelegatedUserID(actorRef string) (string, bool) {
	userID := strings.TrimPrefix(strings.TrimSpace(actorRef), "user:")
	return userID, userID != actorRef && localUserIDPattern.MatchString(userID)
}

func applicationEvaluationScheduleRunnerStoreFailure(err error) string {
	if errors.Is(err, errApplicationEvaluationScheduleStoreContract) {
		return ApplicationEvaluationScheduleFailureStoreContract
	}
	return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
}
