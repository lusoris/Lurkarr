package main

import (
	"os"
	"strings"

	"go.uber.org/fx"
)

func main() {
	opts := []fx.Option{
		configModule,
		databaseModule,
		loggingModule,
		notificationsModule,
	}

	if isRunOnce() {
		opts = append(opts, runOnceModule)
	} else {
		opts = append(opts,
			schedulerModule,
			serverModule,
			servicesModule,
			maintenanceModule,
		)
	}

	fx.New(opts...).Run()
}

func isRunOnce() bool {
	v := strings.ToLower(os.Getenv("RUN_ONCE"))
	return v == "true" || v == "1" || v == "yes"
}
