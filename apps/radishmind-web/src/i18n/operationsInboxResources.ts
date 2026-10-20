import { uiI18n } from "./instance.ts";
import { operationsInbox as zh } from "./locales/zh-CN/operationsInbox.ts";
import { operationsInbox as en } from "./locales/en-US/operationsInbox.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { operationsInbox: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { operationsInbox: en }, true);
