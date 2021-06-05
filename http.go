package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var client *http.Client

func httpClient() *http.Client {
	if client != nil {
		return client
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	client = &http.Client{
		Timeout:   time.Second * 30,
		Transport: t,
	}
	return client
}

func HttpGet(ctx context.Context, url, params string, responseModel interface{}) error {
	url = fmt.Sprintf("%s?%s", url, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	req = req.WithContext(ctx)
	res, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
		return errors.New("status code: " + res.Status)
	}

	if responseModel == nil {
		return nil
	}

	//buf := new(bytes.Buffer)
	//buf.ReadFrom(res.Body)

	//fmt.Println(buf.String())
	//return json.Unmarshal(buf.Bytes(), &responseModel)
	return json.NewDecoder(res.Body).Decode(&responseModel)
}

func HttpPost(ctx context.Context, url, params string, body, responseModel interface{}) error {
	url = fmt.Sprintf("%s?%s", url, params)
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(&body)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	req = req.WithContext(ctx)
	res, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
		return errors.New("status code: " + res.Status)
	}

	if responseModel == nil {
		return nil
	}
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(res.Body)

	//fmt.Println(buf.String())
	//json.Unmarshal(buf.Bytes(), &responseModel)
	return json.NewDecoder(res.Body).Decode(&responseModel)
}
