// Package traefik_oauth_upstream - Traefik plugin to manage upstream OAuth.
package traefik_oauth_upstream //nolint:stylecheck,revive

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

const CALLBACK_PATH = "/_oauth" //nolint:revive,stylecheck

// Config - the plugin configuration.
type Config struct {
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	AuthURL      string   `json:"authUrl"`
	TokenURL     string   `json:"tokenUrl"`
	PersistDir   string   `json:"persistDir"`
	Scopes       []string `json:"scopes"`
}

// CreateConfig - creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Scopes: []string{},
	}
}

// OauthUpstream - information about upstream OAuth.
type OauthUpstream struct {
	next       http.Handler
	config     *oauth2.Config
	name       string
	persistDir string
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.ClientID == "" || config.ClientSecret == "" || config.AuthURL == "" || config.TokenURL == "" || config.PersistDir == "" || len(config.Scopes) == 0 {
		return nil, fmt.Errorf("error loading traefik_oauth_upstream plugin: All of the following config must be defined: clientId, clientSecret, authUrl, tokenUrl, persistDir, scopes")
	}

	return &OauthUpstream{
		config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
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
		//nolint:contextcheck // false positive
		token, err := a.config.Exchange(context.Background(), callbackCode)
		if err != nil {
			http.Error(rw, "Failed to exchange auth code: "+err.Error(), http.StatusInternalServerError)
			return
		}
		Persist(token, a.persistDir)

		rw.WriteHeader(http.StatusOK)
		rw.Header().Add("Content-Type", "text/html")
		_, err = rw.Write([]byte("<html><h1>Tokens persisted</h1><p>You should now be able to access this resource as per usual</p></html>"))
		if err != nil {
			fmt.Printf("%s", err)
		}
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
		//nolint:misspell // UK english
		_, errW := rw.Write([]byte(fmt.Sprintf("<html><h1>Unauthorised</h1><p>This middleware's auth has not been initialised. Visit <a href=\"%s\">this auth link</a> to get things sorted.</p><p>Make sure the redirect URL is allowlisted: <pre>%s</pre></p></html>", url, a.config.RedirectURL)))
		if errW != nil {
			fmt.Printf("%s", errW)
		}
		return
	}

	tokenData, err := LoadTokenData(a.persistDir)
	if err != nil {
		http.Error(rw, "Failed to load persisted data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//nolint:contextcheck // false positive
	tokenSource := a.config.TokenSource(context.Background(), tokenData)
	token, err := tokenSource.Token()
	if err != nil {
		http.Error(rw, "Failed to refresh token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token.SetAuthHeader(req)

	// pass down the middleware chain
	a.next.ServeHTTP(rw, req)
}
