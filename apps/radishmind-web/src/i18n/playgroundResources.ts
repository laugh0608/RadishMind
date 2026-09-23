import { uiI18n } from "./instance.ts";
import { playground as zh } from "./locales/zh-CN/playground.ts";
import { playground as en } from "./locales/en-US/playground.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { playground: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { playground: en }, true);
