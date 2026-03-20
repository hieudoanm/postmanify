package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hieudoanm/free.router/libs/openrouter"
)

const upstreamURL = "https://openrouter.ai/api/v1/chat/completions"

// AliasID returns a Cursor-safe model name by stripping the provider prefix
// and the :free suffix, e.g. "arcee-ai/trinity-large-preview:free" → "trinity-large-preview".
// Cursor rejects slashes and colons in model names.
func AliasID(id string) string {
	// drop provider prefix (everything up to and including the first "/")
	if i := strings.Index(id, "/"); i >= 0 {
		id = id[i+1:]
	}
	// drop :free (or any :<tag>) suffix
	if i := strings.Index(id, ":"); i >= 0 {
		id = id[:i]
	}
	return id
}

// NewHandler returns an http.Handler that:
//   - GET  /v1/models              → returns the pinned model with a Cursor-safe alias
//   - POST /v1/chat/completions    → accepts alias OR full ID, proxies to OpenRouter
//   - Everything else              → 404
func NewHandler(model *openrouter.Model, apiKey string) http.Handler {
	alias := AliasID(model.ID)
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := map[string]any{
			"object": "list",
			"data": []map[string]any{
				{
					// Expose the alias so Cursor accepts it, but also
					// include the full id in a custom field for reference.
					"id":       alias,
					"object":   "model",
					"owned_by": providerOf(model.ID),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})

	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"cannot read body"}`, http.StatusBadRequest)
			return
		}

		var body map[string]any
		if err := json.Unmarshal(rawBody, &body); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Always pin the real OpenRouter model ID regardless of what
		// Cursor sends (it will send the alias).
		body["model"] = model.ID

		// Tell OpenRouter to try all available providers and allow
		// data collection — without this, rate-limited or privacy-
		// restricted providers return 429/404 with no fallback.
		if _, exists := body["provider"]; !exists {
			body["provider"] = map[string]any{
				"allow_fallbacks": true,
				"data_collection": "allow",
			}
		}

		upstreamBody, _ := json.Marshal(body)

		req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("HTTP-Referer", "https://github.com/freerouter/freerouter")
		req.Header.Set("X-Title", "freerouter")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"upstream error: %s"}`, err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/json"
		}
		w.Header().Set("Content-Type", ct)

		isStream := body["stream"] == true
		if isStream {
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Transfer-Encoding", "chunked")
		}

		w.WriteHeader(resp.StatusCode)

		if isStream {
			flusher, ok := w.(http.Flusher)
			buf := make([]byte, 4096)
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					_, _ = w.Write(buf[:n])
					if ok {
						flusher.Flush()
					}
				}
				if readErr != nil {
					break
				}
			}
		} else {
			_, _ = io.Copy(w, resp.Body)
		}
	})

	// 404 for anything else
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error":"not found","path":%q}`, r.URL.Path)
	})

	return mux
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func providerOf(id string) string {
	for j, c := range id {
		if c == '/' {
			return id[:j]
		}
	}
	return "openrouter"
}
