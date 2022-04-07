package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestSpecificImagesMustBeAccepted(t *testing.T) {
	allowedImageList := []string{"ghcr.io/my-other-user/my-other-image:*", "ghcr.io/my-user/my-image:1.2.3"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	// All images in allow list must be accepted
	assert.True(t, helper.Contains("ghcr.io/my-user/my-image:1.2.3"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:3"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:4"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:latest"))

	// Any other image must be rejected
	assert.False(t, helper.Contains("any-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/other-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:1.2.5"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image"))

	// Image but without a tag must not be accepted
	assert.False(t, helper.Contains("ghcr.io/my-other-user/my-other-image"))
}

func TestExplicitAllowlistEntries(t *testing.T) {
	allowedImageList := []string{"docker.io/[AB]/some-image", "docker.io/my-user/my-image:1.2.3"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("docker.io/A/some-image"))
	assert.True(t, helper.Contains("docker.io/B/some-image"))
	assert.False(t, helper.Contains("docker.io/A/some-image:some-specific-tag"))
	assert.False(t, helper.Contains("docker.io/B/some-image:some-specific-tag"))
	assert.True(t, helper.Contains("docker.io/my-user/my-image:1.2.3"))
	assert.False(t, helper.Contains("docker.io/my-user/some-image"))

	// Do not allow images without a registry
	assert.False(t, helper.Contains("B/some-image"))
	assert.False(t, helper.Contains("some-image:123"))
	assert.False(t, helper.Contains("B/some-image:123"))
	assert.False(t, helper.Contains("my-user/some-image"))
	assert.False(t, helper.Contains("my-user/my-image:1.2.3"))

	// Do not allow images from other registries
	assert.False(t, helper.Contains("ghcr.io/my-user/some-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
}

func TestWildcardFormatsRegistry(t *testing.T) {
	allowedImageList := []string{"docker.io/*", "*[AB]/some-image*", "my-docker-mirror.myorg/loadimpact/k6:0.34.*"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("docker.io/A/some-image"))
	assert.True(t, helper.Contains("B/some-image"))
	assert.True(t, helper.Contains("A/some-image:latest"))
	assert.True(t, helper.Contains("docker.io/my-user/my-image:1.2.3"))
	assert.True(t, helper.Contains("ghcr.io/B/some-image"))

	assert.False(t, helper.Contains("some-image"))
	assert.False(t, helper.Contains("ghcr.io/A/my-image:1.2.3"))
	assert.False(t, helper.Contains("ghcr.io/my-user/some-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
	assert.True(t, helper.Contains("my-docker-mirror.myorg/loadimpact/k6:0.34.0"))
	assert.True(t, helper.Contains("my-docker-mirror.myorg/loadimpact/k6:0.34.1"))
	assert.False(t, helper.Contains("my-docker-mirror.myorg/loadimpact/k6:0.35.0"))
	assert.False(t, helper.Contains("my-docker-mirror.myorg/loadimpact/k6"))
}

func TestDifferentContainerRegistryFormats(t *testing.T) {
	allowedImageList := []string{"ghcr.io/*", "a.b.c.d.e.f.g.registry/*", "a.registry:1337/*"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("a.b.c.d.e.f.g.registry/user/a"))
	assert.True(t, helper.Contains("a.registry:1337/user/a"))

	assert.False(t, helper.Contains("a.b.c.d.e.f.g.registry:8080/user/a"))
	assert.False(t, helper.Contains("a.registry:1338/user/a"))
	assert.False(t, helper.Contains("docker.io/user/a"))
}

func TestImageHashFormats(t *testing.T) {
	allowedImageList := []string{"ghcr.io/user/image@sha256*", "image@sha256:a428de4495b23ak*"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("image@sha256:a428de4495b23ak3402k37a5881c2d2cffa93757d34996156e4ea544577ab7f3"))
	assert.True(t, helper.Contains("ghcr.io/user/image@sha256:a428de44a9059fb1a59237a5881c2ddcffa93757d94026156e5ea544577ab3f3"))

	assert.False(t, helper.Contains("user/image@sha256:a42bdea4a4059631459237a5831224f4ffa93757d99056156efea544577ab7f3"))
	assert.False(t, helper.Contains("forbidden.registry/abc/image@sha256:4428de44a9059fb1a49237a5881c2d2cffa9375749902e156ff3a544577ab7f3"))
}

func TestBuildImageAllowListFunction(t *testing.T) {
	allowedImageArray := []string{"ghcr.io/my-other-user/my-other-image:*", "ghcr.io/my-user/my-image:1.2.3"}
	newFunctionResult, err := NewImageFilterList(allowedImageArray)
	require.NoError(t, err)

	allowedImageList := strings.Join(allowedImageArray, ",")
	buildFunctionResult, err := BuildImageAllowList(allowedImageList)
	require.NoError(t, err)

	assert.Equal(t, newFunctionResult, buildFunctionResult)
}

func TestEmptyImageAllowFunction(t *testing.T) {
	var allowedImageArray []string
	newFunctionResult, err := NewImageFilterList(allowedImageArray)
	require.NoError(t, err)

	buildFunctionResult, err := BuildImageAllowList("")
	require.NoError(t, err)

	assert.Len(t, newFunctionResult.patterns, 0)
	assert.Len(t, buildFunctionResult.patterns, 0)
	assert.Equal(t, newFunctionResult, buildFunctionResult)
}

func TestImplicitEmptyImageFilterList(t *testing.T) {
	allowedImageArray := []string{"docker.io/*", "ghcr.io", "*"}
	newFunctionResult, err := NewImageFilterList(allowedImageArray)
	require.NoError(t, err)

	assert.Len(t, newFunctionResult.patterns, 0)

	assert.True(t, newFunctionResult.Contains("<some/random/string._a-~#2ß1 dse"))
}
