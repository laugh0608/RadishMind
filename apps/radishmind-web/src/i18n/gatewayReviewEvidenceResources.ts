import "./auxiliaryEvidenceResources.ts";
import "./gatewayOverviewEvidenceResources.ts";
import "./gatewayRouteEvidenceResources.ts";
import "./gatewayUsageEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { gatewayReviewEvidence as en } from "./locales/en-US/gatewayReviewEvidence.ts";
import { gatewayReviewEvidence as zh } from "./locales/zh-CN/gatewayReviewEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
