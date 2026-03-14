package arrclient

import (
	"context"
	"fmt"
)

// MediaItem represents a library item from any *arr instance for cross-instance matching.
type MediaItem struct {
	ID         int    // Internal arr ID
	Title      string // Display title
	ExternalID string // Canonical external ID (tmdb:123, tvdb:456, etc.)
	HasFile    bool   // Whether files exist on disk
	Monitored  bool   // Whether the item is monitored
}

// RadarrGetAllMovies fetches all movies from a Radarr instance.
func (c *Client) RadarrGetAllMovies(ctx context.Context) ([]MediaItem, error) {
	var movies []struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		TmdbID    int    `json:"tmdbId"`
		HasFile   bool   `json:"hasFile"`
		Monitored bool   `json:"monitored"`
	}
	if err := c.get(ctx, radarrAPI+"/movie", &movies); err != nil {
		return nil, fmt.Errorf("radarr get all movies: %w", err)
	}
	items := make([]MediaItem, 0, len(movies))
	for _, m := range movies {
		if m.TmdbID == 0 {
			continue
		}
		items = append(items, MediaItem{
			ID:         m.ID,
			Title:      m.Title,
			ExternalID: fmt.Sprintf("tmdb:%d", m.TmdbID),
			HasFile:    m.HasFile,
			Monitored:  m.Monitored,
		})
	}
	return items, nil
}

// SonarrGetAllSeries fetches all series from a Sonarr instance.
func (c *Client) SonarrGetAllSeries(ctx context.Context) ([]MediaItem, error) {
	var series []struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		TvdbID     int    `json:"tvdbId"`
		Monitored  bool   `json:"monitored"`
		Statistics struct {
			EpisodeFileCount int `json:"episodeFileCount"`
		} `json:"statistics"`
	}
	if err := c.get(ctx, sonarrAPI+"/series", &series); err != nil {
		return nil, fmt.Errorf("sonarr get all series: %w", err)
	}
	items := make([]MediaItem, 0, len(series))
	for _, s := range series {
		if s.TvdbID == 0 {
			continue
		}
		items = append(items, MediaItem{
			ID:         s.ID,
			Title:      s.Title,
			ExternalID: fmt.Sprintf("tvdb:%d", s.TvdbID),
			HasFile:    s.Statistics.EpisodeFileCount > 0,
			Monitored:  s.Monitored,
		})
	}
	return items, nil
}

// ErosGetAllMovies fetches all movies from an Eros (Whisparr v3) instance.
func (c *Client) ErosGetAllMovies(ctx context.Context) ([]MediaItem, error) {
	var movies []struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		TmdbID    int    `json:"tmdbId"`
		HasFile   bool   `json:"hasFile"`
		Monitored bool   `json:"monitored"`
	}
	if err := c.get(ctx, erosAPI+"/movie", &movies); err != nil {
		return nil, fmt.Errorf("eros get all movies: %w", err)
	}
	items := make([]MediaItem, 0, len(movies))
	for _, m := range movies {
		if m.TmdbID == 0 {
			continue
		}
		items = append(items, MediaItem{
			ID:         m.ID,
			Title:      m.Title,
			ExternalID: fmt.Sprintf("tmdb:%d", m.TmdbID),
			HasFile:    m.HasFile,
			Monitored:  m.Monitored,
		})
	}
	return items, nil
}

// WhisparrGetAllSeries fetches all series from a Whisparr v2 instance.
func (c *Client) WhisparrGetAllSeries(ctx context.Context) ([]MediaItem, error) {
	var series []struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		TvdbID     int    `json:"tvdbId"`
		Monitored  bool   `json:"monitored"`
		Statistics struct {
			EpisodeFileCount int `json:"episodeFileCount"`
		} `json:"statistics"`
	}
	if err := c.get(ctx, whisparrAPI+"/series", &series); err != nil {
		return nil, fmt.Errorf("whisparr get all series: %w", err)
	}
	items := make([]MediaItem, 0, len(series))
	for _, s := range series {
		if s.TvdbID == 0 {
			continue
		}
		items = append(items, MediaItem{
			ID:         s.ID,
			Title:      s.Title,
			ExternalID: fmt.Sprintf("tvdb:%d", s.TvdbID),
			HasFile:    s.Statistics.EpisodeFileCount > 0,
			Monitored:  s.Monitored,
		})
	}
	return items, nil
}

// LidarrGetAllAlbums fetches all albums from a Lidarr instance.
func (c *Client) LidarrGetAllAlbums(ctx context.Context) ([]MediaItem, error) {
	var albums []struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		ForeignAlbumID string `json:"foreignAlbumId"`
		Monitored      bool   `json:"monitored"`
		Statistics     struct {
			TrackFileCount int `json:"trackFileCount"`
		} `json:"statistics"`
	}
	if err := c.get(ctx, lidarrAPI+"/album", &albums); err != nil {
		return nil, fmt.Errorf("lidarr get all albums: %w", err)
	}
	items := make([]MediaItem, 0, len(albums))
	for _, a := range albums {
		if a.ForeignAlbumID == "" {
			continue
		}
		items = append(items, MediaItem{
			ID:         a.ID,
			Title:      a.Title,
			ExternalID: "album:" + a.ForeignAlbumID,
			HasFile:    a.Statistics.TrackFileCount > 0,
			Monitored:  a.Monitored,
		})
	}
	return items, nil
}

// ReadarrGetAllBooks fetches all books from a Readarr instance.
func (c *Client) ReadarrGetAllBooks(ctx context.Context) ([]MediaItem, error) {
	var books []struct {
		ID            int    `json:"id"`
		Title         string `json:"title"`
		ForeignBookID string `json:"foreignBookId"`
		Monitored     bool   `json:"monitored"`
		Statistics    struct {
			BookFileCount int `json:"bookFileCount"`
		} `json:"statistics"`
	}
	if err := c.get(ctx, readarrAPI+"/book", &books); err != nil {
		return nil, fmt.Errorf("readarr get all books: %w", err)
	}
	items := make([]MediaItem, 0, len(books))
	for _, b := range books {
		if b.ForeignBookID == "" {
			continue
		}
		items = append(items, MediaItem{
			ID:         b.ID,
			Title:      b.Title,
			ExternalID: "book:" + b.ForeignBookID,
			HasFile:    b.Statistics.BookFileCount > 0,
			Monitored:  b.Monitored,
		})
	}
	return items, nil
}
