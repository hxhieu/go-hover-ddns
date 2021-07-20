package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func CreateHoverClient(cookies []*http.Cookie) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Jar: jar,
	}

	if cookies != nil {
		url, _ := url.Parse(HoverUrl)
		client.Jar.SetCookies(url, cookies)
	}

	return &client, nil
}
