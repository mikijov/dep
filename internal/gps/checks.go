// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gps

import (
	"github.com/golang/dep/internal/gps/pkgtree"
)

type Ineffectuals struct {
	constraints []ProjectRoot
	ignores     []string
}

func (self Ineffectuals) AddConstraint(constraint ProjectRoot) {
	if self.constraints == nil {
		self.constraints = make([]ProjectRoot, 0, 10)
	}
	self.constraints = append(self.constraints, constraint)
}

func (self Ineffectuals) AddIgnore(ignore string) {
	if self.ignores == nil {
		self.ignores = make([]string, 0, 10)
	}
	self.ignores = append(self.ignores, ignore)
}

func (self Ineffectuals) IsEmpty() bool {
	if self.constraints != nil && len(self.constraints) > 0 {
		return false
	}
	if self.ignores != nil && len(self.ignores) > 0 {
		return false
	}
	return true
}

func FindIneffectualRules(manifest Manifest, packageTree pkgtree.PackageTree, stdLibFn func(string) bool) *Ineffectuals {
	// will track return value here
	// ineffectuals := make([]ProjectRoot, 0, 10)
	ineffectuals := &Ineffectuals{}

	// flatten list of actual imports which should be checked against
	reachmap, _ := packageTree.ToReachMap(true, true, false, nil /*ignores*/)
	reach := reachmap.FlattenFn(stdLibFn)
	imports := make(map[string]bool)
	for _, imp := range reach {
		imports[imp] = true
	}

	// if manifest is a RootManifest, use requires and ignores
	if rootManifest := manifest.(RootManifest); rootManifest != nil {
		// add required packages
		for projectRoot, _ := range rootManifest.RequiredPackages() {
			imports[projectRoot] = true
		}

		// check that ignores actually refer to packages we're importing
		for ignore, _ := range rootManifest.IgnoredPackages() {
			if _, found := imports[ignore]; !found {
				ineffectuals.AddIgnore(ignore)
				// ineffectuals = append(ineffectuals, ProjectRoot(ignore))
			}
		}

		// TODO: Should we remove ignores from the list? Ignores should not be
		// processed, but are constrained ignores ineffectual?
	}

	// at this point we have complete list of imports to test against

	// gather all constraints which should be checked
	constraints := make(map[ProjectRoot]bool) // it's a set to avoid duplicates
	for projectRoot, _ := range manifest.DependencyConstraints() {
		constraints[projectRoot] = true
	}

	// now check the constraints against the packageTree
	for projectRoot, _ := range constraints {
		if imports[string(projectRoot)] {
			ineffectuals.AddConstraint(projectRoot)
			// ineffectuals = append(ineffectuals, projectRoot)
		}
	}

	if ineffectuals.IsEmpty() {
		return nil
	} else {
		return ineffectuals
	}
}
