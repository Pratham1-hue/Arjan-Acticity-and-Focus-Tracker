package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeRules(t *testing.T, json string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "rules.json")
	if err := os.WriteFile(p, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const testRules = `{
  "categories": [
    {"name": "DSA Practice", "kind": "productive", "match": ["leetcode.com"]},
    {"name": "Data Science / AI", "kind": "productive", "match": ["IIT Madras", "huggingface.co", "chroma", "ollama", "jupyter"]},
    {"name": "Web Dev", "kind": "productive", "match": ["OpenCode", "VS Code", "Visual Studio Code", "code.exe", "localhost:3000", "fastapi", "next.js"]},
    {"name": "Productive Research", "kind": "productive", "match": ["gemini.google.com", "claude.ai"]},
    {"name": "Music", "kind": "neutral", "match": ["open.spotify.com"]},
    {"name": "Distractions", "kind": "distraction", "match": ["instagram.com", "youtube.com/shorts"]}
  ]
}`

func TestRuleEngineCategorize(t *testing.T) {
	e := NewRuleEngine(writeRules(t, testRules))
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}

	cases := []struct{ app, title, url, want string }{
		{"chrome.exe", "Two Sum - LeetCode", "https://leetcode.com/problems/two-sum/", "DSA Practice"},
		{"chrome.exe", "IIT Madras - Deep Learning Lecture", "https://www.youtube.com/watch?v=x", "Data Science / AI"},
		{"chrome.exe", "Hugging Face", "https://huggingface.co/datasets", "Data Science / AI"},
		{"chrome.exe", "Jupyter", "http://localhost:8888/tree", "Data Science / AI"},
		{"code.exe", "app.js - Visual Studio Code", "", "Web Dev"},
		{"chrome.exe", "My App", "http://localhost:3000/dashboard", "Web Dev"},
		{"chrome.exe", "Gemini", "https://gemini.google.com/app", "Productive Research"},
		{"chrome.exe", "Claude", "https://claude.ai/new", "Productive Research"},
		{"chrome.exe", "Spotify Web Player", "https://open.spotify.com/", "Music"},
		{"chrome.exe", "Shorts", "https://www.youtube.com/shorts/abc", "Distractions"},
		{"chrome.exe", "Instagram", "https://www.instagram.com/", "Distractions"},
		// plain YouTube watch pages are NOT distractions per the rules
		{"chrome.exe", "Cat Video", "https://www.youtube.com/watch?v=x", "Uncategorized"},
		{"notepad.exe", "Untitled - Notepad", "", "Uncategorized"},
	}
	for _, c := range cases {
		if got := e.Categorize(c.app, c.title, c.url); got != c.want {
			t.Errorf("Categorize(%q, %q, %q) = %q, want %q", c.app, c.title, c.url, got, c.want)
		}
	}

	if k := e.KindOf("Distractions"); k != kindDistraction {
		t.Errorf("KindOf(Distractions) = %q, want %q", k, kindDistraction)
	}
	if k := e.KindOf("Music"); k != "neutral" {
		t.Errorf("KindOf(Music) = %q, want neutral", k)
	}
	if k := e.KindOf("Nope"); k != "neutral" {
		t.Errorf("KindOf(Nope) = %q, want neutral", k)
	}
}

func TestRuleEngineHotReload(t *testing.T) {
	p := writeRules(t, `{"categories":[{"name":"A","kind":"productive","match":["aaa"]}]}`)
	e := NewRuleEngine(p)
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := e.Categorize("chrome.exe", "t", "http://x/aaa"); got != "A" {
		t.Fatalf("got %q, want A", got)
	}

	if err := os.WriteFile(p, []byte(`{"categories":[{"name":"B","kind":"distraction","match":["bbb"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(p, future, future); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := e.Categorize("chrome.exe", "t", "http://x/bbb"); got != "B" {
		t.Fatalf("after reload got %q, want B", got)
	}
}

func TestDistractionNudge(t *testing.T) {
	e := NewRuleEngine(writeRules(t, `{"categories":[{"name":"Distractions","kind":"distraction","match":["instagram.com"]}]}`))
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}

	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var notified []string
	tr := NewTracker(store, e, func(title, msg string) { notified = append(notified, msg) }, nil)

	now := time.Now().Unix()
	tr.mu.Lock()
	tr.fg = windowInfo{exe: "chrome.exe"} // Chrome is the foreground window
	tr.mu.Unlock()

	// Instagram becomes the active tab: the streak starts.
	tr.OnBrowserTab("https://www.instagram.com/", "Instagram", false, "chrome.exe")
	if len(notified) != 0 {
		t.Fatalf("nudged too early: %v", notified)
	}

	// 14 minutes in the same tab: still no nudge.
	tr.mu.Lock()
	tr.distractSince = now - 14*60
	tr.mu.Unlock()
	tr.OnBrowserTab("https://www.instagram.com/", "Instagram", false, "chrome.exe")
	if len(notified) != 0 {
		t.Fatalf("nudged at 14 minutes: %v", notified)
	}

	// 16 minutes: the nudge fires exactly once.
	tr.mu.Lock()
	tr.distractSince = now - 16*60
	tr.mu.Unlock()
	tr.OnBrowserTab("https://www.instagram.com/", "Instagram", false, "chrome.exe")
	if len(notified) != 1 {
		t.Fatalf("nudge count = %d, want 1 (%v)", len(notified), notified)
	}

	// Switching to another site resets the streak.
	tr.OnBrowserTab("https://leetcode.com/", "LeetCode", false, "chrome.exe")
	tr.mu.Lock()
	if tr.distractSince != 0 || tr.nudged {
		t.Fatalf("streak not reset: since=%d nudged=%v", tr.distractSince, tr.nudged)
	}
	tr.mu.Unlock()

	// Back on Instagram: the streak restarts from zero.
	tr.OnBrowserTab("https://www.instagram.com/", "Instagram", false, "chrome.exe")
	if len(notified) != 1 {
		t.Fatalf("nudge count = %d, want 1", len(notified))
	}
}

func TestDailyReport(t *testing.T) {
	e := NewRuleEngine(writeRules(t, `{"categories":[
		{"name":"DSA Practice","kind":"productive","match":["leetcode.com"]},
		{"name":"Distractions","kind":"distraction","match":["instagram.com"]},
		{"name":"Music","kind":"neutral","match":["spotify"]}
	]}`))
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}

	store, err := OpenStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	rows := []session{
		// yesterday: must be ignored
		{start: midnight - 7200, end: midnight - 3600, app: "chrome.exe", url: "https://leetcode.com/", category: "DSA Practice"},
		// today: 1h productive
		{start: midnight + 100, end: midnight + 3700, app: "chrome.exe", url: "https://leetcode.com/", category: "DSA Practice"},
		// today: 30m distraction
		{start: midnight + 200, end: midnight + 2000, app: "chrome.exe", url: "https://instagram.com/", category: "Distractions"},
		// today: 20m other
		{start: midnight + 300, end: midnight + 1500, app: "spotify.exe", title: "Spotify", category: "Music"},
	}
	for i := range rows {
		if _, err := store.Insert(&rows[i]); err != nil {
			t.Fatal(err)
		}
	}

	msg, err := dailyReport(NewAnalyzer(store, e, nil))
	if err != nil {
		t.Fatal(err)
	}
	want := "Today: 1h 00m productive, 30m distraction, 20m other. Top: DSA Practice 1h 00m"
	if msg != want {
		t.Fatalf("dailyReport = %q, want %q", msg, want)
	}
}
