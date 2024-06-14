package pkg

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mcuadros/go-defaults"
)

type APIHandler struct {
	API *API
}
type FileInfo struct {
	FilePath   string `json:"file_path"`
	LinesCount int    `json:"lines_count"`
	FileSize   int64  `json:"file_size"`
	Type       string `json:"type"`
	Host       string `json:"host"`
}

var GlobalFilePaths []FileInfo
var GlobalPipeTmpFilePath string
var GlobalPathSSHConfig []SSHPathConfig

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		API: NewAPI(),
	}
}

type APIRequest struct {
	Query    string `json:"query" query:"query"`
	FilePath string `json:"file_path" query:"file_path"`
	Host     string `json:"host" query:"host"`
	Page     int    `json:"page" query:"page" default:"1" validate:"required,gte=1" message:"page >=1 is required"`
	PerPage  int    `json:"per_page" query:"per_page" default:"15" validate:"required" message:"per_page is required"`
	Reverse  bool   `json:"reverse" query:"reverse" default:"false"`
}

type APIResponse struct {
	Result    ScanResult `json:"result"`
	FilePaths []FileInfo `json:"file_paths"`
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

	if req.FilePath == "" {
		first := GlobalFilePaths[0]
		req.FilePath = first.FilePath
		req.Host = first.Host
	}

	if !FilePathInGlobalFilePaths(req.FilePath) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}
	var watcher *Watcher

	if req.Host != "" {
		sshConfig := h.API.FindSSHConfig(req.Host)
		if sshConfig == nil {
			return echo.NewHTTPError(http.StatusNotFound, "ssh config not found")
		}
		watcher, err = NewWatcher(req.FilePath, req.Query, true, sshConfig.Host, sshConfig.Port, sshConfig.User, sshConfig.Password, sshConfig.PrivateKeyPath)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	if req.Host == "" {
		watcher, err = NewWatcher(req.FilePath, req.Query, false, "", "", "", "", "")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
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
func FilePathInGlobalFilePaths(filePath string) bool {
	for _, fileInfo := range GlobalFilePaths {
		if fileInfo.FilePath == filePath {
			return true
		}
	}
	return false
}
