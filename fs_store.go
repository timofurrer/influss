package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type FSStore struct {
	dir   string
	index *index
}

type index struct {
	LastUpdatedAt time.Time           `json:"last_updated_at"`
	Clips         map[string]clipMeta `json:"clips"`
}

type clipMeta struct {
	Hash      string    `json:"hash"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type fsClip struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Excerpt     string   `json:"excerpt"`
	HTMLContent string   `json:"html_content"`
	Tags        []string `json:"tags"`
}

func NewFSStore(dir string) (*FSStore, error) {
	indexFile := filepath.Join(dir, "index.json")
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return &FSStore{
			dir:   dir,
			index: &index{LastUpdatedAt: time.Now(), Clips: make(map[string]clipMeta)},
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

func (s *FSStore) Store(clip *Clip) error {
	fc := fsClip{
		URL:         clip.URL,
		Title:       clip.Title,
		Author:      clip.Author,
		Excerpt:     clip.Excerpt,
		HTMLContent: clip.HTMLContent,
		Tags:        clip.Tags,
	}

	h := generateClipHash(clip)
	cm := clipMeta{
		Hash:      h,
		Path:      filepath.Join(s.dir, fmt.Sprintf("%s.json", h)),
		Timestamp: time.Now(),
	}

	writeJSON(fc, cm.Path)

	s.index.LastUpdatedAt = cm.Timestamp
	s.index.Clips[h] = cm

	writeJSON(s.index, filepath.Join(s.dir, "index.json"))
	return nil
}

func (s *FSStore) Load(lastN int) []*Clip {
	log := slog.Default()
	cs := slices.Collect(maps.Values(s.index.Clips))
	slices.SortFunc(cs, func(a, b clipMeta) int {
		return a.Timestamp.Compare(b.Timestamp)
	})

	cs = cs[:min(lastN, len(cs))]
	clips := make([]*Clip, 0, len(cs))

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

		clips = append(clips, &Clip{
			URL:         c.URL,
			Title:       c.Title,
			Author:      c.Author,
			Excerpt:     c.Excerpt,
			HTMLContent: c.HTMLContent,
			Tags:        c.Tags,
		})
	}
	return clips
}

func generateClipHash(clip *Clip) string {
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
