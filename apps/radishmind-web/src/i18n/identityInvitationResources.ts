import { uiI18n } from "./instance.ts";
import { identityInvitations as zh } from "./locales/zh-CN/identityInvitations.ts";
import { identityInvitations as en } from "./locales/en-US/identityInvitations.ts";
import "./identityRoleResources.ts";

uiI18n.addResourceBundle("zh-CN", "identity", { invitations: zh }, true);
uiI18n.addResourceBundle("en-US", "identity", { invitations: en }, true);
