/*
Package interp provides a complete Go interpreter.

For the Go language itself, refer to the official Go specification
https://golang.org/ref/spec.

Custom build tags

Custom build tags allow to control which files in imported source
packages are interpreted, in the same way as the "-tags" option of the
"go build" command. Setting a custom build tag spans globally for all
future imports of the session.

A build tag is a line comment that begins

	// yaegi:tags

that lists the build constraints to be satisfied by the further
imports of source packages.

For example the following custom build tag

	// yaegi:tags noasm

Will ensure that an import of a package will exclude files containing

	// +build !noasm

And include files containing

	// +build noasm
*/
package interp

// BUG(marc): Type checking is not implemented yet.
