package main // import "github.com/jmc-audio/kitsession"

import (
	"fmt"
	stdlog "log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/davecgh/go-spew/spew"
	kitlog "github.com/go-kit/kit/log"
	"github.com/jmc-audio/kitsession/bindings"
	"github.com/jmc-audio/kitsession/log"
)

func main() {
	var logger kitlog.Logger
	logger = kitlog.NewLogfmtLogger(os.Stderr)
	logger = kitlog.NewContext(logger).With("ts", kitlog.DefaultTimestampUTC)
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger)) // redirect stdlib logging to us
	stdlog.SetFlags(0)
	logger.Log("info", "kitsession")

	// Mechanical stuff
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()

	errc := make(chan error)
	go func() {
		errc <- interrupt()
	}()

	ctx = context.WithValue(ctx, "logger", logger)
	ctx = context.WithValue(ctx, "errc", errc)

	log.Logger(ctx).Debug().Log("ctx", spew.Sdump(ctx))

	// HTTP REST Endpoint Listeners
	bindings.StartHTTPListener(ctx)

	logger.Log("fatal", <-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
