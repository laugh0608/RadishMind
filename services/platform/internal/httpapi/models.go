package httpapi

import (
	"context"
	"net/http"
	"time"
)

type openAIModelListResponse struct {
	Object string              `json:"object"`
	Data   []openAIModelObject `json:"data"`
}

type openAIModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func handleModels(writer http.ResponseWriter, server *Server) {
	ctx, cancel := context.WithTimeout(context.Background(), server.config.BridgeTimeout)
	defer cancel()

	providers, err := server.bridge.DescribeProviders(ctx)
	if err != nil {
		writeOpenAIError(writer, http.StatusBadGateway, "PROVIDER_REGISTRY_UNAVAILABLE", err.Error())
		return
	}

	models := make([]openAIModelObject, 0, len(providers))
	now := time.Now().Unix()
	for _, provider := range providers {
		models = append(models, openAIModelObject{
			ID:      provider.ProviderID,
			Object:  "model",
			Created: now,
			OwnedBy: "radishmind",
		})
	}

	writeJSON(writer, http.StatusOK, openAIModelListResponse{
		Object: "list",
		Data:   models,
	})
}
