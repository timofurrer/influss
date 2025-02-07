package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/timofurrer/influss/internal/clip"
)

var (
	_ Store = (*FSStore)(nil)
)

type FSStore struct {
	dir   string
	index *index
	m     sync.RWMutex
}

type index struct {
	CreatedAt     time.Time           `json:"created_at"`
	LastUpdatedAt time.Time           `json:"last_updated_at"`
	Clips         map[string]clipMeta `json:"clips"`
}

type clipMeta struct {
	Hash      string    `json:"hash"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type fsClip struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Author      string    `json:"author"`
	Excerpt     string    `json:"excerpt"`
	HTMLContent string    `json:"html_content"`
}

func NewFSStore(dir string) (*FSStore, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, errors.New("when using the local file system store, the given store directory must already exist")
	}

	indexFile := filepath.Join(dir, "index.json")
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return &FSStore{
			dir:   dir,
			index: &index{
				CreatedAt: time.Now(),
				LastUpdatedAt: time.Now(),
				Clips: make(map[string]clipMeta),
			},
		}, nil
	}
	index := &index{}
	err = json.Unmarshal(data, index)
	if err != nil {
		return nil, err
	}

	return &FSStore{
		dir:   dir,
		index: index,
	}, nil
}

func (s *FSStore) CreatedAt() time.Time {
	return s.index.CreatedAt
}

func (s *FSStore) Store(_ context.Context, clip *clip.Clip) error {
	s.m.Lock()
	defer s.m.Unlock()
	fc := fsClip{
		URL:         clip.URL,
		Title:       clip.Title,
		Author:      clip.Author,
		PublishedAt: clip.PublishedAt,
		ModifiedAt:  clip.ModifiedAt,
		Excerpt:     clip.Excerpt,
		HTMLContent: clip.HTMLContent,
	}

	h := generateClipHash(clip)
	cm := clipMeta{
		Hash:      h,
		Path:      filepath.Join(s.dir, fmt.Sprintf("%s.json", h)),
		Timestamp: time.Now(),
	}

	// write clip file
	err := writeJSON(fc, cm.Path)
	if err != nil {
		return fmt.Errorf("failed to store clip: %w", err)
	}
	// write plain text content file for better shell-friendliness
	err = os.WriteFile(filepath.Join(s.dir, fmt.Sprintf("%s.txt", h)), []byte(clip.PlainTextContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to store clip plain text data: %w", err)
	}

	s.index.LastUpdatedAt = cm.Timestamp
	s.index.Clips[h] = cm

	err = writeJSON(s.index, filepath.Join(s.dir, "index.json"))
	if err != nil {
		return fmt.Errorf("failed to store index file after updating clip %s: %w", h, err)
	}
	return nil
}

func (s *FSStore) Load(_ context.Context, lastN int) []*clip.Clip {
	s.m.RLock()
	defer s.m.RUnlock()
	log := slog.Default()
	cs := slices.Collect(maps.Values(s.index.Clips))
	slices.SortFunc(cs, func(a, b clipMeta) int {
		return b.Timestamp.Compare(a.Timestamp)
	})

	cs = cs[:min(lastN, len(cs))]
	clips := make([]*clip.Clip, 0, len(cs))

	for _, cm := range cs {
		data, err := os.ReadFile(cm.Path)
		if err != nil {
			log.Error("Unable to load clip", slog.String("clip_hash", cm.Hash))
			continue
		}
		c := &fsClip{}
		err = json.Unmarshal(data, c)
		if err != nil {
			log.Error("Unable to unmarshal clip", slog.String("clip_hash", cm.Hash))
		}

		clips = append(clips, &clip.Clip{
			URL:         c.URL,
			Title:       c.Title,
			Author:      c.Author,
			PublishedAt: c.PublishedAt,
			ModifiedAt:  c.ModifiedAt,
			Excerpt:     c.Excerpt,
			HTMLContent: c.HTMLContent,
			// NOTE: no need to load the plain text content file
			PlainTextContent: "",
		})
	}
	return clips
}

func generateClipHash(clip *clip.Clip) string {
	h := sha256.New()
	h.Write([]byte(clip.URL))
	return hex.EncodeToString(h.Sum(nil))
}

func writeJSON(data any, filename string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonBytes, 0644)
}
