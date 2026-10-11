import "../../i18n/gatewayRouteEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { evidenceProjectionText, evidenceProjectionStatus } from "./evidenceProjectionCopy.ts";
import type {
  ModelGatewayProviderProfileEvidence,
  ModelGatewayRouteBindingEvidence,
  ModelGatewayRouteEvidenceStatus,
  ModelGatewayRouteEvidenceViewModel,
  ModelGatewayRuntimeGuardEvidence,
  ModelGatewaySelectionCaseEvidence,
} from "./modelGatewayRouteEvidence";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function ModelGatewayRouteEvidencePanel({
  detail,
}: {
  detail: ModelGatewayRouteEvidenceViewModel;
}) {
  const { t } = useTranslation("evaluation");
  return (
    <section
      className="surface-band model-gateway-route-evidence"
      id="model-gateway-route-evidence"
      aria-labelledby="model-gateway-route-evidence-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Model_Gateway)}</p>
          <h3 id="model-gateway-route-evidence-title">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Provider_profile_route_evidence)}</h3>
        </div>
        <StatusBadge tone={detail.canRenderRouteEvidenceDetail ? "neutral" : "bad"}>
          {detail.canRenderRouteEvidenceDetail ? t($ => $.auxiliary.gatewayRouteEvidence_ui_offline_detail) : t($ => $.auxiliary.gatewayRouteEvidence_ui_blocked)}
        </StatusBadge>
      </div>

      <article className="model-gateway-route-evidence-hero">
        <div>
          <p className="eyebrow">{evidenceProjectionStatus(t, detail.routeEvidenceMode)}</p>
          <h4>{detail.tenantRef}</h4>
          <p>{evidenceProjectionText(t, detail, "routeNarrative")}</p>
        </div>
        <dl className="model-gateway-route-evidence-meta">
          <div>
            <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Overview)}</dt>
            <dd>{detail.gatewayOverviewPageId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Request)}</dt>
            <dd>{detail.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Audit)}</dt>
            <dd>{detail.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Live_provider_call)}</dt>
            <dd>{detail.canExecuteProviderCall ? t($ => $.auxiliary.gatewayRouteEvidence_ui_enabled) : t($ => $.auxiliary.gatewayRouteEvidence_ui_locked)}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Provider_Inventory)}</p>
          <h4>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Profiles_and_credential_state)}</h4>
        </div>
        <div className="model-gateway-route-evidence-profile-grid" aria-label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Model_gateway_provider_profiles)}>
          {detail.profiles.map((profile) => (
            <ProviderProfileCard key={profile.profileId} profile={profile} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Route_Binding)}</p>
          <h4>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Northbound_route_metadata)}</h4>
        </div>
        <div className="model-gateway-route-evidence-route-grid" aria-label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Model_gateway_route_bindings)}>
          {detail.routeBindings.map((route) => (
            <RouteBindingCard key={route.routeId} route={route} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Selection_Policy)}</p>
          <h4>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Positive_and_negative_cases)}</h4>
        </div>
        <div className="model-gateway-route-evidence-selection-grid" aria-label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Model_gateway_selection_cases)}>
          {detail.selectionCases.map((selectionCase) => (
            <SelectionCaseCard key={selectionCase.caseId} selectionCase={selectionCase} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Runtime_Guards)}</p>
          <h4>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Execution_boundaries)}</h4>
        </div>
        <div className="model-gateway-route-evidence-guard-grid" aria-label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Model_gateway_runtime_guards)}>
          {detail.runtimeGuards.map((guard) => (
            <RuntimeGuardCard key={guard.guardId} guard={guard} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ProviderProfileCard({ profile }: { profile: ModelGatewayProviderProfileEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{profile.providerId}</p>
          <h5>{profile.profileId}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(profile.status)}>{evidenceProjectionStatus(t, profile.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Profile)}</dt>
          <dd>{profile.profileName}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Resolved_model)}</dt>
          <dd>{evidenceProjectionText(t, profile, "resolvedModel")}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Credential)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.credentialState)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Deployment)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.deploymentMode)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Auth)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.authMode)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Streaming)}</dt>
          <dd>{evidenceProjectionStatus(t, profile.streaming)}</dd>
        </div>
      </dl>
      <ChipList label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Routes)} items={profile.routes} />
      <ChipList label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Protocols)} items={profile.protocols} />
      <ChipList label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Capabilities)} items={profile.capabilities.map(item => evidenceProjectionStatus(t, item))} technical={false} />
      <p className="model-gateway-route-evidence-source">{t($ => $.auxiliary.gatewayRouteEvidence_ui_Evidence)}{profile.evidenceRef}</p>
      <p>{evidenceProjectionText(t, profile, "summary")}</p>
    </article>
  );
}

