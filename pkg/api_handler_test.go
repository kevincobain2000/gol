package pkg

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIHandler(t *testing.T) {
	handler := NewAPIHandler()
	assert.NotNil(t, handler)
}

func TestAPIHandler_Get(t *testing.T) {
	e := echo.New()

	// Set up global variables for testing
	GlobalFilePaths = []FileInfo{
		{
			FilePath:   "test.log",
			LinesCount: 4,
			FileSize:   0,
			Type:       "file",
		},
	}
	GlobalTmpFilePath = "temp.log"

	// Create a temporary log file for testing
	// nolint:goconst
	content := `INFO Starting service
ERROR An error occurred
INFO Service running
ERROR Another error occurred`
	err := os.WriteFile(GlobalFilePaths[0].FilePath, []byte(content), 0600)
	assert.NoError(t, err)
	defer os.Remove(GlobalFilePaths[0].FilePath)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api?query=ERROR&page=1&per_page=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create the API handler and execute the Get method
	handler := NewAPIHandler()
	if assert.NoError(t, handler.Get(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expected := `{
			"result": {
				"file_path": "test.log",
				"match_pattern": "ERROR",
				"total": 2,
				"lines": [
					{
						"line_number": 2,
						"content": "ERROR An error occurred",
						"level": "error"
					},
					{
						"line_number": 4,
						"content": "ERROR Another error occurred",
						"level": "error"
					}
				]
			},
			"file_paths": [
				{
					"file_path": "test.log",
					"lines_count": 4,
					"file_size": 0,
					"type": "file"
				}
			]
		}`
		assert.JSONEq(t, expected, rec.Body.String())
	}

	// TODO fix test (instead of code)
	// req = httptest.NewRequest(http.MethodGet, "/api?file_path=wrong", nil)
	// rec = httptest.NewRecorder()
	// c = e.NewContext(req, rec)
	// assert.Error(t, handler.Get(c))
	// assert.Equal(t, http.StatusNotFound, rec.Code)
}
