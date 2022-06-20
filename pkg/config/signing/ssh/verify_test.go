package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestErrorDecodingSignature(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   []byte
		expectedError error
	}{
		{name: "Error decoding empty file", fileContent: []byte{}, expectedError: ErrorDecodePem},
		{
			name: "Error missing Pem headers", fileContent: []byte("foobar"), expectedError: ErrorDecodePem,
		},
		{
			name: "Error when Pem block is not base64 encoded",
			fileContent: []byte(`
-----BEGIN SOME PEM BLOCK-----
<beautiful base64-encoded signature block here>
-----END SOME PEM BLOCK-----
`),
			expectedError: ErrorDecodePem,
		},
		{
			name: "Error wrong Pem type", fileContent: []byte(`
-----BEGIN SOME PEM BLOCK-----
PGJlYXV0aWZ1bCBiYXNlNjQtZW5jb2RlZCBzaWduYXR1cmUgYmxvY2sgaGVyZT4K
-----END SOME PEM BLOCK-----
`),
			expectedError: ErrorWrongPemBlock,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				signature, err := decodeArmoredSignature(test.fileContent)
				assert.Nil(t, signature)
				assert.ErrorIs(t, err, test.expectedError)
			},
		)
	}
}

func TestDecodingSignatureHappyPath(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   []byte
		pkType        string
		hashAlgorithm string
	}{
		{
			name: "ValidPemBlockWithValidSignature",
			fileContent: []byte(`
-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAgVtIKy9oAt9pd9vbNTHopvBQGb8
0lZHijl0LmeB4KVLkAAAAEZmlsZQAAAAAAAAAGc2hhNTEyAAAAUwAAAAtzc2gtZWQyNTUx
OQAAAEAoAJgfNlw8yoQtGRSMUOK0qFhWlET6a9oE9YDndqCqLy1aYXDV6YkLFOdHyXU85d
UgMCKY5ZztA8F3kD8KhhcO
-----END SSH SIGNATURE-----
				`),
			pkType:        ssh.KeyAlgoED25519,
			hashAlgorithm: "sha512",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				signature, err := decodeArmoredSignature(test.fileContent)
				assert.NotNil(t, signature)
				assert.NoError(t, err)
				assert.NotNil(t, signature.pk)
				assert.Equal(t, signature.pk.Type(), test.pkType)
				assert.NotNil(t, signature.signature)
				assert.Equal(t, signature.hashAlg, test.hashAlgorithm)
			},
		)
	}
}
