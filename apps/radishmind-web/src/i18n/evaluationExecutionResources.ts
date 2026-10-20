import { uiI18n } from "./instance.ts";
import { evaluationExecution as zh } from "./locales/zh-CN/evaluationExecution.ts";
import { evaluationExecution as en } from "./locales/en-US/evaluationExecution.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { campaign: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { campaign: en }, true);
