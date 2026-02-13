package main

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

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

func initDat() {
	dir, _ := os.UserConfigDir()
	appDir := filepath.Join(dir, "app")

	print("appDir::: ", appDir)

	_ = os.MkdirAll(appDir, 0700)

	licenseFile = filepath.Join(appDir, "license.dat")
	loadLicense()
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

	app := NewApp()
	userService := NewUserService()

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
		},
	})

	if err != nil {
		println(err.Error())
	}
}
