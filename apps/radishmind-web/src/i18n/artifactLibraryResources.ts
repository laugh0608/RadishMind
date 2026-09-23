import { uiI18n } from "./instance.ts";
import { artifactLibrary as zh } from "./locales/zh-CN/artifactLibrary.ts";
import { artifactLibrary as en } from "./locales/en-US/artifactLibrary.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { artifactLibrary: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { artifactLibrary: en }, true);
