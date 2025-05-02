package server_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flexer2006/case-person-enrichment-go/internal/service/adapters/server"
	apipeople "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/api/people"
	repopeople "github.com/flexer2006/case-person-enrichment-go/internal/service/ports/repo/people"
	serverconfig "github.com/flexer2006/case-person-enrichment-go/internal/service/setup/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAPI struct {
	mock.Mock
}

func (m *MockAPI) People() apipeople.Services {
	args := m.Called()
	return args.Get(0).(apipeople.Services)
}

type MockPeopleServices struct {
	mock.Mock
}

func (m *MockPeopleServices) Person() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockPeopleServices) Age() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockPeopleServices) Gender() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockPeopleServices) Nationality() interface{} {
	args := m.Called()
	return args.Get(0)
}

type MockRepositories struct {
	mock.Mock
}

func (m *MockRepositories) People() repopeople.Repositories {
	args := m.Called()
	return args.Get(0).(repopeople.Repositories)
}

type MockPeopleRepositories struct {
	mock.Mock
}

func (m *MockPeopleRepositories) Person() interface{} {
	args := m.Called()
	return args.Get(0)
}

func TestNew(t *testing.T) {
	// Arrange
	config := serverconfig.Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	assert.NotNil(t, s, "Server instance should not be nil")
	assert.Equal(t, config, s.GetConfig(), "Server should store the provided config")
}

func TestNewWithEmptyConfig(t *testing.T) {
	config := serverconfig.Config{}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	assert.NotNil(t, s, "Server instance should not be nil")
	assert.Equal(t, config, s.GetConfig(), "Server should store the provided config")
}

type MockApp struct {
	mock.Mock
}

func (m *MockApp) Listen(addr string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockApp) ShutdownWithContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestServerStart(t *testing.T) {
	config := serverconfig.Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		err := s.Start(ctx)

		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			assert.NoError(t, err, "Start should only return context canceled errors")
		}
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for server to stop after context cancellation")
	}
}

func TestServerStartWithEmptyConfig(t *testing.T) {
	config := serverconfig.Config{
		Host: "0.0.0.0",
		Port: 3000,
	}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	// Act
	go func() {
		err := s.Start(ctx)

		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			assert.NoError(t, err, "Start should only return context canceled errors")
		}
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for server to stop after context cancellation")
	}
}

func TestStop(t *testing.T) {
	config := serverconfig.Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	ctx := context.Background()

	err := s.Stop(ctx)

	assert.NoError(t, err, "Stop should not return an error when successful")
}

func TestStopWithCanceledContext(t *testing.T) {
	config := serverconfig.Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mockPeopleServices := new(MockPeopleServices)
	mockPeopleServices.On("Person").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Age").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Gender").Return(mock.Anything).Maybe()
	mockPeopleServices.On("Nationality").Return(mock.Anything).Maybe()

	mockAPI := new(MockAPI)
	mockAPI.On("People").Return(mockPeopleServices).Maybe()

	mockPeopleRepositories := new(MockPeopleRepositories)
	mockPeopleRepositories.On("Person").Return(mock.Anything).Maybe()

	mockRepositories := new(MockRepositories)
	mockRepositories.On("People").Return(mockPeopleRepositories).Maybe()

	s := server.New(config, mockAPI, mockRepositories)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Stop(ctx)

	if err != nil {
		assert.Contains(t, err.Error(), "context canceled", "Error should mention context cancellation")
	} else {
		t.Log("No error returned with canceled context, which is acceptable in some cases")
	}
}
