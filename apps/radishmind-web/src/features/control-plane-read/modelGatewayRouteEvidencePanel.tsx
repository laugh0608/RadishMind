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
  return (
    <section
      className="surface-band model-gateway-route-evidence"
      id="model-gateway-route-evidence"
      aria-labelledby="model-gateway-route-evidence-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Model Gateway</p>
          <h3 id="model-gateway-route-evidence-title">Provider/profile route evidence</h3>
        </div>
        <StatusBadge tone={detail.canRenderRouteEvidenceDetail ? "neutral" : "bad"}>
          {detail.canRenderRouteEvidenceDetail ? "offline detail" : "blocked"}
        </StatusBadge>
      </div>

      <article className="model-gateway-route-evidence-hero">
        <div>
          <p className="eyebrow">{detail.routeEvidenceMode}</p>
          <h4>{detail.tenantRef}</h4>
          <p>{detail.routeNarrative}</p>
        </div>
        <dl className="model-gateway-route-evidence-meta">
          <div>
            <dt>Overview</dt>
            <dd>{detail.gatewayOverviewPageId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{detail.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{detail.auditRef}</dd>
          </div>
          <div>
            <dt>Live provider call</dt>
            <dd>{detail.canExecuteProviderCall ? "enabled" : "locked"}</dd>
          </div>
        </dl>
      </article>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">Provider Inventory</p>
          <h4>Profiles and credential state</h4>
        </div>
        <div className="model-gateway-route-evidence-profile-grid" aria-label="Model gateway provider profiles">
          {detail.profiles.map((profile) => (
            <ProviderProfileCard key={profile.profileId} profile={profile} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">Route Binding</p>
          <h4>Northbound route metadata</h4>
        </div>
        <div className="model-gateway-route-evidence-route-grid" aria-label="Model gateway route bindings">
          {detail.routeBindings.map((route) => (
            <RouteBindingCard key={route.routeId} route={route} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">Selection Policy</p>
          <h4>Positive and negative cases</h4>
        </div>
        <div className="model-gateway-route-evidence-selection-grid" aria-label="Model gateway selection cases">
          {detail.selectionCases.map((selectionCase) => (
            <SelectionCaseCard key={selectionCase.caseId} selectionCase={selectionCase} />
          ))}
        </div>
      </div>

      <div className="model-gateway-route-evidence-section">
        <div className="model-gateway-route-evidence-subheading">
          <p className="eyebrow">Runtime Guards</p>
          <h4>Execution boundaries</h4>
        </div>
        <div className="model-gateway-route-evidence-guard-grid" aria-label="Model gateway runtime guards">
          {detail.runtimeGuards.map((guard) => (
            <RuntimeGuardCard key={guard.guardId} guard={guard} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ProviderProfileCard({ profile }: { profile: ModelGatewayProviderProfileEvidence }) {
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{profile.providerId}</p>
          <h5>{profile.profileId}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(profile.status)}>{profile.status}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>Profile</dt>
          <dd>{profile.profileName}</dd>
        </div>
        <div>
          <dt>Resolved model</dt>
          <dd>{profile.resolvedModel}</dd>
        </div>
        <div>
          <dt>Credential</dt>
          <dd>{profile.credentialState}</dd>
        </div>
        <div>
          <dt>Deployment</dt>
          <dd>{profile.deploymentMode}</dd>
        </div>
        <div>
          <dt>Auth</dt>
          <dd>{profile.authMode}</dd>
        </div>
        <div>
          <dt>Streaming</dt>
          <dd>{profile.streaming}</dd>
        </div>
      </dl>
      <ChipList label="Routes" items={profile.routes} />
      <ChipList label="Protocols" items={profile.protocols} />
      <ChipList label="Capabilities" items={profile.capabilities} />
      <p className="model-gateway-route-evidence-source">Evidence: {profile.evidenceRef}</p>
      <p>{profile.summary}</p>
    </article>
  );
}

function RouteBindingCard({ route }: { route: ModelGatewayRouteBindingEvidence }) {
  return (
    <article className="model-gateway-route-evidence-card">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{route.routeId}</p>
          <h5>{route.protocol}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(route.status)}>{route.status}</StatusBadge>
      </div>
      <p className="route-path">{route.path}</p>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>Selected profile</dt>
          <dd>{route.selectedProfileId}</dd>
        </div>
        <div>
          <dt>Selection source</dt>
          <dd>{route.selectionSource}</dd>
        </div>
        <div>
          <dt>Evidence</dt>
          <dd>{route.evidenceRef}</dd>
        </div>
      </dl>
      <ChipList label="Metadata" items={route.metadataFields} />
      <p>{route.summary}</p>
    </article>
  );
}

function SelectionCaseCard({ selectionCase }: { selectionCase: ModelGatewaySelectionCaseEvidence }) {
  return (
    <article className="model-gateway-route-evidence-case">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{selectionCase.caseId}</p>
          <h5>{selectionCase.label}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(selectionCase.status)}>{selectionCase.status}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>Input</dt>
          <dd>{selectionCase.requestedInput}</dd>
        </div>
        <div>
          <dt>Result</dt>
          <dd>{selectionCase.result}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{selectionCase.failureCode}</dd>
        </div>
        <div>
          <dt>Fallback</dt>
          <dd>{selectionCase.fallbackAllowed ? "allowed" : "disabled"}</dd>
        </div>
        <div>
          <dt>Fast baseline</dt>
          <dd>{selectionCase.defaultFastBaseline ? "yes" : "manual only"}</dd>
        </div>
        <div>
          <dt>Evidence</dt>
          <dd>{selectionCase.evidenceRef}</dd>
        </div>
      </dl>
      <p>{selectionCase.summary}</p>
    </article>
  );
}

function RuntimeGuardCard({ guard }: { guard: ModelGatewayRuntimeGuardEvidence }) {
  return (
    <article className="model-gateway-route-evidence-guard">
      <div className="model-gateway-route-evidence-row-main">
        <div>
          <p className="eyebrow">{guard.guardId}</p>
          <h5>{guard.label}</h5>
        </div>
        <StatusBadge tone={routeEvidenceTone(guard.status)}>{guard.status}</StatusBadge>
      </div>
      <dl className="model-gateway-route-evidence-meta">
        <div>
          <dt>Source</dt>
          <dd>{guard.sourceRef}</dd>
        </div>
        <div>
          <dt>Enforced by</dt>
          <dd>{guard.enforcedBy}</dd>
        </div>
      </dl>
      <p>{guard.summary}</p>
    </article>
  );
}

function ChipList({ items, label }: { items: string[]; label: string }) {
  return (
    <div className="model-gateway-route-evidence-chip-row" aria-label={label}>
      <span>{label}</span>
      <div>
        {items.map((item) => (
          <code key={item}>{item}</code>
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
