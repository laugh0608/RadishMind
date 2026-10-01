import { uiI18n } from "./instance.ts";
import { runReview as zh } from "./locales/zh-CN/runReview.ts";
import { runReview as en } from "./locales/en-US/runReview.ts";
uiI18n.addResourceBundle("zh-CN", "evaluation", { runReview: zh }, true);
uiI18n.addResourceBundle("en-US", "evaluation", { runReview: en }, true);
