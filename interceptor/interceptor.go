package interceptor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/ONSdigital/go-ns/log"
)

const (
	links      = "links"
	dimensions = "dimensions"
	downloads  = "downloads"

	href = "href"
)

var (
	re = regexp.MustCompile(`^(.+:\/\/)(.+$)`)
)

type writer struct {
	http.ResponseWriter
	domain string
}

// WriteHeader wraps the response writer WriteHeader method, but
// removes the content length header so that the correct content
// length is written when the response has been updated
func (w writer) WriteHeader(code int) {
	w.ResponseWriter.Header().Del("Content-Length")

	w.ResponseWriter.WriteHeader(code)
}

// Write intercepts the response writer Write method, parses the json
// and replaces any url domains with the environment host domain
func (w writer) Write(b []byte) (int, error) {
	log.Debug("write method successfully called", nil)

	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		log.Debug("flushing the response writer", nil)
		f.Flush()
	}

	b, err := w.update(b)
	if err != nil {
		log.ErrorC("got an error updating initial response", err, nil)
		return 0, err
	}

	log.Debug("about to call response writer write", nil)

	n, err := w.ResponseWriter.Write(b)

	log.Debug("number of bytes written", log.Data{"n": n})

	return n, err
}

func (w writer) update(b []byte) ([]byte, error) {
	var err error
	document := make(map[string]interface{})
	log.Debug("input bytes", log.Data{"b": string(b)})
	if err = json.Unmarshal(b, &document); err != nil {
		return nil, err
	}

	document, err = w.checkMap(document)
	if err != nil {
		return nil, err
	}

	return json.Marshal(document)
}

func (w writer) checkMap(document map[string]interface{}) (map[string]interface{}, error) {
	var err error

	if docLinks, ok := document[links].(map[string]interface{}); ok {
		document[links], err = updateMap(docLinks, re.ReplaceAllString(w.domain, "${1}api.${2}"))
		if err != nil {
			return nil, err
		}
	}

	if docDownloads, ok := document[downloads].(map[string]interface{}); ok {
		document[downloads], err = updateMap(docDownloads, re.ReplaceAllString(w.domain, "${1}download.${2}"))
		if err != nil {
			return nil, err
		}
	}

	if docDimensions, ok := document[dimensions].([]interface{}); ok {
		document[dimensions], err = updateArray(docDimensions, re.ReplaceAllString(w.domain, "${1}api.${2}"))
		if err != nil {
			return nil, err
		}
	}

	for k, v := range document {
		if subDocument, ok := v.(map[string]interface{}); ok {
			document[k], err = w.checkMap(subDocument)
			if err != nil {
				return nil, err
			}
		}

		if items, ok := v.([]interface{}); ok {
			for i, subVal := range items {
				if subValMap, ok := subVal.(map[string]interface{}); ok {
					items[i], err = w.checkMap(subValMap)
					if err != nil {
						return nil, err
					}
				}
			}
			document[k] = items
		}
	}

	return document, nil
}

func updateMap(docMap map[string]interface{}, domain string) (map[string]interface{}, error) {
	var err error
	for k, v := range docMap {
		if val, ok := v.(map[string]interface{}); ok {
			if field, ok := val[href].(string); ok {
				val[href], err = getLink(field, domain)
				if err != nil {
					return nil, err
				}
			}
			docMap[k] = val
		}
		if val, ok := v.([]interface{}); ok {
			docMap[k], err = updateArray(val, domain)
			if err != nil {
				return nil, err
			}
		}
	}
	return docMap, nil
}

func updateArray(docArray []interface{}, domain string) ([]interface{}, error) {
	var err error
	for i, v := range docArray {
		if val, ok := v.(map[string]interface{}); ok {
			if field, ok := val[href].(string); ok {
				val[href], err = getLink(field, domain)
				if err != nil {
					return nil, err
				}
			}
			docArray[i] = val
		}
	}
	return docArray, nil
}

// Handler takes a given domain to handle the intercepting of http requests with the
// purpose of prepending any api response urls with the correct domain for the environment
func Handler(domain string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h.ServeHTTP(writer{w, domain}, req)
		})
	}
}

func getLink(field, domain string) (string, error) {
	uri, err := url.Parse(field)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", domain, uri.Path), nil
}
