package types

import (
	"github.com/Southclaws/sampctl/versioning"
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
	// Parent indicates that this package is a "working" package that the user has explicitly
	// created and is developing. The opposite of this would be packages that exist in the
	// `dependencies` directory that have been downloaded as a result of an Ensure.
	Parent bool
	// Local path, this indicates the Package object represents a local copy which is a directory
	// containing a `samp.json`/`samp.yaml` file and a set of Pawn source code files.
	// If this field is not set, then the Package is just an in-memory pointer to a remote package.
	Local string
	// The vendor directory - for simple packages with no sub-dependencies, this is simply
	// `<local>/dependencies` but for nested dependencies, this needs to be set.
	Vendor string
	// format stores the original format of the package definition file, either `json` or `yaml`
	Format string
	// allDependencies stores a list of all dependency meta from this package and sub packages
	// this field is only used if `parent` is true.
	AllDependencies []versioning.DependencyMeta

	// Inferred metadata, not always explicitly set via JSON/YAML but inferred from the dependency path
	versioning.DependencyMeta

	// Metadata, set by the package author to describe the package
	Contributors []string `json:"contributors"` // list of contributors
	Website      string   `json:"website"`      // website or forum topic associated with the package

	// Functional, set by the package author to declare relevant files and dependencies
	Entry        string                        `json:"entry"`        // entry point script to compile the project
	Output       string                        `json:"output"`       // output amx file
	Dependencies []versioning.DependencyString `json:"dependencies"` // list of packages that the package depends on
	Builds       []BuildConfig                 `json:"builds"`       // list of build configurations
	Runtime      Runtime                       `json:"runtime"`      // runtime configuration for executing the package code
	Resources    []Resource                    `json:"resources"`    // list of additional resources associated with the package
}

// Resource represents a resource associated with a package
type Resource struct {
	Name     string            `json:"name"`     // filename pattern of the resource
	Platform string            `json:"platform"` // target platform, if empty the resource is always used but if this is set and does not match the runtime OS, the resource is ignored
	Archive  bool              `json:"archive"`  // is this resource an archive file or just a single file?
	Includes []string          `json:"includes"` // if archive: paths to directories containing .inc files for the compiler
	Plugins  []string          `json:"plugins"`  // if archive: paths to plugin binaries, either .so or .dll
	Files    map[string]string `json:"files"`    // if archive: path-to-path map of any other files, keys are paths inside the archive and values are extraction paths relative to the sampctl working directory
}