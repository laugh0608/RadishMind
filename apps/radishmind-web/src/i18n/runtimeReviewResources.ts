import { uiI18n } from "./instance.ts";
import { runtimeReview as zh } from "./locales/zh-CN/runtimeReview.ts";
import { runtimeReview as en } from "./locales/en-US/runtimeReview.ts";
uiI18n.addResourceBundle("zh-CN", "gateway", { runtimeReview: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { runtimeReview: en }, true);
