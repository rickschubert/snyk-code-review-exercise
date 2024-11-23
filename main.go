package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/snyk/snyk-code-review-exercise/api"
	"github.com/snyk/snyk-code-review-exercise/npm"
)

func main() {
	client := npm.New()
	// TODO: We could consider using two different caches for package meta and
	// specific package information.
	c := cache.New(5*time.Minute, 10*time.Minute)

	handler := api.New(client, c)
	fmt.Println("Server running on http://localhost:3000/")
	if err := http.ListenAndServe("localhost:3000", handler); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
