package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const TOKEN_DATA_FILENAME = "token_data.json"

type TokenData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

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

func Persist(tokenData *TokenData, persistDir string) {
	encoded, _ := json.Marshal(tokenData)
	_ = ioutil.WriteFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME), encoded, 0600)
}

func LoadTokenData(persistDir string) (*TokenData, error) {
	encoded, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", persistDir, TOKEN_DATA_FILENAME))
	if err != nil {
		return nil, err
	}
	var tokenData *TokenData
	err = json.Unmarshal(encoded, tokenData)
	return tokenData, err
}
