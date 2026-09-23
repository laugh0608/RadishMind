import { uiI18n } from "./instance.ts";
import { prompt as zhPrompt } from "./locales/zh-CN/prompt.ts";
import { prompt as enPrompt } from "./locales/en-US/prompt.ts";

// Loaded with the Prompt workspace, after startup initialization and before its panels render.
uiI18n.addResourceBundle("zh-CN", "prompt", zhPrompt);
uiI18n.addResourceBundle("en-US", "prompt", enPrompt);
