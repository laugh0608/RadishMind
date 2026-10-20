import { uiI18n } from "./instance.ts";
import { evaluationPair as zh } from "./locales/zh-CN/evaluationPair.ts";
import { evaluationPair as en } from "./locales/en-US/evaluationPair.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { campaign: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { campaign: en }, true);
