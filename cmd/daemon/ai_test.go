package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeLLM is a minimal OpenAI-compatible endpoint. Like some real
// providers, it rejects response_format, so the client's plain
// retry path is exercised too.
func fakeLLM(t *testing.T, calls *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, `{"error":{"message":"bad key"}}`, http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/v1/models":
			io.WriteString(w, `{"data":[{"id":"models/b-model"},{"id":"a-model"}]}`)
		case "/v1/chat/completions":
			calls.Add(1)
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			json.Unmarshal(raw, &body)
			if _, ok := body["response_format"]; ok {
				http.Error(w, `{"error":{"message":"response_format unsupported"}}`, http.StatusBadRequest)
				return
			}
			msgs := body["messages"].([]any)
			user := msgs[1].(map[string]any)["content"].(string)
			var items []string
			if strings.Contains(user, "Gradient") {
				items = append(items, `{"id":1,"kind":"productive","category":"Lecture","subject":"Machine Learning","topic":"Gradient Descent"}`)
			}
			if strings.Contains(user, "Instagram") {
				items = append(items, `{"id":2,"kind":"distraction","category":"Social Media","subject":"","topic":""}`)
			}
			content, _ := json.Marshal("{\"items\":[" + strings.Join(items, ",") + "]}")
			io.WriteString(w, `{"choices":[{"message":{"content":`+string(content)+`}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestClassifierEndToEnd(t *testing.T) {
	t.Setenv("STUDYTRACKER_AI_KEY", "test-key")
	var calls atomic.Int32
	llm := fakeLLM(t, &calls)
	defer llm.Close()

	dir := t.TempDir()
	cfg, _ := LoadConfig(filepath.Join(dir, "config.json"))
	if err := cfg.Update(func(c *Config) {
		c.AI = AIConfig{Enabled: true, BaseURL: llm.URL + "/v1", Model: "a-model"}
	}); err != nil {
		t.Fatal(err)
	}

	e := NewRuleEngine(writeRules(t, testRules))
	e.Reload()
	store, err := OpenStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Now().Unix()
	for _, s := range []session{
		{start: now - 900, end: now - 300, app: "chrome.exe", url: "https://www.youtube.com/watch?v=g", title: "Gradient Descent, Step-by-Step - YouTube"},
		{start: now - 300, end: now - 100, app: "chrome.exe", url: "https://www.instagram.com/", title: "Instagram"}, // rules: Distractions -> skipped
		{start: now - 100, end: now - 90, app: "notepad.exe", title: "tiny"},                                         // < 60s -> skipped
	} {
		s := s
		s.category = e.Categorize(s.app, s.title, s.url)
		if _, err := store.Insert(&s); err != nil {
			t.Fatal(err)
		}
	}

	labels := NewLabelCache()
	c := NewClassifier(store, e, labels, cfg, log.New(io.Discard, "", 0))

	pending, err := c.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || !strings.Contains(pending[0].Title, "Gradient") {
		t.Fatalf("pending = %+v, want only the lecture", pending)
	}

	c.runOnce(context.Background())
	st := c.Status()
	if st.LastError != "" || st.Labeled != 1 || st.Pending != 0 || !st.Configured {
		t.Fatalf("status = %+v", st)
	}
	l, ok := labels.Get(pending[0].Key)
	if !ok || l.Subject != "Machine Learning" || l.Topic != "Gradient Descent" || l.Kind != kindProductive {
		t.Fatalf("label = %+v (ok=%v)", l, ok)
	}
	saved, _ := store.Labels()
	if len(saved) != 1 {
		t.Fatalf("labels persisted = %d, want 1", len(saved))
	}
	if calls.Load() != 2 { // first attempt rejected (response_format), retried plain
		t.Fatalf("chat calls = %d, want 2", calls.Load())
	}

	// Already labeled: a second run sends nothing.
	c.runOnce(context.Background())
	if calls.Load() != 2 {
		t.Fatalf("relabeled an already-labeled activity")
	}

	results, err := c.Test(context.Background())
	if err != nil || len(results) != 2 || results[1]["kind"] != kindDistraction {
		t.Fatalf("Test() = %+v, %v", results, err)
	}

	models, err := c.Models(context.Background())
	if err != nil || len(models) != 2 || models[0] != "a-model" || models[1] != "b-model" {
		t.Fatalf("Models() = %v, %v", models, err)
	}

	if err := c.Clear(); err != nil || labels.Len() != 0 {
		t.Fatalf("Clear() left %d labels (%v)", labels.Len(), err)
	}
}

// A provider that is "experiencing high demand" (503) for a while
// must be retried, not reported as a failure.
func TestRetriesTransientErrors(t *testing.T) {
	saved := retryDelays
	retryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	defer func() { retryDelays = saved }()

	var calls, failFirst atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= failFirst.Load() {
			w.Header().Set("Retry-After", "0")
			http.Error(w, `{"error":{"message":"This model is currently experiencing high demand."}}`, http.StatusServiceUnavailable)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"{\"items\":[{\"id\":1,\"kind\":\"productive\",\"subject\":\"Statistics\"}]}"}}]}`)
	}))
	defer srv.Close()
	cfg := AIConfig{BaseURL: srv.URL, Model: "m"}

	// Two 503s, then success: the caller never sees an error.
	failFirst.Store(2)
	content, err := chatComplete(context.Background(), srv.Client(), cfg, "k", "sys", "user")
	if err != nil || !strings.Contains(content, "Statistics") {
		t.Fatalf("after transient 503s: content=%q err=%v", content, err)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
	}

	// Always overloaded: give up after the retries, with a helpful hint.
	calls.Store(0)
	failFirst.Store(100)
	_, err = chatComplete(context.Background(), srv.Client(), cfg, "k", "sys", "user")
	if err == nil || !strings.Contains(err.Error(), "high demand") || !strings.Contains(err.Error(), "retry automatically") {
		t.Fatalf("persistent 503 error = %v", err)
	}
	if calls.Load() != int32(len(retryDelays)+1) {
		t.Fatalf("calls = %d, want %d", calls.Load(), len(retryDelays)+1)
	}
}

