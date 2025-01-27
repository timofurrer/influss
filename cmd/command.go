package cmd

import (
	"flag"
	"log/slog"
	"net/http"

	"github.com/timofurrer/influss/internal/api"
	"github.com/timofurrer/influss/internal/feed"
	"github.com/timofurrer/influss/internal/store"
)

type cmdConfig struct {
	listenAddr      string
	useLocalStore   bool
	localStoreDir   string
	feedTitle       string
	feedLink        string
	feedDescription string
	feedAuthorName  string
	feedAuthorEmail string
	feedCategory    string
}

type Cmd struct {
	log    *slog.Logger
	config cmdConfig
}

func NewCommand(log *slog.Logger) *Cmd {
	return &Cmd{log: log}
}

func (c *Cmd) Parse() {
	flag.StringVar(&c.config.listenAddr, "listen-addr", ":8080", "the address to listen on")

	flag.BoolVar(&c.config.useLocalStore, "use-local-store", true, "enable local file system store")
	flag.StringVar(&c.config.localStoreDir, "local-store-dir", "store", "the path to the local store root directory")

	flag.StringVar(&c.config.feedTitle, "feed-title", "Influss", "the RSS feed title")
	flag.StringVar(&c.config.feedLink, "feed-link", "", "the external URL to the RSS feed")
	flag.StringVar(&c.config.feedDescription, "feed-description", "Influss Feed", "the description of the RSS feed")
	flag.StringVar(&c.config.feedAuthorName, "feed-author-name", "", "the RSS feed author name (your name probably)")
	flag.StringVar(&c.config.feedAuthorEmail, "feed-author-email", "", "the RSS feed author email (your email probably)")
	flag.StringVar(&c.config.feedCategory, "feed-category", "Read It Later", "the RSS feed category")

	flag.Parse()
}

func (c *Cmd) Run() {
	var s store.Store
	switch {
	case c.config.useLocalStore:
		var err error
		s, err = store.NewFSStore(c.config.localStoreDir)
		if err != nil {
			c.log.Error("failed to create store", slog.String("error", err.Error()))
			return
		}
	default:
		c.log.Error("No store provider chosen")
		return
	}

	feedConfig := feed.Config{
		Title:       c.config.feedTitle,
		Link:        c.config.feedLink,
		Description: c.config.feedDescription,
		AuthorName:  c.config.feedAuthorName,
		AuthorEmail: c.config.feedAuthorEmail,
		Category:    c.config.feedCategory,
		CreatedAt:   s.CreatedAt(),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /clips", api.GetFeedFunc(feedConfig, s))
	mux.HandleFunc("POST /clips", api.ClipURLFunc(c.log, s))

	c.log.Info("Serving ...", slog.String("listen_addr", c.config.listenAddr))
	if err := http.ListenAndServe(c.config.listenAddr, mux); err != nil {
		c.log.Error("Failed to serve", slog.String("error", err.Error()))
	}
}
