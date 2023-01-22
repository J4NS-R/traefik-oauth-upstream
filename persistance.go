package traefik_oauth_upstream //nolint:stylecheck,revive

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"golang.org/x/oauth2"
)

const TOKEN_DATA_FILENAME = "token_data.json" //nolint:revive,stylecheck,gosec // Filename is hardcoded, but not the contents.

// TokenDataExists - figures out whether token data exists on disk.
func TokenDataExists(persistDir string) (bool, error) {
	//nolint:gocritic // not the place for a switch
	if _, err := os.Stat(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME)); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		return false, err
	}
}

// Persist a token to a file.
func Persist(tokenData *oauth2.Token, persistDir string) {
	encoded, err := json.Marshal(tokenData)
	if err != nil {
		fmt.Printf("%s", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME), encoded, 0600)
	if err != nil {
		fmt.Printf("%s", err)
	}
}

// LoadTokenData - load token info from a file.
func LoadTokenData(persistDir string) (*oauth2.Token, error) {
	encoded, err := os.ReadFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME))
	if err != nil {
		return nil, err
	}
	tokenData := oauth2.Token{}
	err = json.Unmarshal(encoded, &tokenData)
	if err != nil {
		return nil, err
	}
	return &tokenData, nil
}

// CalcRefreshTimestamp - calculate at what point the token should be refreshed.
func CalcRefreshTimestamp(expiryUnix int64) int64 {
	nowUnix := time.Now().Unix()
	diff := expiryUnix - nowUnix
	return nowUnix + int64(math.Round(0.9*float64(diff)))
}
