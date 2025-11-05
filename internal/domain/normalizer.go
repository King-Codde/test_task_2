package domain

import (
	"net/url"
	"sort"
	"strings"
)

type StreamNormalizer struct{}

func NewStreamNormalizer() *StreamNormalizer {
	return &StreamNormalizer{}
}

func (sn *StreamNormalizer) Normalize(uri string) (string, error) {
	uri = strings.TrimSpace(uri)

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		password, hasPassword := parsedURL.User.Password()
		if hasPassword {
			parsedURL.User = url.UserPassword(strings.ToLower(username), strings.ToLower(password))
		} else {
			parsedURL.User = url.User(strings.ToLower(username))
		}
	}

	parsedURL.Host = strings.ToLower(parsedURL.Host)

	parsedURL.Path = strings.ToLower(strings.TrimSuffix(parsedURL.Path, "/"))

	if parsedURL.RawQuery != "" {
		query := parsedURL.Query()
		keys := make([]string, 0, len(query))
		for k := range query {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		normalizedQuery := url.Values{}
		for _, k := range keys {
			normalizedQuery[k] = query[k]
		}
		parsedURL.RawQuery = normalizedQuery.Encode()
	}

	return parsedURL.String(), nil
}

func (sn *StreamNormalizer) AreIdentical(uri1, uri2 string) (bool, error) {
	normalized1, err := sn.Normalize(uri1)
	if err != nil {
		return false, err
	}

	normalized2, err := sn.Normalize(uri2)
	if err != nil {
		return false, err
	}

	return normalized1 == normalized2, nil
}
