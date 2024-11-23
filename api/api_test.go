package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/snyk/snyk-code-review-exercise/api"
	"github.com/snyk/snyk-code-review-exercise/api/testdata"
	"github.com/snyk/snyk-code-review-exercise/npm"
	mock_npm "github.com/snyk/snyk-code-review-exercise/npm/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// In this test, both React and prop-types depend on object-assign
func TestPackageHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock_npm.NewMockClient(ctrl)

	client.EXPECT().FetchPackageMeta("react").Return(
		testdata.ReactPackageMeta(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("react", "16.13.0").Return(
		testdata.ReactPackageResponse(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackageMeta("object-assign").Return(
		testdata.ObjectAssignPackageMeta(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("object-assign", "4.1.1").Return(
		testdata.ObjectAssignPackageResponse(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackageMeta("prop-types").Return(
		testdata.PropTypesPackageMeta(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("prop-types", "15.6.2").Return(
		testdata.PropTypesPackageResponse(),
		nil,
	).AnyTimes()

	api := api.New(client, cache.New(5*time.Minute, 10*time.Minute))
	server := httptest.NewServer(api)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/package/react/16.13.0")
	require.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	var data npm.NpmPackageVersion
	err = json.Unmarshal(body, &data)
	require.Nil(t, err)

	assert.Equal(t, "react", data.Name)
	assert.Equal(t, "16.13.0", data.Version)

	expectedData := npm.NpmPackageVersion{
		Name:    "react",
		Version: "16.13.0",
		Dependencies: map[string]*npm.NpmPackageVersion{
			"object-assign": &npm.NpmPackageVersion{
				Name:         "object-assign",
				Version:      "4.1.1",
				Dependencies: map[string]*npm.NpmPackageVersion{},
			},
			"prop-types": &npm.NpmPackageVersion{
				Name:    "prop-types",
				Version: "15.6.2",
				Dependencies: map[string]*npm.NpmPackageVersion{
					"object-assign": &npm.NpmPackageVersion{
						Name:         "object-assign",
						Version:      "4.1.1",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
				},
			},
		},
	}

	assert.Equal(t, expectedData, data)
}

// In this test, both React and prop-types depend on object-assign, and on top
// of that prop-types creates a circular dependency back to React.
func TestPackageHandlerCircularDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock_npm.NewMockClient(ctrl)

	client.EXPECT().FetchPackageMeta("react").Return(
		testdata.ReactPackageMeta(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("react", "16.13.0").Return(
		testdata.ReactPackageResponse(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackageMeta("object-assign").Return(
		testdata.ObjectAssignPackageMeta(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("object-assign", "4.1.1").Return(
		testdata.ObjectAssignPackageResponse(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackageMeta("prop-types").Return(
		testdata.PropTypesPackageMetaCircularToReact(),
		nil,
	).AnyTimes()
	client.EXPECT().FetchPackage("prop-types", "15.6.2").Return(
		testdata.PropTypesPackageResponseCircularToReact(),
		nil,
	).AnyTimes()

	api := api.New(client, cache.New(5*time.Minute, 10*time.Minute))
	server := httptest.NewServer(api)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/package/react/16.13.0")
	require.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	var data npm.NpmPackageVersion
	err = json.Unmarshal(body, &data)
	require.Nil(t, err)

	assert.Equal(t, "react", data.Name)
	assert.Equal(t, "16.13.0", data.Version)

	expectedData := npm.NpmPackageVersion{
		Name:    "react",
		Version: "16.13.0",
		Dependencies: map[string]*npm.NpmPackageVersion{
			"object-assign": &npm.NpmPackageVersion{
				Name:         "object-assign",
				Version:      "4.1.1",
				Dependencies: map[string]*npm.NpmPackageVersion{},
			},
			"prop-types": &npm.NpmPackageVersion{
				Name:    "prop-types",
				Version: "15.6.2",
				Dependencies: map[string]*npm.NpmPackageVersion{
					"react": &npm.NpmPackageVersion{
						Name:         "react",
						Version:      "16.13.0",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
					"object-assign": &npm.NpmPackageVersion{
						Name:         "object-assign",
						Version:      "4.1.1",
						Dependencies: map[string]*npm.NpmPackageVersion{},
					},
				},
			},
		},
	}

	assert.Equal(t, expectedData, data)
}
