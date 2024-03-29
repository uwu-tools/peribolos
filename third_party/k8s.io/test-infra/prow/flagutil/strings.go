/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flagutil

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Strings represents the value of a flag that accept multiple strings.
type Strings struct {
	vals    []string
	beenSet bool
}

// NewStrings returns a Strings struct that defaults to the value of def if left unset.
func NewStrings(def ...string) Strings {
	return Strings{
		vals:    def,
		beenSet: false,
	}
}

// NewStringsBeenSet returns a Strings struct with beenSet: true
func NewStringsBeenSet(def ...string) Strings {
	return Strings{
		vals:    def,
		beenSet: true,
	}
}

// Strings returns the slice of strings set for this value instance.
func (s *Strings) Strings() []string {
	return s.vals
}

// StringSet returns a sets.Set[string] of strings set for this value instance.
func (s *Strings) StringSet() sets.Set[string] {
	return sets.New[string](s.Strings()...)
}

// String returns a concatenated string of all the values joined by commas.
func (s *Strings) String() string {
	return strings.Join(s.vals, ",")
}

// Set records the value passed, overwriting the defaults (if any)
func (s *Strings) Set(value string) error {
	if !s.beenSet {
		s.beenSet = true
		// Value is being set, don't use default.
		s.vals = nil
	}
	s.vals = append(s.vals, value)
	return nil
}

// Add records the value passes, adding to the defaults (if any)
func (s *Strings) Add(value string) {
	s.vals = append(s.vals, value)
}
