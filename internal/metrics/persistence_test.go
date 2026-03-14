package metrics

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

func TestBuildLabelKey(t *testing.T) {
	tests := []struct {
		name   string
		labels []*dto.LabelPair
		want   string
	}{
		{
			name:   "single label",
			labels: []*dto.LabelPair{{Name: proto.String("app_type"), Value: proto.String("sonarr")}},
			want:   "sonarr",
		},
		{
			name: "two labels",
			labels: []*dto.LabelPair{
				{Name: proto.String("app_type"), Value: proto.String("radarr")},
				{Name: proto.String("instance"), Value: proto.String("Movies")},
			},
			want: "radarr/Movies",
		},
		{
			name:   "no labels",
			labels: nil,
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLabelKey(tt.labels)
			if got != tt.want {
				t.Errorf("buildLabelKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLabelKey(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		labelNames []string
		wantNil    bool
		wantLabels map[string]string
	}{
		{
			name:       "two labels",
			key:        "sonarr/TV",
			labelNames: []string{"app_type", "instance"},
			wantLabels: map[string]string{"app_type": "sonarr", "instance": "TV"},
		},
		{
			name:       "mismatch count",
			key:        "sonarr",
			labelNames: []string{"app_type", "instance"},
			wantNil:    true,
		},
		{
			name:       "empty key empty labels",
			key:        "",
			labelNames: []string{},
			wantLabels: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabelKey(tt.key, tt.labelNames)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseLabelKey() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("parseLabelKey() = nil, want non-nil")
			}
			for k, v := range tt.wantLabels {
				if got[k] != v {
					t.Errorf("parseLabelKey()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestFindDef(t *testing.T) {
	found := findDef("lurkarr_queue_cleaner_items_removed_total")
	if found == nil {
		t.Fatal("expected to find counter def")
	}
	if found.counter != QueueCleanerItemsRemoved {
		t.Error("wrong counter returned")
	}

	if findDef("nonexistent") != nil {
		t.Error("expected nil for unknown metric")
	}
}

func TestCounterKey(t *testing.T) {
	got := counterKey("lurkarr_lurk_searches_total", "sonarr/TV")
	want := "lurkarr_lurk_searches_total|sonarr/TV"
	if got != want {
		t.Errorf("counterKey() = %q, want %q", got, want)
	}
}

func TestCollectCounterMetrics(t *testing.T) {
	LurkSearchesTotal.WithLabelValues("sonarr", "TestCollect").Add(5)
	m := collectCounterMetrics(LurkSearchesTotal)

	key := "sonarr/TestCollect"
	if v, ok := m[key]; !ok {
		t.Errorf("missing key %q in collected metrics", key)
	} else if v < 5 {
		t.Errorf("collected value = %f, want >= 5", v)
	}
}
