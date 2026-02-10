package main

import (
	"database/sql"
	"embed"

	_ "github.com/mattn/go-sqlite3"

	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./user_data.db")
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		active BOOLEAN NULL,
		expiration_date TEXT NULL
	)`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
}

func main() {
	// Initialize the database
	initDB()

	// Create an instance of the app structure
	app := NewApp()
	userService := NewUserService()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "app",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			userService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func saveUserData(id string, email string) error {
	query := `INSERT INTO users (id, email) VALUES (?, ?)`
	_, err := db.Exec(query, id, email)
	if err != nil {
		return fmt.Errorf("Erro ao salvar dados do usuário: %v", err)
	}
	return nil
}

func updateUserData(id string, active bool, expirationDate string) error {
	query := `UPDATE users SET active = ?, expiration_date = ? WHERE id = ?`
	_, err := db.Exec(query, active, expirationDate, id)
	if err != nil {
		return fmt.Errorf("Erro ao atualizar dados do usuário: %v", err)
	}
	return nil
}

func getUserStatus() (bool, error) {
	query := `SELECT id, email, active, expiration_date FROM users LIMIT 1`
	row := db.QueryRow(query)

	var id, email, expirationDate string
	var active bool

	err := row.Scan(&id, &email, &active, &expirationDate)
	if err == sql.ErrNoRows {
		return false, nil // Nenhum usuário encontrado
	} else if err != nil {
		return false, fmt.Errorf("Erro ao buscar usuário: %v", err)
	}

	// Verifica se o usuário está ativo e se a licença é válida
	if active {
		expirationTime, err := time.Parse(time.RFC3339, expirationDate)
		if err != nil {
			return false, fmt.Errorf("Erro ao analisar data de expiração: %v", err)
		}
		if time.Now().Before(expirationTime) {
			return true, nil // Usuário ativo e licença válida
		}
	}

	return false, nil // Usuário não está ativo ou licença expirada
}
