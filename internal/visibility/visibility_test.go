package visibility

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/toboshii/hajimari/internal/models"
)

func TestParseGroups(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty string", "", []string{}},
		{"single group", "family", []string{"family"}},
		{"comma separated", "family,friends,admins", []string{"family", "friends", "admins"}},
		{"whitespace padded", " family , friends ", []string{"family", "friends"}},
		{"uppercase normalized", "Family,ADMINS", []string{"family", "admins"}},
		{"empty segments dropped", ",,family,,", []string{"family"}},
		{"whitespace only", " , , ", []string{}},
		{"wildcard preserved", "*,family", []string{"*", "family"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseGroups(tt.raw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseGroups(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNewUserFromRequest(t *testing.T) {
	req := httptestRequest(map[string]string{
		"Remote-Groups": "family, admins",
	})

	tests := []struct {
		name       string
		headerName string
		req        *http.Request
		want       []string
	}{
		{"default header", "", req, []string{"family", "admins"}},
		{"custom header name", "X-Custom-Groups", httptestRequest(map[string]string{"X-Custom-Groups": "friends"}), []string{"friends"}},
		{"missing header", "", httptestRequest(nil), []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserFromRequest(tt.headerName, tt.req)
			if !reflect.DeepEqual(u.groups, setOf(tt.want...)) {
				t.Errorf("groups = %v, want %v", u.groups, tt.want)
			}
		})
	}
}

func TestUserCanSee(t *testing.T) {
	tests := []struct {
		name          string
		userGroups    []string
		visibleGroups []string
		want          bool
	}{
		{"no metadata visible to everyone", nil, nil, true},
		{"no metadata visible to groupless", []string{}, []string{}, true},
		{"member sees annotated app", []string{"family"}, []string{"family", "friends"}, true},
		{"non member cannot see annotated app", []string{"admins"}, []string{"family", "friends"}, false},
		{"groupless user cannot see annotated app", nil, []string{"family"}, false},
		{"wildcard grants any user", []string{"friends"}, []string{"*"}, true},
		{"wildcard requires membership somewhere", nil, []string{"*"}, false},
		{"case insensitive match", []string{"Family"}, []string{"FAMILY"}, true},
		{"whitespace tolerant", []string{"family"}, []string{" family , friends "}, true},
		{"multi membership union", []string{"friends", "admins"}, []string{"admins"}, true},
		{"typo in annotation hides from all", []string{"family"}, []string{"famly"}, false},
		{"garbage annotation treated as absent", []string{"family"}, []string{" , , "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUser(tt.userGroups)
			if got := u.CanSee(tt.visibleGroups); got != tt.want {
				t.Errorf("CanSee(%v) for user %v = %v, want %v", tt.visibleGroups, tt.userGroups, got, tt.want)
			}
		})
	}
}

func TestFilterAppGroups(t *testing.T) {
	input := []models.AppGroup{
		{Group: "media", Apps: []models.App{
			{Name: "jellyfin", VisibleGroups: []string{"family", "friends"}},
			{Name: "immich", VisibleGroups: []string{"family"}},
			{Name: "sonarr"},
		}},
		{Group: "infra", Apps: []models.App{
			{Name: "proxmox", VisibleGroups: []string{"admins"}},
		}},
		{Group: "books", Apps: []models.App{
			{Name: "calibre", VisibleGroups: []string{"admins"}},
		}},
	}

	user := NewUser([]string{"family"})
	got := FilterAppGroups(user, input)

	want := []models.AppGroup{
		{Group: "media", Apps: []models.App{
			{Name: "jellyfin", VisibleGroups: []string{"family", "friends"}},
			{Name: "immich", VisibleGroups: []string{"family"}},
			{Name: "sonarr"},
		}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("FilterAppGroups(family) = %+v, want %+v", got, want)
	}

	admin := NewUser([]string{"admins"})
	gotAdmin := FilterAppGroups(admin, input)
	if len(gotAdmin) != 3 || len(gotAdmin[0].Apps) != 1 || gotAdmin[0].Apps[0].Name != "sonarr" {
		t.Errorf("FilterAppGroups(admins) = %+v, want sonarr in media + infra + books", gotAdmin)
	}

	groupless := NewUser(nil)
	gotGroupless := FilterAppGroups(groupless, input)
	if len(gotGroupless) != 1 || gotGroupless[0].Group != "media" || len(gotGroupless[0].Apps) != 1 {
		t.Errorf("FilterAppGroups(groupless) = %+v, want only unannotated sonarr", gotGroupless)
	}
}

func TestFilterAppGroupsDoesNotMutateInput(t *testing.T) {
	input := []models.AppGroup{
		{Group: "media", Apps: []models.App{
			{Name: "jellyfin", VisibleGroups: []string{"family"}},
			{Name: "sonarr"},
		}},
	}
	original := []models.AppGroup{
		{Group: "media", Apps: []models.App{
			{Name: "jellyfin", VisibleGroups: []string{"family"}},
			{Name: "sonarr"},
		}},
	}

	FilterAppGroups(NewUser([]string{"admins"}), input)

	if !reflect.DeepEqual(input, original) {
		t.Errorf("input mutated by FilterAppGroups: %+v", input)
	}
	if cap(input[0].Apps) != cap(original[0].Apps) || len(input[0].Apps) != 2 {
		t.Errorf("input backing array modified: %+v", input[0].Apps)
	}
}

func TestFilterBookmarkGroups(t *testing.T) {
	input := []models.BookmarkGroup{
		{Group: "Communicate", Bookmarks: []models.Bookmark{{Name: "Discord"}}},
		{Group: "Admin Links", VisibleGroups: []string{"admins"}, Bookmarks: []models.Bookmark{{Name: "Proxmox"}}},
	}

	got := FilterBookmarkGroups(NewUser([]string{"family"}), input)
	if len(got) != 1 || got[0].Group != "Communicate" {
		t.Errorf("FilterBookmarkGroups(family) = %+v, want only Communicate", got)
	}

	got = FilterBookmarkGroups(NewUser([]string{"admins"}), input)
	if len(got) != 2 {
		t.Errorf("FilterBookmarkGroups(admins) = %+v, want both groups", got)
	}

	got = FilterBookmarkGroups(NewUser(nil), input)
	if len(got) != 1 || got[0].Group != "Communicate" {
		t.Errorf("FilterBookmarkGroups(groupless) = %+v, want only Communicate", got)
	}
}

func setOf(groups ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		set[g] = struct{}{}
	}
	return set
}

func httptestRequest(headers map[string]string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/apps", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}
