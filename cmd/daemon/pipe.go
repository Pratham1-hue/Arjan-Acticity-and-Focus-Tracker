package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// The browser launches nmhost.exe (a separate, short-lived process) for the
// extension. nmhost forwards each message to this long-running daemon over a
// local named pipe, one JSON object per line.
const pipeName = `\\.\pipe\StudyTrackerBrowser`

// inbound is the wire format between nmhost and the daemon. It is also the
// format the Phase 2 extension sends.
//
//	{"type":"tab","url":"https://...","title":"...","incognito":false,"browser":"chrome.exe"}
//	{"type":"shutdown"}                     (used by `tracker.exe --stop`)
//
// "browser" is added by nmhost (it detects which browser launched it).
type inbound struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Incognito bool   `json:"incognito"`
	Browser   string `json:"browser"`
}

func serveBrowserPipe(ctx context.Context, t *Tracker, onShutdown func(), lg *log.Logger) error {
	l, err := winio.ListenPipe(pipeName, &winio.PipeConfig{
		InputBufferSize:  64 * 1024,
		OutputBufferSize: 4 * 1024,
	})
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		c, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			lg.Printf("pipe accept: %v", err)
			time.Sleep(time.Second) // avoid a hot loop on a persistent error
			continue
		}
		go handlePipeConn(c, t, onShutdown, lg)
	}
}

func handlePipeConn(c net.Conn, t *Tracker, onShutdown func(), lg *log.Logger) {
	defer c.Close()
	sc := bufio.NewScanner(c)
	sc.Buffer(make([]byte, 0, 4*1024), 1<<20)
	for sc.Scan() {
		var m inbound
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			continue
		}
		switch m.Type {
		case "tab":
			t.OnBrowserTab(m.URL, m.Title, m.Incognito, strings.ToLower(m.Browser))
		case "shutdown":
			lg.Printf("shutdown requested")
			onShutdown()
			return
		}
	}
}

// sendStop asks a running daemon to exit cleanly. Returns a process exit code.
func sendStop() int {
	timeout := time.Second
	c, err := winio.DialPipe(pipeName, &timeout)
	if err != nil {
		return 1
	}
	defer c.Close()
	if _, err := c.Write([]byte("{\"type\":\"shutdown\"}\n")); err != nil {
		return 1
	}
	return 0
}
