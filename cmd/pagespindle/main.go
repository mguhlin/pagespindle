package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/mguhlin/pagespindle/internal/app"
	"github.com/mguhlin/pagespindle/internal/book"
	"github.com/mguhlin/pagespindle/internal/engine"
	"io"
	"net"
	"net/http"
	"os"

	"os/signal"
	"path/filepath"

	"github.com/mguhlin/pagespindle/internal/provision"
	"sync"
	"time"
)

var nativeTestDirectory string

func main() {
	if os.Getenv("PAGESPINDLE_ENGINE_SUPERVISOR") == "1" {
		os.Exit(engine.Supervise(os.Stdin))
	}
	extract := flag.Bool("extract-pdf", false, "internal bounded PDF extraction worker")
	root := flag.String("assets", "", "folder containing runtime/ and models/")
	noBrowser := flag.Bool("no-browser", false, "print the local URL without opening a browser")
	port := flag.Int("port", 0, "local port (default: choose an available port)")
	testDir := flag.String("native-self-test", "", "run a fixed synthetic workspace pilot in the native window")
	cpuOnly := flag.Bool("cpu", false, "use CPU for AI preparation instead of automatic GPU selection")
	modelsDir := flag.String("models", "", "use an already installed local model folder")
	installModels := flag.String("install-models-dir", "", "install pinned models into this folder")
	flag.Parse()
	if *installModels != "" {
		target, err := filepath.Abs(*installModels)
		if err != nil {
			os.Exit(2)
		}
		if provision.Ready(target) {
			return
		}
		manager := provision.New(target)
		manager.Start()
		last := ""
		for {
			state := manager.Snapshot()
			message := fmt.Sprintf("%s: %d%%", state.Phase, state.Bytes*100/provision.PackBytes)
			if message != last {
				fmt.Println(message)
				last = message
			}
			if state.Phase == "complete" {
				manager.Close()
				return
			}
			if state.Phase == "failed" || state.Phase == "cancelled" {
				fmt.Fprintln(os.Stderr, state.Message)
				manager.Close()
				os.Exit(2)
			}
			time.Sleep(time.Second)
		}
	}
	if *testDir != "" {
		nativeTestDirectory, _ = filepath.Abs(*testDir)
	}
	if *extract {
		b, e := io.ReadAll(io.LimitReader(os.Stdin, book.MaxInput+1))
		if e != nil || len(b) > book.MaxInput {
			os.Exit(2)
		}
		text, e := book.PDF(b)
		if e != nil {
			os.Exit(3)
		}
		fmt.Print(text)
		return
	}
	exe, e := os.Executable()
	if e != nil {
		fmt.Fprintln(os.Stderr, "Cannot locate PageSpindle")
		os.Exit(1)
	}
	if *root == "" {
		*root = filepath.Dir(exe)
		resources := filepath.Join(*root, "..", "Resources")
		if info, err := os.Stat(filepath.Join(resources, "runtime")); err == nil && info.IsDir() {
			*root = resources
		}
	}
	*root, e = filepath.Abs(*root)
	if e != nil {
		os.Exit(1)
	}
	l, e := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", *port))
	if e != nil {
		fmt.Fprintln(os.Stderr, "Cannot open a local port. Close another PageSpindle instance and retry.")
		os.Exit(1)
	}
	key := make([]byte, 32)
	if _, e = rand.Read(key); e != nil {
		os.Exit(1)
	}
	token := hex.EncodeToString(key)
	host := l.Addr().String()
	s := app.New(token, host, *root)
	s.Engine.ForceCPU = *cpuOnly
	if *modelsDir != "" {
		s.Engine.ModelRoot, _ = filepath.Abs(*modelsDir)
	}
	httpServer := &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	url := "http://" + host + "/#" + token
	var once sync.Once
	stop := func() {
		once.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			httpServer.Shutdown(ctx)
			s.Close()
		})
	}
	if !*noBrowser {
		if err := runNativeWindow(url, s, func() { go httpServer.Serve(l) }, stop); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		stop()
		return
	}
	fmt.Println("PageSpindle — Spin text into listening")
	fmt.Println(url)
	fmt.Println("Headless diagnostic session. Press Ctrl+C to stop.")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
		s.Close()
	}()
	if e = httpServer.Serve(l); e != nil && e != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "The local interface stopped unexpectedly")
	}
	s.Close()
}
