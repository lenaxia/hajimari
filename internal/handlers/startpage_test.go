package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/toboshii/hajimari/internal/config"
	"github.com/toboshii/hajimari/internal/models"
)

type fakeStartpageService struct{}

func (f *fakeStartpageService) NewStartpage(startpage *models.Startpage) (string, error) {
	return startpage.ID, nil
}

func (f *fakeStartpageService) GetStartpage(id string) (*models.Startpage, error) {
	return &models.Startpage{ID: id, Name: "test"}, nil
}

func (f *fakeStartpageService) UpdateStartpage(id string, startpage *models.Startpage) (*models.Startpage, error) {
	return startpage, nil
}

func (f *fakeStartpageService) RemoveStartpage(id string) (*models.Startpage, error) {
	return &models.Startpage{ID: id, Name: "test"}, nil
}

func (f *fakeStartpageService) ConvertConfigToStartpage(appConfig *config.Config, startpage *models.Startpage) {
}

func resetStartpageConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.Set("GroupsHeader", "Remote-Groups")
	viper.Set("AdminGroups", []string{"admins"})
}

func doStartpageRequest(t *testing.T, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	resetStartpageConfig(t)

	var body io.Reader
	if method == http.MethodPost || method == http.MethodPut {
		body = strings.NewReader(`{"name":"test"}`)
	}

	req, _ := http.NewRequest(method, target, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	recorder := httptest.NewRecorder()
	NewStartpageResource(&fakeStartpageService{}).StartpageRoutes().ServeHTTP(recorder, req)
	return recorder
}

func TestStartpageMutationsDeniedForNonAdmin(t *testing.T) {
	mutations := []struct {
		name   string
		method string
		target string
	}{
		{"create", http.MethodPost, "/"},
		{"update", http.MethodPut, "/startpage123"},
		{"delete", http.MethodDelete, "/startpage123"},
	}
	cases := []struct {
		name    string
		headers map[string]string
	}{
		{"family user", map[string]string{"Remote-Groups": "family"}},
		{"friend user", map[string]string{"Remote-Groups": "friends"}},
		{"no header", nil},
	}
	for _, m := range mutations {
		for _, c := range cases {
			t.Run(m.name+" denied for "+c.name, func(t *testing.T) {
				recorder := doStartpageRequest(t, m.method, m.target, c.headers)
				if recorder.Code != http.StatusForbidden {
					t.Errorf("expected 403, got %d: %s", recorder.Code, recorder.Body.String())
				}
				if !strings.Contains(recorder.Body.String(), "startpage mutations require admin group membership") {
					t.Errorf("expected forbidden message in body, got: %s", recorder.Body.String())
				}
			})
		}
	}
}

func TestStartpageMutationsAllowedForAdmin(t *testing.T) {
	mutations := []struct {
		name   string
		method string
		target string
	}{
		{"create", http.MethodPost, "/"},
		{"update", http.MethodPut, "/startpage123"},
		{"delete", http.MethodDelete, "/startpage123"},
	}
	admins := []struct {
		name    string
		headers map[string]string
	}{
		{"admin", map[string]string{"Remote-Groups": "admins"}},
		{"multi group admin", map[string]string{"Remote-Groups": "family,admins"}},
	}
	for _, m := range mutations {
		for _, a := range admins {
			t.Run(m.name+" allowed for "+a.name, func(t *testing.T) {
				recorder := doStartpageRequest(t, m.method, m.target, a.headers)
				if recorder.Code == http.StatusForbidden {
					t.Errorf("%s should pass the admin guard, got 403: %s", a.name, recorder.Body.String())
				}
			})
		}
	}
}

func TestStartpageMutationsCustomHeaderName(t *testing.T) {
	viper.Reset()
	viper.Set("GroupsHeader", "X-Team-Groups")
	viper.Set("AdminGroups", []string{"admins"})

	req, _ := http.NewRequest(http.MethodDelete, "/startpage123", nil)
	req.Header.Set("X-Team-Groups", "admins")

	recorder := httptest.NewRecorder()
	NewStartpageResource(&fakeStartpageService{}).StartpageRoutes().ServeHTTP(recorder, req)

	if recorder.Code == http.StatusForbidden {
		t.Errorf("admin via custom header should pass the guard, got 403: %s", recorder.Body.String())
	}
}

func TestStartpageGetRemainsOpen(t *testing.T) {
	gets := []struct {
		name   string
		method string
		target string
	}{
		{"default startpage", http.MethodGet, "/"},
		{"startpage by id", http.MethodGet, "/startpage123"},
	}
	for _, g := range gets {
		t.Run(g.name, func(t *testing.T) {
			recorder := doStartpageRequest(t, g.method, g.target, nil)
			if recorder.Code != http.StatusOK {
				t.Errorf("GET should stay open, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
