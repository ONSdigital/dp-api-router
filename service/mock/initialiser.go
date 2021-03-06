// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/dp-api-router/service"
	"github.com/ONSdigital/dp-kafka/v2"
	"sync"
)

var (
	lockInitialiserMockDoGetHealthCheck   sync.RWMutex
	lockInitialiserMockDoGetKafkaProducer sync.RWMutex
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
//     func TestSomethingThatUsesInitialiser(t *testing.T) {
//
//         // make and configure a mocked service.Initialiser
//         mockedInitialiser := &InitialiserMock{
//             DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 	               panic("mock out the DoGetHealthCheck method")
//             },
//             DoGetKafkaProducerFunc: func(ctx context.Context, brokers []string, kafkaVersion string, topic string, maxBytes int) (kafka.IProducer, error) {
// 	               panic("mock out the DoGetKafkaProducer method")
//             },
//         }
//
//         // use mockedInitialiser in code that requires service.Initialiser
//         // and then make assertions.
//
//     }
type InitialiserMock struct {
	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error)

	// DoGetKafkaProducerFunc mocks the DoGetKafkaProducer method.
	DoGetKafkaProducerFunc func(ctx context.Context, brokers []string, kafkaVersion string, topic string, maxBytes int) (kafka.IProducer, error)

	// calls tracks calls to the methods.
	calls struct {
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
		// DoGetKafkaProducer holds details about calls to the DoGetKafkaProducer method.
		DoGetKafkaProducer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Brokers is the brokers argument value.
			Brokers []string
			// KafkaVersion is the kafkaVersion argument value.
			KafkaVersion string
			// Topic is the topic argument value.
			Topic string
			// MaxBytes is the maxBytes argument value.
			MaxBytes int
		}
	}
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	lockInitialiserMockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	lockInitialiserMockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	lockInitialiserMockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	lockInitialiserMockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetKafkaProducer calls DoGetKafkaProducerFunc.
func (mock *InitialiserMock) DoGetKafkaProducer(ctx context.Context, brokers []string, kafkaVersion string, topic string, maxBytes int) (kafka.IProducer, error) {
	if mock.DoGetKafkaProducerFunc == nil {
		panic("InitialiserMock.DoGetKafkaProducerFunc: method is nil but Initialiser.DoGetKafkaProducer was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		Brokers      []string
		KafkaVersion string
		Topic        string
		MaxBytes     int
	}{
		Ctx:          ctx,
		Brokers:      brokers,
		KafkaVersion: kafkaVersion,
		Topic:        topic,
		MaxBytes:     maxBytes,
	}
	lockInitialiserMockDoGetKafkaProducer.Lock()
	mock.calls.DoGetKafkaProducer = append(mock.calls.DoGetKafkaProducer, callInfo)
	lockInitialiserMockDoGetKafkaProducer.Unlock()
	return mock.DoGetKafkaProducerFunc(ctx, brokers, kafkaVersion, topic, maxBytes)
}

// DoGetKafkaProducerCalls gets all the calls that were made to DoGetKafkaProducer.
// Check the length with:
//     len(mockedInitialiser.DoGetKafkaProducerCalls())
func (mock *InitialiserMock) DoGetKafkaProducerCalls() []struct {
	Ctx          context.Context
	Brokers      []string
	KafkaVersion string
	Topic        string
	MaxBytes     int
} {
	var calls []struct {
		Ctx          context.Context
		Brokers      []string
		KafkaVersion string
		Topic        string
		MaxBytes     int
	}
	lockInitialiserMockDoGetKafkaProducer.RLock()
	calls = mock.calls.DoGetKafkaProducer
	lockInitialiserMockDoGetKafkaProducer.RUnlock()
	return calls
}
