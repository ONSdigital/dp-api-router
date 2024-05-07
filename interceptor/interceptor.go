package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-router/config"
	"github.com/ONSdigital/log.go/v2/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	links        = "links"
	datasetLinks = "dataset_links"
	dimensions   = "dimensions"
	downloads    = "downloads"
	href         = "href"

	// NOTE: Don't change 'maxBodyLengthToLog' value too much from '20' as it's used to generate boundary test cases.
	maxBodyLengthToLog = 20 // only log a small part of the body to help any problem diagnosis, as the full body length could be many Megabytes
)

// keyVal used as a helper to encode strings for output as JSON
type keyVal struct {
	K string
}

// Transport implements the http RoundTripper method and allows the
// response body to be post processed
type Transport struct {
	domain      string
	contextURL  string
	apiURL      string
	downloadURL string
	http.RoundTripper
}

var _ http.RoundTripper = &Transport{}

// NewRoundTripper creates a Transport instance with configured domain
func NewRoundTripper(domain, contextURL string, rt http.RoundTripper) *Transport {
	cfg, err := config.Get()
	if err != nil {
		log.Error(context.Background(), "Unable to retrieve config'", err)
	}

	if cfg.OtelEnabled {
		return &Transport{domain, contextURL, otelhttp.NewTransport(rt)}
	}

	return &Transport{domain, contextURL, rt}
}

var (
	re                          = regexp.MustCompile(`^(.+://)(.+)(/v\d)$`)
	reIsChars                   = regexp.MustCompile(`^[- a-zA-Z/0-9_?=+!@$%&*()\[\]{}|':;?/<>.,]*$`)
	_         http.RoundTripper = &Transport{}
)

// NewRoundTripper creates a Transport instance with configured domain
func NewRoundTripper(domain, contextURL string, rt http.RoundTripper) *Transport {
	apiURL := re.ReplaceAllString(domain, "${1}api.${2}${3}")
	downloadURL := re.ReplaceAllString(domain, "${1}download.${2}")
	return &Transport{domain, contextURL, apiURL, downloadURL, otelhttp.NewTransport(rt)}
}

