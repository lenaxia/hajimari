package visibility

import (
	"net/http"
	"strings"

	"github.com/toboshii/hajimari/internal/models"
)

// DefaultGroupsHeader is the forward-auth header group membership is read from
const DefaultGroupsHeader = "Remote-Groups"

// WildcardGroup matches any authenticated user when listed in visibility metadata
const WildcardGroup = "*"

// User struct holds normalized group memberships of the requesting user
type User struct {
	groups map[string]struct{}
}

// NewUser creates a User from a list of group names
func NewUser(groups []string) *User {
	normalized := ParseGroupsList(groups)
	set := make(map[string]struct{}, len(normalized))
	for _, g := range normalized {
		set[g] = struct{}{}
	}
	return &User{groups: set}
}

// NewUserFromRequest builds a User from the trusted forward-auth header of a request
func NewUserFromRequest(headerName string, r *http.Request) *User {
	if headerName == "" {
		headerName = DefaultGroupsHeader
	}
	return NewUser(ParseGroups(r.Header.Get(headerName)))
}

// MemberOfAny reports whether the user belongs to at least one of the given groups
func (u *User) MemberOfAny(groups []string) bool {
	if u == nil {
		return false
	}
	for _, g := range groups {
		if _, ok := u.groups[g]; ok {
			return true
		}
	}
	return false
}

// HasGroups reports whether the user is a member of any group at all
func (u *User) HasGroups() bool {
	return u != nil && len(u.groups) > 0
}

// CanSee reports whether visibility metadata grants the user access.
// Absence of metadata means visible to everyone (INV1); the wildcard group
// grants access to any authenticated user; otherwise group intersection applies.
func (u *User) CanSee(visibleGroups []string) bool {
	normalized := ParseGroupsList(visibleGroups)
	if len(normalized) == 0 {
		return true
	}
	for _, g := range normalized {
		if g == WildcardGroup {
			return u.HasGroups()
		}
	}
	return u.MemberOfAny(normalized)
}

// ParseGroups splits a comma separated group list into normalized group names
func ParseGroups(raw string) []string {
	return ParseGroupsList([]string{raw})
}

// ParseGroupsList normalizes a group list (split on commas, trim, lowercase), dropping empties
func ParseGroupsList(list []string) []string {
	normalized := make([]string, 0, len(list))
	for _, item := range list {
		for _, g := range strings.Split(item, ",") {
			g = strings.ToLower(strings.TrimSpace(g))
			if g != "" {
				normalized = append(normalized, g)
			}
		}
	}
	return normalized
}

// FilterAppGroups returns the app groups visible to the user without mutating the input
func FilterAppGroups(u *User, appGroups []models.AppGroup) []models.AppGroup {
	filtered := make([]models.AppGroup, 0, len(appGroups))
	for _, ag := range appGroups {
		apps := make([]models.App, 0, len(ag.Apps))
		for _, app := range ag.Apps {
			if u.CanSee(app.VisibleGroups) {
				apps = append(apps, app)
			}
		}
		if len(apps) > 0 {
			ag.Apps = apps
			filtered = append(filtered, ag)
		}
	}
	return filtered
}

// FilterBookmarkGroups returns the bookmark groups visible to the user without mutating the input
func FilterBookmarkGroups(u *User, bookmarkGroups []models.BookmarkGroup) []models.BookmarkGroup {
	filtered := make([]models.BookmarkGroup, 0, len(bookmarkGroups))
	for _, bg := range bookmarkGroups {
		if u.CanSee(bg.VisibleGroups) {
			filtered = append(filtered, bg)
		}
	}
	return filtered
}
