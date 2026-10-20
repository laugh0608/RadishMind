import { uiI18n } from "./instance.ts";
import { usageQuota as zh } from "./locales/zh-CN/usageQuota.ts";
import { usageQuota as en } from "./locales/en-US/usageQuota.ts";
uiI18n.addResourceBundle("zh-CN", "gateway", { usageQuota: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { usageQuota: en }, true);
