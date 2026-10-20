import { uiI18n } from "./instance.ts";
import { requestHistory as zh } from "./locales/zh-CN/requestHistory.ts";
import { requestHistory as en } from "./locales/en-US/requestHistory.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { requestHistory: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { requestHistory: en }, true);
