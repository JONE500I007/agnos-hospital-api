package his

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type HTTPClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *HTTPClient) FindPatient(ctx context.Context, id string) (*model.Patient, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/patient/search/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrHISUnavailable, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrHISUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var p model.Patient
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return nil, fmt.Errorf("%w: invalid response: %v", model.ErrHISUnavailable, err)
		}
		return &p, nil
	case http.StatusNotFound:
		return nil, model.ErrPatientNotFound
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", model.ErrHISUnavailable, resp.StatusCode)
	}
}
