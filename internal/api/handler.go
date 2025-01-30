package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/timofurrer/influss/internal/clip"
	"github.com/timofurrer/influss/internal/feed"
	"github.com/timofurrer/influss/internal/store"
)

type clipRequest struct {
	URL string `json:"url"`
}

func ClipURLFunc(log *slog.Logger, store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		log.Info("Received request to clip URL", slog.String("url", req.URL))

		clip, err := clip.ClipURL(req.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error clipping URL: %s", err), http.StatusInternalServerError)
			return
		}

		err = store.Store(r.Context(), clip)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error storing clipped URL: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func GetFeedFunc(config feed.Config, itemsLimit int, store store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clips := store.Load(r.Context(), itemsLimit)

		fb := feed.NewBuidler(config)
		for _, c := range clips {
			fb.WithClip(c)
		}

		data, err := fb.ToXML()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	}
}
