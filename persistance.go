package traefik_oauth_upstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"

	"golang.org/x/oauth2"
)

const TOKEN_DATA_FILENAME = "token_data.json"

func TokenDataExists(persistDir string) (bool, error) {
	if _, err := os.Stat(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME)); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		return false, err
	}
}

func Persist(tokenData *oauth2.Token, persistDir string) {
	encoded, _ := json.Marshal(tokenData)
	_ = ioutil.WriteFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME), encoded, 0600)
}

func LoadTokenData(persistDir string) (*oauth2.Token, error) {
	encoded, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME))
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

func CalcRefreshTimestamp(expiryUnix int64) int64 {
	nowUnix := time.Now().Unix()
	diff := expiryUnix - nowUnix
	return nowUnix + int64(math.Round(0.9*float64(diff)))
}
