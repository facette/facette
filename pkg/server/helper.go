package server

import (
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/utils"
)

func (server *Server) applyResponseLimit(writer http.ResponseWriter, request *http.Request, response *listResponse) {
	writer.Header().Add("X-Total-Records", strconv.Itoa(response.list.Len()))

	if response.list.Len() == 0 {
		return
	}

	sort.Sort(response.list)

	if response.limit != 0 && response.list.Len() > response.offset+response.limit {
		response.list = response.list.slice(response.offset, response.offset+response.limit).(sortableListResponse)
	} else if response.offset != 0 {
		response.list = response.list.slice(response.offset, response.list.Len()).(sortableListResponse)
	}
}

func (server *Server) parseError(writer http.ResponseWriter, request *http.Request, err error) (*serverResponse, int) {
	if err == os.ErrInvalid {
		return &serverResponse{mesgResourceInvalid}, http.StatusBadRequest
	} else if os.IsExist(err) {
		return &serverResponse{mesgResourceConflict}, http.StatusConflict
	} else if os.IsNotExist(err) {
		return &serverResponse{mesgResourceNotFound}, http.StatusNotFound
	} else if err != nil {
		return &serverResponse{mesgUnhandledError}, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

func (server *Server) parseListRequest(writer http.ResponseWriter, request *http.Request,
	offset, limit *int) (*serverResponse, int) {

	var err error

	if request.Method != "GET" && request.Method != "HEAD" {
		return &serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed
	}

	if offset != nil && request.FormValue("offset") != "" {
		if *offset, err = strconv.Atoi(request.FormValue("offset")); err != nil || *offset < 0 {
			return &serverResponse{mesgFormOffsetInvalid}, http.StatusBadRequest
		}
	}

	if limit != nil && request.FormValue("limit") != "" {
		if *limit, err = strconv.Atoi(request.FormValue("limit")); err != nil || *limit < 0 {
			return &serverResponse{mesgFormLimitInvalid}, http.StatusBadRequest
		}
	}

	return nil, http.StatusOK
}

func (server *Server) parseShowRequest(writer http.ResponseWriter, request *http.Request) (*serverResponse, int) {
	return server.parseListRequest(writer, request, nil, nil)
}

func (server *Server) parseStoreRequest(writer http.ResponseWriter, request *http.Request,
	id string) (*serverResponse, int) {

	if request.Method == "POST" && id != "" || request.Method == "PUT" && id == "" {
		return &serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed
	} else if utils.HTTPGetContentType(request) != "application/json" {
		return &serverResponse{mesgUnsupportedMediaType}, http.StatusUnsupportedMediaType
	}

	return nil, http.StatusOK
}

func routeMatch(path, prefix string) bool {
	// Ensure path ends with a slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func routeTrimPrefix(path, prefix string) string {
	return strings.Trim(strings.TrimPrefix(path, prefix), "/")
}

func setHTTPCacheHeaders(writer http.ResponseWriter) {
	date := time.Now().UTC().Format(http.TimeFormat)

	writer.Header().Set("Cache-Control", "private, max-age=0")
	writer.Header().Set("Date", date)
	writer.Header().Set("Expires", date)
}
