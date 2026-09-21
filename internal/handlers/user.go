package handlers

import (
	"net/http"
	"strings"

	"github.com/toboshii/hajimari/internal/config"
	"github.com/toboshii/hajimari/internal/visibility"
)

// resolveVisibilityUser returns the effective user for the request. Normally
// this is the user resolved from the trusted forward-auth header. When a
// `group` query parameter is present, admins (members of a configured admin
// group) impersonate the requested group membership instead.
func resolveVisibilityUser(appConfig *config.Config, r *http.Request) (*visibility.User, int, bool) {
	requester := visibility.NewUserFromRequest(appConfig.GroupsHeader, r)

	requested := r.URL.Query()["group"]
	if len(requested) == 0 {
		return requester, 0, true
	}

	if !requester.MemberOfAny(appConfig.AdminGroups) {
		return nil, http.StatusForbidden, false
	}

	var groups []string
	for _, g := range requested {
		groups = append(groups, strings.Split(g, ",")...)
	}
	return visibility.NewUser(groups), 0, true
}

// requireAdmin reports whether the requesting user is a member of one of the
// configured admin groups, as read from the trusted forward-auth header.
func requireAdmin(appConfig *config.Config, r *http.Request) bool {
	return visibility.NewUserFromRequest(appConfig.GroupsHeader, r).MemberOfAny(appConfig.AdminGroups)
}
