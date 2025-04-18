package beatport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	beatportBaseUrl   = "https://api.beatport.com/v4"
	beatsourceBaseUrl = "https://api.beatsource.com/v4"
)

type Beatport struct {
	store   Store
	client  *http.Client
	headers map[string]string
	auth    *Auth
}

type ServerError struct {
	Code    int
	Message string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

type Paginated[T any] struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Count    int     `json:"count"`
	Page     string  `json:"page"`
	PerPage  int     `json:"per_page"`
	Results  []T     `json:"results"`
}

func New(store Store, proxyUrl string, auth *Auth) *Beatport {
	transport := &http.Transport{}
	if proxyUrl != "" {
		proxyURL, _ := url.Parse(proxyUrl)
		proxy := http.ProxyURL(proxyURL)
		transport.Proxy = proxy
	}
	headers := map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language": "en-US,en;q=0.9",
		"cache-control":   "max-age=0",
		"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	}
	f := Beatport{
		store: store,
		auth:  auth,
		client: &http.Client{
			Timeout:   time.Duration(40) * time.Second,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		headers: headers,
	}
	return &f
}

func (b *Beatport) fetch(method, endpoint string, payload interface{}, contentType string) (*http.Response, error) {
	var body bytes.Buffer

	if endpoint != tokenEndpoint && endpoint != authEndpoint && endpoint != loginEndpoint {
		if err := b.auth.Check(b); err != nil {
			return nil, err
		}
	}

	if payload != nil {
		switch contentType {
		case "application/json":
			if err := json.NewEncoder(&body).Encode(payload); err != nil {
				return nil, &ServerError{Code: http.StatusBadRequest, Message: fmt.Sprintf("failed to encode json payload: %v", err)}
			}
		case "application/x-www-form-urlencoded":
			formData, err := encodeFormPayload(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to encode form payload: %w", err)
			}
			body.WriteString(formData.Encode())
		default:
			return nil, &ServerError{Code: http.StatusBadRequest, Message: fmt.Sprintf("unsupported content type: %s", contentType)}
		}
	}

	var baseUrl string
	switch b.store {
	default:
		baseUrl = beatportBaseUrl
	case StoreBeatsource:
		baseUrl = beatsourceBaseUrl
	}

	req, err := http.NewRequest(method, baseUrl+endpoint, &body)
	if err != nil {
		return nil, &ServerError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	for key, value := range b.headers {
		req.Header.Add(key, value)
	}

	if payload != nil {
		req.Header.Set("Content-Type", contentType)
	}

	if b.auth.tokenPair != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.auth.tokenPair.AccessToken))
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, &ServerError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("request failed: %v", err)}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		if resp.StatusCode == http.StatusUnauthorized && endpoint != tokenEndpoint && endpoint != authEndpoint && endpoint != loginEndpoint {
			b.auth.Invalidate()
			return b.fetch(method, endpoint, payload, contentType)
		}
		body, err := readResponseBody(resp)
		resp.Body.Close()
		message := fmt.Sprintf("request failed with status code: %d", resp.StatusCode)
		if err == nil && body != "" {
			message += fmt.Sprintf(", response body: %s", body)
		}
		return nil, &ServerError{Code: resp.StatusCode, Message: message}
	}

	return resp, nil
}

func encodeFormPayload(payload interface{}) (url.Values, error) {
	values := url.Values{}

	switch p := payload.(type) {
	case map[string]string:
		for key, value := range p {
			values.Set(key, value)
		}
	case url.Values:
		values = p
	default:
		return nil, errors.New("invalid payload")
	}

	return values, nil
}

func readResponseBody(resp *http.Response) (string, error) {
	if resp == nil || resp.Body == nil {
		return "", nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}
