package application

import (
	"context"
	"log/slog"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/framework/service"
)

// Execute starts and runs the default application, and always waits for checked
// shutdown before returning (including startup/run failure and panic). Each
// drain attempt has a fresh 30-second budget independent of ctx cancellation.
// An incomplete shutdown is retried, not treated as permission to exit. Owners
// must cooperate with cancellation; this function cannot force callbacks to end.
func Execute(ctx context.Context, svc service.Service) *cd.Error {
	return execute(ctx, Get(), svc, 30*time.Second, time.Second)
}

func execute(ctx context.Context, app Application, svc service.Service, budget, retryDelay time.Duration) *cd.Error {
	defer func() {
		for {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), budget)
			err := app.ShutdownChecked(shutdownCtx)
			cancel()
			if err == nil {
				return
			}
			slog.Error("shutdown incomplete; retaining runtime and retrying", "error", err.Error())
			time.Sleep(retryDelay)
		}
	}()
	if err := app.Startup(ctx, svc); err != nil {
		return err
	}
	return app.Run(ctx)
}
