import { uiI18n } from "./instance.ts";
import { workspaceSurface as zh } from "./locales/zh-CN/workspaceSurface.ts";
import { workspaceSurface as en } from "./locales/en-US/workspaceSurface.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { workspaceSurface: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { workspaceSurface: en }, true);
