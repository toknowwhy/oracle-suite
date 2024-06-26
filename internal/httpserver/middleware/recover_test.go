package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover_WithCallback(t *testing.T) {
	var panicVal interface{}
	defer func() {
		assert.Nil(t, recover())
		assert.Equal(t, panicVal, "panic")
	}()
	h := (&Recover{Recover: func(err interface{}) {
		panicVal = err
	}}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic("panic")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
}

func TestRecover_WithoutCallback(t *testing.T) {
	defer func() {
		assert.Nil(t, recover())
	}()
	h := (&Recover{}).Handle(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic("panic")
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
}
