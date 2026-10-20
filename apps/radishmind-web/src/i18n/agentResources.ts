import { uiI18n } from "./instance.ts";
import { agent as zh } from "./locales/zh-CN/agent.ts";
import { agent as en } from "./locales/en-US/agent.ts";

uiI18n.addResourceBundle("zh-CN", "agent", zh, true);
uiI18n.addResourceBundle("en-US", "agent", en, true);
