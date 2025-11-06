package server

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/riyadennis/identity-server/business/store"
)

func TestNewServerPortValidation(t *testing.T) {
	sc := []struct {
		name          string
		port          string
		expectedError error
	}{
		{
			name:          "emprty port",
			expectedError: errEmptyPort,
		},
		{
			name:          "string port",
			port:          "INVALID",
			expectedError: errPortNotAValidNumber,
		},
		{
			name:          "reserved port",
			port:          "1023",
			expectedError: errPortReserved,
		},
		{
			name:          "beyond range",
			port:          "65536",
			expectedError: errPortBeyondRange,
		},
		{
			name: "valid port",
			port: "8080",
		},
	}
	for _, tc := range sc {
		t.Run(tc.name, func(t *testing.T) {
			se := NewServer(tc.port)
			// using select to listen to channel
			select {
			case err := <-se.ServerError:
				assert.Equal(t, tc.expectedError, err)
			default:
				t.Log("reached default for test " + tc.name)
				close(se.ServerError)
			}
		})
	}
}

// mock http.Server to replace ListenAndServe and Shutdown
// since http.Server is a struct, we'll simulate the error via goroutine and channels
func TestServer_Run_Error(t *testing.T) {
	s := NewServer("8081")
	//var buf bytes.Buffer
	//logger := log.New(&buf, "", 0)
	logger := logrus.New()
	// Simulate error from ListenAndServe
	go func() {
		time.Sleep(10 * time.Millisecond)
		s.ServerError <- errors.New("listen error")
	}()
	err := s.Run(&sql.DB{}, &store.TokenConfig{}, logger)
	assert.EqualError(t, err, "listen error")
}

func TestServer_Run_Shutdown(t *testing.T) {
	s := NewServer("8082")
	logger := logrus.New()
	// Simulate shutdown signal after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		s.ShutDown <- os.Kill
	}()
	err := s.Run(&sql.DB{}, &store.TokenConfig{}, logger)
	assert.NoError(t, err)
}
