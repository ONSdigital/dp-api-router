// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/ONSdigital/dp-api-router/middleware"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

// Ensure, that RouterMock does implement middleware.Router.
// If this is not the case, regenerate this file with moq.
var _ middleware.Router = &RouterMock{}

// RouterMock is a mock implementation of middleware.Router.
//
//	func TestSomethingThatUsesRouter(t *testing.T) {
//
//		// make and configure a mocked middleware.Router
//		mockedRouter := &RouterMock{
//			MatchFunc: func(req *http.Request, match *mux.RouteMatch) bool {
//				panic("mock out the Match method")
//			},
//		}
//
//		// use mockedRouter in code that requires middleware.Router
//		// and then make assertions.
//
//	}
type RouterMock struct {
	// MatchFunc mocks the Match method.
	MatchFunc func(req *http.Request, match *mux.RouteMatch) bool

	// calls tracks calls to the methods.
	calls struct {
		// Match holds details about calls to the Match method.
		Match []struct {
			// Req is the req argument value.
			Req *http.Request
			// Match is the match argument value.
			Match *mux.RouteMatch
		}
	}
	lockMatch sync.RWMutex
}

// Match calls MatchFunc.
func (mock *RouterMock) Match(req *http.Request, match *mux.RouteMatch) bool {
	if mock.MatchFunc == nil {
		panic("RouterMock.MatchFunc: method is nil but Router.Match was just called")
	}
	callInfo := struct {
		Req   *http.Request
		Match *mux.RouteMatch
	}{
		Req:   req,
		Match: match,
	}
	mock.lockMatch.Lock()
	mock.calls.Match = append(mock.calls.Match, callInfo)
	mock.lockMatch.Unlock()
	return mock.MatchFunc(req, match)
}

// MatchCalls gets all the calls that were made to Match.
// Check the length with:
//
//	len(mockedRouter.MatchCalls())
func (mock *RouterMock) MatchCalls() []struct {
	Req   *http.Request
	Match *mux.RouteMatch
} {
	var calls []struct {
		Req   *http.Request
		Match *mux.RouteMatch
	}
	mock.lockMatch.RLock()
	calls = mock.calls.Match
	mock.lockMatch.RUnlock()
	return calls
}
