import { uiI18n } from "./instance.ts";
import { operations as zh } from "./locales/zh-CN/operations.ts";
import { operations as en } from "./locales/en-US/operations.ts";

uiI18n.addResourceBundle("zh-CN", "gateway", { operations: zh }, true);
uiI18n.addResourceBundle("en-US", "gateway", { operations: en }, true);
