import { uiI18n } from "./instance.ts";
import { identitySecurity as zh } from "./locales/zh-CN/identitySecurity.ts";
import { identitySecurity as en } from "./locales/en-US/identitySecurity.ts";

uiI18n.addResourceBundle("zh-CN", "identity", { security: zh }, true);
uiI18n.addResourceBundle("en-US", "identity", { security: en }, true);
