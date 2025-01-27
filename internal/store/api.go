package store

import (
	"time"

	"github.com/timofurrer/influss/internal/clip"
)

type Store interface {
	Store(clip *clip.Clip) error
	Load(lastN int) []*clip.Clip
	CreatedAt() time.Time
}
