import { uiI18n } from "./instance.ts";
import { artifact as zh } from "./locales/zh-CN/artifact.ts";
import { artifact as en } from "./locales/en-US/artifact.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { artifact: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { artifact: en }, true);
