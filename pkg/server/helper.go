package server

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/facette/facette/pkg/utils"
)

func (server *Server) applyCollectionListResponse(writer http.ResponseWriter, request *http.Request,
	response CollectionListResponse, offset, limit int) {

	writer.Header().Add("X-Total-Records", strconv.Itoa(len(response)))

	sort.Sort(response)

	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
	}
}

func (server *Server) applyItemListResponse(writer http.ResponseWriter, request *http.Request,
	response ItemListResponse, offset, limit int) {

	writer.Header().Add("X-Total-Records", strconv.Itoa(len(response)))

	sort.Sort(response)

	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
	}
}

func (server *Server) applyStringListResponse(writer http.ResponseWriter, request *http.Request, response []string,
	offset, limit int) {

	writer.Header().Add("X-Total-Records", strconv.Itoa(len(response)))

	sort.Strings(response)

	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
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
		if *offset, err = strconv.Atoi(request.FormValue("offset")); err != nil {
			return &serverResponse{mesgFormOffsetInvalid}, http.StatusBadRequest
		}
	}

	if limit != nil && request.FormValue("limit") != "" {
		if *limit, err = strconv.Atoi(request.FormValue("limit")); err != nil {
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
	} else if utils.RequestGetContentType(request) != "application/json" {
		return &serverResponse{mesgUnsupportedMediaType}, http.StatusUnsupportedMediaType
	} else if !server.handleAuth(writer, request) {
		return &serverResponse{mesgAuthenticationRequired}, http.StatusUnauthorized
	}

	fmt.Println(server.handleAuth(writer, request))

	return nil, http.StatusOK
}
