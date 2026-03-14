package lurking

//go:generate mockgen -destination=mock_arrlurker_test.go -package=lurking github.com/lusoris/lurkarr/internal/lurking ArrLurker

import (
	"context"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ArrLurker abstracts the arr-specific API calls for missing/upgrade/search/queue.
type ArrLurker interface {
	GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error)
	GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error)
	Search(ctx context.Context, c *arrclient.Client, mediaID int) error
	GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error)
}

// --- Sonarr ---

type sonarrLurker struct{}

func (sonarrLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	eps, err := c.SonarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(eps))
	for i, ep := range eps {
		items[i] = lurkableItem{ID: ep.ID, Title: ep.Title, SortDate: parseArrDate(ep.AirDateUtc)}
	}
	return items, nil
}

func (sonarrLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	eps, err := c.SonarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(eps))
	for i, ep := range eps {
		items[i] = lurkableItem{ID: ep.ID, Title: ep.Title, SortDate: parseArrDate(ep.AirDateUtc)}
	}
	return items, nil
}

func (sonarrLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.SonarrSearchEpisode(ctx, []int{mediaID})
	return err
}

func (sonarrLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.SonarrGetQueue(ctx)
}

// --- Radarr ---

type radarrLurker struct{}

func (radarrLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	movies, err := c.RadarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(movies))
	for i, m := range movies {
		items[i] = lurkableItem{ID: m.ID, Title: m.Title, SortDate: parseArrDate(m.Added)}
	}
	return items, nil
}

func (radarrLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	movies, err := c.RadarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(movies))
	for i, m := range movies {
		items[i] = lurkableItem{ID: m.ID, Title: m.Title, SortDate: parseArrDate(m.Added)}
	}
	return items, nil
}

func (radarrLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.RadarrSearchMovie(ctx, []int{mediaID})
	return err
}

func (radarrLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.RadarrGetQueue(ctx)
}

// --- Lidarr ---

type lidarrLurker struct{}

func (lidarrLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	albums, err := c.LidarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(albums))
	for i, a := range albums {
		items[i] = lurkableItem{ID: a.ID, Title: a.Title, SortDate: parseArrDate(a.ReleaseDate)}
	}
	return items, nil
}

func (lidarrLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	albums, err := c.LidarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(albums))
	for i, a := range albums {
		items[i] = lurkableItem{ID: a.ID, Title: a.Title, SortDate: parseArrDate(a.ReleaseDate)}
	}
	return items, nil
}

func (lidarrLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.LidarrSearchAlbum(ctx, []int{mediaID})
	return err
}

func (lidarrLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.LidarrGetQueue(ctx)
}

// --- Readarr ---

type readarrLurker struct{}

func (readarrLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	books, err := c.ReadarrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(books))
	for i, b := range books {
		items[i] = lurkableItem{ID: b.ID, Title: b.Title, SortDate: parseArrDate(b.ReleaseDate)}
	}
	return items, nil
}

func (readarrLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	books, err := c.ReadarrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(books))
	for i, b := range books {
		items[i] = lurkableItem{ID: b.ID, Title: b.Title, SortDate: parseArrDate(b.ReleaseDate)}
	}
	return items, nil
}

func (readarrLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.ReadarrSearchBook(ctx, []int{mediaID})
	return err
}

func (readarrLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.ReadarrGetQueue(ctx)
}

// --- Whisparr ---

type whisparrLurker struct{}

func (whisparrLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	episodes, err := c.WhisparrGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(episodes))
	for i, ep := range episodes {
		items[i] = lurkableItem{ID: ep.ID, Title: ep.Title, SortDate: parseArrDate(ep.ReleaseDate)}
	}
	return items, nil
}

func (whisparrLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	episodes, err := c.WhisparrGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(episodes))
	for i, ep := range episodes {
		items[i] = lurkableItem{ID: ep.ID, Title: ep.Title, SortDate: parseArrDate(ep.ReleaseDate)}
	}
	return items, nil
}

func (whisparrLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.WhisparrSearchEpisode(ctx, []int{mediaID})
	return err
}

func (whisparrLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.WhisparrGetQueue(ctx)
}

// --- Eros ---

type erosLurker struct{}

func (erosLurker) GetMissing(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	movies, err := c.ErosGetMissing(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(movies))
	for i, m := range movies {
		items[i] = lurkableItem{ID: m.ID, Title: m.Title, SortDate: parseArrDate(m.Added)}
	}
	return items, nil
}

func (erosLurker) GetUpgrades(ctx context.Context, c *arrclient.Client) ([]lurkableItem, error) {
	movies, err := c.ErosGetCutoffUnmet(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]lurkableItem, len(movies))
	for i, m := range movies {
		items[i] = lurkableItem{ID: m.ID, Title: m.Title, SortDate: parseArrDate(m.Added)}
	}
	return items, nil
}

func (erosLurker) Search(ctx context.Context, c *arrclient.Client, mediaID int) error {
	_, err := c.ErosSearchMovie(ctx, []int{mediaID})
	return err
}

func (erosLurker) GetQueue(ctx context.Context, c *arrclient.Client) (*arrclient.QueueResponse, error) {
	return c.ErosGetQueue(ctx)
}

// --- Registry ---

var lurkerRegistry = map[database.AppType]ArrLurker{
	database.AppSonarr:   sonarrLurker{},
	database.AppRadarr:   radarrLurker{},
	database.AppLidarr:   lidarrLurker{},
	database.AppReadarr:  readarrLurker{},
	database.AppWhisparr: whisparrLurker{},
	database.AppEros:     erosLurker{},
}

// LurkerFor returns the ArrLurker for the given app type, or nil if unsupported.
func LurkerFor(appType database.AppType) ArrLurker {
	return lurkerRegistry[appType]
}
