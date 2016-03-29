package log

import (
	kitlog "github.com/go-kit/kit/log"
	levlog "github.com/go-kit/kit/log/levels"
	"golang.org/x/net/context"
)

func Logger(ctx context.Context) levlog.Levels {
	return levlog.New(ctx.Value("logger").(kitlog.Logger))
}
