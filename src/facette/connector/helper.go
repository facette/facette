package connector

import "strings"

func normalizeURL(url *string) {
	v := *url
	v = strings.TrimRight(v, "/")

	if !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
		v = "http://" + v
	}

	*url = v
}
