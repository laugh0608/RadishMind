import { uiI18n } from "./instance.ts";
import { promptEvaluationPlan as zh } from "./locales/zh-CN/promptEvaluationPlan.ts";
import { promptEvaluationPlan as en } from "./locales/en-US/promptEvaluationPlan.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { promptPlan: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { promptPlan: en }, true);
