package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAllowAllList(t *testing.T) {
	helper, err := NewAllowAllImageFilterList()
	require.NoError(t, err)

	assert.Len(t, helper.patterns, 0)
	assert.True(t, helper.Contains("my-testimage:123"))
	assert.True(t, helper.Contains("custom.registry/user/my-testimage:latest"))

	helper, err = NewImageFilterList([]string{"*"})
	require.NoError(t, err)

	assert.Len(t, helper.patterns, 0)
	assert.True(t, helper.Contains("my-testimage:123"))
	assert.True(t, helper.Contains("custom.registry/user/my-testimage:latest"))

	helper, err = NewImageFilterList([]string{})
	require.NoError(t, err)

	assert.Len(t, helper.patterns, 0)
	assert.True(t, helper.Contains("my-testimage:123"))
	assert.True(t, helper.Contains("custom.registry/user/my-testimage:latest"))
}

func TestSpecificImagesMustBeAccepted(t *testing.T) {
	allowedImageList := []string{"ghcr.io/my-other-user/my-other-image:*", "ghcr.io/my-user/my-image:1.2.3"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	// All images in allow list must be accepted
	assert.True(t, helper.Contains("ghcr.io/my-user/my-image:1.2.3"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:3"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:4"))
	assert.True(t, helper.Contains("ghcr.io/my-other-user/my-other-image:latest"))

	// Any other image must be rejected
	assert.False(t, helper.Contains("any-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/other-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:1.2.5"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image"))
}

func TestDefaultRegistry(t *testing.T) {
	allowedImageList := []string{"some-image", "my-user/my-image:1.2.3"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("some-image"))
	assert.True(t, helper.Contains("B/some-image"))
	assert.True(t, helper.Contains("some-image:123"))
	assert.True(t, helper.Contains("B/some-image:123"))
	assert.True(t, helper.Contains("docker.io/A/some-image"))
	assert.True(t, helper.Contains("docker.io/B/some-image"))
	assert.True(t, helper.Contains("docker.io/B/some-image:latest"))
	assert.True(t, helper.Contains("docker.io/B/some-image:1.2.3"))
	assert.True(t, helper.Contains("docker.io/my-user/some-image"))
	assert.True(t, helper.Contains("docker.io/my-user/my-image:1.2.3"))

	assert.False(t, helper.Contains("ghcr.io/my-user/some-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
}

func TestAcceptEverythingFromDefaultRegistry(t *testing.T) {
	allowedImageList := []string{"docker.io/*"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("docker.io/A/some-image"))
	assert.True(t, helper.Contains("B/some-image"))
	assert.True(t, helper.Contains("A/some-image:latest"))
	assert.True(t, helper.Contains("some-image"))
	assert.True(t, helper.Contains("docker.io/my-user/my-image:1.2.3"))

	assert.False(t, helper.Contains("ghcr.io/my-user/some-image"))
	assert.False(t, helper.Contains("ghcr.io/my-user/my-image:latest"))
}

func TestVariousImageFormats(t *testing.T) {
	allowedImageList := []string{"image", "docker.io/*", "ghcr.io/*", "a.b.c.d.e.f.g.registry/*", "a.registry:1337/*"}
	helper, err := NewImageFilterList(allowedImageList)
	require.NoError(t, err)

	assert.True(t, helper.Contains("image@sha256:a428de4495b23ak3402k37a5881c2d2cffa93757d34996156e4ea544577ab7f3"))
	assert.True(t, helper.Contains("user/image@sha256:a42bdea4a4059631459237a5831224f4ffa93757d99056156efea544577ab7f3"))
	assert.True(t, helper.Contains("ghcr.io/user/image@sha256:a428de44a9059fb1a59237a5881c2ddcffa93757d94026156e5ea544577ab3f3"))
	assert.True(t, helper.Contains("a.b.c.d.e.f.g.registry/user/a"))
	assert.True(t, helper.Contains("a.registry:1337/user/a"))

	assert.False(t, helper.Contains("forbidden.registry/abc/image@sha256:4428de44a9059fb1a49237a5881c2d2cffa9375749902e156ff3a544577ab7f3"))
	assert.False(t, helper.Contains("a.b.c.d.e.f.g.registry:8080/user/a"))
	assert.False(t, helper.Contains("a.registry:1338/user/a"))
}