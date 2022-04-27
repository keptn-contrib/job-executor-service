# Integration test guide

## Running integration tests

Integration tests can be run in two different ways: locally with `go test` or via triggering the "Integration Tests" GitHub Workflow.

### Locally
For running the integration tests locally, you have to clone the repository and install the required Go version.
Additionally, the following environment variables have to be defined: 
* `KEPTN_ENDPOINT` - The API endpoint from Keptn (e.g.: `http://10.201.4.2/api`)
* `KEPTN_API_TOKEN` - The auth token for Keptn
* `JES_E2E_TEST` - Must be set to `true` to enable the integration tests
* `JES_NAMESPACE` - The namespace in which the job-executor-service was installed (e.g.: `keptn-jes`)

After these environment variables have been defined, the test can be run with `go test -v test/e2e/...` or you can
pick specific tests by defining a filter via the `-run` parameter:

```bash
go test -v -run "^\QTestEnvironmentVariables\E$"
```

*Note: Be aware that running `go test ./...` while `JES_E2E_TEST` is set to `true` will also run the integration tests!*

### GitHub Actions

Depending on the repository, executing the [Integration Tests](https://github.com/keptn-contrib/job-executor-service/actions/workflows/integration-tests.yaml)
GitHub action may require additional setup. On the *keptn-contrib/job-executor-service* repository this action should be able to 
run on any given branch that has at least one successful CI build.

If you forked the repository you have to configure / change the following properties:
* You have to change the `DOCKER_ORGANIZATION` in [.ci_env](../.ci_env) to a docker hub organization / account where
you can push images to
* Create the `REGISTRY_USER` and `REGISTRY_PASSWORD` secrets, which should contain the Docker user and PAT token
* Change the value of `image.repository` in the [helm values](../chart/values.yaml) to reflect the change of the docker organization
* Change the value of `jobexecutorserviceinitcontainer.image.repository` in the [helm values](../chart/values.yaml) to reflect the change of the docker organization

After these modifications you should be able to run the CI and the Integration tests workflow.

## Adding integration tests

To add integration tests to the job-executor-service, you have to add your test into the integration test folder 
(`test/e2e`) such hat it can be executed by `go test`.

## Skipping integration tests

Integration tests can be skipped entirely (`t.Skip(...)`) or just for specific versions of Keptn. For skipping the integration tests
for a specific version, the `ShouldRun` function can be used to determine if the given Keptn version satisfies the version
requirements.
```go
    testEnv, err := newTestEnvironment(/* ... */)    

	// Skip this test conditionally depending on the version
    err = testEnv.ShouldRun(">=0.12.4")
	if err != nil {
		t.Skip(err.Error())
	}
	
	testEnv.SetupTestEnvironment()
```