package main

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
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
	dir, _ := os.UserConfigDir()
	appDir := filepath.Join(dir, "app")

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
	productService := service.NewProductService(db)
	saleService := service.NewSaleService(db)
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
			productService,
			saleService,
			databaseService,
		},
	})

	if err != nil {
		println(err.Error())
	}
}
