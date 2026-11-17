// RoutingA:
// https://github.com/v2rayA/RoutingA
// Use of this source code is governed by MIT license that can be found in the LICENSE file.
package RoutingA

import (
	"fmt"
	"path"
	"strings"
)

// ProcessMatcher is the process(...) condition: the three ways a program can
// be named in a rule.
//
// The parser accepts any function name, so process(...) parsed long before it
// meant anything. This type, and Function.ProcessMatcher, make the condition
// part of the language: a consumer reads a typed value instead of matching the
// name process and hoping the parameters are shaped as it expects.
type ProcessMatcher struct {
	// Names hold executable file names, e.g. firefox. A bare parameter is a
	// name, because a name is what people write first.
	Names []string
	// Paths hold full paths to an executable file.
	Paths []string
	// Folders hold directories: an executable matches when it lives in one or
	// below it.
	Folders []string
}

// ProcessMatcher reports whether f is process(...) and, if it is, its
// condition. The bool is false for every other function, which is what a
// consumer switching over condition names needs.
//
// An unknown parameter name is an error rather than something ignored: a rule
// whose condition cannot be enforced must be refused, not applied more widely
// than written. The same goes for process() with no value at all — it would
// otherwise match every program.
func (f Function) ProcessMatcher() (ProcessMatcher, bool, error) {
	if f.Name != "process" {
		return ProcessMatcher{}, false, nil
	}
	var m ProcessMatcher
	for _, name := range f.Params {
		if name = strings.TrimSpace(name); name != "" {
			m.Names = append(m.Names, name)
		}
	}
	for key, values := range f.NamedParams {
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			switch key {
			case "name":
				m.Names = append(m.Names, value)
			case "path":
				m.Paths = append(m.Paths, value)
			case "folder":
				m.Folders = append(m.Folders, value)
			default:
				return ProcessMatcher{}, true, fmt.Errorf(
					"process(%s: %s): the only parameters are name, path and folder", key, value)
			}
		}
	}
	if m.IsEmpty() {
		return ProcessMatcher{}, true, fmt.Errorf("process(): no name, path or folder to match")
	}
	return m, true, nil
}

// IsEmpty reports a condition that names nothing.
func (m ProcessMatcher) IsEmpty() bool {
	return len(m.Names) == 0 && len(m.Paths) == 0 && len(m.Folders) == 0
}

// Match reports whether the program with this file name and full path is
// described by the condition. Any of the three lists matching is enough, so
// process(firefox, path: /usr/lib/firefox/firefox) matches twice over.
//
// A value may hold the shell-style wildcards path.Match understands: * stays
// inside one path element, ? is one character. A malformed pattern matches
// nothing rather than everything.
func (m ProcessMatcher) Match(name, fullPath string) bool {
	for _, pattern := range m.Names {
		if globMatch(pattern, name) {
			return true
		}
	}
	for _, pattern := range m.Paths {
		if globMatch(pattern, fullPath) {
			return true
		}
	}
	dir := path.Dir(fullPath)
	for _, folder := range m.Folders {
		if withinFolder(dir, folder) {
			return true
		}
	}
	return false
}

func globMatch(pattern, value string) bool {
	ok, err := path.Match(pattern, value)
	return err == nil && ok
}

// withinFolder reports whether dir is folder or below it. Both sides are
// cleaned first, so folder: /opt/app/ and /opt/app agree, and /opt/application
// is not inside /opt/app.
func withinFolder(dir, folder string) bool {
	dir, folder = path.Clean(dir), path.Clean(folder)
	if dir == folder {
		return true
	}
	if folder == "/" {
		return true
	}
	return strings.HasPrefix(dir, folder+"/")
}