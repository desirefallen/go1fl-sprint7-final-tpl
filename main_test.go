package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	city := "moscow"
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList[city]))},
	}

	for _, v := range requests {
		reqURL := fmt.Sprintf("/cafe?count=%d&city=%s", v.count, city)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", reqURL, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, fmt.Sprintf("count: %d", v.count))

		trimBody := strings.TrimSpace(response.Body.String())
		var cafesCount int
		if trimBody == "" {
			cafesCount = 0
		} else {
			cafesCount = len(strings.Split(trimBody, ","))
		}

		assert.Equal(t, v.want, cafesCount, fmt.Sprintf("count: %d", v.count))
	}
}

func TestCafeSearch(t *testing.T) {
	city := "moscow"

	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		reqURL := fmt.Sprintf("/cafe?city=%s&search=%s", city, v.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", reqURL, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, fmt.Sprintf("search: %s", v.search))

		trimBody := strings.TrimSpace(response.Body.String())

		var cafesCount int
		if trimBody == "" {
			cafesCount = 0
			assert.Equal(t, v.wantCount, 0, fmt.Sprintf("search: %s", v.search))
			continue
		}
		cafesCount = len(strings.Split(trimBody, ","))
		for cafe := range strings.SplitSeq(trimBody, ",") {
			assert.Contains(t, strings.ToLower(cafe), strings.ToLower(v.search), fmt.Sprintf("cafe: %s", cafe))
			assert.Equal(t, v.wantCount, cafesCount, fmt.Sprintf("search: %s, cafesCount: %d", v.search, cafesCount))
		}
	}
}
