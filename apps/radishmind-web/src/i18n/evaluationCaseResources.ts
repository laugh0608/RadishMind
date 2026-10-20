import { uiI18n } from "./instance.ts";
import { evaluationCases as zh } from "./locales/zh-CN/evaluationCases.ts";
import { evaluationCases as en } from "./locales/en-US/evaluationCases.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { cases: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { cases: en }, true);
