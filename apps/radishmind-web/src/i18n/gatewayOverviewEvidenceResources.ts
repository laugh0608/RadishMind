import "./auxiliaryEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { gatewayOverviewEvidence as en } from "./locales/en-US/gatewayOverviewEvidence.ts";
import { gatewayOverviewEvidence as zh } from "./locales/zh-CN/gatewayOverviewEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
