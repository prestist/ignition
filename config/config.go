// Copyright 2019 CoreOS, Inc.
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

package config

import (
	"bytes"

	butaneconfig "github.com/coreos/ignition/v2/butane/config"
	"github.com/coreos/ignition/v2/butane/config/common"
	exp "github.com/coreos/ignition/v2/config/v3_6_experimental"
	types_exp "github.com/coreos/ignition/v2/config/v3_6_experimental/types"

	"github.com/coreos/vcontext/report"
)

// Parse parses a config of any supported version and returns the equivalent config at the latest
// supported version. It automatically detects and transpiles Butane (YAML) configs.
func Parse(raw []byte) (types_exp.Config, report.Report, error) {
	// Try to detect if this is a YAML (Butane) config
	if isYAML(raw) {
		// Attempt to parse and transpile as Butane config
		ignitionJSON, rpt, err := butaneconfig.TranslateBytes(raw, common.TranslateBytesOptions{})
		if err != nil {
			// If Butane parsing failed, fall back to normal Ignition parsing
			// This handles the case where someone has a YAML-looking Ignition config
			return exp.ParseCompatibleVersion(raw)
		}

		// Parse the resulting Ignition JSON
		cfg, parseRpt, parseErr := exp.ParseCompatibleVersion(ignitionJSON)

		// Merge reports from translation and parsing
		rpt.Merge(parseRpt)

		return cfg, rpt, parseErr
	}

	// Fall back to normal Ignition JSON parsing
	return exp.ParseCompatibleVersion(raw)
}

// isYAML checks if the input looks like YAML rather than JSON
func isYAML(raw []byte) bool {
	// Trim leading whitespace
	trimmed := bytes.TrimLeft(raw, " \t\n\r")

	if len(trimmed) == 0 {
		return false
	}

	// JSON configs start with '{'
	// YAML configs typically start with a letter (like 'variant:')
	// or with '---' (YAML document separator)
	firstChar := trimmed[0]

	// If it starts with '{', it's definitely JSON
	if firstChar == '{' {
		return false
	}

	// If it starts with anything else (letters, '---', etc), assume YAML
	return true
}
