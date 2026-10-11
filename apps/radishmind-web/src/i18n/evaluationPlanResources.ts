import { uiI18n } from "./instance.ts";
import { evaluationPlan as zh } from "./locales/zh-CN/evaluationPlan.ts";
import { evaluationPlan as en } from "./locales/en-US/evaluationPlan.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { campaign: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { campaign: en }, true);
