package translation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func claudeResponse(titleEN, bodyEN string) map[string]interface{} {
	translated, _ := json.Marshal(map[string]string{"title": titleEN, "body": bodyEN})
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": string(translated)},
		},
	}
}

// AC-I18N-004: on success, titleEN and bodyEN are returned, no error
func TestClaudeTranslator_TranslatesSuccessfully(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		assert.Equal(t, "application/json", r.Header.Get("content-type"))
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse("Hello World", "Body in English"))
	}))
	defer server.Close()

	tr := newTranslatorWithClient("test-key", server.URL, server.Client())
	titleEN, bodyEN, err := tr.Translate(context.Background(), "Xin chào", "Nội dung tiếng Việt")

	require.NoError(t, err)
	assert.Equal(t, "Hello World", titleEN)
	assert.Equal(t, "Body in English", bodyEN)
}

// AC-I18N-004: API failure → error returned; blog save is caller's responsibility
func TestClaudeTranslator_APIError_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tr := newTranslatorWithClient("test-key", server.URL, server.Client())
	_, _, err := tr.Translate(context.Background(), "Title", "Body")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// AC-I18N-004: malformed JSON in response → error
func TestClaudeTranslator_BadResponseJSON_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"content":[{"type":"text","text":"not-json-object"}]}`))
	}))
	defer server.Close()

	tr := newTranslatorWithClient("test-key", server.URL, server.Client())
	_, _, err := tr.Translate(context.Background(), "Title", "Body")

	assert.Error(t, err)
}

// AC-I18N-004: empty content array → error
func TestClaudeTranslator_EmptyContent_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"content": []interface{}{}})
	}))
	defer server.Close()

	tr := newTranslatorWithClient("test-key", server.URL, server.Client())
	_, _, err := tr.Translate(context.Background(), "Title", "Body")

	assert.Error(t, err)
}
