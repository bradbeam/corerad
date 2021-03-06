// Copyright 2020 Matt Layher
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package crtest provides testing facilities for CoreRAD.
package crtest

import (
	"fmt"

	"inet.af/netaddr"
)

// MustIP parses a netaddr.IP from s or panics.
func MustIP(s string) netaddr.IP {
	ip, err := netaddr.ParseIP(s)
	if err != nil {
		panicf("crtest: failed to parse IP address: %v", err)
	}

	return ip
}

// MustIPPrefix parses a netaddr.IPPrefix from s or panics.
func MustIPPrefix(s string) netaddr.IPPrefix {
	p, err := netaddr.ParseIPPrefix(s)
	if err != nil {
		panicf("crtest: failed to parse IP prefix: %v", err)
	}

	return p
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
