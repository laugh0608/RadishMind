import "./auxiliaryEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { gatewayUsageEvidence as en } from "./locales/en-US/gatewayUsageEvidence.ts";
import { gatewayUsageEvidence as zh } from "./locales/zh-CN/gatewayUsageEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
