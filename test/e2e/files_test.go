package e2e

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/Masterminds/goutils"
	"github.com/keptn/go-utils/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

func TestResourceFiles(t *testing.T) {
	if !isE2ETestingAllowed() {
		t.Skip("Skipping TestResourceFiles, not allowed by environment")
	}

	testEnv := setupE2ETTestEnvironment(t,
		"../events/e2e.jes.triggered.json",
		"../shipyard/e2e.deployment.yaml",
		"../data/files.config.yaml",
	)

	defer testEnv.CleanupFunc()

	// Generate and upload some resource files:
	files := map[string]randomResourceFile{
		"small.file":     newRandomResourceFile(t, 1024),
		"folder/file.py": newRandomResourceFile(t, 128*1024),

		// NOTE: This seems to be the max file size that we can push to the API endpoint:
		// NOTE: Even via keptn add-resource, resource files can only be 767KiB big!
		"folder/big.file": newRandomResourceFile(t, 767*1024),
	}

	for path, file := range files {
		err := testEnv.API.AddServiceResource(testEnv.EventData.Project, testEnv.EventData.Stage,
			testEnv.EventData.Service, path, file.content)

		require.NoErrorf(t, err, "unable to create file %s with %d bytes", path, file.size)
	}

	// Send the event to keptn
	keptnContext, err := testEnv.API.SendEvent(testEnv.Event)
	require.NoError(t, err)

	// Checking if the job executor service responded with a .started event
	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.started",
		func(_ *models.KeptnContextExtendedCE) bool {
			return true
		},
	)

	// If the started event was sent by the job executor we wait for a .finished with the following data:
	expectedEventData := eventData{
		Project: testEnv.EventData.Project,
		Result:  "pass",
		Service: testEnv.EventData.Service,
		Stage:   testEnv.EventData.Stage,
		Status:  "succeeded",
	}

	filesRegex, err := regexp.Compile(`(?P<hash>[a-f0-9]+) {2}/keptn/(?P<file>[/a-z.]+)\n`)

	requireWaitForEvent(t,
		testEnv.API,
		2*time.Minute,
		1*time.Second,
		keptnContext,
		"sh.keptn.event.deployment.finished",
		func(event *models.KeptnContextExtendedCE) bool {
			responseEventData, err := parseKeptnEventData(event)
			require.NoError(t, err)

			// Gather all files from the log output
			matches := filesRegex.FindAllStringSubmatch(responseEventData.Message, -1)
			foundFiles := make(map[string]string)
			for _, match := range matches {
				foundFiles[match[2]] = match[1]
			}

			// transform the expected files into the same format
			expectedFiles := make(map[string]string)
			for name, file := range files {
				expectedFiles[name] = file.sha1
			}

			responseEventData.Message = ""

			// Assert that logging content and response data is as we expect
			assert.Equal(t, expectedFiles, foundFiles)
			assert.Equal(t, expectedEventData, *responseEventData)
			return true
		},
	)
}

type randomResourceFile struct {
	size    int
	content string
	sha1    string
}

// newRandomResourceFile generates a new randomResourceFile struct and fills the content with random bytes, the size
// is of the resulting file will be slightly bigger or smaller depending on the encoding of the random bytes in the
// string datatype
func newRandomResourceFile(t *testing.T, size int) randomResourceFile {
	// Since RandomNonAlphaNumeric generates a string from random data, the size in bytes on the string depends
	// on the contents. E.g.: if a lot of symbols are generated with runes the size is 3-4 times bigger then a single
	// byte. size / 3 is an approximation to generate files with around the requested size
	fileContent, err := goutils.RandomNonAlphaNumeric(size / 3)
	require.NoError(t, err)

	hasher := sha1.New()
	hasher.Write([]byte(fileContent))
	hash := hasher.Sum(nil)

	return randomResourceFile{
		size:    len(fileContent),
		content: fileContent,
		sha1:    hex.EncodeToString(hash),
	}
}
