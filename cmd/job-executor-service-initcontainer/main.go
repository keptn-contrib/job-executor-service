package main

import (
	"context"
	oauthutils "github.com/keptn/go-utils/pkg/common/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"keptn-contrib/job-executor-service/pkg/file"
	"keptn-contrib/job-executor-service/pkg/keptn"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	api "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/spf13/afero"
)

type envConfig struct {
	// Whether we are running locally (e.g., for testing) or on production
	Env string `envconfig:"ENV" default:"local"`
	// URL of the Keptn configuration service (this is where we can fetch files from the config repo)
	ConfigurationServiceURL string `envconfig:"CONFIGURATION_SERVICE" required:"true"`
	// The token of the keptn API
	KeptnAPIToken string `envconfig:"KEPTN_API_TOKEN" required:"true"`
	// The keptn project contained in the initial cloud event
	Project string `envconfig:"KEPTN_PROJECT" required:"true"`
	// The keptn stage contained in the initial cloud event
	Stage string `envconfig:"KEPTN_STAGE" required:"true"`
	// The keptn service contained in the initial cloud event
	Service string `envconfig:"KEPTN_SERVICE" required:"true"`
	// The keptn service contained in the initial cloud event
	Action string `envconfig:"JOB_ACTION" required:"true"`
	// The name of the config task which triggered the init container run
	Task string `envconfig:"JOB_TASK" required:"true"`
	// The authentication mode that should be used by the init container, must match the
	// authentication mode of the job executor, otherwise communication with keptn might not work
	AuthMode string `envconfig:"AUTH_MODE" required:"false" default:"token"`
	// The OAuth client id
	OAuthClientID string `envconfig:"OAUTH_CLIENT_ID" required:"false"`
	// The OAuth client secret
	OAuthClientSecret string `envconfig:"OAUTH_CLIENT_SECRET" required:"false"`
	// The OAuth scopes, must be defined in a comma separated list
	OAuthScopes []string `envconfig:"OAUTH_SCOPES" required:"false"`
	// The well known oauth discovery url for the init container
	OAuthDiscovery string `envconfig:"OAUTH_DISCOVERY" required:"false"`
}

func main() {

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	fs := afero.NewOsFs()

	// TODO: When using the new api interface we only need the base url and not the configuration url
	configurationServiceURL, _ := url.Parse(env.ConfigurationServiceURL)
	baseURL := strings.TrimSuffix(configurationServiceURL.String(), "configuration-service")

	apiOptions := []func(*api.APISet){
		api.WithScheme(configurationServiceURL.Scheme),
	}

	// Chose the authentication method from the give environment; This can either be token or oauth
	if env.AuthMode == "token" {

		// We only append a token when needing / having one, no token is needed
		// if the ini-container lives in the same namespace as keptn, if the task
		// is part of a remote execution plane then a token is needed
		if env.KeptnAPIToken != "" {
			apiOptions = append(apiOptions, api.WithAuthToken(env.KeptnAPIToken, "x-token"))
		}

	} else if env.AuthMode == "oauth" {

		// To avoid stalling the Job for too long we wait at max 10 seconds to query the token endpoint
		// from the given discovery url. This is the same timeout that is used in the distributor.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		oauthDiscovery := oauthutils.NewOauthDiscovery(&http.Client{})
		discoveryRes, err := oauthDiscovery.Discover(ctx, env.OAuthDiscovery)
		if err != nil {
			log.Fatalf("unable to query information from well known oauth URL: %s", err)
		}

		conf := clientcredentials.Config{
			ClientID:     env.OAuthClientID,
			ClientSecret: env.OAuthClientSecret,
			Scopes:       env.OAuthScopes,
			TokenURL:     discoveryRes.TokenEndpoint,
		}

		apiOptions = append(apiOptions, api.WithHTTPClient(conf.Client(context.Background())))
	} else {
		log.Fatalf("unkown authentication mode: %s", env.AuthMode)
	}

	keptnApi, err := api.New(baseURL, apiOptions...)
	if err != nil {
		log.Fatalf("unable to create keptn API: %s", err)
	}

	useLocalFileSystem := false

	// configure keptn options
	if env.Env == "local" {
		log.Println("env=local: Running with local filesystem to fetch resources")
		useLocalFileSystem = true
	}

	configService := keptn.NewConfigService(useLocalFileSystem, env.Project, env.Stage, env.Service, keptnApi.ResourcesV1())

	err = file.MountFiles(env.Action, env.Task, fs, configService)
	if err != nil {
		log.Printf("Error while copying files: %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}
