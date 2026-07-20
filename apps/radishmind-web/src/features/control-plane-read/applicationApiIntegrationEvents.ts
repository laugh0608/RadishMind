import type {
  ApplicationApiProtocol,
  ApplicationModelCatalogItem,
} from "./applicationApiIntegrationConsumer.ts";

export const APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT = "radishmind:application-api-integration-draft-handoff";
export const APPLICATION_MODEL_CATALOG_READY_EVENT = "radishmind:application-model-catalog-ready";

export type ApplicationApiIntegrationDraftHandoffDetail = {
  applicationId: string;
  protocol: ApplicationApiProtocol;
  model: string;
};

export type ApplicationModelCatalogReadyDetail = {
  applicationId: string;
  models: ApplicationModelCatalogItem[];
  selectedModel: string;
};

export function createApplicationApiIntegrationDraftHandoffDetail(
  applicationId: string,
  protocol: ApplicationApiProtocol,
  model: string,
): ApplicationApiIntegrationDraftHandoffDetail {
  return { applicationId: applicationId.trim(), protocol, model: model.trim() };
}

export function requestApplicationApiIntegrationDraftHandoff(
  applicationId: string,
  protocol: ApplicationApiProtocol,
  model: string,
): void {
  window.dispatchEvent(new CustomEvent(APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT, {
    detail: createApplicationApiIntegrationDraftHandoffDetail(applicationId, protocol, model),
  }));
}

export function createApplicationModelCatalogReadyDetail(
  applicationId: string,
  models: ApplicationModelCatalogItem[],
  selectedModel: string,
): ApplicationModelCatalogReadyDetail {
  const normalizedApplicationId = applicationId.trim();
  const normalizedSelectedModel = selectedModel.trim();
  if (!/^[A-Za-z0-9._:-]{1,160}$/u.test(normalizedApplicationId)) {
    throw new Error("Application model catalog handoff scope is invalid.");
  }
  const seen = new Set<string>();
  const normalizedModels = models.map((model) => {
    const id = model.id.trim();
    if (!/^[A-Za-z0-9._:/-]{1,160}$/u.test(id) || seen.has(id) ||
      model.ownedBy.length > 160 || /[\r\n\0]/u.test(model.ownedBy) ||
      !model.protocols.every((protocol) => protocol === "chat_completions" || protocol === "responses" || protocol === "messages")) {
      throw new Error("Application model catalog handoff is invalid.");
    }
    seen.add(id);
    return { id, ownedBy: model.ownedBy, protocols: [...new Set(model.protocols)] };
  });
  if (!normalizedSelectedModel || !seen.has(normalizedSelectedModel)) {
    throw new Error("Application model catalog handoff selection is invalid.");
  }
  return { applicationId: normalizedApplicationId, models: normalizedModels, selectedModel: normalizedSelectedModel };
}

export function requestApplicationModelCatalogReady(
  applicationId: string,
  models: ApplicationModelCatalogItem[],
  selectedModel: string,
): void {
  window.dispatchEvent(new CustomEvent<ApplicationModelCatalogReadyDetail>(APPLICATION_MODEL_CATALOG_READY_EVENT, {
    detail: createApplicationModelCatalogReadyDetail(applicationId, models, selectedModel),
  }));
}
