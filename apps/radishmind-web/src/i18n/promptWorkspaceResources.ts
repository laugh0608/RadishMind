import { uiI18n } from "./instance.ts";
import { promptWorkspace as zh } from "./locales/zh-CN/promptWorkspace.ts";
import { promptWorkspace as en } from "./locales/en-US/promptWorkspace.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { promptWorkspace: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { promptWorkspace: en }, true);
