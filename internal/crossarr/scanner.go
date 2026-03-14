package crossarr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// Store defines the database methods used by the cross-arr scanner.
type Store interface {
	ListInstanceGroups(ctx context.Context, appType database.AppType) ([]database.InstanceGroup, error)
	GetInstance(ctx context.Context, id uuid.UUID) (*database.AppInstance, error)
	UpsertCrossInstanceMedia(ctx context.Context, groupID uuid.UUID, externalID, title string) (*database.CrossInstanceMedia, error)
	SetCrossInstancePresence(ctx context.Context, mediaID uuid.UUID, presence []database.CrossInstancePresence) error
	DeleteCrossInstanceMediaByGroup(ctx context.Context, groupID uuid.UUID) error
}

// instanceMedia holds a media item with the instance it came from.
type instanceMedia struct {
	InstanceID uuid.UUID
	Item       arrclient.MediaItem
}

// Scanner detects cross-instance media overlaps.
type Scanner struct {
	DB            Store
	ClientTimeout time.Duration
}

// ScanResult summarizes the outcome of a scan.
type ScanResult struct {
	GroupID   uuid.UUID
	GroupName string
	Overlaps  int
	Errors    []string
}

// ScanAllGroups scans all instance groups for all app types.
func (s *Scanner) ScanAllGroups(ctx context.Context) []ScanResult {
	appTypes := []database.AppType{"sonarr", "radarr", "lidarr", "readarr", "whisparr", "eros"}
	var results []ScanResult

	for _, appType := range appTypes {
		groups, err := s.DB.ListInstanceGroups(ctx, appType)
		if err != nil {
			slog.Error("cross-arr: failed to list groups", "app_type", appType, "error", err)
			continue
		}
		for i := range groups {
			result := s.ScanGroup(ctx, &groups[i])
			results = append(results, result)
		}
	}
	return results
}

// ScanGroup scans a single instance group for media overlaps.
func (s *Scanner) ScanGroup(ctx context.Context, group *database.InstanceGroup) ScanResult {
	result := ScanResult{
		GroupID:   group.ID,
		GroupName: group.Name,
	}

	// Filter out independent members — they don't participate in overlap detection
	var members []database.InstanceGroupMember
	for _, m := range group.Members {
		if !m.IsIndependent {
			members = append(members, m)
		}
	}
	if len(members) < 2 {
		return result
	}

	// Fetch all media from each participating instance
	allMedia := make(map[uuid.UUID][]arrclient.MediaItem)
	for _, member := range members {
		inst, err := s.DB.GetInstance(ctx, member.InstanceID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("get instance %s: %v", member.InstanceName, err))
			continue
		}
		if !inst.Enabled {
			continue
		}

		client := arrclient.NewClient(inst.APIURL, inst.APIKey, s.clientTimeout(), true)
		items, err := s.fetchMedia(ctx, client, string(group.AppType))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("fetch media from %s: %v", inst.Name, err))
			continue
		}
		allMedia[member.InstanceID] = items
		slog.Debug("cross-arr: fetched media", "instance", inst.Name, "count", len(items))
	}

	// Build index by external ID → list of (instanceID, item)
	extIndex := make(map[string][]instanceMedia)
	for instID, items := range allMedia {
		for _, item := range items {
			extIndex[item.ExternalID] = append(extIndex[item.ExternalID], instanceMedia{
				InstanceID: instID,
				Item:       item,
			})
		}
	}

	// Clear old data for this group
	if err := s.DB.DeleteCrossInstanceMediaByGroup(ctx, group.ID); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("clear old data: %v", err))
		return result
	}

	// Store overlaps (items present in 2+ instances)
	for externalID, entries := range extIndex {
		if len(entries) < 2 {
			continue
		}

		// Use the first entry's title as the display title
		title := entries[0].Item.Title
		media, err := s.DB.UpsertCrossInstanceMedia(ctx, group.ID, externalID, title)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("upsert media %s: %v", externalID, err))
			continue
		}

		presence := make([]database.CrossInstancePresence, len(entries))
		for i, e := range entries {
			presence[i] = database.CrossInstancePresence{
				MediaID:    media.ID,
				InstanceID: e.InstanceID,
				Monitored:  e.Item.Monitored,
				HasFile:    e.Item.HasFile,
			}
		}
		if err := s.DB.SetCrossInstancePresence(ctx, media.ID, presence); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("set presence %s: %v", externalID, err))
			continue
		}
		result.Overlaps++
	}

	slog.Info("cross-arr: scan complete", "group", group.Name, "overlaps", result.Overlaps, "errors", len(result.Errors))
	return result
}

func (s *Scanner) clientTimeout() time.Duration {
	if s.ClientTimeout > 0 {
		return s.ClientTimeout
	}
	return 30 * time.Second
}

func (s *Scanner) fetchMedia(ctx context.Context, client *arrclient.Client, appType string) ([]arrclient.MediaItem, error) {
	switch appType {
	case "radarr":
		return client.RadarrGetAllMovies(ctx)
	case "sonarr":
		return client.SonarrGetAllSeries(ctx)
	case "lidarr":
		return client.LidarrGetAllAlbums(ctx)
	case "readarr":
		return client.ReadarrGetAllBooks(ctx)
	case "whisparr":
		return client.WhisparrGetAllSeries(ctx)
	case "eros":
		return client.ErosGetAllMovies(ctx)
	default:
		return nil, fmt.Errorf("unsupported app type: %s", appType)
	}
}
