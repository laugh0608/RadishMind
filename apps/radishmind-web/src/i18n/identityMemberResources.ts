import { uiI18n } from "./instance.ts";
import { identityMembers as zh } from "./locales/zh-CN/identityMembers.ts";
import { identityMembers as en } from "./locales/en-US/identityMembers.ts";
import "./identityRoleResources.ts";

uiI18n.addResourceBundle("zh-CN", "identity", { members: zh }, true);
uiI18n.addResourceBundle("en-US", "identity", { members: en }, true);