function RouteBindingCard({ route }: { route: ModelGatewayRouteBindingEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{route.routeId}</p>
          <h5>{evidenceProjectionText(t, route, "protocol")}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(route.status)}>{evidenceProjectionStatus(t, route.status)}</StatusBadge>
      </div>
      <p className="route-path">{route.path}</p>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Selected_profile)}</dt>
          <dd>{route.selectedProfileId}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Selection_source)}</dt>
          <dd>{evidenceProjectionStatus(t, route.selectionSource)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Evidence_2)}</dt>
          <dd>{route.evidenceRef}</dd>
        </div>
      </dl>
      <ChipList label={t($ => $.auxiliary.gatewayRouteEvidence_ui_Metadata)} items={route.metadataFields} />
      <p>{evidenceProjectionText(t, route, "summary")}</p>
    </article>
  );
}

function SelectionCaseCard({ selectionCase }: { selectionCase: ModelGatewaySelectionCaseEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-route-evidence-case">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{selectionCase.caseId}</p>
          <h5>{evidenceProjectionText(t, selectionCase, "label")}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(selectionCase.status)}>{evidenceProjectionStatus(t, selectionCase.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Input)}</dt>
          <dd>{selectionCase.requestedInput}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Result)}</dt>
          <dd>{evidenceProjectionStatus(t, selectionCase.result)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Failure)}</dt>
          <dd>{selectionCase.failureCode}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Fallback)}</dt>
          <dd>{selectionCase.fallbackAllowed ? t($ => $.auxiliary.gatewayRouteEvidence_ui_allowed) : t($ => $.auxiliary.gatewayRouteEvidence_ui_disabled)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Fast_baseline)}</dt>
          <dd>{selectionCase.defaultFastBaseline ? t($ => $.auxiliary.gatewayRouteEvidence_ui_yes) : t($ => $.auxiliary.gatewayRouteEvidence_ui_manual_only)}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Evidence_2)}</dt>
          <dd>{selectionCase.evidenceRef}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, selectionCase, "summary")}</p>
    </article>
  );
}

function RuntimeGuardCard({ guard }: { guard: ModelGatewayRuntimeGuardEvidence }) {
  const { t } = useTranslation("evaluation");
  return (
    <article className="model-gateway-route-evidence-guard">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{guard.guardId}</p>
          <h5>{evidenceProjectionText(t, guard, "label")}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(guard.status)}>{evidenceProjectionStatus(t, guard.status)}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Source)}</dt>
          <dd>{guard.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.auxiliary.gatewayRouteEvidence_ui_Enforced_by)}</dt>
          <dd>{evidenceProjectionText(t, guard, "enforcedBy")}</dd>
        </div>
      </dl>
      <p>{evidenceProjectionText(t, guard, "summary")}</p>
    </article>
  );
}

function ChipList({ items, label, technical = true }: { items: string[]; label: string; technical?: boolean }) {
  return (
    <div className="model-gateway-route-evidence-chip-row" aria-label={label}>
      <span>{label}</span>
      <div>
        {items.map((item) => (
          technical ? <code key={item}>{item}</code> : <span className="capability-label" key={item}>{item}</span>
        ))}
      </div>
    </div>
  );
}

function routeEvidenceTone(status: ModelGatewayRouteEvidenceStatus): StatusBadgeTone {
  if (status === "ready") {
    return "good";
  }
  if (status === "blocked" || status === "locked" || status === "review_required") {
    return "bad";
  }
  return "neutral";
}

function StatusBadge({ children, tone }: { children: string; tone: StatusBadgeTone }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
