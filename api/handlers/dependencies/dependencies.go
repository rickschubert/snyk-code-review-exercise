package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/snyk/snyk-code-review-exercise/api/handlers"
	"github.com/snyk/snyk-code-review-exercise/npm"
)

type handleImpl struct {
	client npm.Client
	c      *cache.Cache
}

type visited struct {
	visitedItems map[string]struct{}
	mutex        sync.Mutex
}

func New(client npm.Client, c *cache.Cache) handlers.Handler {
	return &handleImpl{
		client: client,
		c:      c,
	}
}

func (h handleImpl) Handle() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		pkgName := vars["package"]
		pkgVersion := vars["version"]

		rootPkg := &npm.NpmPackageVersion{Name: pkgName, Dependencies: map[string]*npm.NpmPackageVersion{}}
		if err := h.resolveDependencies(rootPkg, pkgVersion); err != nil {
			handlers.SendErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		stringified, err := json.MarshalIndent(rootPkg, "", "  ")
		if err != nil {
			handlers.SendErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(200)
		w.Write(stringified)
	}
}

func (h handleImpl) resolveDependencies(pkg *npm.NpmPackageVersion, versionConstraint string) error {
	visits := visited{
		mutex:        sync.Mutex{},
		visitedItems: map[string]struct{}{},
	}

	var encounteredErrors []error

	var wg sync.WaitGroup
	wg.Add(1)

	errorsChannel := make(chan error)

	go func() {
		for encounteredError := range errorsChannel {
			if encounteredError != nil {
				encounteredErrors = append(encounteredErrors, encounteredError)
			}
		}
	}()
	go h.resolveDependenciesInParallel(pkg, versionConstraint, &visits, &wg, errorsChannel)

	wg.Wait()
	close(errorsChannel)

	if len(encounteredErrors) == 0 {
		return nil
	}
	return errors.Join(encounteredErrors...)
}

func (h handleImpl) resolveDependenciesInParallel(
	pkg *npm.NpmPackageVersion,
	versionConstraint string,
	visits *visited,
	wg *sync.WaitGroup,
	errorsChannel chan<- error,
) {
	defer wg.Done()
	alreadyVisited := markPackageConstraintAsVisited(
		visits,
		pkg.Name,
		versionConstraint,
	)
	if alreadyVisited {
		pkg.Version = versionConstraint
		fmt.Printf("WARNING: Detected circular dependency for package '%s', version %s", pkg.Name, pkg.Version)
		return
	}

	pkgMeta, err := h.getPackageMeta(pkg.Name)
	if err != nil {
		errorsChannel <- fmt.Errorf("retrieving package meta: %w", err)
		return
	}

	concreteVersion, err := highestCompatibleVersion(versionConstraint, pkgMeta)
	if err != nil {
		errorsChannel <- fmt.Errorf("getting highest compatible version: %w", err)
		return
	}
	pkg.Version = concreteVersion

	packageInfo, err := h.getPackageInfo(pkg.Name, pkg.Version)
	if err != nil {
		errorsChannel <- fmt.Errorf("retrieving package info: %w", err)
		return
	}

	var innerWg sync.WaitGroup
	innerWg.Add(len(packageInfo.Dependencies))
	for dependencyName, dependencyVersionConstraint := range packageInfo.Dependencies {
		dep := &npm.NpmPackageVersion{Name: dependencyName, Dependencies: map[string]*npm.NpmPackageVersion{}}
		pkg.Dependencies[dependencyName] = dep
		go h.resolveDependenciesInParallel(dep, dependencyVersionConstraint, visits, &innerWg, errorsChannel)
	}
	innerWg.Wait()
}

func highestCompatibleVersion(constraintStr string, versions *npm.NpmPackageMetaResponse) (string, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return "", fmt.Errorf("creating new semver constraint: %w", err)
	}
	filtered := filterCompatibleVersions(constraint, versions)
	sort.Sort(filtered)
	if len(filtered) == 0 {
		return "", errors.New("no compatible versions found")
	}
	return filtered[len(filtered)-1].String(), nil
}

func filterCompatibleVersions(constraint *semver.Constraints, pkgMeta *npm.NpmPackageMetaResponse) semver.Collection {
	var compatible semver.Collection
	for version := range pkgMeta.Versions {
		semVer, err := semver.NewVersion(version)
		if err != nil {
			continue
		}
		if constraint.Check(semVer) {
			compatible = append(compatible, semVer)
		}
	}
	return compatible
}

func (h handleImpl) getPackageMeta(packageName string) (*npm.NpmPackageMetaResponse, error) {
	if metaResponse, found := h.c.Get(packageName); found {
		return metaResponse.(*npm.NpmPackageMetaResponse), nil
	} else {
		fetchedMeta, err := h.client.FetchPackageMeta(packageName)
		if err != nil {
			return nil, fmt.Errorf("fetching package meta: %w", err)
		}
		h.c.Set(packageName, fetchedMeta, cache.NoExpiration)
		return fetchedMeta, nil
	}
}

func (h handleImpl) getPackageInfo(packageName, packageVersion string) (*npm.NpmPackageResponse, error) {
	packageAndConcreteVersion := fmt.Sprintf("%s@%s", packageName, packageVersion)
	if packageResponse, found := h.c.Get(packageAndConcreteVersion); found {
		return packageResponse.(*npm.NpmPackageResponse), nil
	} else {
		npmPkg, err := h.client.FetchPackage(packageName, packageVersion)
		if err != nil {
			return nil, fmt.Errorf("fetching package: %w", err)
		}
		h.c.Set(packageAndConcreteVersion, npmPkg, cache.NoExpiration)
		return npmPkg, nil
	}
}

func markPackageConstraintAsVisited(visits *visited, packageName, versionConstraint string) (alreadyVisited bool) {
	visits.mutex.Lock()
	defer visits.mutex.Unlock()

	packageAndVersionWithConstraint := packageName + "@" + versionConstraint
	if _, ok := visits.visitedItems[packageAndVersionWithConstraint]; ok {
		return true
	}
	visits.visitedItems[packageAndVersionWithConstraint] = struct{}{}
	return false
}
