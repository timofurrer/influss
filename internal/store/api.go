package store

import (
	"context"
	"time"

	"github.com/timofurrer/influss/internal/clip"
)

type Store interface {
	CreatedAt() time.Time
	Store(ctx context.Context, clip *clip.Clip) error
	Load(ctx context.Context, lastN int) []*clip.Clip
}
