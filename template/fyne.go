package template

import (
	"cheetah/config"
	"cheetah/internal/handler/gui"
	"cheetah/service"
	"context"
	"time"
)

// ShowMain shows the main GUI using the new handler architecture
func ShowMain(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	// Create orchestration service
	serviceFactory := service.NewFactory(cfg)
	orchestrationService := serviceFactory.CreateOrchestrationService()

	// Create GUI handler
	guiFactory := gui.NewFactory(cfg)
	guiHandler := guiFactory.CreateDefaultHandler(orchestrationService)

	// Show main window
	if err := guiHandler.ShowMainWindow(ctx); err != nil {
		panic(err)
	}
}
