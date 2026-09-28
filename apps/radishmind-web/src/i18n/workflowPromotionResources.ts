import { uiI18n } from "./instance.ts";
import { workflowPromotion as zh } from "./locales/zh-CN/workflowPromotion.ts";
import { workflowPromotion as en } from "./locales/en-US/workflowPromotion.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { promotion: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { promotion: en }, true);
