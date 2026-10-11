import { uiI18n } from "./instance.ts";
import { runEvidence as zh } from "./locales/zh-CN/runEvidence.ts";
import { runEvidence as en } from "./locales/en-US/runEvidence.ts";

uiI18n.addResourceBundle("zh-CN", "evaluation", { evidence: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { evidence: en }, true);
