package domain

import (
	"context"
)

type RepositoryType interface {
	CreateEntry(ctx context.Context, entry *UserRecord) (*UserRecord, error)
	GetEntry(ctx context.Context, id string) (*UserRecord, error)
	GetEntries(ctx context.Context, limit, offset int) ([]*UserRecord, error)
	UpdateEntry(ctx context.Context, entry *UserRecord) (*UserRecord, error)
	DeleteEntry(ctx context.Context, id string) (*UserRecord, error)
}
