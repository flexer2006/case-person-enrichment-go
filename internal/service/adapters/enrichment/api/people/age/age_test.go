package age_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"encoding/json"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/age"
	agemodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func createMockResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestNewAgeAPIClient(t *testing.T) {
	t.Run("with nil client", func(t *testing.T) {
		client := age.NewAgeAPIClient(nil)
		require.NotNil(t, client, "client should not be nil")
	})

	t.Run("with custom client", func(t *testing.T) {
		mockClient := &MockHTTPClient{}
		client := age.NewAgeAPIClient(mockClient)
		require.NotNil(t, client, "client should not be nil")
	})
}

func TestAPIClient_GetAgeByName(t *testing.T) {
	testCases := []struct {
		name           string
		inputName      string
		mockResponse   func() (*http.Response, error)
		expectedAge    int
		expectedProb   float64
		expectedErrMsg string
	}{
		{
			name:           "empty name",
			inputName:      "",
			mockResponse:   func() (*http.Response, error) { return nil, nil },
			expectedAge:    0,
			expectedProb:   0,
			expectedErrMsg: "name cannot be empty",
		},
		{
			name:      "http client error",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return nil, errors.New("network error")
			},
			expectedAge:    0,
			expectedProb:   0,
			expectedErrMsg: "failed to execute request: network error",
		},
		{
			name:      "non 200 status",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(404, []byte(`{"error":"not found"}`)), nil
			},
			expectedAge:    0,
			expectedProb:   0,
			expectedErrMsg: "API returned non-200 status code: status 404",
		},
		{
			name:      "invalid json response",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(200, []byte(`invalid json`)), nil
			},
			expectedAge:    0,
			expectedProb:   0,
			expectedErrMsg: "failed to decode API response",
		},
		{
			name:      "successful response with low count",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				resp := agemodels.Response{
					Name:  "John",
					Age:   35,
					Count: 500,
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedAge:    35,
			expectedProb:   0.5,
			expectedErrMsg: "",
		},
		{
			name:      "successful response with high count",
			inputName: "Maria",
			mockResponse: func() (*http.Response, error) {
				resp := agemodels.Response{
					Name:  "Maria",
					Age:   28,
					Count: 2000,
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedAge:    28,
			expectedProb:   1.0,
			expectedErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return tc.mockResponse()
				},
			}

			client := age.NewAgeAPIClient(mockClient)
			require.NotNil(t, client)

			age, prob, err := client.GetAgeByName(context.Background(), tc.inputName)

			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAge, age)
				assert.Equal(t, tc.expectedProb, prob)
			}
		})
	}
}

func TestAPIClient_GetAgeByName_RequestParameters(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "https://api.agify.io", req.URL.Scheme+"://"+req.URL.Host)
			assert.Equal(t, "TestName", req.URL.Query().Get("name"))

			resp := agemodels.Response{
				Name:  "TestName",
				Age:   30,
				Count: 100,
			}
			body, _ := json.Marshal(resp)
			return createMockResponse(200, body), nil
		},
	}

	client := age.NewAgeAPIClient(mockClient)
	_, _, err := client.GetAgeByName(context.Background(), "TestName")
	assert.NoError(t, err)
}
