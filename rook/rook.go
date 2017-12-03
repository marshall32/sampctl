// Package rook defines an API for working with Pawn libraries as 'packages' similar to how Go
// handles packages. It uses GitHub as a backend and tries to infer as much as possible from a repo
// such as where source files are located. Tags are encouraged for versioning but if absent, the git
// SHA1 hash is used.
package rook

import (
	"fmt"
	"path/filepath"

	"github.com/Southclaws/sampctl/compiler"
	"github.com/Southclaws/sampctl/util"
	"github.com/pkg/errors"
)

// Package represents a definition for a Pawn package and can either be used to define a build or
// as a description of a package in a repository. This is akin to npm's package.json and combines
// a project's dependencies with a description of that project.
//
// For example, a gamemode that includes a library does not need to define the User, Repo, Version,
// Contributors and Include fields at all, it can just define the Dependencies list in order to
// build correctly.
//
// On the flip side, a library written in pure Pawn should define some contributors and a web URL
// but, being written in pure Pawn, has no dependencies.
//
// Finally, if a repository stores its package source files in a subdirectory, that directory should
// be specified in the Include field. This is common practice for plugins that store the plugin
// source code in the root and the Pawn source in a subdirectory called 'include'.
type Package struct {
	// Local path, this indicates the Package object represents a local copy which is a directory
	// containing a `samp.json`/`samp.yaml` file and a set of Pawn source code files.
	// If this field is not set, then the Package is just an in-memory pointer to a remote package.
	local string

	// Inferred metadata, not always explicitly set via JSON/YAML but inferred from the dependency path
	PackageMeta

	// Metadata, set by the package author to describe the package
	Contributors []string `json:"contributors"` // list of contributors
	Website      string   `json:"website"`      // website or forum topic associated with the package

	// Functional, set by the package author to declare relevant files and dependencies
	Entry        string             `json:"entry"`        // entry point script to compile the project
	Output       string             `json:"output"`       // output amx file
	Dependencies []DependencyString `json:"dependencies"` // list of packages that the package depends on
	Builds       []compiler.Config  `json:"builds"`       // list of build configurations
	Resources    []Resource         `json:"resources"`    // list of additional resources associated with the package
}

// PackageMeta represents all the components required to locate a package version
type PackageMeta struct {
	User    string `json:"user"`    // Owner of the project repository
	Repo    string `json:"repo"`    // GitHub repository name
	Path    string `json:"path"`    // Subdirectory that contains .inc files (if any)
	Version string `json:"version"` // Version string (git tag, preferably a semantic version)
}

// Resource represents a resource associated with a package
type Resource struct {
	Name     string            `json:"name"`     // filename pattern of the resource
	Platform string            `json:"platform"` // target platform, if empty the resource is always used but if this is set and does not match the runtime OS, the resource is ignored
	Archive  bool              `json:"release"`  // is this resource an archive file or just a single file?
	Includes []string          `json:"includes"` // if archive: paths to directories containing .inc files for the compiler
	Plugins  []string          `json:"plugins"`  // if archive: paths to plugin binaries, either .so or .dll
	Files    map[string]string `json:"files"`    // if archive: path-to-path map of any other files, keys are paths inside the archive and values are extraction paths relative to the sampctl working directory
}

func (pkg Package) String() string {
	return fmt.Sprintf("%s/%s:%s", pkg.User, pkg.Repo, pkg.Version)
}

// Validate checks a package for missing fields
func (pkg Package) Validate() (err error) {
	if pkg.Entry == "" {
		return errors.New("package does not define an entry point")
	}

	if pkg.Output == "" {
		return errors.New("package does not define an output file")
	}

	if pkg.Entry == pkg.Output {
		return errors.New("package entry and output point to the same file")
	}

	return
}

// GetURL generates a GitHub URL for a package - it does not test the validity of the URL
func (pkg Package) GetURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", pkg.User, pkg.Repo)
}

// PackageFromDep creates a Package object from a Dependency String
func PackageFromDep(depString DependencyString) (pkg Package, err error) {
	dep, err := depString.Explode()
	pkg.User, pkg.Repo, pkg.Path, pkg.Version = dep.User, dep.Repo, dep.Path, dep.Version
	return
}

// EnsureDependencies traverses package dependencies and ensures they are up to date in `Package.local`/vendor
func (pkg Package) EnsureDependencies() (err error) {
	if pkg.local == "" {
		return errors.New("package does not represent a locally stored package")
	}

	if !util.Exists(pkg.local) {
		return errors.New("package local path does not exist")
	}

	vendorDir := filepath.Join(pkg.local, "dependencies")

	for _, depString := range pkg.Dependencies {
		dep, err := PackageFromDep(depString)
		if err != nil {
			return errors.Errorf("package dependency '%s' is invalid: %v", depString, err)
		}

		err = EnsurePackage(vendorDir, dep)
		if err != nil {
			return errors.Wrapf(err, "failed to ensure package %s", dep)
		}
	}
	return
}
