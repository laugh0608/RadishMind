import { uiI18n } from "./instance.ts";
import { adminRead as zh } from "./locales/zh-CN/adminRead.ts";
import { adminRead as en } from "./locales/en-US/adminRead.ts";

uiI18n.addResourceBundle("zh-CN", "admin", { read: zh }, true);
uiI18n.addResourceBundle("en-US", "admin", { read: en }, true);
