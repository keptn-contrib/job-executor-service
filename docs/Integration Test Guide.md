# Integration test guide

## Running integration tests

The integration tests can be run in two different ways: locally with `go test` or via a GitHub Workflow.

### Locally
For running the integration tests locally, you have to clone the repository and install the required Go version.
Additionally, the following environment variables have to be defined: 
* `KEPTN_ENDPOINT` - The API endpint from Keptn (e.g.: `http://10.201.4.2/api`)
* `KEPTN_API_TOKEN` - The auth token for Keptn
* `JES_E2E_TEST` - Must be set to `true` to enable the integration tests
* `JES_NAMESPACE` - The namespace in which the job-executor-service was installed (e.g.: `keptn-jes`)

After these environment variables have been defined, the test can be run with `go test -v test/e2e/*` or you can
pick specific tests by defining a filter via the `-run` parameter:

```bash
go test -v -run "^\QTestEnvironmentVariables\E$"
```

*Note: Be aware that running `go test ./...` while `JES_E2E_TEST` is set to `true` will also run the integration tests!*

### GitHub Actions

Depending on the repository, executing the [Integration Tests](https://github.com/keptn-contrib/job-executor-service/actions/workflows/integration-tests.yaml)
GitHub action may require additional setup. On the *keptn-contrib/job-executor-service* repository this action should be able to 
run on any given branch that has at least one successful CI build.

However, if you forked the repository you have to configure / change the following properties:
* You have to change the `DOCKER_ORGANIZATION` in [.ci_env](../.ci_env) to a docker hub organization / account where
you can push images to
* Create the `REGISTRY_USER` and `REGISTRY_PASSWORD` secrets, which should contain the Docker user and PAT token
* Change the value of `image.repository` in the [helm values](../chart/values.yaml) to reflect the change of the docker organization
* Change the value of `jobexecutorserviceinitcontainer.image.repository` in the [helm values](../chart/values.yaml) to reflect the change of the docker organization

After this modifications you should be able to run the CI and the Integration tests workflow.

## Adding integration tests

To add integration tests to the job-executor-service, a few steps are necessary:
* Add your test into the integration test folder (`test/e2e`) such hat it can be executed by `go test`
* Add your test to the [integration-tests.yaml](../.github/workflows/integration-tests.yaml) by using the following
template:
```yaml
      - name: <Name of the test in the workflow>
        id: test_<name of the test in the report>
        continue-on-error: true
        working-directory: test/e2e
        run: go test -v -run <Function to test>
```

## Skipping integration tests

Integration tests can be skipped entirely or just for specific versions of Keptn. For that the `if` property of the step
has to be configured in the GitHub workflow: 
```yaml
  # This test will be skipped entirely 
  - name: Run "Hello World" test
    if: false
    id: test_hello_world
    continue-on-error: true
    working-directory: test/e2e
    run: go test -v -run "^\QTestHelloWorldDeployment\E$"

  # This test will be skipped for a specific Keptn version
  - name: Run "Files" test
    if: ${{ matrix.keptn-version != '0.12.4' }}
    id: test_files
    continue-on-error: true
    working-directory: test/e2e
    run: go test -v -run "^\QTestResourceFiles\E$"
```