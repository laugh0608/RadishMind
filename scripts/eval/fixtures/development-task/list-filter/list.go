// Package tasklist is synthetic input for the development-task probe.
// The limit boundary defect is intentional; it is not a platform implementation.
package tasklist

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Task struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func Handler(tasks []Task) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != "" && state != "open" && state != "done" {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		limit := 0
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 100 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = parsed
		}
		result := make([]Task, 0)
		for _, task := range tasks {
			if state == "" || task.State == state {
				result = append(result, task)
			}
		}
		if limit > 0 && len(result) > limit {
			result = result[:limit-1]
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			return
		}
	})
}
