package interceptor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
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

	// NOTE: Don't go changing 'maxBodyLengthToLog' value to omuch from '20' as its used to generate boundary test cases.
	maxBodyLengthToLog = 20 // only log a small part of the body to help any problem diagnosis, as the full body length could be many Megabytes
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

var pathsToUse = []string{
	"/datasets",
	"/v1/filter-outputs",
	"/v1/filters",
	"/code-lists",
	"/v1/hierarchies",
	"/v1/dimension-search",
}

// Check to see whether the response should be remapped
func shallIgnore(path string) bool {
	for _, pathToIgnore := range pathsToIgnore {
		if strings.HasPrefix(path, pathToIgnore) {
			return true
		}
	}
	return false
}

func shallUse(path string) bool {
	for _, pathToUse := range pathsToUse {
		if strings.HasPrefix(path, pathToUse) {
			return true
		}
	}
	return false
}

// RoundTrip intercepts the response body and post processes to add the correct environment
// host to links
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type") // get canonical form

	if strings.Contains(contentType, "gzip") {
		return resp, nil
	}

	if shallUse(req.RequestURI) {

		// get small number of bytes from resp
		readdata, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxBodyLengthToLog))
		if err != nil {
			rawQuery := ""
			if resp.Request != nil && resp.Request.URL != nil {
				rawQuery = resp.Request.URL.RawQuery
			}
			log.Error(req.Context(), "Problem reading first part of resp'", err, log.Data{
				"content_type":     contentType,                         // needed to further identify content types that need to be rejected similarly to 'gzip' above
				"content_encoding": resp.Header.Get("Content-Encoding"), // as above
				"raw_query":        rawQuery,                            // as above
			})
			return nil, err
		}
		if len(readdata) == 0 {
			err = resp.Body.Close()
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
			return resp, nil
		}

		if readdata[0] != '{' && readdata[0] != '[' {
			// quickly reject non json or map files such as .zip's, to avoid reading in the body of potentially very large objects
			rawQuery := ""
			if resp.Request != nil && resp.Request.URL != nil {
				rawQuery = resp.Request.URL.RawQuery
			}
			log.Error(req.Context(), "Not a JSON file", err, log.Data{
				"body":             string(readdata),
				"content_type":     contentType,                         // needed to further identify content types that need to be rejected similarly to 'gzip' above
				"content_encoding": resp.Header.Get("Content-Encoding"), // as above
				"raw_query":        rawQuery,                            // as above
			})
			// recombine the buffered part of the body with the rest of the stream
			first := ioutil.NopCloser(bytes.NewReader(readdata))
			resp2 := resp
			// NOTE: MultiReader will do the "resp.Body.Close()"
			resp2.Body = ioutil.NopCloser(io.MultiReader(first, resp.Body))
			return resp2, nil
		}

		// get the rest of the stream, which should be of reasonable size
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// prefix the first chunk read above back into the stream
		b = append(readdata, b...)

		bodyLength := len(b)

		updatedB, err := t.update(b)
		if err != nil {
			limitedBodyLength := bodyLength
			if limitedBodyLength > maxBodyLengthToLog {
				limitedBodyLength = maxBodyLengthToLog
			}
			rawQuery := ""
			if resp.Request != nil && resp.Request.URL != nil {
				rawQuery = resp.Request.URL.RawQuery
			}
			log.Error(req.Context(), "could not update response body with correct links", err, log.Data{
				"body":             string(b[0:limitedBodyLength]),
				"content_type":     contentType,                         // needed to further identify content types that need to be rejected similarly to 'gzip' above
				"body_length":      bodyLength,                          // as above
				"content_encoding": resp.Header.Get("Content-Encoding"), // as above
				"raw_query":        rawQuery,                            // as above
			})
			resp.Body = ioutil.NopCloser(bytes.NewReader(b))
			return resp, nil
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(updatedB))
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
	case reflect.Map: // starts with {
		// Assert type onto document
		return t.updateMap(resource.(map[string]interface{}))
	case reflect.Slice: // starts with [
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
