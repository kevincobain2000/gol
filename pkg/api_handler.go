package pkg

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mcuadros/go-defaults"
)

type APIHandler struct {
}

var GlobalFilePaths []string
var GlobalTempFilePath string

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

type APIRequest struct {
	Query    string `json:"query" query:"query" message:"query is required"`
	FilePath string `json:"file_path" query:"file_path" message:"file_path is required"`
	Page     int    `json:"page" query:"page" default:"1" validate:"required,gte=1" message:"page >=1 is required"`
	PerPage  int    `json:"per_page" query:"per_page" default:"15" validate:"required" message:"per_page is required"`
	Reverse  bool   `json:"reverse" query:"reverse" default:"false"`
}

type APIResponse struct {
	Result    ScanResult `json:"result"`
	FilePaths []string   `json:"file_paths"`
}

func (h *APIHandler) Get(c echo.Context) error {
	req := new(APIRequest)
	if err := BindRequest(c, req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err)
	}
	defaults.SetDefaults(req)
	msgs, err := ValidateRequest(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, msgs)
	}
	if len(GlobalFilePaths) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "filepath not found")
	}

	if !StringInSlice(GlobalTempFilePath, GlobalFilePaths) {
		GlobalFilePaths = append([]string{GlobalTempFilePath}, GlobalFilePaths...)
	}

	if req.FilePath == "" {
		req.FilePath = GlobalFilePaths[0]
	}

	if !StringInSlice(req.FilePath, GlobalFilePaths) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	watcher, err := NewWatcher(req.FilePath, req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	result, err := watcher.Scan(req.Page, req.PerPage, req.Reverse)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, APIResponse{
		Result:    *result,
		FilePaths: GlobalFilePaths,
	})
}
