package utils

import (
	"github.com/gobwas/glob"
	"log"
	"strings"
)

const defaultContainerRegistry = "docker.io"

// ImageFilterList represents a list of glob filters that can be used to check if an image matches one of these filters
type ImageFilterList struct {
	patterns []glob.Glob
}

// BuildImageAllowList creates a ImageFilterList from a comma separated string that is present as environment variable
func BuildImageAllowList(envVariable string) (*ImageFilterList, error) {
	// Extract allow list from env variable, strip empty strings from the
	// list, since they are useless, and we really don't want them
	var allowListStrings []string
	for _, str := range strings.Split(envVariable, ",") {
		if str != "" {
			allowListStrings = append(allowListStrings, str)
		}
	}

	// Remind the user that he is probably running an unsafe configuration
	if len(allowListStrings) == 0 {
		log.Println("Found empty allowlist for images, all images are allowed!")
	}

	return NewImageFilterList(allowListStrings)
}

// NewImageFilterList creates a new list of wildcards
func NewImageFilterList(patterns []string) (*ImageFilterList, error) {

	// transform all patterns to a compiled regex instance
	globPatterns := make([]glob.Glob, len(patterns))
	for index, pattern := range patterns {

		// If the pattern is a * we can skip everything and just create an empty list
		if pattern == "*" {
			log.Println("Warning: Found '*' in the allowlist, all images will be accepted!")
			return &ImageFilterList{
				patterns: []glob.Glob{},
			}, nil
		}

		compiledGlob, err := glob.Compile(pattern)
		if err != nil {
			return nil, err
		}

		globPatterns[index] = compiledGlob
	}

	return &ImageFilterList{
		patterns: globPatterns,
	}, nil
}

// Contains returns true if the entry matches a single element the list or if the list is empty
func (w ImageFilterList) Contains(entry string) bool {

	// Empty list equals to accept everything
	if len(w.patterns) == 0 {
		return true
	}

	// First try to match the filter list as the user has defined it
	for _, pattern := range w.patterns {
		if matches := pattern.Match(entry); matches {
			return true
		}
	}

	return false
}
