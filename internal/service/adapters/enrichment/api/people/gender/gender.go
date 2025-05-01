// Package gender предоставляет реализацию интерфейса gender.Service для получения пола по имени.
package gender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	gendermodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/gender"
	genderservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/gender"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"go.uber.org/zap"
)

// Ошибки, которые могут возникнуть при работе с API.
var (
	ErrEmptyName      = errors.New("name cannot be empty")
	ErrNon200Response = errors.New("API returned non-200 status code")
)

// Проверка, что APIClient реализует интерфейс gender.Service.
var _ genderservice.Service = (*APIClient)(nil)

// APIClient реализует интерфейс gender.Service через вызов API genderize.io.
type APIClient struct {
	baseURL    string
	httpClient HTTPClient
}

// HTTPClient - интерфейс для выполнения HTTP-запросов.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewGenderAPIClient создает новый экземпляр APIClient.
func NewGenderAPIClient(client HTTPClient) *APIClient {
	if client == nil {
		client = &http.Client{}
	}

	return &APIClient{
		baseURL:    "https://api.genderize.io",
		httpClient: client,
	}
}

// GetGenderByName возвращает предполагаемый пол и вероятность для указанного имени.
func (c *APIClient) GetGenderByName(ctx context.Context, name string) (string, float64, error) {
	logger.Debug(ctx, "getting gender for name", zap.String("name", name))

	if name == "" {
		logger.Error(ctx, "empty name provided for gender prediction")
		return "", 0, ErrEmptyName
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		logger.Error(ctx, "failed to parse base URL", zap.Error(err))
		return "", 0, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	q.Add("name", name)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		logger.Error(ctx, "failed to create request", zap.Error(err))
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to execute request", zap.Error(err))
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn(ctx, "failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "API returned non-200 status code",
			zap.Int("status_code", resp.StatusCode))
		return "", 0, fmt.Errorf("%w: status %d", ErrNon200Response, resp.StatusCode)
	}

	var genderResp gendermodels.Response
	if err := json.NewDecoder(resp.Body).Decode(&genderResp); err != nil {
		logger.Error(ctx, "failed to decode API response", zap.Error(err))
		return "", 0, fmt.Errorf("failed to decode API response: %w", err)
	}

	logger.Debug(ctx, "received gender from API",
		zap.String("name", name),
		zap.String("gender", genderResp.Gender),
		zap.Float64("probability", genderResp.Probability))

	return genderResp.Gender, genderResp.Probability, nil
}
