package main

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"
)

type feedConfig struct {
	title       string
	link        string
	description string
	authorName  string
	authorEmail string
	category    string
	createdAt   time.Time
}

type feedBuilder struct {
	feed    *feeds.RssFeed
	pubDate time.Time
}

func newFeedBuidler(cfg feedConfig) *feedBuilder {
	f := &feeds.RssFeed{
		Title:          cfg.title,
		Link:           cfg.link,
		Description:    cfg.description,
		ManagingEditor: fmt.Sprintf("%s (%s)", cfg.authorName, cfg.authorEmail),
		LastBuildDate:  cfg.createdAt.Format(time.RFC1123Z),
		Category:       cfg.category,
		Copyright:      fmt.Sprintf("Influss and %s", cfg.authorName),
	}
	return &feedBuilder{
		feed:    f,
		pubDate: cfg.createdAt,
	}
}

func (f *feedBuilder) withClip(c *Clip) {
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

func (f *feedBuilder) toXML() ([]byte, error) {
	f.feed.PubDate = f.pubDate.Format(time.RFC1123Z)
	data, err := feeds.ToXML(f.feed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate feed XML: %w", err)
	}

	return []byte(data), nil
}
