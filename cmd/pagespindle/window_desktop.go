//go:build desktop

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mguhlin/pagespindle/internal/app"
	"github.com/mguhlin/pagespindle/internal/provision"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
)

func runNativeWindow(url string, server *app.Server, start func(), stop func()) error {
	var force, prompt atomic.Bool
	confirm := func() bool { return true }
	profile, err := os.MkdirTemp("", "pagespindle-view-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(profile)
	win := application.WindowsOptions{WebviewUserDataPath: profile}
	if runtime.GOOS == "windows" {
		fixed := filepath.Join(server.Engine.Root, "gui", "webview2")
		if info, e := os.Stat(fixed); e == nil && info.IsDir() {
			win.WebviewBrowserPath = fixed
		}
	}
	models := filepath.Join(server.Engine.Root, "models")
	if server.Engine.ModelRoot != "" {
		models = server.Engine.ModelRoot
	}
	if info, e := os.Stat(filepath.Join(models, "Qwen3-1.7B-Q8_0.gguf")); e != nil || !info.Mode().IsRegular() {
		var data string
		if runtime.GOOS == "linux" {
			data = os.Getenv("XDG_DATA_HOME")
			if !filepath.IsAbs(data) {
				home, _ := os.UserHomeDir()
				data = filepath.Join(home, ".local", "share")
			}
		} else if runtime.GOOS == "windows" {
			data = os.Getenv("LOCALAPPDATA")
			if data == "" {
				data, _ = os.UserCacheDir()
			}
		} else {
			data, _ = os.UserConfigDir()
		}
		models = filepath.Join(data, "PageSpindle", "model-pack-"+provision.PackSHA[:12], "models")
		if !provision.Ready(models) {
			server.Setup = provision.New(models)
		}
	}
	server.Engine.ModelRoot = models
	if server.Setup != nil {
		defer server.Setup.Close()
		url = strings.Replace(url, "/#", "/setup.html#", 1)
	}
	shutdown := func() {
		if server.Setup != nil {
			server.Setup.Close()
		}
		stop()
	}
	native := application.New(application.Options{Icon: nativeIcon, Name: "PageSpindle", Description: "Spin text into listening", Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), DisableDefaultSignalHandler: false, OnShutdown: shutdown, Windows: win, ShouldQuit: func() bool { return confirm() }, Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true}})
	window := native.Window.NewWithOptions(application.WebviewWindowOptions{Name: "PageSpindle", Title: "PageSpindle", URL: url, Width: 1280, Height: 900, MinWidth: 620, MinHeight: 540, EnableFileDrop: false})
	server.Desktop = &app.DesktopHooks{OpenHelp: native.Browser.OpenURL, ChooseSave: func(name string) (string, error) {
		return native.Dialog.SaveFile().AttachToWindow(window).SetFilename(name).SetMessage("Save a new PageSpindle file. Existing files are preserved.").PromptForSingleSelection()
	}}
	confirm = func() bool {
		if force.Load() || !server.DesktopDirty() {
			return true
		}
		if prompt.CompareAndSwap(false, true) {
			dialog := native.Dialog.Question().SetTitle("Close PageSpindle?").SetMessage("Unsaved text and temporary audio will be cleared. Save a project backup first if you want to keep your work.").AttachToWindow(window)
			dialog.AddButton("Keep working").SetAsDefault().SetAsCancel().OnClick(func() { prompt.Store(false) })
			dialog.AddButton("Close PageSpindle").OnClick(func() { force.Store(true); native.Quit() })
			dialog.Show()
		}
		return false
	}
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if !confirm() {
			e.Cancel()
		}
	})
	if nativeTestDirectory != "" {
		entries, e := os.ReadDir(nativeTestDirectory)
		if e != nil || len(entries) != 0 {
			return fmt.Errorf("native self-test needs an existing empty output directory")
		}
		server.Desktop.ChooseSave = func(name string) (string, error) { return filepath.Join(nativeTestDirectory, filepath.Base(name)), nil }
		server.SelfTestScript = "const PILOT_KEY=location.hash.slice(1);window.addEventListener('load',()=>{" + nativeWorkspacePilot + "});"
		server.SelfTest = func(ok bool, message string) {
			b, _ := json.Marshal(map[string]any{"ok": ok, "message": message})
			os.WriteFile(filepath.Join(nativeTestDirectory, "native-window-report.json"), b, 0600)
			force.Store(true)
			go native.Quit()
		}
	}
	start()
	return native.Run()
}
