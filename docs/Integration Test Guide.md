# Integration test guide

## Running integration tests

For running the integration tests, following environment variables have to be defined: 
* `KEPTN_ENDPOINT` - The API endpint from Keptn (e.g.: `http://10.201.4.2/api`)
* `KEPTN_API_TOKEN` - The auth token for Keptn
* `JES_E2E_TEST` - Must be set to `true` to enable the integration tests
* `JES_NAMESPACE` - The namespace in which the job-executor-service was installed (e.g.: `keptn-jes`)

After these environment variables have been defined, the test can be run with `go test -v test/e2e/*`

## Adding integration tests