package handlers

import (
	"net/http"
	"time"

	loggerMiddleware "github.com/chi-middleware/logrus-logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	"github.com/toboshii/hajimari/internal/log"
	"github.com/toboshii/hajimari/internal/services"
	"github.com/toboshii/hajimari/internal/stores"
)

var (
	logger = log.New()
)

// contentSecurityPolicy allows the same-origin SPA, inline styles (Svelte
// component CSS), external app icons (https images), and the runtime iconify
// API used to resolve mdi/simple-icons names.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' https: data:; font-src 'self' data:; connect-src 'self' https://api.iconify.design; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'self'"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func NewHandler() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(loggerMiddleware.Logger("router", logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(securityHeaders)

	router.MethodNotAllowed(methodNotAllowedHandler)
	router.NotFound(notFoundHandler)

	var store stores.StartpageStore

	if viper.GetBool("memory") {
		store = stores.NewMemoryStore()
	} else {
		store = stores.NewFileStore()
	}

	startpageService := services.NewStartpageService(store, logger)
	appService := services.NewAppService(logger)

	router.Mount("/apps", NewAppResource(appService).AppRoutes())
	router.Mount("/bookmarks", NewBookmarkResource().BookmarkRoutes())
	router.Mount("/startpage", NewStartpageResource(startpageService).StartpageRoutes())

	return router
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(405)
	render.Render(w, r, ErrMethodNotAllowed)
}
