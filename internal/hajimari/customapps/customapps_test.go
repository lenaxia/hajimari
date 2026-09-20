package customapps

import (
	"reflect"
	"testing"

	"github.com/toboshii/hajimari/internal/config"
	"github.com/toboshii/hajimari/internal/models"
)

func TestPopulatePreservesVisibleGroups(t *testing.T) {
	appConfig := config.Config{
		CustomApps: []models.AppGroup{
			{
				Group: "media",
				Apps: []models.App{
					{Name: "jellyfin", URL: "https://jellyfin.example.com", VisibleGroups: []string{"family", "friends"}},
					{Name: "sonarr", URL: "https://sonarr.example.com"},
				},
			},
		},
	}

	got, err := NewList(appConfig).Populate().Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []models.AppGroup{
		{
			Group: "media",
			Apps: []models.App{
				{Name: "jellyfin", URL: "https://jellyfin.example.com", VisibleGroups: []string{"family", "friends"}},
				{Name: "sonarr", URL: "https://sonarr.example.com"},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Populate() = %+v, want %+v", got, want)
	}
}
