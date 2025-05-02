package nationality_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/nationality"
	nationalitymodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/nationality"
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

func TestNewNationalityAPIClient(t *testing.T) {
	t.Run("with nil client", func(t *testing.T) {
		client := nationality.NewNationalityAPIClient(nil)
		require.NotNil(t, client, "client should not be nil")
	})

	t.Run("with custom client", func(t *testing.T) {
		mockClient := &MockHTTPClient{}
		client := nationality.NewNationalityAPIClient(mockClient)
		require.NotNil(t, client, "client should not be nil")
	})
}

func TestAPIClient_GetNationalityByName(t *testing.T) {
	testCases := []struct {
		name                string
		inputName           string
		mockResponse        func() (*http.Response, error)
		expectedNationality string
		expectedProb        float64
		expectedErrMsg      string
	}{
		{
			name:                "empty name",
			inputName:           "",
			mockResponse:        func() (*http.Response, error) { return nil, nil },
			expectedNationality: "",
			expectedProb:        0,
			expectedErrMsg:      "name cannot be empty",
		},
		{
			name:      "http client error",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return nil, errors.New("network error")
			},
			expectedNationality: "",
			expectedProb:        0,
			expectedErrMsg:      "failed to execute request: network error",
		},
		{
			name:      "non 200 status",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(404, []byte(`{"error":"not found"}`)), nil
			},
			expectedNationality: "",
			expectedProb:        0,
			expectedErrMsg:      "API returned non-200 status code: status 404",
		},
		{
			name:      "invalid json response",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(200, []byte(`invalid json`)), nil
			},
			expectedNationality: "",
			expectedProb:        0,
			expectedErrMsg:      "failed to decode API response",
		},
		{
			name:      "empty countries list",
			inputName: "UnknownName",
			mockResponse: func() (*http.Response, error) {
				resp := nationalitymodels.Response{
					Name:      "UnknownName",
					Countries: []nationalitymodels.Country{},
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedNationality: "",
			expectedProb:        0,
			expectedErrMsg:      "",
		},
		{
			name:      "successful response with single country",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				resp := nationalitymodels.Response{
					Name: "John",
					Countries: []nationalitymodels.Country{
						{CountryID: "US", Probability: 0.8},
					},
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedNationality: "US",
			expectedProb:        0.8,
			expectedErrMsg:      "",
		},
		{
			name:      "successful response with multiple countries",
			inputName: "Maria",
			mockResponse: func() (*http.Response, error) {
				resp := nationalitymodels.Response{
					Name: "Maria",
					Countries: []nationalitymodels.Country{
						{CountryID: "ES", Probability: 0.4},
						{CountryID: "PT", Probability: 0.2},
						{CountryID: "IT", Probability: 0.7},
						{CountryID: "MX", Probability: 0.3},
					},
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedNationality: "IT",
			expectedProb:        0.7,
			expectedErrMsg:      "",
		},
		{
			name:      "successful response with first country being most probable",
			inputName: "Akira",
			mockResponse: func() (*http.Response, error) {
				resp := nationalitymodels.Response{
					Name: "Akira",
					Countries: []nationalitymodels.Country{
						{CountryID: "JP", Probability: 0.9},
						{CountryID: "KR", Probability: 0.1},
						{CountryID: "CN", Probability: 0.05},
					},
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedNationality: "JP",
			expectedProb:        0.9,
			expectedErrMsg:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return tc.mockResponse()
				},
			}

			client := nationality.NewNationalityAPIClient(mockClient)
			require.NotNil(t, client)

			nationalityResult, prob, err := client.GetNationalityByName(context.Background(), tc.inputName)

			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedNationality, nationalityResult)
				assert.Equal(t, tc.expectedProb, prob)
			}
		})
	}
}

func TestAPIClient_GetNationalityByName_RequestParameters(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Validate the request parameters
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "https://api.nationalize.io", req.URL.Scheme+"://"+req.URL.Host)
			assert.Equal(t, "TestName", req.URL.Query().Get("name"))

			// Return successful response
			resp := nationalitymodels.Response{
				Name: "TestName",
				Countries: []nationalitymodels.Country{
					{CountryID: "US", Probability: 0.8},
				},
			}
			body, _ := json.Marshal(resp)
			return createMockResponse(200, body), nil
		},
	}

	client := nationality.NewNationalityAPIClient(mockClient)
	_, _, err := client.GetNationalityByName(context.Background(), "TestName")
	assert.NoError(t, err)
}
