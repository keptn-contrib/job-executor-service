package main

import (
	"flag"
	"io/ioutil"
	"keptn-contrib/job-executor-service/pkg/config"
	"keptn-contrib/job-executor-service/pkg/utils"
	"log"
)

func main() {

	// Parse the allowPrivilegedJobs flag that can be changed to match the behavior of the job-executor-service
	allowPrivilegedJobs := flag.Bool("allow-privileged-jobs", false,
		"Set to true if you want to allow privileged job workloads")

	flag.Parse()

	args := flag.Args()
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

	err = utils.VerifySecurityConfiguration(conf, *allowPrivilegedJobs)
	if err != nil {
		log.Fatalf("error processing security context: %v", err)
	}

	log.Printf("config %v is valid", jobConfigName)
}
