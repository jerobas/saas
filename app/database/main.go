package main

import (
	"log"
	"path/filepath"
)

func main() {
	dbPath := filepath.Join(".", "app.db")
	_, err := NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
}