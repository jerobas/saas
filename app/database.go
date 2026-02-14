package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{conn: db}
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

	if _, err := d.conn.Exec(string(schema)); err != nil {
		log.Printf("Erro ao criar tabelas: %v", err)
		return err
	}

	return nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) GetConnection() *sql.DB {
	return d.conn
}