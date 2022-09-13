// Package main a demo plugin.
package traefik_oauth_upstream

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

const CALLBACK_PATH = "/_oauth"

// Config - the plugin configuration.
type Config struct {
	ClientId     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	AuthUrl      string   `json:"authUrl"`
	TokenUrl     string   `json:"tokenUrl"`
	PersistDir   string   `json:"persistDir"`
	Scopes       []string `json:"scopes"`
}

// CreateConfig - creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Scopes: []string{},
	}
}

// Demo a Demo plugin.
type OauthUpstream struct {
	next       http.Handler
	config     *oauth2.Config
	name       string
	persistDir string
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.ClientId == "" || config.ClientSecret == "" || config.AuthUrl == "" || config.TokenUrl == "" || config.PersistDir == "" || len(config.Scopes) == 0 {
		return nil, fmt.Errorf("Error loading traefik_oauth_upstream plugin: All of the following config must be defined: clientId, clientSecret, authUrl, tokenUrl, persistDir, scopes")
	}

	return &OauthUpstream{
		config: &oauth2.Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthUrl,
				TokenURL: config.TokenUrl,
			},
		},
		persistDir: config.PersistDir,
		next:       next,
		name:       name,
	}, nil
}

func (a *OauthUpstream) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if strings.HasPrefix(req.URL.Path, CALLBACK_PATH) {
		// Handle token exchange
		callbackCode := req.URL.Query().Get("code")
		token, err := a.config.Exchange(oauth2.NoContext, callbackCode)
		if err != nil {
			http.Error(rw, "Failed to exchange auth code: "+err.Error(), http.StatusInternalServerError)
			return
		}
		Persist(token, a.persistDir)

		rw.WriteHeader(http.StatusOK)
		rw.Header().Add("Content-Type", "text/html")
		rw.Write([]byte("<html><h1>Tokens persisted</h1><p>You should now be able to access this resource as per usual</p></html>"))
		return
	}

	tokenExists, err := TokenDataExists(a.persistDir)
	if err != nil {
		http.Error(rw, "Failed to access persisted data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !tokenExists {
		// auth redirect
		a.config.RedirectURL = fmt.Sprintf("https://%s%s", req.Host, CALLBACK_PATH)
		url := a.config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
		rw.WriteHeader(420)
		rw.Header().Add("Content-Type", "text/html")
		rw.Write([]byte(fmt.Sprintf("<html><h1>Unauthorised</h1><p>This middleware's auth has not been initialised. Visit <a href=\"%s\">this auth link</a> to get things sorted.</p><p>Make sure the redirect URL is allowlisted: <pre>%s</pre></p></html>", url, a.config.RedirectURL)))
		return
	}

	tokenData, err := LoadTokenData(a.persistDir)
	if err != nil {
		http.Error(rw, "Failed to load persisted data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenSource := a.config.TokenSource(oauth2.NoContext, tokenData)
	token, err := tokenSource.Token()
	if err != nil {
		http.Error(rw, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token.SetAuthHeader(req)

	// pass down the middleware chain
	a.next.ServeHTTP(rw, req)
}
