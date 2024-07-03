package storage

import (
	"metrics/internal/server/adapters/storage/database"
	"metrics/internal/server/adapters/storage/file"
	"metrics/internal/server/adapters/storage/memory"
)

type Config struct {
	Memory   *memory.Config
	File     *file.Config
	Database *database.Config
}
