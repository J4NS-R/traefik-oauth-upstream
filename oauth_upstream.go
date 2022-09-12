// Package main a demo plugin.
package main

import (
	"context"
	"fmt"
	"net/http"
)

// Config - the plugin configuration.
type Config struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AuthUrl      string `json:"authUrl"`
	TokenUrl     string `json:"tokenUrl"`
	PersistDir   string `json:"persistDir"`
}

// CreateConfig - creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Demo a Demo plugin.
type OauthUpstream struct {
	next   http.Handler
	config *Config
	name   string
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.ClientId == "" || config.ClientSecret == "" || config.AuthUrl == "" || config.TokenUrl == "" || config.PersistDir == "" {
		return nil, fmt.Errorf("All of the following config must be defined: clientId, clientSecret, authUrl, tokenUrl, persistDir")
	}

	return &OauthUpstream{
		config: config,
		next:   next,
		name:   name,
	}, nil
}

func (a *OauthUpstream) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	tokenExists, err := TokenDataExists(a.config.PersistDir)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if !tokenExists {
		// TODO return 302 auth redirect
		http.Error(rw, "This should be a 302 to the authUrl", http.StatusNotImplemented)
		return
	}

	tokenData, err := LoadTokenData(a.config.PersistDir)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: check token expiry, and refresh if necessary

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.Token))

	// pass down the middleware chain
	a.next.ServeHTTP(rw, req)
}
