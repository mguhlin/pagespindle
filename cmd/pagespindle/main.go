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
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	if os.Getenv("PAGESPINDLE_ENGINE_SUPERVISOR") == "1" {
		os.Exit(engine.Supervise(os.Stdin))
	}
	extract := flag.Bool("extract-pdf", false, "internal bounded PDF extraction worker")
	root := flag.String("assets", "", "folder containing runtime/ and models/")
	noBrowser := flag.Bool("no-browser", false, "print the local URL without opening a browser")
	port := flag.Int("port", 0, "local port (default: choose an available port)")
	flag.Parse()
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
	httpServer := &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	url := "http://" + host + "/#" + token
	fmt.Println("PageSpindle — Spin text into listening")
	fmt.Println("Open this private local session in your browser:")
	fmt.Println(url)
	fmt.Println("Keep this launcher open. Press Ctrl+C here to stop and remove temporary audio.")
	if !*noBrowser {
		var c *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "darwin":
			c = exec.Command("open", url)
		default:
			c = exec.Command("xdg-open", url)
		}
		if e = c.Start(); e == nil {
			go c.Wait()
		}
	}
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
