import "./auxiliaryEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { gatewayRouteEvidence as en } from "./locales/en-US/gatewayRouteEvidence.ts";
import { gatewayRouteEvidence as zh } from "./locales/zh-CN/gatewayRouteEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
