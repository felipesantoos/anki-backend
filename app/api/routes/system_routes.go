package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterSystemRoutes registers system-related routes (addons, backups, media, sync)
func (r *Router) RegisterSystemRoutes() {
	addOnService := dicontainer.GetAddOnService()
	backupService := dicontainer.GetBackupService()
	mediaService := dicontainer.GetMediaService()
	syncMetaService := dicontainer.GetSyncMetaService()

	addOnHandler := handlers.NewAddOnHandler(addOnService)
	backupHandler := handlers.NewBackupHandler(backupService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	syncMetaHandler := handlers.NewSyncMetaHandler(syncMetaService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(r.jwtSvc, r.rdb)

	// System group
	v1 := r.echo.Group("/api/v1", authMiddleware)

	// Add-ons
	addons := v1.Group("/addons")
	addons.POST("", addOnHandler.Install)
	addons.GET("", addOnHandler.FindAll)
	addons.PUT("/:code/config", addOnHandler.UpdateConfig)
	addons.POST("/:code/toggle", addOnHandler.Toggle)
	addons.DELETE("/:code", addOnHandler.Uninstall)

	// Backups
	backups := v1.Group("/backups")
	backups.POST("", backupHandler.Create)
	backups.GET("", backupHandler.FindAll)
	backups.DELETE("/:id", backupHandler.Delete)

	// Media
	media := v1.Group("/media")
	media.POST("", mediaHandler.Create)
	media.GET("", mediaHandler.FindAll)
	media.GET("/:id", mediaHandler.FindByID)
	media.DELETE("/:id", mediaHandler.Delete)

	// Sync
	sync := v1.Group("/sync")
	sync.GET("/meta", syncMetaHandler.FindMe)
	sync.PUT("/meta", syncMetaHandler.Update)
}

