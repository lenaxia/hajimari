package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/toboshii/hajimari/internal/models"
)

type fakeAppService struct {
	apps []models.AppGroup
}

func (f *fakeAppService) GetCachedKubeApps() []models.AppGroup {
	return f.apps
}

func testApps() []models.AppGroup {
	return []models.AppGroup{
		{Group: "media", Apps: []models.App{
			{Name: "jellyfin"},
			{Name: "immich", VisibleGroups: []string{"family"}},
		}},
		{Group: "infra", Apps: []models.App{
			{Name: "proxmox", VisibleGroups: []string{"admins"}},
		}},
	}
}

func listAppsFor(t *testing.T, headers map[string]string) []map[string]interface{} {
	t.Helper()
	return listAppsForQuery(t, headers, "")
}

func listAppsForQuery(t *testing.T, headers map[string]string, query string) []map[string]interface{} {
	t.Helper()
	viper.Reset()
	viper.Set("GroupsHeader", "Remote-Groups")
	viper.Set("AdminGroups", []string{"admins"})

	url := "/apps"
	if query != "" {
		url += "?" + query
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	recorder := httptest.NewRecorder()
	NewAppResource(&fakeAppService{apps: testApps()}).ListApps(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return response
}

func listAppsForQueryStatus(t *testing.T, headers map[string]string, query string) int {
	t.Helper()
	viper.Reset()
	viper.Set("GroupsHeader", "Remote-Groups")
	viper.Set("AdminGroups", []string{"admins"})

	req, _ := http.NewRequest(http.MethodGet, "/apps?"+query, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	recorder := httptest.NewRecorder()
	NewAppResource(&fakeAppService{apps: testApps()}).ListApps(recorder, req)
	return recorder.Code
}

func appNames(response []map[string]interface{}) map[string][]string {
	names := map[string][]string{}
	for _, group := range response {
		groupName, _ := group["group"].(string)
		apps, _ := group["apps"].([]interface{})
		for _, app := range apps {
			appMap, _ := app.(map[string]interface{})
			name, _ := appMap["name"].(string)
			names[groupName] = append(names[groupName], name)
		}
	}
	return names
}

func TestListAppsFiltersPerRequest(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    map[string][]string
	}{
		{
			name:    "family user",
			headers: map[string]string{"Remote-Groups": "family"},
			want:    map[string][]string{"media": {"jellyfin", "immich"}},
		},
		{
			name:    "admin user",
			headers: map[string]string{"Remote-Groups": "admins"},
			want:    map[string][]string{"media": {"jellyfin"}, "infra": {"proxmox"}},
		},
		{
			name:    "multi group user gets union",
			headers: map[string]string{"Remote-Groups": "admins,family"},
			want:    map[string][]string{"media": {"jellyfin", "immich"}, "infra": {"proxmox"}},
		},
		{
			name:    "friend user sees only unannotated",
			headers: map[string]string{"Remote-Groups": "friends"},
			want:    map[string][]string{"media": {"jellyfin"}},
		},
		{
			name:    "no header sees only unannotated",
			headers: nil,
			want:    map[string][]string{"media": {"jellyfin"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appNames(listAppsFor(t, tt.headers))
			if len(got) != len(tt.want) {
				t.Fatalf("groups = %v, want %v", got, tt.want)
			}
			for group, wantApps := range tt.want {
				if len(got[group]) != len(wantApps) {
					t.Fatalf("group %q apps = %v, want %v", group, got[group], wantApps)
				}
				for i, want := range wantApps {
					if got[group][i] != want {
						t.Fatalf("group %q apps = %v, want %v", group, got[group], wantApps)
					}
				}
			}
		})
	}
}

func TestListAppsDoesNotLeakGroupMetadata(t *testing.T) {
	response := listAppsFor(t, map[string]string{"Remote-Groups": "admins"})
	for _, group := range response {
		apps, _ := group["apps"].([]interface{})
		for _, app := range apps {
			if appMap, ok := app.(map[string]interface{}); ok {
				if _, leaked := appMap["visibleGroups"]; leaked {
					t.Errorf("visibleGroups leaked in API response: %v", appMap)
				}
			}
		}
	}
}

func TestListAppsCustomHeaderName(t *testing.T) {
	viper.Reset()
	viper.Set("GroupsHeader", "X-Team-Groups")

	req, _ := http.NewRequest(http.MethodGet, "/apps", nil)
	req.Header.Set("X-Team-Groups", "admins")

	recorder := httptest.NewRecorder()
	NewAppResource(&fakeAppService{apps: testApps()}).ListApps(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	names := appNames(response)
	if len(names["infra"]) != 1 || names["infra"][0] != "proxmox" {
		t.Errorf("custom header groups not applied: %v", names)
	}
}

func TestListAppsAdminImpersonation(t *testing.T) {
	adminHeader := map[string]string{"Remote-Groups": "admins"}

	tests := []struct {
		name    string
		headers map[string]string
		query   string
		want    map[string][]string
	}{
		{
			name:    "admin impersonates family",
			headers: adminHeader,
			query:   "group=family",
			want:    map[string][]string{"media": {"jellyfin", "immich"}},
		},
		{
			name:    "admin impersonates friend sees only unannotated",
			headers: adminHeader,
			query:   "group=friends",
			want:    map[string][]string{"media": {"jellyfin"}},
		},
		{
			name:    "admin impersonates groupless with empty param",
			headers: adminHeader,
			query:   "group=",
			want:    map[string][]string{"media": {"jellyfin"}},
		},
		{
			name:    "admin impersonates comma separated union",
			headers: adminHeader,
			query:   "group=family,admins",
			want:    map[string][]string{"media": {"jellyfin", "immich"}, "infra": {"proxmox"}},
		},
		{
			name:    "g alias impersonates family",
			headers: adminHeader,
			query:   "g=family",
			want:    map[string][]string{"media": {"jellyfin", "immich"}},
		},
		{
			name:    "g alias comma separated union",
			headers: adminHeader,
			query:   "g=family,admins",
			want:    map[string][]string{"media": {"jellyfin", "immich"}, "infra": {"proxmox"}},
		},
		{
			name:    "comma separated with whitespace",
			headers: adminHeader,
			query:   "group=family, admins",
			want:    map[string][]string{"media": {"jellyfin", "immich"}, "infra": {"proxmox"}},
		},
		{
			name:    "group and g params merge with duplicates collapsed",
			headers: adminHeader,
			query:   "group=family&g=admins,family",
			want:    map[string][]string{"media": {"jellyfin", "immich"}, "infra": {"proxmox"}},
		},
		{
			name:    "admin without param keeps own view",
			headers: adminHeader,
			query:   "",
			want:    map[string][]string{"media": {"jellyfin"}, "infra": {"proxmox"}},
		},
		{
			name:    "case insensitive group value",
			headers: adminHeader,
			query:   "group=Family",
			want:    map[string][]string{"media": {"jellyfin", "immich"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appNames(listAppsForQuery(t, tt.headers, tt.query))
			if len(got) != len(tt.want) {
				t.Fatalf("groups = %v, want %v", got, tt.want)
			}
			for group, wantApps := range tt.want {
				if len(got[group]) != len(wantApps) {
					t.Fatalf("group %q apps = %v, want %v", group, got[group], wantApps)
				}
				for i, want := range wantApps {
					if got[group][i] != want {
						t.Fatalf("group %q apps = %v, want %v", group, got[group], wantApps)
					}
				}
			}
		})
	}
}

func TestListAppsImpersonationDeniedForNonAdmin(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{"family user", map[string]string{"Remote-Groups": "family"}},
		{"groupless user", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := listAppsForQueryStatus(t, tt.headers, "group=family"); code != http.StatusForbidden {
				t.Errorf("expected 403, got %d", code)
			}
		})
	}
}