func TestSettingsValidation(t *testing.T) {
	bad := []string{
		"503 Service Unavailable: This model is currently experiencing high demand. Please try again later.",
		"short",
		"https://aistudio.google.com/apikey extra",
	}
	for _, k := range bad {
		if looksLikeAPIKey(k) {
			t.Errorf("accepted bogus key %q", k)
		}
	}
	if !looksLikeAPIKey("AIzaSyD-abcdefghijklmnopqrstuvwxyz01234") || !looksLikeAPIKey("sk-proj-abc_DEF-123456") {
		t.Error("rejected a real-looking key")
	}
	for _, m := range []string{"gemini-3.8-flash-lite-tts", "gemini-embedding-001", "veo-3.1-generate-preview", "gemini-2.5-flash-image"} {
		if !isNonChatModel(m) {
			t.Errorf("%s should be rejected as non-chat", m)
		}
	}
	for _, m := range []string{"gemini-3.8-flash", "gemini-3.8-flash-lite", "gpt-5-mini", "llama3.2"} {
		if isNonChatModel(m) {
			t.Errorf("%s wrongly rejected", m)
		}
	}
}

func TestAPIKeyEncryptedAtRest(t *testing.T) {
	t.Setenv("STUDYTRACKER_AI_KEY", "")
	path := filepath.Join(t.TempDir(), "config.json")
	cfg, _ := LoadConfig(path)
	enc, err := encryptAPIKey("AIza-secret-1234")
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Update(func(c *Config) { c.AI.APIKeyEnc = enc }); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.APIKey(); got != "AIza-secret-1234" {
		t.Fatalf("APIKey() = %q", got)
	}
	if strings.Contains(reloaded.Get().AI.APIKeyEnc, "secret") {
		t.Fatal("key stored in plain text")
	}
	if keyHint("AIza-secret-1234") != "••••1234" {
		t.Fatalf("keyHint = %q", keyHint("AIza-secret-1234"))
	}
}
