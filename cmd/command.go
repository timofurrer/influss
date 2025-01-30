package cmd

import (
	"errors"
	"flag"
	"log/slog"
	"net/http"

	"github.com/timofurrer/influss/internal/api"
	"github.com/timofurrer/influss/internal/feed"
	"github.com/timofurrer/influss/internal/store"
)

type cmdConfig struct {
	listenAddr          string
	useLocalStore       bool
	localStoreDir       string
	useSqlStore         bool
	sqlConnectionString string
	feedTitle           string
	feedLink            string
	feedDescription     string
	feedAuthorName      string
	feedAuthorEmail     string
	feedCategory        string
	feedItemsLimit      int64
}

type Cmd struct {
	log    *slog.Logger
	config cmdConfig
}

func NewCommand(log *slog.Logger) *Cmd {
	return &Cmd{log: log}
}

func (c *Cmd) Parse() error {
	flag.StringVar(&c.config.listenAddr, "listen-addr", ":8080", "the address to listen on")

	flag.BoolVar(&c.config.useLocalStore, "use-local-store", false, "enable local file system store")
	flag.StringVar(&c.config.localStoreDir, "local-store-dir", "store", "the path to the local store root directory")

	flag.BoolVar(&c.config.useSqlStore, "use-sql-store", false, "enable SQL store")
	flag.StringVar(&c.config.sqlConnectionString, "sql-connection-string", "", "the SQL connection string for the SQL store")

	flag.StringVar(&c.config.feedTitle, "feed-title", "influss", "the RSS feed title")
	flag.StringVar(&c.config.feedLink, "feed-link", "", "the external URL to the RSS feed")
	flag.StringVar(&c.config.feedDescription, "feed-description", "influss RSS feed", "the description of the RSS feed")
	flag.StringVar(&c.config.feedAuthorName, "feed-author-name", "", "the RSS feed author name (your name probably)")
	flag.StringVar(&c.config.feedAuthorEmail, "feed-author-email", "", "the RSS feed author email (your email probably)")
	flag.StringVar(&c.config.feedCategory, "feed-category", "Read It Later", "the RSS feed category")

	flag.Int64Var(&c.config.feedItemsLimit, "feed-items-limit", 20, "the number of feed items to put in the RSS feed")

	flag.Parse()

	if !c.config.useLocalStore && !c.config.useSqlStore {
		return errors.New("choose between using a local store or sql store")
	}

	if c.config.useLocalStore && c.config.useSqlStore {
		return errors.New("choose between using a local store or sql store, but not both")
	}

	if c.config.useSqlStore && c.config.sqlConnectionString == "" {
		return errors.New("when using the sql store a connection string is required")
	}

	if c.config.useLocalStore && c.config.sqlConnectionString != "" {
		return errors.New("when using the local store the connection string is ignored, don't specify it")
	}

	return nil
}

func (c *Cmd) Run() {
	var s store.Store
	var err error
	switch {
	case c.config.useLocalStore:
		s, err = store.NewFSStore(c.config.localStoreDir)
	case c.config.useSqlStore:
		s, err = store.NewSqlStore(c.log, c.config.sqlConnectionString)
	default:
		c.log.Error("No store provider chosen")
		return
	}
	if err != nil {
		c.log.Error("failed to create store", slog.String("error", err.Error()))
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

	mux.HandleFunc("GET /clips", api.GetFeedFunc(feedConfig, int(c.config.feedItemsLimit), s))
	mux.HandleFunc("POST /clips", api.ClipURLFunc(c.log, s))

	c.log.Info("Serving ...", slog.String("listen_addr", c.config.listenAddr))
	if err := http.ListenAndServe(c.config.listenAddr, mux); err != nil {
		c.log.Error("Failed to serve", slog.String("error", err.Error()))
	}
}
