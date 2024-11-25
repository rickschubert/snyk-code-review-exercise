package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NpmPackageMetaResponse struct {
	Versions map[string]NpmPackageResponse `json:"versions"`
}

type NpmPackageResponse struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type NpmPackageVersion struct {
	Name         string                        `json:"name"`
	Version      string                        `json:"version"`
	Dependencies map[string]*NpmPackageVersion `json:"dependencies"`
}

type Client interface {
	FetchPackageMeta(p string) (*NpmPackageMetaResponse, error)
	FetchPackage(name, version string) (*NpmPackageResponse, error)
}

type realClient struct{}

func (r realClient) FetchPackageMeta(p string) (*NpmPackageMetaResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", p))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed NpmPackageMetaResponse
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (r realClient) FetchPackage(name, version string) (*NpmPackageResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", name, version))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed NpmPackageResponse
	_ = json.Unmarshal(body, &parsed)
	return &parsed, nil
}

func New() Client {
	return realClient{}
}
