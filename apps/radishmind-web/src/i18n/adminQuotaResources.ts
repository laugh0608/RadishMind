import { uiI18n } from "./instance.ts";
import { adminQuota as zh } from "./locales/zh-CN/adminQuota.ts";
import { adminQuota as en } from "./locales/en-US/adminQuota.ts";

uiI18n.addResourceBundle("zh-CN", "admin", { quota: zh }, true);
uiI18n.addResourceBundle("en-US", "admin", { quota: en }, true);
