import "./auxiliaryEvidenceResources.ts";
import "./gatewayReviewEvidenceResources.ts";
import { uiI18n } from "./instance.ts";
import { adminOperationsEvidence as en } from "./locales/en-US/adminOperationsEvidence.ts";
import { adminOperationsEvidence as zh } from "./locales/zh-CN/adminOperationsEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
