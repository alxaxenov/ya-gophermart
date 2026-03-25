package client

import (
	"net/url"
	"path"
)

func UrlJoin(basePath string, paths ...string) (*url.URL, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}

	p2 := append([]string{u.Path}, paths...)
	u.Path = path.Join(p2...)

	return u, nil
}
