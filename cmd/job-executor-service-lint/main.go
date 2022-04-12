package main

import (
	"io/ioutil"
	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/utils"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("exactly one argument needed")
	}

	jobConfigName := args[0]
	jobConfig, err := ioutil.ReadFile(jobConfigName)
	if err != nil {
		log.Fatalf("could not read job config %v: %v", jobConfigName, err)
	}

	conf, err := config.NewConfig(jobConfig)
	if err != nil {
		log.Fatalf("error parsing %v: %v", string(jobConfig), err)
	}

	err = utils.VerifySecurityConfiguration(conf, true)
	if err != nil {
		log.Fatalf("error processing security context: %v", err)
	}

	log.Printf("config %v is valid", jobConfigName)
}
