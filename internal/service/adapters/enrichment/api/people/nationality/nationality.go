// Package nationality предоставляет реализацию интерфейса nationality.Service для получения национальности по имени.
package nationality

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	nationalitymodels "github.com/flexer2006/case-person-enrichment-go/internal/service/domain/models/api/nationality"
	nationalityservice "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people/nationality"
	"github.com/flexer2006/case-person-enrichment-go/pkg/logger"
	"go.uber.org/zap"
)

// Ошибки, которые могут возникнуть при работе с API.
var (
	ErrEmptyName      = errors.New("name cannot be empty")
	ErrNon200Response = errors.New("API returned non-200 status code")
)

// Проверка, что APIClient реализует интерфейс nationality.Service.
var _ nationalityservice.Service = (*APIClient)(nil)

// APIClient реализует интерфейс nationality.Service через вызов API nationalize.io.
type APIClient struct {
	baseURL    string
	httpClient HTTPClient
}

// HTTPClient - интерфейс для выполнения HTTP-запросов.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewNationalityAPIClient создает новый экземпляр APIClient.
func NewNationalityAPIClient(client HTTPClient) *APIClient {
	if client == nil {
		client = &http.Client{}
	}

	return &APIClient{
		baseURL:    "https://api.nationalize.io",
		httpClient: client,
	}
}

// GetNationalityByName возвращает предполагаемую национальность и вероятность для указанного имени.
func (c *APIClient) GetNationalityByName(ctx context.Context, name string) (string, float64, error) {
	logger.Debug(ctx, "getting nationality for name", zap.String("name", name))

	if name == "" {
		logger.Error(ctx, "empty name provided for nationality prediction")
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

	var nationalityResp nationalitymodels.Response
	if err := json.NewDecoder(resp.Body).Decode(&nationalityResp); err != nil {
		logger.Error(ctx, "failed to decode API response", zap.Error(err))
		return "", 0, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Если список стран пуст, возвращаем пустую строку и нулевую вероятность
	if len(nationalityResp.Countries) == 0 {
		logger.Debug(ctx, "no nationality data found for name", zap.String("name", name))
		return "", 0, nil
	}

	// Находим наиболее вероятную страну
	mostProbableCountry := nationalityResp.Countries[0]
	for _, country := range nationalityResp.Countries {
		if country.Probability > mostProbableCountry.Probability {
			mostProbableCountry = country
		}
	}

	logger.Debug(ctx, "received nationality from API",
		zap.String("name", name),
		zap.String("country_id", mostProbableCountry.CountryID),
		zap.Float64("probability", mostProbableCountry.Probability))

	return mostProbableCountry.CountryID, mostProbableCountry.Probability, nil
}
