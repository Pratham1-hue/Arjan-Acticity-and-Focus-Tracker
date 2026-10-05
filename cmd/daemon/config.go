package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const defaultPort = 4000

// AIConfig selects an OpenAI-compatible chat completions endpoint.
// One code path covers Gemini, OpenAI, OpenRouter, Groq and a
// local Ollama.
type AIConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"` // preset id shown in the UI
	BaseURL  string `json:"baseURL"`
	Model    string `json:"model"`
	// APIKeyEnc is the API key encrypted with Windows DPAPI
	// (only this Windows user can decrypt it), base64-encoded.
	APIKeyEnc string `json:"apiKeyEnc,omitempty"`
}

// Config lives in %LOCALAPPDATA%\StudyTracker\config.json.
type Config struct {
	Port int      `json:"port"`
	AI   AIConfig `json:"ai"`
}

func defaultConfig() Config {
	return Config{
		Port: defaultPort,
		AI: AIConfig{
			Provider: "gemini",
			BaseURL:  "https://generativelanguage.googleapis.com/v1beta/openai",
			Model:    "gemini-3.8-flash",
		},
	}
}

// ConfigStore is a thread-safe, file-backed Config.
type ConfigStore struct {
	path string
	mu   sync.RWMutex
	cfg  Config
}

// LoadConfig reads path, falling back to defaults when the file
// does not exist yet.
func LoadConfig(path string) (*ConfigStore, error) {
	cs := &ConfigStore{path: path, cfg: defaultConfig()}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cs, nil
	}
	if err != nil {
		return cs, err
	}
	if err := json.Unmarshal(data, &cs.cfg); err != nil {
		return cs, err
	}
	if cs.cfg.Port <= 0 {
		cs.cfg.Port = defaultPort
	}
	return cs, nil
}

func (cs *ConfigStore) Get() Config {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.cfg
}

// Update applies fn to a copy of the config and persists it
// atomically (write temp file, then rename).
func (cs *ConfigStore) Update(fn func(*Config)) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	next := cs.cfg
	fn(&next)
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	tmp := cs.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, cs.path); err != nil {
		return err
	}
	cs.cfg = next
	return nil
}

// APIKey returns the decrypted API key ("" if none).
// The STUDYTRACKER_AI_KEY environment variable takes precedence.
func (cs *ConfigStore) APIKey() string {
	if v := os.Getenv("STUDYTRACKER_AI_KEY"); v != "" {
		return v
	}
	enc := cs.Get().AI.APIKeyEnc
	if enc == "" {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return ""
	}
	plain, err := dpapiDecrypt(raw)
	if err != nil {
		return ""
	}
	return string(plain)
}

// keyHint is a masked version of the key for display.
func keyHint(key string) string {
	if key == "" {
		return ""
	}
	r := []rune(key)
	if len(r) <= 4 {
		return "••••"
	}
	return "••••" + string(r[len(r)-4:])
}

func encryptAPIKey(key string) (string, error) {
	if key == "" {
		return "", nil
	}
	enc, err := dpapiEncrypt([]byte(key))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(enc), nil
}

func dpapiEncrypt(data []byte) ([]byte, error) {
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

func dpapiDecrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty ciphertext")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}
