package views

import (
	"fmt"
	"net/url"
)

var base *url.URL

func u(str string, args ...any) string {
	res := fmt.Sprintf(str, args...)
	if base == nil {
		return res
	}
	return base.JoinPath(res).Path
}

func SetBase(site *url.URL) {
	base = site
}
