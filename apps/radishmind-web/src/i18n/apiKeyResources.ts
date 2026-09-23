import { uiI18n } from "./instance.ts";
import { apiKey as zh } from "./locales/zh-CN/apiKey.ts";
import { apiKey as en } from "./locales/en-US/apiKey.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { apiKey: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { apiKey: en }, true);
