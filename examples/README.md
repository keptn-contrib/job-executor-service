# Examples

*Note*: Unless specified otherwise, all commands should be executed from within this folder. It is assumed that job-executor-service is arleady installed and running (in remote execution plane).

## Hello World

**Setup**
```bash
keptn create project jes-hello-world --shipyard hello-world/shipyard.yaml
keptn create service foobar --project jes-hello-world
keptn add-resource --project jes-hello-world --service foobar --all-stages --resource hello-world/job-config.yaml --resourceUri job/config.yaml
keptn add-resource --project jes-hello-world --service foobar --all-stages --resource hello-world/greetings.txt --resourceUri greetings.txt
```

**Execution**
```bash
keptn trigger delivery --project jes-hello-world --service foobar --image foobar:1.2.3
```

**Expected Output**
You should see that the sequence has finished in Bridge with the following Cloud Event and output:

```
Job job-executor-service-job-a2585307-57be-40af-9c4f-eb33-1 finished successfully!

Logs:
Hello jes-hello-world:foobar:dev


Job job-executor-service-job-a2585307-57be-40af-9c4f-eb33-2 finished successfully!

Logs:
Spanish: hola.
French: bonjour.
German: guten tag.
Italian: salve.
Chinese: nǐn hǎo.
Portuguese: olá
Arabic: asalaam alaikum.
Japanese: konnichiwa.
```

## Locust

**Note**: Make sure jmeter-service is not running (e.g., uninstall using `helm -n keptn uninstall jmeter-service`)

**Setup**
```bash
keptn create project jes-locust --shipyard locust/shipyard.yaml
keptn create service helloservice --project jes-locust
keptn add-resource --project jes-locust --service helloservice --stage=qa --resource locust/job-config.yaml --resourceUri job/config.yaml
keptn add-resource --project jes-locust --service helloservice --stage=qa --resource locust/basic.py
keptn add-resource --project jes-locust --service helloservice --stage=qa --resource locust/locust.conf
```

**Execution**
```bash
keptn send event -f locust/cloud-event.json
```



## Kubectl Example

**Setup**
```bash
keptn create project jes-kubectl --shipyard kubectl/shipyard.yaml
keptn create service foobar --project jes-kubectl
keptn add-resource --project jes-kubectl --service foobar --all-stages --resource kubectl/job-config.yaml --resourceUri job/config.yaml
```

**Execution**
```bash
keptn trigger delivery --project jes-kubectl --service foobar --image foobar:1.2.3
```

**Expected Output**
In the default setup of job-executor-service this will fail, because jobs/workloads cannot access the Kubernetes API:
```
Job job-executor-service-job-700e65cb-0f85-46fb-94e2-266b-1 failed: job job-executor-service-job-700e65cb-0f85-46fb-94e2-266b-1 failed. Reason: BackoffLimitExceeded, Message: Job has reached the specified backoff limit

Logs: 
The connection to the server localhost:8080 was refused - did you specify the right host or port?
```

