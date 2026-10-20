import { uiI18n } from "./instance.ts";
import { adminProvider as zh } from "./locales/zh-CN/adminProvider.ts";
import { adminProvider as en } from "./locales/en-US/adminProvider.ts";

uiI18n.addResourceBundle("zh-CN", "admin", { provider: zh }, true);
uiI18n.addResourceBundle("en-US", "admin", { provider: en }, true);
