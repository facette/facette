package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"facette.io/httproute"
	"facette.io/httputil"
)

func httpBuildMessage(err error) map[string]string {
	return map[string]string{
		"message": fmt.Sprintf("%s", err),
	}
}

func httpGetBoolParam(r *http.Request, name string) bool {
	vs := httproute.QueryParam(r, name)
	return vs == "1" || vs == "true"
}

func httpGetIntParam(r *http.Request, name string) (int, error) {
	vs := httproute.QueryParam(r, name)

	v, err := strconv.Atoi(vs)
	if vs != "" && err != nil {
		return 0, os.ErrInvalid
	}

	return v, nil
}

func httpGetListParam(r *http.Request, name string, fallback []string) []string {
	list := strings.Split(r.URL.Query().Get(name), ",")
	if len(list) == 1 && list[0] == "" {
		return fallback
	}

	return list
}

func httpHandleNotFound(rw http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(rw, httpBuildMessage(ErrUnknownEndpoint), http.StatusNotFound)
}
