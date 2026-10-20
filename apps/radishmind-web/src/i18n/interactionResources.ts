import { uiI18n } from "./instance.ts";
import { interaction as zh } from "./locales/zh-CN/interaction.ts";
import { interaction as en } from "./locales/en-US/interaction.ts";
uiI18n.addResourceBundle("zh-CN", "applications", { interaction: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { interaction: en }, true);
