package v0_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jgfranco17/echoris/api/events"
	v0 "github.com/jgfranco17/echoris/api/router/v0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockLogClient_ForwardLogs(t *testing.T) {
	t.Run("successful forward", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		batch := []events.Entry{
			{Service: "test", Level: "info", Message: "test1"},
			{Service: "test", Level: "debug", Message: "test2"},
		}

		mockClient.On("ForwardLogs", mock.Anything, batch).Return(nil)

		err := mockClient.ForwardLogs(context.Background(), batch)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("forward with error", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		batch := []events.Entry{
			{Service: "test", Level: "error", Message: "test"},
		}
		expectedError := errors.New("failed to forward logs")

		mockClient.On("ForwardLogs", mock.Anything, batch).Return(expectedError)

		err := mockClient.ForwardLogs(context.Background(), batch)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("forward with any batch", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		mockClient.On("ForwardLogs", mock.Anything, mock.AnythingOfType("[]events.Entry")).Return(nil)

		batch := []events.Entry{{Service: "test", Level: "info", Message: "test"}}
		err := mockClient.ForwardLogs(context.Background(), batch)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestMockLogClient_FetchLogs(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		expectedLogs := []events.Entry{
			{Service: "test", Level: "info", Message: "log1"},
			{Service: "test", Level: "info", Message: "log2"},
		}

		mockClient.On("FetchLogs", mock.Anything, "test", "info").Return(expectedLogs, nil)

		logs, err := mockClient.FetchLogs(context.Background(), "test", "info")
		assert.NoError(t, err)
		assert.Equal(t, expectedLogs, logs)
		assert.Len(t, logs, 2)
		mockClient.AssertExpectations(t)
	})

	t.Run("fetch with error", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		expectedError := errors.New("failed to fetch logs")
		mockClient.On("FetchLogs", mock.Anything, "test", "error").Return(nil, expectedError)

		logs, err := mockClient.FetchLogs(context.Background(), "test", "error")
		assert.Error(t, err)
		assert.Nil(t, logs)
		assert.Equal(t, expectedError, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("fetch with empty results", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		emptyLogs := []events.Entry{}
		mockClient.On("FetchLogs", mock.Anything, "test", "debug").Return(emptyLogs, nil)

		logs, err := mockClient.FetchLogs(context.Background(), "test", "debug")
		assert.NoError(t, err)
		assert.Empty(t, logs)
		mockClient.AssertExpectations(t)
	})
}

func TestMockLogClient_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		mockClient.On("Close").Return(nil)

		err := mockClient.Close()
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("close with error", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		expectedError := errors.New("failed to close connection")
		mockClient.On("Close").Return(expectedError)

		err := mockClient.Close()
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockClient.AssertExpectations(t)
	})
}

func TestMockLogClient_AssertCalled(t *testing.T) {
	t.Run("assert number of calls", func(t *testing.T) {
		mockClient := new(v0.MockLogClient)

		batch := []events.Entry{{Service: "test", Level: "info", Message: "test"}}
		mockClient.On("ForwardLogs", mock.Anything, batch).Return(nil).Times(3)

		// Call it 3 times
		for i := 0; i < 3; i++ {
			err := mockClient.ForwardLogs(context.Background(), batch)
			assert.NoError(t, err)
		}

		mockClient.AssertExpectations(t)
		mockClient.AssertNumberOfCalls(t, "ForwardLogs", 3)
	})
}
