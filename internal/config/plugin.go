// Copyright 2019 Matt Layher
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

package config

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/corerad/internal/plugin"
)

// parsePlugin parses raw plugin configuration into a slice of plugins.
func parsePlugins(ifi rawInterface, maxInterval time.Duration) ([]plugin.Plugin, error) {
	var plugins []plugin.Plugin

	for _, p := range ifi.Prefixes {
		pfx, err := parsePrefix(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prefix %q: %v", p.Prefix, err)
		}

		plugins = append(plugins, pfx)
	}

	for _, r := range ifi.RDNSS {
		rdnss, err := parseRDNSS(r, maxInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RDNSS: %v", err)
		}

		plugins = append(plugins, rdnss)
	}

	for _, d := range ifi.DNSSL {
		dnssl, err := parseDNSSL(d, maxInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DNSSL: %v", err)
		}

		plugins = append(plugins, dnssl)
	}

	// Loopback has an MTU of 65536 on Linux. Good enough?
	if ifi.MTU < 0 || ifi.MTU > 65536 {
		return nil, fmt.Errorf("MTU (%d) must be between 0 and 65536", ifi.MTU)
	}
	if ifi.MTU != 0 {
		m := plugin.MTU(ifi.MTU)
		plugins = append(plugins, &m)
	}

	return plugins, nil
}

// parseDNSSL parses a DNSSL plugin.
func parseDNSSL(d rawDNSSL, maxInterval time.Duration) (*plugin.DNSSL, error) {
	lifetime, err := parseDuration(d.Lifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid lifetime: %v", err)
	}

	// If auto, compute lifetime as recommended by radvd.
	if lifetime == durationAuto {
		lifetime = 2 * maxInterval
	}

	return &plugin.DNSSL{
		Lifetime:    lifetime,
		DomainNames: d.DomainNames,
	}, nil
}

// parsePrefix parses a Prefix plugin.
func parsePrefix(p rawPrefix) (*plugin.Prefix, error) {
	ip, prefix, err := net.ParseCIDR(p.Prefix)
	if err != nil {
		return nil, err
	}

	// Don't allow individual IP addresses.
	if !prefix.IP.Equal(ip) {
		return nil, fmt.Errorf("%q is not a CIDR prefix", ip)
	}

	// Only allow IPv6 addresses.
	if prefix.IP.To16() != nil && prefix.IP.To4() != nil {
		return nil, fmt.Errorf("%q is not an IPv6 CIDR prefix", prefix.IP)
	}

	valid, err := parseDuration(p.ValidLifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid valid lifetime: %v", err)
	}

	// Use defaults for auto values.
	switch valid {
	case 0:
		return nil, errors.New("valid lifetime must be non-zero")
	case durationAuto:
		valid = 24 * time.Hour
	}

	preferred, err := parseDuration(p.PreferredLifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid preferred lifetime: %v", err)
	}

	// Use defaults for auto values.
	switch preferred {
	case 0:
		return nil, errors.New("preferred lifetime must be non-zero")
	case durationAuto:
		preferred = 4 * time.Hour
	}

	// See: https://tools.ietf.org/html/rfc4861#section-4.6.2.
	if preferred > valid {
		return nil, fmt.Errorf("preferred lifetime of %s exceeds valid lifetime of %s",
			preferred, valid)
	}

	onLink := true
	if p.OnLink != nil {
		onLink = *p.OnLink
	}

	auto := true
	if p.Autonomous != nil {
		auto = *p.Autonomous
	}

	return &plugin.Prefix{
		Prefix:            prefix,
		OnLink:            onLink,
		Autonomous:        auto,
		ValidLifetime:     valid,
		PreferredLifetime: preferred,
	}, nil
}

// parseDNSSL parses a DNSSL plugin.
func parseRDNSS(d rawRDNSS, maxInterval time.Duration) (*plugin.RDNSS, error) {
	lifetime, err := parseDuration(d.Lifetime)
	if err != nil {
		return nil, fmt.Errorf("invalid lifetime: %v", err)
	}

	// If auto, compute lifetime as recommended by radvd.
	if lifetime == durationAuto {
		lifetime = 2 * maxInterval
	}

	// Parse all server addresses as IPv6 addresses.
	servers := make([]net.IP, 0, len(d.Servers))
	for _, s := range d.Servers {
		ip := net.ParseIP(s)
		if ip == nil || (ip.To16() != nil && ip.To4() != nil) {
			return nil, fmt.Errorf("string %q is not an IPv6 address", s)
		}

		servers = append(servers, ip)
	}

	return &plugin.RDNSS{
		Lifetime: lifetime,
		Servers:  servers,
	}, nil
}
