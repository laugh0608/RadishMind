import { uiI18n } from "./instance.ts";
import { adminShell as zh } from "./locales/zh-CN/adminShell.ts";
import { adminShell as en } from "./locales/en-US/adminShell.ts";

uiI18n.addResourceBundle("zh-CN", "admin", zh, true);
uiI18n.addResourceBundle("en-US", "admin", en, true);
