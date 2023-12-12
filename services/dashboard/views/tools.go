package views

import "net/url"

var base *url.URL

func u(str string) string {
	return base.JoinPath(str).Path
}

func SetBase(site *url.URL) {
	base = site
}
