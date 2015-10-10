package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
)

// GetToken from existing data
func GetToken() (*oauth2.Token, error) {
	tokenStr, err := readTokenFile()
	if err != nil {
		return nil, err
	}
	token, err := tokenFromJSON(tokenStr)
	return token, err
}

// SaveToken persit token to json file
func SaveToken(token *oauth2.Token) error {
	tokenJSON, _ := tokenToJSON(token)
	f, err := os.Create("access-token.json")
	if err != nil {
		return err
	}

	_, writeErr := f.WriteString(tokenJSON)
	return writeErr
}

func readTokenFile() (string, error) {
	bytes, err := ioutil.ReadFile("access-token.json")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func tokenToJSON(token *oauth2.Token) (string, error) {
	d, err := json.Marshal(token)
	if err == nil {
		return string(d), nil
	}
	return "", err
}

func tokenFromJSON(jsonStr string) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(jsonStr), &token); err != nil {
		return nil, err
	}
	return &token, nil
}
