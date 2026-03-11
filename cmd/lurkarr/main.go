package main

import "go.uber.org/fx"

func main() {
	fx.New(
		configModule,
		databaseModule,
		loggingModule,
		notificationsModule,
		schedulerModule,
		serverModule,
		servicesModule,
		maintenanceModule,
	).Run()
}
