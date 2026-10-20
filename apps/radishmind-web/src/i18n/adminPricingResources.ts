import { uiI18n } from "./instance.ts";
import { adminPricing as zh } from "./locales/zh-CN/adminPricing.ts";
import { adminPricing as en } from "./locales/en-US/adminPricing.ts";

uiI18n.addResourceBundle("zh-CN", "admin", { pricing: zh }, true);
uiI18n.addResourceBundle("en-US", "admin", { pricing: en }, true);
