package interceptor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/log"
)

// Transport implements the http RoundTripper method and allows the
// response body to be post processed
type Transport struct {
	domain     string
	contextURL string
	http.RoundTripper
}

var _ http.RoundTripper = &Transport{}

// NewRoundTripper creates a Transport instance with configured domain
func NewRoundTripper(domain, contextURL string, rt http.RoundTripper) *Transport {
	return &Transport{domain, contextURL, rt}
}

const (
	links      = "links"
	dimensions = "dimensions"
	downloads  = "downloads"

	href = "href"
)

var (
	re = regexp.MustCompile(`^(.+://)(.+)(/v\d)$`)
)

var pathsToIgnore = []string{
	"/v1/tokens",
	"/v1/users",
	"/v1/groups",
	"/v1/password-reset",
}

func shallIgnore(path string) bool {
	for _, pathToIgnore := range pathsToIgnore {
		if strings.HasPrefix(path, pathToIgnore) {
			return true
		}
	}
	return false
}

// RoundTrip intercepts the response body and post processes to add the correct enviornment
// host to links
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if !shallIgnore(req.RequestURI){
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if len(b) == 0 {
			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
			return resp, nil
		}

		updatedB, err := t.update(b)
		if err != nil {
			log.Event(req.Context(), "could not update response body with correct links", log.ERROR, log.Error(err))
			body := ioutil.NopCloser(bytes.NewReader(b))

			resp.Body = body
			return resp, nil
		}

		body := ioutil.NopCloser(bytes.NewReader(updatedB))

		resp.Body = body
		resp.ContentLength = int64(len(updatedB))
		resp.Header.Set("Content-Length", strconv.Itoa(len(updatedB)))
	}
	return resp, nil
}

func (t *Transport) update(b []byte) ([]byte, error) {

	var (
		err      error
		resource interface{}
	)

	if err = json.Unmarshal(b, &resource); err != nil {
		return nil, err
	}

	resourceType := reflect.TypeOf(resource)
	if resourceType == nil {
		return nil, errors.New("nil resource type")
	}

	switch resourceType.Kind() {
	case reflect.Map:
		// Assert type onto document
		return t.updateMap(resource.(map[string]interface{}))
	case reflect.Slice:
		// Assert type onto documents
		return t.updateSlice(resource.([]interface{}))
	default:
		return nil, errors.New("unknown resource type")
	}
}

func (t *Transport) updateMap(document map[string]interface{}) ([]byte, error) {
	var err error

	document, err = t.checkMap(document)
	if err != nil {
		return nil, err
	}

	if len(t.contextURL) > 0 {
		document["@context"] = t.contextURL
	}
	var updatedB []byte
	buf := bytes.NewBuffer(updatedB)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	err = enc.Encode(document)

	return buf.Bytes(), err
}

func (t *Transport) updateSlice(documents []interface{}) ([]byte, error) {
	var (
		documentList []map[string]interface{}
		err          error
	)

	for i := range documents {
		document := documents[i].(map[string]interface{})
		document, err = t.checkMap(document)
		if err != nil {
			return nil, err
		}
		documentList = append(documentList, document)
	}

	var updatedB []byte
	buf := bytes.NewBuffer(updatedB)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	err = enc.Encode(documentList)

	return buf.Bytes(), err
}

func (t *Transport) checkMap(document map[string]interface{}) (map[string]interface{}, error) {
	var err error

	if docLinks, ok := document[links].(map[string]interface{}); ok {
		document[links], err = updateMap(docLinks, re.ReplaceAllString(t.domain, "${1}api.${2}${3}"))
		if err != nil {
			return nil, err
		}
	}

	if docDownloads, ok := document[downloads].(map[string]interface{}); ok {
		document[downloads], err = updateMap(docDownloads, re.ReplaceAllString(t.domain, "${1}download.${2}"))
		if err != nil {
			return nil, err
		}
	}

	// Dataset api versions endpoint treats dimensions as an array
	if docDimensions, ok := document[dimensions].([]interface{}); ok {
		document[dimensions], err = updateArray(docDimensions, re.ReplaceAllString(t.domain, "${1}api.${2}${3}"))
		if err != nil {
			return nil, err
		}
	}

	// Dataset api observations endpoint treats dimensions as a nested list
	if docDimensions, ok := document[dimensions].(map[string]interface{}); ok {
		document[dimensions], err = updateMap(docDimensions, re.ReplaceAllString(t.domain, "${1}api.${2}${3}"))
		if err != nil {
			return nil, err
		}
	}

	for k, v := range document {
		if subDocument, ok := v.(map[string]interface{}); ok {
			document[k], err = t.checkMap(subDocument)
			if err != nil {
				return nil, err
			}
		}

		if items, ok := v.([]interface{}); ok {
			for i, subVal := range items {
				if subValMap, ok := subVal.(map[string]interface{}); ok {
					items[i], err = t.checkMap(subValMap)
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
			} else {
				val, err = updateMap(val, domain)
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

func getLink(field, domain string) (string, error) {

	// if the URL is already correct, return it
	if strings.HasPrefix(field, domain) {
		return field, nil
	}

	uri, err := url.Parse(field)
	if err != nil {
		return "", err
	}

	queries := uri.RawQuery

	if len(queries) == 0 {
		return fmt.Sprintf("%s%s", domain, uri.Path), nil
	}
	return fmt.Sprintf("%s%s?%s", domain, uri.Path, queries), nil

}
