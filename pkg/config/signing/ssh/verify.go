package ssh

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"log"

	"golang.org/x/crypto/ssh"
)

const (
	magicHeader                         = "SSHSIG"
	signatureFileNamespace              = "file"
	sshSignaturePemType                 = "SSH SIGNATURE"
	supportedSignatureVersion           = 1
	jobConfigSignatureResourceName      = "job/config.yaml.sig"
	jobConfigAllowedSignersResourceName = "job/config.yaml.allowed_signers"
)

var ErrorDecodePem = errors.New("unable to decode pem file")
var ErrorWrongPemBlock = errors.New("unsupported pem block type found, expected " + sshSignaturePemType)
var supportedHashAlgorithms = map[string]func() hash.Hash{
	"sha256": sha256.New,
	"sha512": sha512.New,
}

type Signature struct {
	signature *ssh.Signature
	pk        ssh.PublicKey
	hashAlg   string
}

// https://github.com/openssh/openssh-portable/blob/master/PROTOCOL.sshsig#L34
type ArmoredSignature struct {
	MagicHeader   [6]byte
	Version       uint32
	PublicKey     string
	Namespace     string
	Reserved      string
	HashAlgorithm string
	Signature     string
}

// https://github.com/openssh/openssh-portable/blob/master/PROTOCOL.sshsig#L81
type MessageWrapper struct {
	Namespace     string
	Reserved      string
	HashAlgorithm string
	Hash          string
}

// KeptnResourceService defines the contract used by JobConfigReader to retrieve a resource from keptn (using project,
// service, stage from context)
type KeptnResourceService interface {
	GetKeptnResource(resource string) ([]byte, error)
}

type SignatureVerifier struct {
	ResourceService KeptnResourceService
}

func decodeArmoredSignature(armoredSignature []byte) (*Signature, error) {

	pemBlock, _ := pem.Decode(armoredSignature)
	if pemBlock == nil {
		return nil, ErrorDecodePem
	}

	if pemBlock.Type != sshSignaturePemType {
		return nil, fmt.Errorf("wrong block type %s: %w", pemBlock.Type, ErrorWrongPemBlock)
	}

	// Now we unmarshal it into the ArmoredSignature struct
	sig := ArmoredSignature{}
	if err := ssh.Unmarshal(pemBlock.Bytes, &sig); err != nil {
		return nil, err
	}

	if sig.Version != supportedSignatureVersion {
		return nil, fmt.Errorf("unsupported signature version: %d", sig.Version)
	}
	if string(sig.MagicHeader[:]) != magicHeader {
		return nil, fmt.Errorf("invalid magic header: %s", sig.MagicHeader[:])
	}
	if sig.Namespace != signatureFileNamespace {
		return nil, fmt.Errorf("invalid signature namespace: %s", sig.Namespace)
	}
	if _, ok := supportedHashAlgorithms[sig.HashAlgorithm]; !ok {
		return nil, fmt.Errorf("unsupported hash algorithm: %s", sig.HashAlgorithm)
	}

	// Now we can unpack the Signature and PublicKey blocks
	sshSig := ssh.Signature{}
	if err := ssh.Unmarshal([]byte(sig.Signature), &sshSig); err != nil {
		return nil, err
	}

	pk, err := ssh.ParsePublicKey([]byte(sig.PublicKey))
	if err != nil {
		return nil, err
	}

	return &Signature{
		signature: &sshSig,
		pk:        pk,
		hashAlg:   sig.HashAlgorithm,
	}, nil
}

func (sv *SignatureVerifier) VerifyJobConfigBytes(
	jobConfigData []byte,
	signatureData []byte,
	allowedSignersData []byte,
) error {
	signature, err := decodeArmoredSignature(signatureData)

	if err != nil {
		return fmt.Errorf("error decoding job config signature: %w", err)
	}

	h := supportedHashAlgorithms[signature.hashAlg]()
	h.Write(jobConfigData)
	jobConfigHash := h.Sum(nil)

	toVerify := MessageWrapper{
		Namespace:     "file",
		HashAlgorithm: signature.hashAlg,
		Hash:          string(jobConfigHash),
	}

	signedMessage := ssh.Marshal(toVerify)
	signedMessage = append([]byte(magicHeader), signedMessage...)

	// basic sanity test... does it verify with its own public key
	err = signature.pk.Verify(signedMessage, signature.signature)
	if err != nil {
		log.Printf("job config fails validating against its own pk!")
	}

	var publicKey ssh.PublicKey
	var keyComment string
	var keyOptions []string

	for {
		publicKey, keyComment, keyOptions, allowedSignersData, err = ssh.ParseAuthorizedKey(allowedSignersData)
		if err != nil {
			// by looking at the impl we get an error only when we don't find any more SSH keys
			return fmt.Errorf("error looking for pk that validates signature: %w", err)
		}

		// TODO we should probably filter on key type
		// (makes no sense to try and validate an rsa signature with an ed25519 key)

		log.Printf(
			"attempting job config signature verification with key %s, comment: %s, options: %s",
			publicKey.Type(), keyComment, keyOptions,
		)
		verificationErr := publicKey.Verify(signedMessage, signature.signature)
		if verificationErr == nil {
			log.Print("Job config signature verification succeeded")
			// Verification succeeded :)
			return nil
		}
		log.Print("Job config signature verification failed")
	}
}

func (sv *SignatureVerifier) VerifyJobConfig(
	jobConfigData []byte,
) error {

	jobConfigSignatureBytes, err := sv.ResourceService.GetKeptnResource(jobConfigSignatureResourceName)
	if err != nil {
		return fmt.Errorf("error retrieving job config signature: %w", err)
	}

	// TODO this probably should not be stored together with the signature (
	// if git repo is compromised an attacker can easily attach another key here)
	// it should probably be stored within the k8s cluster (configMap probably)

	jobConfigAllowedSignersBytes, err := sv.ResourceService.GetKeptnResource(jobConfigAllowedSignersResourceName)
	if err != nil {
		return fmt.Errorf("error retrieving job config allowed signers: %w", err)
	}

	return sv.VerifyJobConfigBytes(jobConfigData, jobConfigSignatureBytes, jobConfigAllowedSignersBytes)
}
