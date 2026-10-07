import { uiI18n } from "./instance.ts";
import { identityRoles as zh } from "./locales/zh-CN/identityRoles.ts";
import { identityRoles as en } from "./locales/en-US/identityRoles.ts";

uiI18n.addResourceBundle("zh-CN", "identity", { roles: zh }, true);
uiI18n.addResourceBundle("en-US", "identity", { roles: en }, true);
