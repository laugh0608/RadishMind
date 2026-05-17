package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const sseContentType = "text/event-stream; charset=utf-8"

func prepareSSEHeaders(writer http.ResponseWriter) {
	header := writer.Header()
	header.Set("Content-Type", sseContentType)
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
}

func writeSSEEvent(writer http.ResponseWriter, event string, data any) error {
	if event != "" {
		if _, err := fmt.Fprintf(writer, "event: %s\n", event); err != nil {
			return err
		}
	}

	switch payload := data.(type) {
	case nil:
		if _, err := fmt.Fprint(writer, "data: {}\n\n"); err != nil {
			return err
		}
	case string:
		if payload == "[DONE]" {
			if _, err := fmt.Fprint(writer, "data: [DONE]\n\n"); err != nil {
				return err
			}
		} else {
			for _, line := range strings.Split(payload, "\n") {
				if _, err := fmt.Fprintf(writer, "data: %s\n", line); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(writer, "\n"); err != nil {
				return err
			}
		}
	default:
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(encoded), "\n") {
			if _, err := fmt.Fprintf(writer, "data: %s\n", line); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(writer, "\n"); err != nil {
			return err
		}
	}

	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func splitTextForStreaming(text string, maxRunes int) []string {
	if maxRunes <= 0 {
		maxRunes = 80
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{text}
	}

	chunks := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}
