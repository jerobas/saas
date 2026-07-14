package service

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DatabaseService struct {
	db  *Database
	ctx context.Context
}

func NewDatabaseService(db *Database) *DatabaseService {
	return &DatabaseService{db: db}
}

func (s *DatabaseService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *DatabaseService) Export() error {
	if s.ctx == nil {
		return errors.New("context not set")
	}

	destPath, err := runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Salvar backup",
		DefaultFilename: "backup.db",
		Filters: []runtime.FileFilter{{
			DisplayName: "SQLite",
			Pattern:     "*.db",
		}},
	})
	if err != nil {
		return err
	}
	if destPath == "" {
		return errors.New("export cancelled")
	}

	return s.db.Export(destPath)
}

func (s *DatabaseService) Import() error {
	if s.ctx == nil {
		return errors.New("context not set")
	}
	return s.db.Import("")
}
