package database

import (
	"os"
)

func (d *Database) Export(destPath string) error {
	_, err := d.Conn.Exec("VACUUM INTO ?", destPath)
	return err
}

func (d *Database) Import(srcPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := d.Conn.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(d.path, data, 0600); err != nil {
		return err
	}
	db, err := openConnection(d.path)
	if err != nil {
		return err
	}
	d.Conn = db
	return d.createTables()
}
