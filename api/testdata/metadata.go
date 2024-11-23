package testdata

import "github.com/snyk/snyk-code-review-exercise/npm"

/*
Dependency outline:

React --> object-assign, prop-types
object-assign --> no dependencies
prop-types --> object-assign, React if circular
*/

func ReactPackageMeta() *npm.NpmPackageMetaResponse {
	return &npm.NpmPackageMetaResponse{
		Versions: map[string]npm.NpmPackageResponse{
			"16.13.0": *ReactPackageResponse(),
		},
	}
}

func ReactPackageResponse() *npm.NpmPackageResponse {
	return &npm.NpmPackageResponse{
		Name:    "react",
		Version: "16.13.0",
		Dependencies: map[string]string{
			"object-assign": "^4.1.1",
			"prop-types":    "^15.6.2",
		},
	}
}

func ObjectAssignPackageMeta() *npm.NpmPackageMetaResponse {
	return &npm.NpmPackageMetaResponse{
		Versions: map[string]npm.NpmPackageResponse{
			"4.1.1": *ObjectAssignPackageResponse(),
		},
	}
}

func ObjectAssignPackageResponse() *npm.NpmPackageResponse {
	return &npm.NpmPackageResponse{
		Name:         "object-assign",
		Version:      "4.1.1",
		Dependencies: map[string]string{},
	}
}

func PropTypesPackageMeta() *npm.NpmPackageMetaResponse {
	return &npm.NpmPackageMetaResponse{
		Versions: map[string]npm.NpmPackageResponse{
			"15.6.2": *PropTypesPackageResponse(),
		},
	}
}

func PropTypesPackageResponse() *npm.NpmPackageResponse {
	return &npm.NpmPackageResponse{
		Name:    "prop-types",
		Version: "15.6.2",
		Dependencies: map[string]string{
			"object-assign": "4.1.1",
		},
	}
}

func PropTypesPackageMetaCircularToReact() *npm.NpmPackageMetaResponse {
	return &npm.NpmPackageMetaResponse{
		Versions: map[string]npm.NpmPackageResponse{
			"15.6.2": *PropTypesPackageResponseCircularToReact(),
		},
	}
}

func PropTypesPackageResponseCircularToReact() *npm.NpmPackageResponse {
	return &npm.NpmPackageResponse{
		Name:    "prop-types",
		Version: "15.6.2",
		Dependencies: map[string]string{
			"react":         "16.13.0",
			"object-assign": "4.1.1",
		},
	}
}
