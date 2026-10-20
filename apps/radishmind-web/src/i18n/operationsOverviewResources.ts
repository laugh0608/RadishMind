import { uiI18n } from "./instance.ts";
import { operationsOverview as zh } from "./locales/zh-CN/operationsOverview.ts";
import { operationsOverview as en } from "./locales/en-US/operationsOverview.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { operationsOverview: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { operationsOverview: en }, true);
