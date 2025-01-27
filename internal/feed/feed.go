package feed

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/timofurrer/influss/internal/clip"
)

type Config struct {
	Title       string
	Link        string
	Description string
	AuthorName  string
	AuthorEmail string
	Category    string
	CreatedAt   time.Time
}

type Builder struct {
	feed    *feeds.RssFeed
	pubDate time.Time
}

func NewBuidler(cfg Config) *Builder {
	f := &feeds.RssFeed{
		Title:          cfg.Title,
		Link:           cfg.Link,
		Description:    cfg.Description,
		ManagingEditor: fmt.Sprintf("%s (%s)", cfg.AuthorName, cfg.AuthorEmail),
		LastBuildDate:  cfg.CreatedAt.Format(time.RFC1123Z),
		Category:       cfg.Category,
		Copyright:      fmt.Sprintf("Influss and %s", cfg.AuthorName),
	}
	return &Builder{
		feed:    f,
		pubDate: cfg.CreatedAt,
	}
}

func (f *Builder) WithClip(c *clip.Clip) {
	f.feed.Items = append(f.feed.Items, &feeds.RssItem{
		Guid: &feeds.RssGuid{
			Id:          c.URL,
			IsPermaLink: "true",
		},
		Title:       c.Title,
		Link:        c.URL,
		Source:      c.URL,
		Author:      c.Author,
		Description: c.Excerpt,
		PubDate:     c.ModifiedAt.Format(time.RFC1123Z),
		Content: &feeds.RssContent{
			Content: c.HTMLContent,
		},
	})
	if c.ModifiedAt.After(f.pubDate) {
		f.pubDate = c.ModifiedAt
	}
}

func (f *Builder) ToXML() ([]byte, error) {
	f.feed.PubDate = f.pubDate.Format(time.RFC1123Z)
	data, err := feeds.ToXML(f.feed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate feed XML: %w", err)
	}

	return []byte(data), nil
}
