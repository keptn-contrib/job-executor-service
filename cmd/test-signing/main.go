package main

import (
	"flag"
	"log"
	"os"

	"keptn-contrib/job-executor-service/pkg/config/signing/ssh"
)

var jobConfigSignatureFile = flag.String(
	"signature", "jobconfig.yaml.sig",
	"file containing the signature of the job configuration",
)

var jobConfigAllowedSignersFile = flag.String(
	"signers", "jobconfig.yaml.allowed_signers",
	"file containing the public keys of the allowed signers",
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		log.Fatal("exactly one argument needed")
	}

	jobConfigName := args[0]

	jobconfigbytes, err := os.ReadFile(jobConfigName)
	if err != nil {
		log.Fatalf("Error reading job configuration file %s: %v", jobConfigName, err)
	}

	jobconfigsignaturebytes, err := os.ReadFile(*jobConfigSignatureFile)
	if err != nil {
		log.Fatalf("Error reading signature file %s: %v", *jobConfigSignatureFile, err)
	}

	jobconfigallowedsignersbytes, err := os.ReadFile(*jobConfigAllowedSignersFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", *jobConfigAllowedSignersFile, err)
	}

	sv := new(ssh.SignatureVerifier)
	err = sv.VerifyJobConfigBytes(
		jobconfigbytes, jobconfigsignaturebytes,
		jobconfigallowedsignersbytes,
	)

	if err != nil {
		log.Fatalf("Error validating: %v", err)
	}

	log.Print("Job config validated successfully")
}
