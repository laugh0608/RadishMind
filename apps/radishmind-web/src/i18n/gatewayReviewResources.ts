import { uiI18n } from "./instance.ts";
import { gatewayReview as zh } from "./locales/zh-CN/gatewayReview.ts";
import { gatewayReview as en } from "./locales/en-US/gatewayReview.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { review: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { review: en }, true);
