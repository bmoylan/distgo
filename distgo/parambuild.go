// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package distgo

import (
	"fmt"

	"github.com/palantir/godel/v2/pkg/osarch"
	"github.com/pkg/errors"
)

// BuildOSArchID identifies the output of a build. Must be the string representation of an osarch.OSArch.
type BuildOSArchID string

type ByBuildOSArchID []BuildOSArchID

func (a ByBuildOSArchID) Len() int           { return len(a) }
func (a ByBuildOSArchID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBuildOSArchID) Less(i, j int) bool { return a[i] < a[j] }

type BuildParam struct {
	// NameTemplate is the template used for the executable output. The following template parameters can be used in the
	// template:
	//   * {{Product}}: the name of the product
	//   * {{Version}}: the version of the project
	NameTemplate string

	// OutputDir specifies the default build output directory for products executables built by the "build" task. The
	// executables generated by "build" are written to "{{OutputDir}}/{{ID}}/{{Version}}/{{OSArch}}/{{NameTemplate}}".
	OutputDir string

	// MainPkg is the location of the main package for the product relative to the project root directory. For example,
	// "distgo/main".
	MainPkg string

	// BuildArgsScript is the content of a script that is written to a file and run before this product is built
	// to provide supplemental build arguments for the product. The content of this value is written to a file and
	// executed. The script process uses the project directory as its working directory and inherits the environment
	// variables of the Go process. Each line of output of the script is provided to the "build" command as a separate
	// argument. For example, the following script would add the arguments "-ldflags" "-X" "main.year=$YEAR" to the
	// build command:
	//
	//   #!/usr/bin/env bash
	//   YEAR=$(date +%Y)
	//   echo "-ldflags"
	//   echo "-X"
	//   echo "main.year=$YEAR"
	BuildArgsScript string

	// VersionVar is the path to a variable that is set with the version information for the build. For example,
	// "github.com/palantir/godel/v2/cmd/godel.Version". If specified, it is provided to the "build" command as an
	// ldflag.
	VersionVar string

	// Environment specifies values for the environment variables that should be set for the build. For example,
	// a value of map[string]string{"CGO_ENABLED": "0"} would build with CGo disabled.
	Environment map[string]string

	// Script is the content of a script that is written to a file and run before the build processes start. The script
	// process inherits the environment variables of the Go process and also has project-related environment variables.
	// Refer to the documentation for the distgo.BuildScriptEnvVariables function for the extra environment variables.
	Script string

	// OSArchs specifies the GOOS and GOARCH pairs for which the product is built.
	OSArchs []osarch.OSArch
}

type BuildOutputInfo struct {
	BuildNameTemplateRendered string          `json:"buildNameTemplateRendered"`
	BuildOutputDir            string          `json:"buildOutputDir"`
	MainPkg                   string          `json:"mainPkg"`
	OSArchs                   []osarch.OSArch `json:"osArchs"`
}

func (p *BuildParam) ToBuildOutputInfo(productID ProductID, version string) (BuildOutputInfo, error) {
	renderedName, err := renderNameTemplate(p.NameTemplate, productID, version)
	if err != nil {
		return BuildOutputInfo{}, errors.Wrapf(err, "failed to render name template")
	}
	return BuildOutputInfo{
		BuildNameTemplateRendered: renderedName,
		BuildOutputDir:            p.OutputDir,
		MainPkg:                   p.MainPkg,
		OSArchs:                   p.OSArchs,
	}, nil
}

func (p *BuildParam) BuildArgs(productTaskOutputInfo ProductTaskOutputInfo) ([]string, error) {
	buildArgs, err := BuildArgsFromScript(productTaskOutputInfo, p.BuildArgsScript)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute script to generate build arguments")
	}
	if versionVar := p.VersionVar; versionVar != "" {
		buildArgs = append(buildArgs, "-ldflags", fmt.Sprintf("-X %s=%s", versionVar, productTaskOutputInfo.Project.Version))
	}
	return buildArgs, nil
}
