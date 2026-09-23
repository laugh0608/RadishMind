import { uiI18n } from "./instance.ts";
import { apiIntegration as zh } from "./locales/zh-CN/apiIntegration.ts";
import { apiIntegration as en } from "./locales/en-US/apiIntegration.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { apiIntegration: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { apiIntegration: en }, true);
