package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gojektech/heimdall/v6/httpclient"
	"net/http"
)

func Post(
	ctx context.Context,
	url string,
	requestBody []byte,
	headers map[string]string,
	httpClient *httpclient.Client,
) (response *http.Response, err error) {
	if httpClient == nil {
		err = errors.New("no http client provided")
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	response, err = httpClient.Do(req)
	if err != nil {
		fmt.Println("error while api", err.Error())
		return
	}

	return
}
