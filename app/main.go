package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var db *database.Database

func initDat() {
	if bindingGeneration {
		var err error
		db, err = database.NewDatabase(":memory:")
		if err != nil {
			log.Fatalf("failed to initialise binding-generation database: %v", err)
		}
		return
	}
	appDir := dataDirectory()

	if err := os.MkdirAll(appDir, 0700); err != nil {
		log.Fatalf("failed to create application data directory: %v", err)
	}

	dbPath := filepath.Join(appDir, "app.db")
	var err error
	db, err = database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
	log.Println("Banco de dados inicializado com sucesso")
}

func dataDirectory() string {
	if configured := strings.TrimSpace(os.Getenv("SAAS_DATA_DIR")); configured != "" {
		return configured
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "app")
}

func main() {
	initDat()
	defer db.Close()

	app := NewApp()
	itemService := service.NewItemService(db)
	batchService := service.NewBatchService(db)
	recipeService := service.NewRecipeService(db)
	purchaseService := service.NewPurchaseService(db)
	databaseService := service.NewDatabaseService(db)
	app.DatabaseService = databaseService

	err := wails.Run(&options.App{
		Title:  "app",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
			itemService,
			batchService,
			recipeService,
			purchaseService,
			databaseService,
		},
	})

	if err != nil {
		println(err.Error())
	}
}
