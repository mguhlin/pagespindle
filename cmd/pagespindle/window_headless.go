//go:build !desktop

package main

import (
	"errors"
	"github.com/mguhlin/pagespindle/internal/app"
)

func runNativeWindow(url string, server *app.Server, start func(), stop func()) error {
	return errors.New("This is the headless worker. Launch the installed PageSpindle desktop application")
}
