package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/toboshii/hajimari/internal/models"
)

func listBookmarksFor(t *testing.T, headers map[string]string, query string) []string {
	t.Helper()
	viper.Reset()
	viper.Set("GroupsHeader", "Remote-Groups")
	viper.Set("AdminGroups", []string{"admins"})
	viper.Set("GlobalBookmarks", []models.BookmarkGroup{
		{Group: "Communicate", Bookmarks: []models.Bookmark{{Name: "Discord"}}},
		{Group: "Admin Links", VisibleGroups: []string{"admins"}, Bookmarks: []models.Bookmark{{Name: "Proxmox"}}},
	})

	url := "/bookmarks"
	if query != "" {
		url += "?" + query
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	recorder := httptest.NewRecorder()
	NewBookmarkResource().ListBookmarks(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	groups := []string{}
	for _, group := range response {
		name, _ := group["group"].(string)
		groups = append(groups, name)
	}
	return groups
}

func listBookmarksForStatus(t *testing.T, headers map[string]string, query string) int {
	t.Helper()
	viper.Reset()
	viper.Set("GroupsHeader", "Remote-Groups")
	viper.Set("AdminGroups", []string{"admins"})
	viper.Set("GlobalBookmarks", []models.BookmarkGroup{})

	req, _ := http.NewRequest(http.MethodGet, "/bookmarks?"+query, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	recorder := httptest.NewRecorder()
	NewBookmarkResource().ListBookmarks(recorder, req)
	return recorder.Code
}

func TestListBookmarksFiltersPerRequest(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    []string
	}{
		{"admin sees both groups", map[string]string{"Remote-Groups": "admins"}, []string{"Communicate", "Admin Links"}},
		{"family sees public group only", map[string]string{"Remote-Groups": "family"}, []string{"Communicate"}},
		{"no header sees public group only", nil, []string{"Communicate"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listBookmarksFor(t, tt.headers, "")
			if len(got) != len(tt.want) {
				t.Fatalf("groups = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("groups = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestListBookmarksAdminImpersonation(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		query   string
		want    []string
	}{
		{"admin impersonates family", map[string]string{"Remote-Groups": "admins"}, "group=family", []string{"Communicate"}},
		{"admin impersonates admins unchanged", map[string]string{"Remote-Groups": "admins"}, "group=admins", []string{"Communicate", "Admin Links"}},
		{"non admin denied", map[string]string{"Remote-Groups": "family"}, "group=admins", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				if code := listBookmarksForStatus(t, tt.headers, tt.query); code != http.StatusForbidden {
					t.Fatalf("expected 403, got %d", code)
				}
				return
			}
			got := listBookmarksFor(t, tt.headers, tt.query)
			if len(got) != len(tt.want) {
				t.Fatalf("groups = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("groups = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
