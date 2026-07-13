package main

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type LicenseData struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Active         bool   `json:"active"`
	ExpirationDate string `json:"expiration_date"`
}

var licenseFile string
var license LicenseData
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

	_ = os.MkdirAll(appDir, 0700)

	licenseFile = filepath.Join(appDir, "license.dat")
	loadLicense()

	dbPath := filepath.Join(appDir, "app.db")
	var err error
	db, err = database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
	log.Println("Banco de dados inicializado com sucesso")
}

func developmentMode() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("SAAS_DEV_MODE")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func dataDirectory() string {
	if configured := strings.TrimSpace(os.Getenv("SAAS_DATA_DIR")); configured != "" {
		return configured
	}
	dir, _ := os.UserConfigDir()
	if developmentMode() {
		return filepath.Join(dir, "saas-dev")
	}
	return filepath.Join(dir, "app")
}

func loadLicense() {
	file, err := os.Open(licenseFile)
	if os.IsNotExist(err) {
		saveLicense()
		return
	} else if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&license)
}

func saveLicense() {
	file, err := os.Create(licenseFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	json.NewEncoder(file).Encode(license)
}

func saveUserData(id string, email string) error {
	license = LicenseData{
		ID:             id,
		Email:          email,
		Active:         false,
		ExpirationDate: "",
	}

	saveLicense()
	return nil
}

func updateUserData(id string, active bool, expirationDate string) error {
	if license.ID != id {
		return nil
	}

	license.Active = active
	license.ExpirationDate = expirationDate

	saveLicense()
	return nil
}

func getUserStatus() (bool, error) {
	if developmentMode() {
		return true, nil
	}
	if !license.Active {
		return false, nil
	}

	exp, err := time.Parse(time.RFC3339, license.ExpirationDate)
	if err != nil {
		return false, err
	}

	if time.Now().After(exp) {
		return false, nil
	}

	return true, nil
}

func main() {
	initDat()
	defer db.Close()

	app := NewApp()
	userService := NewUserService()
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
			userService,
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
