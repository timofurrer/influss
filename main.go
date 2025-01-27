package main

import (
	"cmp"
	"encoding/json"
	"flag"
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
	LastUpdatedAt() time.Time
}

type config struct {
	feedTitle       string
	feedLink        string
	feedDescription string
	feedAuthorName  string
	feedAuthorEmail string
	feedCategory    string
}

func URLField(url string) slog.Attr {
	return slog.String("url", url)
}

func clipHandlerFunc(config config, store Store) http.HandlerFunc {
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

			log.Info("Received request to clip URL", URLField(req.URL))

			err = clip(store, req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error clipping URL: %s", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			clips := store.Load(20)

			feed := &feeds.RssFeed{
				Title:          config.feedTitle,
				Link:           config.feedLink,
				Description:    config.feedDescription,
				ManagingEditor: fmt.Sprintf("%s (%s)", config.feedAuthorName, config.feedAuthorEmail),
				LastBuildDate:  store.LastUpdatedAt().Format(time.RFC1123Z),
				PubDate:        store.CreatedAt().Format(time.RFC1123Z),
				Category:       config.feedCategory,
				Copyright:      fmt.Sprintf("Influss and %s", config.feedAuthorName),
			}
			for _, c := range clips {
				item := &feeds.RssItem{
					Guid: &feeds.RssGuid{
						Id:          c.URL,
						IsPermaLink: "true",
					},
					Title: c.Title,
					Link: c.URL,
					Source: c.URL,
					Author: c.Author,
					Description: c.Excerpt,
					PubDate: c.ModifiedAt.Format(time.RFC1123Z),
					Content: &feeds.RssContent{
						Content: c.HTMLContent,
					},
				}
				feed.Items = append(feed.Items, item)
			}

			data, err := feeds.ToXML(feed)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to generate RSS: %s", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(data))
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
