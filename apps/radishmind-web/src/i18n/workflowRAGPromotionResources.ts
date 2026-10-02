import { uiI18n } from "./instance.ts";
import { workflowRAGPromotion as zh } from "./locales/zh-CN/workflowRAGPromotion.ts";
import { workflowRAGPromotion as en } from "./locales/en-US/workflowRAGPromotion.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { ragPromotion: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { ragPromotion: en }, true);
