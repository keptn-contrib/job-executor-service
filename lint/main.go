package main

import (
	"io/ioutil"
	"keptn-sandbox/job-executor-service/pkg/config"
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

	_, err = config.NewConfig(jobConfig)
	if err != nil {
		log.Fatalf("error parsing %v: %v", string(jobConfig), err)
	}

	log.Printf("config %v is valid", jobConfigName)
}
