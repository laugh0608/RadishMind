import { uiI18n } from "./instance.ts";
import { catalog as zh } from "./locales/zh-CN/catalog.ts";
import { catalog as en } from "./locales/en-US/catalog.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { catalog: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { catalog: en }, true);
