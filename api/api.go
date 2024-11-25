package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/snyk/snyk-code-review-exercise/api/handlers/dependencies"
	"github.com/snyk/snyk-code-review-exercise/npm"
)

func New(client npm.Client, c *cache.Cache) http.Handler {
	router := mux.NewRouter()
	packageHandler := dependencies.New(client, c)
	router.Handle("/package/{package}/{version}", http.HandlerFunc(packageHandler.Handle()))
	return router
}
