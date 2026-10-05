// Command nmhost is the Native Messaging bridge between the browser
// extension and the StudyTracker daemon.
//
// Chrome/Edge launches this executable (see host/com.studytracker.json)
// whenever the extension sends a message. Messages arrive on stdin in
// the Chrome native-messaging framing: a 4-byte little-endian length
// followed by that many bytes of JSON. Each message is forwarded, as a
// single newline-terminated JSON object, to the daemon's named pipe.
//
// The process stays alive as long as Chrome keeps stdin open, so a
// long-lived `connectNative()` port costs only one process launch.
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

// Must match pipeName in cmd/daemon/pipe.go.
const pipeName = `\\.\pipe\StudyTrackerBrowser`

const maxMessageBytes = 1 << 20 // Chrome caps messages at 1 MB

func main() {
	// Silent by default; errors go to a log file so a failure never
	// surfaces as a browser error popup.
	lg := log.New(io.Discard, "", 0)
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		if f, err := os.OpenFile(filepath.Join(dir, "StudyTracker", "nmhost.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
			defer f.Close()
			lg = log.New(f, "nmhost ", log.LstdFlags)
		}
	}

	browser := detectBrowser()
	lg.Printf("started for browser=%q", browser)

	for {
		payload, err := readMessage(os.Stdin)
		if err != nil {
			if err != io.EOF {
				lg.Printf("read: %v", err)
			}
			return
		}
		payload = tagBrowser(payload, browser)
		if err := forward(payload); err != nil {
			lg.Printf("forward: %v", err) // daemon not running; Chrome retries on next message
			return
		}
	}
}

// detectBrowser walks up the process tree to find which browser
// started us (Chrome launches native hosts through cmd.exe), so the
// daemon can tell Chrome tabs from Edge tabs. "" if unknown.
func detectBrowser() string {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(snap)

	type proc struct {
		parent uint32
		exe    string
	}
	procs := map[uint32]proc{}
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		procs[e.ProcessID] = proc{e.ParentProcessID, strings.ToLower(windows.UTF16ToString(e.ExeFile[:]))}
	}

	pid := uint32(os.Getpid())
	for i := 0; i < 5; i++ {
		self, ok := procs[pid]
		if !ok {
			return ""
		}
		parent, ok := procs[self.parent]
		if !ok {
			return ""
		}
		if parent.exe == "cmd.exe" || parent.exe == "conhost.exe" {
			pid = self.parent
			continue
		}
		return parent.exe
	}
	return ""
}

// tagBrowser adds "browser":"chrome.exe" to a JSON message.
func tagBrowser(payload []byte, browser string) []byte {
	if browser == "" {
		return payload
	}
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return payload
	}
	m["browser"] = browser
	if out, err := json.Marshal(m); err == nil {
		return out
	}
	return payload
}

func readMessage(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint32(hdr[:])
	if n == 0 || n > maxMessageBytes {
		return nil, fmt.Errorf("bad message length %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func forward(payload []byte) error {
	timeout := 2 * time.Second
	c, err := winio.DialPipe(pipeName, &timeout)
	if err != nil {
		return fmt.Errorf("daemon not running: %w", err)
	}
	defer c.Close()
	// The daemon reads one JSON object per line.
	if _, err := c.Write(append(payload, '\n')); err != nil {
		return err
	}
	return nil
}
