package clip

import (
	"cmp"
	"fmt"
	"time"

	"github.com/go-shiori/go-readability"
)

type Clip struct {
	URL              string
	Title            string
	Author           string
	PublishedAt      time.Time
	ModifiedAt       time.Time
	Excerpt          string
	HTMLContent      string
	PlainTextContent string
}

func ClipURL(url string) (*Clip, error) {
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	now := time.Now()
	clip := &Clip{
		URL:              url,
		Title:            article.Title,
		Author:           article.Byline,
		PublishedAt:      *cmp.Or(article.PublishedTime, &now),
		ModifiedAt:       *cmp.Or(article.ModifiedTime, &now),
		Excerpt:          article.Excerpt,
		HTMLContent:      article.Content,
		PlainTextContent: article.TextContent,
	}

	return clip, nil
}
