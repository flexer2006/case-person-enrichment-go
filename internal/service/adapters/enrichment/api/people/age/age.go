// Package age предоставляет реализацию интерфейса age.AgeService для получения возраста по имени.
package age

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	agemodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/age"
	ageservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/age"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"go.uber.org/zap"
)

// Ошибки, которые могут возникнуть при работе с API.
var (
	ErrEmptyName      = errors.New("name cannot be empty")
	ErrNon200Response = errors.New("API returned non-200 status code")
)

// Проверка, что APIClient реализует интерфейс age.Service.
var _ ageservice.Service = (*APIClient)(nil)

// APIClient реализует интерфейс age.Service через вызов API agify.io.
type APIClient struct {
	baseURL    string
	httpClient HTTPClient
}

// HTTPClient - интерфейс для выполнения HTTP-запросов.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewAgeAPIClient создает новый экземпляр APIClient.
func NewAgeAPIClient(client HTTPClient) *APIClient {
	if client == nil {
		client = &http.Client{}
	}

	return &APIClient{
		baseURL:    "https://api.agify.io",
		httpClient: client,
	}
}

// GetAgeByName возвращает предполагаемый возраст и вероятность для указанного имени.
func (c *APIClient) GetAgeByName(ctx context.Context, name string) (int, float64, error) {
	logger.Debug(ctx, "getting age for name", zap.String("name", name))

	if name == "" {
		logger.Error(ctx, "empty name provided for age prediction")
		return 0, 0, ErrEmptyName
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		logger.Error(ctx, "failed to parse base URL", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	q.Add("name", name)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		logger.Error(ctx, "failed to create request", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to execute request", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn(ctx, "failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "API returned non-200 status code",
			zap.Int("status_code", resp.StatusCode))
		return 0, 0, fmt.Errorf("%w: status %d", ErrNon200Response, resp.StatusCode)
	}

	var ageResp agemodels.Response
	if err := json.NewDecoder(resp.Body).Decode(&ageResp); err != nil {
		logger.Error(ctx, "failed to decode API response", zap.Error(err))
		return 0, 0, fmt.Errorf("failed to decode API response: %w", err)
	}

	logger.Debug(ctx, "received age from API",
		zap.String("name", name),
		zap.Int("age", ageResp.Age),
		zap.Int("count", ageResp.Count))

	probability := min(float64(ageResp.Count)/1000.0, 1.0)

	return ageResp.Age, probability, nil
}
