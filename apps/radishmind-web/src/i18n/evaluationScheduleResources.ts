import { uiI18n } from "./instance.ts";
import { evaluationSchedule as zh } from "./locales/zh-CN/evaluationSchedule.ts";
import { evaluationSchedule as en } from "./locales/en-US/evaluationSchedule.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { campaign: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { campaign: en }, true);
