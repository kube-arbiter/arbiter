/*
Copyright 2022 The Arbiter Authors.

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

package utils

import (
	"reflect"
	"testing"
)

func TestRemoveFinalizers(t *testing.T) {
	type input struct {
		finalizers []string
		f          string
		exp        []string
	}
	for _, tc := range []input{
		{
			finalizers: []string{},
			f:          "a",
			exp:        []string{},
		},
		{
			finalizers: []string{"a"},
			f:          "a",
			exp:        []string{},
		},
		{
			finalizers: []string{"a"},
			f:          "b",
			exp:        []string{"a"},
		},

		{
			finalizers: []string{"a", "b"},
			f:          "a",
			exp:        []string{"b"},
		},
		{
			finalizers: []string{"a", "b"},
			f:          "c",
			exp:        []string{"a", "b"},
		},
		{
			finalizers: []string{"a", "b"},
			f:          "b",
			exp:        []string{"a"},
		},

		{
			finalizers: []string{"a", "b", "c"},
			f:          "b",
			exp:        []string{"a", "c"},
		},
		{
			finalizers: nil,
			f:          "a",
			exp:        nil,
		},
		{
			finalizers: []string{"a", "b", "a", "c"},
			f:          "a",
			exp:        []string{"b", "c"},
		},
		{
			finalizers: []string{"a", "a", "a"},
			f:          "a",
			exp:        []string{},
		},
	} {
		if r := RemoveFinalizer(tc.finalizers, tc.f); !reflect.DeepEqual(tc.exp, r) {
			t.Fatalf("expect %v get %v", tc.exp, r)
		}
	}
}

func TestContainFinalizers(t *testing.T) {
	type input struct {
		finalizers []string
		f          string
		exp        bool
	}
	for _, tc := range []input{
		{
			finalizers: nil,
			f:          "a",
			exp:        false,
		},
		{
			finalizers: []string{},
			f:          "a",
			exp:        false,
		},
		{
			finalizers: []string{"a"},
			f:          "a",
			exp:        true,
		},
		{
			finalizers: []string{"a"},
			f:          "b",
			exp:        false,
		},
		{
			finalizers: []string{"a", "b"},
			f:          "b",
			exp:        true,
		},
	} {
		if r := ContainFinalizers(tc.finalizers, tc.f); !reflect.DeepEqual(tc.exp, r) {
			t.Fatalf("expect %v get %v", tc.exp, r)
		}
	}
}

func TestCommonOpt(t *testing.T) {
	type input struct {
		s, t, exp []string
	}

	for _, tc := range []input{
		{
			s:   []string{},
			t:   []string{},
			exp: []string{},
		},
		{
			s:   nil,
			t:   nil,
			exp: []string{},
		},
		{
			s:   []string{},
			t:   nil,
			exp: []string{},
		},
		{
			s:   nil,
			t:   []string{},
			exp: []string{},
		},
		{
			s:   []string{"a"},
			t:   nil,
			exp: []string{},
		},
		{
			s:   []string{"a"},
			t:   []string{"a"},
			exp: []string{"a"},
		},
		{
			s:   []string{"a"},
			t:   []string{"b"},
			exp: []string{},
		},
		{
			s:   []string{"a", "b"},
			t:   []string{"a", "b"},
			exp: []string{"a", "b"},
		},
		{
			s:   []string{"a", "b"},
			t:   []string{"a", "b", "c"},
			exp: []string{"a", "b"},
		},
		{
			s:   []string{"a", "b", "c"},
			t:   []string{"a", "b"},
			exp: []string{"a", "b"},
		},
	} {
		if r := CommonOpt(tc.t, tc.s); !reflect.DeepEqual(tc.exp, r) {
			t.Fatalf("expect %v get %v", tc.exp, r)
		}
	}
}

func TestAddFinalizers(t *testing.T) {
	type input struct {
		finalizers []string
		f          string
		exp        []string
	}
	for _, tc := range []input{
		{
			finalizers: []string{},
			f:          "a",
			exp:        []string{"a"},
		},
		{
			finalizers: nil,
			f:          "a",
			exp:        []string{"a"},
		},
		{
			finalizers: []string{"a"},
			f:          "a",
			exp:        []string{"a"},
		},
		{
			finalizers: []string{"a"},
			f:          "b",
			exp:        []string{"a", "b"},
		},
	} {
		if r := AddFinalizers(tc.finalizers, tc.f); !reflect.DeepEqual(tc.exp, r) {
			t.Fatalf("expect %v get %v", tc.exp, r)
		}
	}
}
