package v1

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vbatoufflet/httproute"
)

func parseBoolParam(r *http.Request, name string) bool {
	vs := httproute.QueryParam(r, name)
	return vs == "1" || vs == "true"
}

func parseIntParam(r *http.Request, name string) (int, error) {
	vs := httproute.QueryParam(r, name)

	v, err := strconv.Atoi(vs)
	if vs != "" && err != nil {
		return 0, os.ErrInvalid
	}

	return v, nil
}

func parseListParam(r *http.Request, name string, fallback []string) []string {
	list := strings.Split(r.URL.Query().Get(name), ",")
	if len(list) == 1 && list[0] == "" {
		return fallback
	}

	return list
}
