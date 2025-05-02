package gender_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/enrichment/api/people/gender"
	gendermodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/gender"
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

func TestNewGenderAPIClient(t *testing.T) {
	t.Run("with nil client", func(t *testing.T) {
		client := gender.NewGenderAPIClient(nil)
		require.NotNil(t, client, "client should not be nil")
	})

	t.Run("with custom client", func(t *testing.T) {
		mockClient := &MockHTTPClient{}
		client := gender.NewGenderAPIClient(mockClient)
		require.NotNil(t, client, "client should not be nil")
	})
}

func TestAPIClient_GetGenderByName(t *testing.T) {
	testCases := []struct {
		name           string
		inputName      string
		mockResponse   func() (*http.Response, error)
		expectedGender string
		expectedProb   float64
		expectedErrMsg string
	}{
		{
			name:           "empty name",
			inputName:      "",
			mockResponse:   func() (*http.Response, error) { return nil, nil },
			expectedGender: "",
			expectedProb:   0,
			expectedErrMsg: "name cannot be empty",
		},
		{
			name:      "http client error",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return nil, errors.New("network error")
			},
			expectedGender: "",
			expectedProb:   0,
			expectedErrMsg: "failed to execute request: network error",
		},
		{
			name:      "non 200 status",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(404, []byte(`{"error":"not found"}`)), nil
			},
			expectedGender: "",
			expectedProb:   0,
			expectedErrMsg: "API returned non-200 status code: status 404",
		},
		{
			name:      "invalid json response",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				return createMockResponse(200, []byte(`invalid json`)), nil
			},
			expectedGender: "",
			expectedProb:   0,
			expectedErrMsg: "failed to decode API response",
		},
		{
			name:      "successful response - male gender",
			inputName: "John",
			mockResponse: func() (*http.Response, error) {
				resp := gendermodels.Response{
					Name:        "John",
					Gender:      "male",
					Probability: 0.95,
					Count:       1000,
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedGender: "male",
			expectedProb:   0.95,
			expectedErrMsg: "",
		},
		{
			name:      "successful response - female gender",
			inputName: "Maria",
			mockResponse: func() (*http.Response, error) {
				resp := gendermodels.Response{
					Name:        "Maria",
					Gender:      "female",
					Probability: 0.98,
					Count:       2000,
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedGender: "female",
			expectedProb:   0.98,
			expectedErrMsg: "",
		},
		{
			name:      "successful response - null gender",
			inputName: "Riley",
			mockResponse: func() (*http.Response, error) {
				resp := gendermodels.Response{
					Name:        "Riley",
					Gender:      "",
					Probability: 0,
					Count:       500,
				}
				body, _ := json.Marshal(resp)
				return createMockResponse(200, body), nil
			},
			expectedGender: "",
			expectedProb:   0,
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

			client := gender.NewGenderAPIClient(mockClient)
			require.NotNil(t, client)

			genderResult, prob, err := client.GetGenderByName(context.Background(), tc.inputName)

			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGender, genderResult)
				assert.Equal(t, tc.expectedProb, prob)
			}
		})
	}
}

func TestAPIClient_GetGenderByName_RequestParameters(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "https://api.genderize.io", req.URL.Scheme+"://"+req.URL.Host)
			assert.Equal(t, "TestName", req.URL.Query().Get("name"))

			resp := gendermodels.Response{
				Name:        "TestName",
				Gender:      "male",
				Probability: 0.85,
				Count:       100,
			}
			body, _ := json.Marshal(resp)
			return createMockResponse(200, body), nil
		},
	}

	client := gender.NewGenderAPIClient(mockClient)
	_, _, err := client.GetGenderByName(context.Background(), "TestName")
	assert.NoError(t, err)
}
