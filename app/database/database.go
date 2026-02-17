package database

import (
	"embed"
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

type Database struct {
	Conn *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{Conn: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		log.Printf("Erro ao ler schema.sql: %v", err)
		return err
	}

	if _, err := d.Conn.Exec(string(schema)); err != nil {
		log.Printf("Erro ao criar tabelas: %v", err)
		return err
	}

	return nil
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

func (d *Database) GetConnection() *sql.DB {
	return d.Conn
}