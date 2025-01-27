package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func URLField(url string) slog.Attr {
	return slog.String("url", url)
}

func clipHandlerFunc(config config, store Store) http.HandlerFunc {
	feedConfig := feedConfig{
		title:       config.feedTitle,
		link:        config.feedLink,
		description: config.feedDescription,
		authorName:  config.feedAuthorName,
		authorEmail: config.feedAuthorEmail,
		category:    config.feedCategory,
		createdAt:   store.CreatedAt(),
	}
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

			fb := newFeedBuidler(feedConfig)
			for _, c := range clips {
				fb.withClip(c)
			}

			data, err := fb.toXML()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
