package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type githubAPI struct {
	URL    string
	Token  string
	Client *http.Client
}

func githubRequest(ctx context.Context, api githubAPI, method, path string, payload, result any) (err error) {
	var body []byte
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	request, err := http.NewRequestWithContext(ctx, method, api.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+api.Token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "migration-lab-ci")
	response, err := api.Client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, response.Body.Close())
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if response.StatusCode == http.StatusForbidden {
			return fmt.Errorf("GitHub %s %s: HTTP 403; the caller must grant pull-requests: write and repository policies must allow comments", method, path)
		}
		return fmt.Errorf("GitHub %s %s: HTTP %d", method, path, response.StatusCode)
	}
	if result == nil {
		_, err := io.Copy(io.Discard, io.LimitReader(response.Body, 8<<20))
		return err
	}
	return json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(result)
}
