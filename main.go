package main

import (
	"cmp"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-shiori/go-readability"
)

type clipRequest struct {
	URL string `json:"url"`
}

type Clip struct {
	URL         string
	Title       string
	Author      string
	PublishedAt time.Time
	ModifiedAt  time.Time
	Excerpt     string
	HTMLContent string
}

type Store interface {
	Store(clip *Clip) error
	Load(lastN int) []*Clip
	CreatedAt() time.Time
}

type config struct {
	feedTitle       string
	feedLink        string
	feedDescription string
	feedAuthorName  string
	feedAuthorEmail string
	feedCategory    string
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
	}
	store.Store(clip)

	return nil
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	config := config{}
	flag.StringVar(&config.feedTitle, "feed-title", "Influss", "the RSS feed title")
	flag.StringVar(&config.feedLink, "feed-link", "", "the external URL to the RSS feed")
	flag.StringVar(&config.feedDescription, "feed-description", "Influss Feed", "the description of the RSS feed")
	flag.StringVar(&config.feedAuthorName, "feed-author-name", "", "the RSS feed author name (your name probably)")
	flag.StringVar(&config.feedAuthorEmail, "feed-author-email", "", "the RSS feed author email (your email probably)")
	flag.StringVar(&config.feedCategory, "feed-category", "Read It Later", "the RSS feed category")

	flag.Parse()

	storeDir := cmp.Or(os.Getenv("INFLUSS_STORE_DIR"), "store")

	store, err := NewFSStore(storeDir)
	if err != nil {
		log.Error("failed to create store", slog.String("error", err.Error()))
		return
	}

	http.HandleFunc("/clips", clipHandlerFunc(config, store))

	port := ":8080"
	log.Info("Serving ...", slog.String("port", port))
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Error("Failed to serve", slog.String("error", err.Error()))
	}
}
