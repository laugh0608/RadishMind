import { uiI18n } from "./instance.ts";
import { evaluationCampaign as zh } from "./locales/zh-CN/evaluationCampaign.ts";
import { evaluationCampaign as en } from "./locales/en-US/evaluationCampaign.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { campaign: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { campaign: en }, true);
