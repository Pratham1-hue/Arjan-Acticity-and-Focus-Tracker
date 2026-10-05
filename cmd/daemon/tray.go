package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

// dashboardTitle must match the page <title>; it's how an already
// open dashboard window is found and brought to the front.
const dashboardTitle = "Arjan Dashboard"

var (
	colorActive = color.NRGBA{0x34, 0xd3, 0x99, 0xff} // emerald: tracking
	colorPaused = color.NRGBA{0x71, 0x71, 0x7a, 0xff} // zinc: paused
)

// Tray is the system-tray UI: the app's only visible surface
// besides the dashboard window.
type Tray struct {
	tracker  *Tracker
	analyzer *Analyzer
	server   *Server
	notifier Notifier
	dataDir  string
	lg       *log.Logger
	quit     func()
}

func (t *Tray) onReady() {
	systray.SetIcon(trayIcon(colorActive))
	systray.SetTooltip("Arjan — tracking")

	mStatus := systray.AddMenuItem("Arjan", "")
	mStatus.Disable()
	systray.AddSeparator()
	mOpen := systray.AddMenuItem("Open Dashboard", "Open the Arjan dashboard")
	systray.AddSeparator()
	mPause := systray.AddMenuItemCheckbox("Pause tracking", "Stop recording until resumed", t.tracker.Paused())
	mAuto := systray.AddMenuItemCheckbox("Start with Windows", "Launch Arjan when you sign in", autostartEnabled())
	systray.AddSeparator()
	mReport := systray.AddMenuItem("Show today's report", "Show today's summary as a notification")
	mData := systray.AddMenuItem("Open data folder", "Database, logs and settings")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Arjan", "Stop tracking and exit")

	// Left click opens the dashboard; right click shows the menu.
	systray.SetOnTapped(t.openDashboard)

	applyPaused := func(p bool) {
		if p {
			mPause.Check()
			systray.SetIcon(trayIcon(colorPaused))
		} else {
			mPause.Uncheck()
			systray.SetIcon(trayIcon(colorActive))
		}
		t.refreshStatus(mStatus)
	}
	t.tracker.OnPauseChange(applyPaused)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		t.refreshStatus(mStatus)
		for {
			select {
			case <-ticker.C:
				t.refreshStatus(mStatus)
			case <-mOpen.ClickedCh:
				t.openDashboard()
			case <-mPause.ClickedCh:
				t.tracker.SetPaused(!t.tracker.Paused())
			case <-mAuto.ClickedCh:
				on := !autostartEnabled()
				if err := setAutostart(on); err != nil {
					t.lg.Printf("autostart: %v", err)
				}
				if autostartEnabled() {
					mAuto.Check()
				} else {
					mAuto.Uncheck()
				}
			case <-mReport.ClickedCh:
				go func() {
					msg, err := dailyReport(t.analyzer)
					if err != nil {
						msg = err.Error()
					}
					t.notifier("Arjan — today so far", msg)
				}()
			case <-mData.ClickedCh:
				exec.Command("explorer.exe", t.dataDir).Start()
			case <-mQuit.ClickedCh:
				t.quit()
				return
			}
		}
	}()
}

// refreshStatus updates the first menu line and the tooltip with
// today's totals.
func (t *Tray) refreshStatus(item *systray.MenuItem) {
	state := "tracking"
	if t.tracker.Paused() {
		state = "paused"
	}
	rep, err := t.analyzer.Day(time.Now())
	if err != nil || rep.Total == 0 {
		item.SetTitle("Arjan — " + state)
		systray.SetTooltip("Arjan — " + state)
		return
	}
	line := fmt.Sprintf("Today %s", fmtDur(rep.Total))
	if rep.FocusScore != nil {
		line += fmt.Sprintf(" · Focus %d", *rep.FocusScore)
	}
	item.SetTitle(line)
	systray.SetTooltip(fmt.Sprintf("Arjan — %s\n%s", state, line))
}

func (t *Tray) openDashboard() { openDashboardURL(t.server.URL()) }

// openDashboardURL focuses an open dashboard window, or opens one
// as a chromeless Edge "app" window (Edge ships with Windows), or
// falls back to the default browser.
func openDashboardURL(url string) {
	if focusWindowByTitle(dashboardTitle) {
		return
	}
	if edge := findEdge(); edge != "" {
		if exec.Command(edge, "--app="+url, "--window-size=1320,900").Start() == nil {
			return
		}
	}
	exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}

func findEdge() string {
	for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
		if base == "" {
			continue
		}
		p := filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// -------------------------------------------------------------- autostart ---

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue   = "Arjan"
)

// autostartEnabled reports whether the HKCU Run entry exists (it
// also shows up in Task Manager → Startup apps).
func autostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetStringValue(runValue)
	return err == nil && v != ""
}

func setAutostart(on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		if err := k.DeleteValue(runValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(runValue, `"`+exe+`"`)
}

// ------------------------------------------------------------------- icon ---

// trayIcon draws a "focus target" (ring + dot) in c and returns it
// as a multi-size .ico, so no icon file has to ship with the app.
func trayIcon(c color.NRGBA) []byte {
	sizes := []int{16, 24, 32, 48}
	pngs := make([][]byte, len(sizes))
	for i, s := range sizes {
		pngs[i] = renderIcon(s, c)
	}
	return encodeICO(sizes, pngs)
}

func renderIcon(size int, c color.NRGBA) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	center := float64(size) / 2
	outer := center - 0.5
	ring := float64(size) * 0.16
	dot := float64(size) * 0.2
	const ss = 4 // 4x4 supersampling for smooth edges
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			hits := 0
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					px := float64(x) + (float64(sx)+0.5)/ss
					py := float64(y) + (float64(sy)+0.5)/ss
					d := math.Hypot(px-center, py-center)
					if (d <= outer && d >= outer-ring) || d <= dot {
						hits++
					}
				}
			}
			if hits > 0 {
				a := uint8(int(c.A) * hits / (ss * ss))
				img.SetNRGBA(x, y, color.NRGBA{c.R, c.G, c.B, a})
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// encodeICO wraps PNG images in an ICO container (PNG-compressed
// icons are supported since Windows Vista).
func encodeICO(sizes []int, pngs [][]byte) []byte {
	var b bytes.Buffer
	le := binary.LittleEndian
	binary.Write(&b, le, uint16(0))          // reserved
	binary.Write(&b, le, uint16(1))          // type: icon
	binary.Write(&b, le, uint16(len(sizes))) // image count
	offset := 6 + 16*len(sizes)
	for i, s := range sizes {
		dim := byte(s)
		if s >= 256 {
			dim = 0
		}
		b.Write([]byte{dim, dim, 0, 0})
		binary.Write(&b, le, uint16(1))  // color planes
		binary.Write(&b, le, uint16(32)) // bits per pixel
		binary.Write(&b, le, uint32(len(pngs[i])))
		binary.Write(&b, le, uint32(offset))
		offset += len(pngs[i])
	}
	for _, p := range pngs {
		b.Write(p)
	}
	return b.Bytes()
}
