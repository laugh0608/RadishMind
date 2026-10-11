import "./auxiliaryEvidenceResources.ts";
import "./adminOperationsEvidenceResources.ts";
import "./gatewayRouteEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { adminDeploymentEvidence as en } from "./locales/en-US/adminDeploymentEvidence.ts";
import { adminDeploymentEvidence as zh } from "./locales/zh-CN/adminDeploymentEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
