import { uiI18n } from "./instance.ts";
import { workflowDraft as zhDraft } from "./locales/zh-CN/workflowDraft.ts";
import { workflowDraft as enDraft } from "./locales/en-US/workflowDraft.ts";

// Registered at the existing draft panel boundary, without changing editor identity.
uiI18n.addResourceBundle("zh-CN", "workflow", { draft: zhDraft }, true);
uiI18n.addResourceBundle("en-US", "workflow", { draft: enDraft }, true);
