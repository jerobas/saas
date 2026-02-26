package database
// package main

import (
	"database/sql"
	"embed"
	"log"
	"slices"
	"io/fs"

	_ "modernc.org/sqlite"
)

//go:embed schemas/*.sql
var schemaFS embed.FS

type Database struct {
	Conn *sql.DB
	path string
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{Conn: db, path: dbPath}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
	schema, err := schemaFS.ReadDir("schemas")
	if err != nil {
		log.Printf("Erro ao ler diretório schemas: %v", err)
		return err
	}

	slices.SortFunc(schema, func(a, b fs.DirEntry) int {
		if a.Name() < b.Name() {
			return -1
		}
		if a.Name() > b.Name() {
			return 1
		}
		return 0
	})

	for _, file := range schema {
		content, err := schemaFS.ReadFile("schemas/" + file.Name())
		if err != nil {
			log.Printf("Erro ao ler arquivo %s: %v", file.Name(), err)
			return err
		}
		if _, err := d.Conn.Exec(string(content)); err != nil {
			log.Printf("Erro ao criar tabela do arquivo %s: %v", file.Name(), err)
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

func (d *Database) GetConnection() *sql.DB {
	return d.Conn
}
