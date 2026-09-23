import { uiI18n } from "./instance.ts";
import { publish as zh } from "./locales/zh-CN/publish.ts";
import { publish as en } from "./locales/en-US/publish.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { publish: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { publish: en }, true);
