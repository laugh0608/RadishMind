import { uiI18n } from "./instance.ts";
import { configuration as zh } from "./locales/zh-CN/configuration.ts";
import { configuration as en } from "./locales/en-US/configuration.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { configuration: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { configuration: en }, true);
