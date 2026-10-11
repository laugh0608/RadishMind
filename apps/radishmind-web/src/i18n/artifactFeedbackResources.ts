import { uiI18n } from "./instance.ts";
import { artifactFeedback as en } from "./locales/en-US/artifactFeedback.ts";
import { artifactFeedback as zh } from "./locales/zh-CN/artifactFeedback.ts";
uiI18n.addResourceBundle("en-US", "applications", { artifactFeedback: en }, true);
uiI18n.addResourceBundle("zh-CN", "applications", { artifactFeedback: zh }, true);
