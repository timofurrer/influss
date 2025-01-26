package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/gorilla/feeds"
)

type clipRequest struct {
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type Clip struct {
	URL         string
	Title       string
	Author      string
	PublishedAt time.Time
	ModifiedAt  time.Time
	Excerpt     string
	HTMLContent string
	Tags        []string
}

type Store interface {
	Store(clip *Clip) error
	Load(lastN int) []*Clip
}

func URLField(url string) slog.Attr {
	return slog.String("url", url)
}

func TagsField(tags []string) slog.Attr {
	return slog.Any("tags", tags)
}

func clipHandlerFunc(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := slog.Default()

		switch r.Method {
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading request body: %s", err), http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			req := clipRequest{}
			err = json.Unmarshal(body, &req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing request body: %s", err), http.StatusBadRequest)
				return
			}

			log.Info("Received request to clip URL", URLField(req.URL), TagsField(req.Tags))

			err = clip(store, req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error clipping URL: %s", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			clips := store.Load(20)

			feed := &feeds.Feed{
				Title:       "Influss",
				Link:        &feeds.Link{},
				Description: "Influss Feed",
				Author: &feeds.Author{
					Name:  "Timo Furrer",
					Email: "timo@furrer.life",
				},
				Updated: time.Now(),
				Created: time.Now(),
				// Id:          "",
				// Subtitle:    "",
				// Items:       []*feeds.Item{},
				// Copyright:   "",
				// Image:       &feeds.Image{},
			}
			for _, c := range clips {
				item := &feeds.Item{
					Title: c.Title,
					Link: &feeds.Link{
						Href: c.URL,
					},
					Source: &feeds.Link{
						Href: c.URL,
					},
					Author: &feeds.Author{
						Name: c.Author,
						// Email: "",
					},
					Description: c.Excerpt,
					// Id:          "",
					// IsPermaLink: "",
					Updated: c.ModifiedAt,
					Created: c.PublishedAt,
					// Enclosure:   &feeds.Enclosure{},
					Content: c.HTMLContent,
				}
				feed.Items = append(feed.Items, item)
			}

			rss, err := feed.ToRss()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to generate RSS: %s", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(rss))
			return
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func clip(store Store, req clipRequest) error {
	article, err := readability.FromURL(req.URL, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	now := time.Now()
	clip := &Clip{
		URL:         req.URL,
		Title:       article.Title,
		Author:      article.Byline,
		PublishedAt: *cmp.Or(article.PublishedTime, &now),
		ModifiedAt:  *cmp.Or(article.ModifiedTime, &now),
		Excerpt:     article.Excerpt,
		HTMLContent: article.Content,
		Tags:        req.Tags,
	}
	store.Store(clip)

	return nil
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	storeDir := cmp.Or(os.Getenv("INFLUSS_STORE_DIR"), "store")

	store, err := NewFSStore(storeDir)
	if err != nil {
		log.Error("failed to create store", slog.String("error", err.Error()))
		return
	}

	http.HandleFunc("/clips", clipHandlerFunc(store))

	port := ":8080"
	log.Info("Serving ...", slog.String("port", port))
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Error("Failed to serve", slog.String("error", err.Error()))
	}
}
