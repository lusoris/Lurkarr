package hunting

import (
	"context"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ArrHunter abstracts the arr-specific API calls for missing/upgrade/search/queue.
type ArrHunter interface {
	GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error)
	GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error)
	Search(ctx context.Context, c *arrclient.Client, mediaID int) error
	GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error)
}

// --- Sonarr ---

type sonarrHunter struct{}

func (sonarrHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	eps, err := c.SonarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(eps))
	for i, ep := range eps {
		items[i] = huntableItem{ID: ep.ID, Title: ep.Title}
	}
	return items, nil
}

func (sonarrHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	eps, err := c.SonarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(eps))
	for i, ep := range eps {
		items[i] = huntableItem{ID: ep.ID, Title: ep.Title}
	}
	return items, nil
}

func (sonarrHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.SonarrSearchEpisode(ctx, []int{mediaID})
	return err
}

func (sonarrHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.SonarrGetQueue(ctx)
}

// --- Radarr ---

type radarrHunter struct{}

func (radarrHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.RadarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (radarrHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.RadarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (radarrHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.RadarrSearchMovie(ctx, []int{mediaID})
	return err
}

func (radarrHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.RadarrGetQueue(ctx)
}

// --- Lidarr ---

type lidarrHunter struct{}

func (lidarrHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	albums, err := c.LidarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(albums))
	for i, a := range albums {
		items[i] = huntableItem{ID: a.ID, Title: a.Title}
	}
	return items, nil
}

func (lidarrHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	albums, err := c.LidarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(albums))
	for i, a := range albums {
		items[i] = huntableItem{ID: a.ID, Title: a.Title}
	}
	return items, nil
}

func (lidarrHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.LidarrSearchAlbum(ctx, []int{mediaID})
	return err
}

func (lidarrHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.LidarrGetQueue(ctx)
}

// --- Readarr ---

type readarrHunter struct{}

func (readarrHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	books, err := c.ReadarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(books))
	for i, b := range books {
		items[i] = huntableItem{ID: b.ID, Title: b.Title}
	}
	return items, nil
}

func (readarrHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	books, err := c.ReadarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(books))
	for i, b := range books {
		items[i] = huntableItem{ID: b.ID, Title: b.Title}
	}
	return items, nil
}

func (readarrHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.ReadarrSearchBook(ctx, []int{mediaID})
	return err
}

func (readarrHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.ReadarrGetQueue(ctx)
}

// --- Whisparr ---

type whisparrHunter struct{}

func (whisparrHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.WhisparrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (whisparrHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.WhisparrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (whisparrHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.WhisparrSearchMovie(ctx, []int{mediaID})
	return err
}

func (whisparrHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.WhisparrGetQueue(ctx)
}

// --- Eros ---

type erosHunter struct{}

func (erosHunter) GetMissing(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.ErosGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (erosHunter) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]huntableItem, error) {
	movies, err := c.ErosGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]huntableItem, len(movies))
	for i, m := range movies {
		items[i] = huntableItem{ID: m.ID, Title: m.Title}
	}
	return items, nil
}

func (erosHunter) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.ErosSearchMovie(ctx, []int{mediaID})
	return err
}

func (erosHunter) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.ErosGetQueue(ctx)
}

// --- Registry ---

var hunterRegistry = map[database.AppType]ArrHunter{
	database.AppSonarr:   sonarrHunter{},
	database.AppRadarr:   radarrHunter{},
	database.AppLidarr:   lidarrHunter{},
	database.AppReadarr:  readarrHunter{},
	database.AppWhisparr: whisparrHunter{},
	database.AppEros:     erosHunter{},
}

// HunterFor returns the ArrHunter for the given app type, or nil if unsupported.
func HunterFor(appType database.AppType) ArrHunter {
	return hunterRegistry[appType]
}
