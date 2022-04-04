package utils

import (
	"errors"
	"fmt"
	"github.com/gobwas/glob"
	"regexp"
	"strings"
)

const defaultContainerRegistry = "docker.io"

type ImageFilterList struct {
	registryMatcher *regexp.Regexp
	userMatcher     *regexp.Regexp
	patterns        []glob.Glob
}

func NewAllowAllImageFilterList() (*ImageFilterList, error) {
	return NewImageFilterList([]string{})
}

// NewImageFilterList creates a new list of wildcards
func NewImageFilterList(patterns []string) (*ImageFilterList, error) {

	registryMatcher, err := regexp.Compile("^.*\\..*/")
	if err != nil {
		return nil, err
	}

	userMatcher, err := regexp.Compile("^.*\\.*/.*/")
	if err != nil {
		return nil, err
	}

	// transform all patterns to a compiled regex instance
	globPatterns := make([]glob.Glob, len(patterns))
	for index, pattern := range patterns {

		// If the pattern is a * we can skip everything and just create an empty list
		if pattern == "*" {
			return &ImageFilterList{
				patterns: []glob.Glob{},
			}, nil
		}

		// Check if pattern contains a registry, if that's not the case
		// then we prepend the default registry to the pattern
		if !registryMatcher.MatchString(pattern) {
			pattern = defaultContainerRegistry + "/" + pattern
		}

		// Check if a user is contained in the pattern, if not we insert an implicit * at the location
		// except if we have already a pattern that ends with * (user or registry), then we shouldn't do
		// anything
		if !userMatcher.MatchString(pattern) && !strings.HasSuffix(pattern, "/*") {
			parts := strings.Split(pattern, "/")
			if len(parts) < 2 {
				return nil, errors.New("unable to separate registry and image from pattern: " + pattern)
			}

			registry := parts[0]
			image := strings.Join(parts[1:], "")

			pattern = registry + "/*/" + image
		}
		fmt.Printf("%s", pattern)

		// If the pattern ends with <image> or <image>:* we drop the suffix (:*) and replace it with an implicit *
		// to be able to match <image>, <image>:latest, <image>@sha, <image>:1.2.3
		if strings.HasSuffix(pattern, ":*") || !strings.Contains(pattern, ":") {
			pattern = strings.TrimSuffix(pattern, ":*")
			pattern = pattern + "*"
		}

		compiledGlob, err := glob.Compile(pattern)
		if err != nil {
			return nil, err
		}

		globPatterns[index] = compiledGlob
	}

	return &ImageFilterList{
		registryMatcher: registryMatcher,
		userMatcher:     userMatcher,
		patterns:        globPatterns,
	}, nil
}

// Contains returns true if the entry matches a single element the list or if the list is empty
func (w ImageFilterList) Contains(entry string) bool {

	// Empty list equals to accept everything
	if len(w.patterns) == 0 {
		return true
	}

	for _, pattern := range w.patterns {

		// Since we build the full URI for the docker image names, we have to make sure
		// that they are also extended in the entry which we try to check
		if !w.registryMatcher.MatchString(entry) {
			entry = defaultContainerRegistry + "/" + entry
		}

		if !w.userMatcher.MatchString(entry) {
			parts := strings.Split(entry, "/")
			if len(parts) >= 2 {
				registry := parts[0]
				image := strings.Join(parts[1:], "")
				entry = registry + "/<any-user>/" + image
			}
		}

		if matches := pattern.Match(entry); matches {
			return true
		}
	}

	return false
}