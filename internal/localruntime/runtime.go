package localruntime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yazanabuashour/openbudget/internal/app"
	"github.com/yazanabuashour/openbudget/internal/storage/sqlite"
)

const EnvDatabasePath = app.EnvDatabasePath

type Config struct {
	DatabasePath string
	Timeout      time.Duration
}

type Paths struct {
	DataDir      string
	DatabasePath string
}

type Session struct {
	Paths      Paths
	Repository *sqlite.Repository

	close func() error
}

func ResolvePaths(config Config) (Paths, error) {
	dataDir, databasePath, err := app.ResolveLocalPaths(app.LocalPathConfig{
		DatabasePath: config.DatabasePath,
	})
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		DataDir:      dataDir,
		DatabasePath: databasePath,
	}, nil
}

func Open(config Config) (*Session, error) {
	paths, err := ResolvePaths(config)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(paths.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create local data directory %s: %w", paths.DataDir, err)
	}

	db, err := sqlite.Open(paths.DatabasePath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	if err := sqlite.ApplyMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Session{
		Paths:      paths,
		Repository: sqlite.NewRepository(db),
		close:      db.Close,
	}, nil
}

func (s *Session) Close() error {
	if s == nil || s.close == nil {
		return nil
	}
	return s.close()
}
