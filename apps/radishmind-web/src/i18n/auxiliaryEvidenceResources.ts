import { uiI18n } from "./instance.ts";
import { auxiliaryEvidence as en } from "./locales/en-US/auxiliaryEvidence.ts";
import { auxiliaryEvidence as zh } from "./locales/zh-CN/auxiliaryEvidence.ts";
uiI18n.addResourceBundle("en-US", "evaluation", { auxiliary: en }, true);
uiI18n.addResourceBundle("zh-CN", "evaluation", { auxiliary: zh }, true);
