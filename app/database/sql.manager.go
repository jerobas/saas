package database

import (
	"os"
)

func (d *Database) Export(destPath string) error {
	_, err := d.Conn.Exec("VACUUM INTO ?", destPath)
	return err
}

func (d *Database) Import(srcPath string) error {
	d.Conn.Close()

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	return os.WriteFile(d.path, data, 0644)
}