// RoundTrip intercepts the response body and post processes to add the correct environment
// host to links
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Make the request to the server
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type") // get canonical form

	if strings.Contains(contentType, "gzip") {
		return resp, nil
	}

	// "contentEncoding": "gzip" ... might need to exclude these things at some point

	// get small number of bytes from resp
	readdata, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyLengthToLog))
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
		resp.Body = io.NopCloser(bytes.NewReader([]byte{}))
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
		// recombine the buffered 'first' part of the body with any remaining part of the stream
		resp.Body = NewMultiReadCloser(bytes.NewReader(readdata), resp.Body)
		return resp, nil
	}

	updatedBody := ""
	depthType := []string{}
	hasLen := []int64{}
	path := []string{}
	depth := -1
	mReader := NewMultiReadCloser(bytes.NewReader(readdata), resp.Body)
	jsDecoder := json.NewDecoder(mReader)
	jsDecoder.UseNumber()

	outputNextTokenFunc := func(s string, isOpening, isClosing bool) {
		hasMore := jsDecoder.More()
		depType := "-"
		if depth >= 0 {
			depType = depthType[depth]
		}
		appendChars := ","
		if !hasMore || isOpening {
			appendChars = ""
		}
		if isClosing {
			if depth < 0 {
				path = []string{}
			} else {
				path = path[0 : depth+1]
			}
		} else if isOpening {
			if depth-1 < len(path) {
				path = append(path, ".")
			}
			if depType == `[` {
				path[depth] = "[]"
			} else {
				// is `{` opener
				if depth == 0 && len(t.contextURL) > 0 {
					appendChars = `"@context":"` + t.contextURL + `"`
					if hasMore {
						appendChars += ","
					}
				}
			}
		} else if hasMore && depType == `{` && hasLen[depth]%2 == 0 {
			appendChars = ":"
			path[depth] = s
		}
		updatedBody += appendChars
	}

	skipNextToken := false
	for {
		token, err := jsDecoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			resp.Body = NewMultiReadCloser(bytes.NewReader([]byte(readdata)), mReader)
			return resp, nil
		}
		switch tk := token.(type) {
		case json.Delim:
			tokenStr := tk.String()
			if tokenStr == `{` || tokenStr == `[` {
				depth++
				hasLen = append(hasLen, 0)
				depthType = append(depthType, tokenStr)
				updatedBody += tk.String()
				outputNextTokenFunc(tokenStr, true, false)
			} else if tokenStr == `}` || tokenStr == `]` {
				depth--
				if depth >= 0 {
					hasLen = hasLen[0 : depth+1]
					hasLen[depth]++
					depthType = depthType[0 : depth+1]
				} else if depth < 0 {
					hasLen = []int64{}
					depthType = []string{}
				}
				updatedBody += tk.String()
				outputNextTokenFunc(tokenStr, false, true)
			} else {
				return nil, fmt.Errorf("delim: bad type `%T` for `%v` at depth %d", token, token, depth)
			}
		case string:
			skipThisKey := false
			if depthType[depth] == "{" {
				if hasLen[depth]%2 == 0 { // key
					if depth == 0 && tk == "@context" && len(t.contextURL) > 0 {
						skipThisKey = true
						skipNextToken = true
					}
				} else if len(path) >= 3 && path[len(path)-1] == href {
					// we have the value for an object and it is an href with at least two (more) ancestors
					if path[len(path)-3] == links ||
						(len(path) >= 4 && path[len(path)-4] == links && path[len(path)-2] == "[]") {

						tk, err = reLink(tk, t.apiURL)
						if err != nil {
							return nil, err
						}
					}
					if path[len(path)-3] == datasetLinks {
						tk, err = reLink(tk, t.apiURL)
						if err != nil {
							return nil, err
						}
					}
					if path[len(path)-3] == downloads {
						tk, err = reLink(tk, t.downloadURL)
						if err != nil {
							return nil, err
						}
					}
					if path[len(path)-3] == dimensions ||
						(len(path) >= 4 && path[len(path)-4] == dimensions && path[len(path)-2] != "[]") {

						// Dataset api versions endpoint treats dimensions as an array
						tk, err = reLink(tk, t.apiURL)
						if err != nil {
							return nil, err
						}
					}
				}
			}

			if skipThisKey {
				// nought
			} else if skipNextToken {
				skipNextToken = false
				if !jsDecoder.More() && updatedBody[len(updatedBody)-1] == ',' {
					updatedBody = updatedBody[0 : len(updatedBody)-1]
				}
			} else {
				if len(tk) > 1024 || !reIsChars.Match([]byte(tk)) {
					// encode tk
					var encodedTk []byte
					encodedTk, err = json.Marshal(keyVal{K: tk})
					if err != nil {
						return nil, err
					}
					updatedBody += string(encodedTk[5 : len(encodedTk)-1])
				} else {
					updatedBody += `"` + tk + `"`
				}
				outputNextTokenFunc(tk, false, false)
				hasLen[depth]++
			}
		case json.Number:
			numStr := tk.String()
			updatedBody += numStr
			outputNextTokenFunc(numStr, false, false)
			hasLen[depth]++
		case bool:
			valStr := "true"
			if !tk {
				valStr = "false"
			}
			updatedBody += valStr
			outputNextTokenFunc(valStr, false, false)
			hasLen[depth]++
		case nil:
			updatedBody += `null`
			outputNextTokenFunc("null", false, false)
			hasLen[depth]++
		default:
			return nil, fmt.Errorf("other: bad type `%T` for `%v` at depth %d", token, token, depth)
		}
	}

	// return updated body
	resp.Body = io.NopCloser(strings.NewReader(updatedBody))
	resp.ContentLength = int64(len(updatedBody))
	resp.Header.Set("Content-Length", strconv.Itoa(len(updatedBody)))

	return resp, nil
}

type multiReadCloser struct {
	readers     []io.Reader
	multiReader io.Reader
}

func NewMultiReadCloser(readers ...io.Reader) io.ReadCloser {
	return &multiReadCloser{
		readers:     readers,
		multiReader: io.MultiReader(readers...),
	}
}

func (r *multiReadCloser) Read(p []byte) (n int, err error) {
	return r.multiReader.Read(p)
}

func (r *multiReadCloser) Close() (err error) {
	for _, r := range r.readers {
		if c, ok := r.(io.Closer); ok {
			if e := c.Close(); e != nil {
				err = e
			}
		}
	}

	return err
}

func reLink(field, domain string) (string, error) {
	// if the URL is already correct, return it
	if strings.HasPrefix(field, domain) {
		return field, nil
	}

	uri, err := url.Parse(field)
	if err != nil {
		return "", err
	}

	queries := uri.RawQuery

	if queries == "" {
		return fmt.Sprintf("%s%s", domain, uri.Path), nil
	}
	return fmt.Sprintf("%s%s?%s", domain, uri.Path, queries), nil
}
