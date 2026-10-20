import { uiI18n } from "./instance.ts";
import { actionSafety as zh } from "./locales/zh-CN/actionSafety.ts";
import { actionSafety as en } from "./locales/en-US/actionSafety.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { actionSafety: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { actionSafety: en }, true);
