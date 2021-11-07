package interceptor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"testing"
)

// run with:
// go test -run=interceptor_roundtrip_benchmark_test.go -bench=Test1 -memprofile=mem0.out
//
// examine result file:
// go tool pprof --alloc_space interceptor.test mem0.out
//
// With code as of 4th Nov produces:
/*
Benchmarking: 'roundTrip'
test interceptor correctly updates a href in links subdocs within an array
goos: darwin
goarch: amd64
pkg: github.com/ONSdigital/dp-api-router/interceptor
cpu: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
BenchmarkTest1-12    	Benchmarking: 'roundTrip'
test interceptor correctly updates a href in links subdocs within an array
Benchmarking: 'roundTrip'
test interceptor correctly updates a href in links subdocs within an array
Benchmarking: 'roundTrip'
test interceptor correctly updates a href in links subdocs within an array
   97032	     11283 ns/op	    7265 B/op	     117 allocs/op
PASS
ok  	github.com/ONSdigital/dp-api-router/interceptor	1.318s
*/
// enter command:
// top30
//
// produces:
/*
Showing nodes accounting for 709.79MB

      flat  flat%   sum%        cum   cum%
  204.05MB 28.06% 28.06%   216.05MB 29.71%  encoding/json.(*decodeState).objectInterface
  101.55MB 13.96% 42.02%   101.55MB 13.96%  io.ReadAll
   57.01MB  7.84% 49.86%    57.01MB  7.84%  reflect.mapiterinit
   39.01MB  5.36% 55.22%    39.01MB  5.36%  net/textproto.MIMEHeader.Set (inline)
   33.50MB  4.61% 59.83%   122.51MB 16.84%  encoding/json.mapEncoder.encode
   30.50MB  4.19% 64.02%    30.50MB  4.19%  net/url.parse
   29.01MB  3.99% 68.01%   727.30MB   100%  github.com/ONSdigital/dp-api-router/interceptor.BenchmarkTest1
      19MB  2.61% 70.62%       19MB  2.61%  reflect.copyVal
      17MB  2.34% 72.96%    31.50MB  4.33%  net/http/httptest.(*ResponseRecorder).Result
   16.50MB  2.27% 75.23%    16.50MB  2.27%  bytes.NewReader (inline)
   14.50MB  1.99% 77.22%    20.16MB  2.77%  regexp.(*Regexp).backtrack
   13.50MB  1.86% 79.08%    13.50MB  1.86%  bytes.makeSlice
      13MB  1.79% 80.87%    79.51MB 10.93%  reflect.Value.MapKeys
   11.50MB  1.58% 82.45%   241.55MB 33.21%  encoding/json.Unmarshal
   11.50MB  1.58% 84.03%    11.50MB  1.58%  net/http/httptest.NewRecorder (inline)
      11MB  1.51% 85.54%       11MB  1.51%  fmt.Sprintf
      11MB  1.51% 87.05%       11MB  1.51%  regexp.(*Regexp).expand
       8MB  1.10% 88.15%       48MB  6.60%  github.com/ONSdigital/dp-api-router/interceptor.getLink
    7.50MB  1.03% 89.18%   223.55MB 30.74%  encoding/json.(*decodeState).arrayInterface
       7MB  0.96% 90.15%        7MB  0.96%  bytes.NewBuffer (inline)
       7MB  0.96% 91.11%        7MB  0.96%  io.MultiReader (inline)
    6.50MB  0.89% 92.00%     6.50MB  0.89%  encoding/json.unquote (inline)
       6MB  0.82% 92.83%    39.16MB  5.38%  regexp.(*Regexp).ReplaceAllString
       6MB  0.82% 93.65%   239.68MB 32.95%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).updateSlice
    5.50MB  0.76% 94.41%     5.50MB  0.76%  net/http.Header.Clone
    5.50MB  0.76% 95.17%     9.50MB  1.31%  encoding/json.(*decodeState).literalInterface
    5.16MB  0.71% 95.87%     5.16MB  0.71%  regexp.(*bitState).reset
       5MB  0.69% 96.56%        5MB  0.69%  encoding/json.(*scanner).pushParseState
       4MB  0.55% 97.11%   698.29MB 96.01%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).RoundTrip
    3.50MB  0.48% 97.59%    10.50MB  1.44%  github.com/ONSdigital/dp-api-router/interceptor.NewMultiReadCloser (inline)
*/

var testJSON = `[{"links":{"self":{"href":"/datasets/12345"}}}, {"links":{"self":{"href":"/datasets/12345"}}}]`

func BenchmarkTest1(b *testing.B) {
	fmt.Println("Benchmarking: 'roundTrip', using code from unit test that is known to work")

	fmt.Println("test interceptor correctly updates a href in links subdocs within an array")
	//	testJSON := `[{"links":{"self":{"href":"/datasets/12345"}}}, {"links":{"self":{"href":"/datasets/12345"}}}]`
	transp := dummyRT{testJSON}

	t := NewRoundTripper(testDomain, "", transp)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {

		/*resp, err := */
		t.RoundTrip(&http.Request{RequestURI: "/datasets"})
	}

	/*So(err, ShouldBeNil)

	b, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)

	err = resp.Body.Close()
	So(err, ShouldBeNil)
	So(len(b), ShouldEqual, 154)
	So(string(b), ShouldEqual, `[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}},{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}]`+"\n")
	*/
}

// to help analysis, in directory of dp-api-router/interceptor, do:
//
// go build -gcflags='-m -m' ./...
//
// and examine the output to see what 'escapes to heap'
//

// 5th Nov, adding some sync.pool where possible gives minor improvement:
/*
Benchmarking: 'roundTrip'
test interceptor correctly updates a href in links subdocs within an array
   95220	     11306 ns/op	    7061 B/op	     115 allocs/op
PASS


File: interceptor.test
Type: alloc_space
Time: Nov 5, 2021 at 10:46am (GMT)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top30
Showing nodes accounting for 666.78MB, 97.37% of 684.79MB total
Dropped 15 nodes (cum <= 3.42MB)
Showing top 30 nodes out of 51
      flat  flat%   sum%        cum   cum%
  197.05MB 28.77% 28.77%   202.55MB 29.58%  encoding/json.(*decodeState).objectInterface
   99.05MB 14.46% 43.24%    99.05MB 14.46%  io.ReadAll
      53MB  7.74% 50.98%       53MB  7.74%  reflect.mapiterinit
   40.50MB  5.91% 56.89%   126.01MB 18.40%  encoding/json.mapEncoder.encode
   35.51MB  5.19% 62.08%    35.51MB  5.19%  net/textproto.MIMEHeader.Set (inline)
   25.51MB  3.72% 65.80%   684.79MB   100%  github.com/ONSdigital/dp-api-router/interceptor.BenchmarkTest1
      21MB  3.07% 68.87%       21MB  3.07%  reflect.copyVal
      20MB  2.92% 71.79%       20MB  2.92%  net/url.parse
   19.50MB  2.85% 74.64%   238.55MB 34.84%  encoding/json.Unmarshal
   17.50MB  2.56% 77.20%       27MB  3.94%  net/http/httptest.(*ResponseRecorder).Result
      14MB  2.04% 79.24%       14MB  2.04%  bytes.NewReader (inline)
   11.50MB  1.68% 80.92%    74.01MB 10.81%  reflect.Value.MapKeys
      11MB  1.61% 82.53%    17.65MB  2.58%  regexp.(*Regexp).backtrack
   10.50MB  1.53% 84.06%    10.50MB  1.53%  net/http/httptest.NewRecorder (inline)
   10.50MB  1.53% 85.59%    10.50MB  1.53%  regexp.(*Regexp).expand
      10MB  1.46% 87.05%       11MB  1.61%  fmt.Sprintf
    8.50MB  1.24% 88.29%     8.50MB  1.24%  encoding/json.(*scanner).pushParseState
       8MB  1.17% 89.46%       37MB  5.40%  github.com/ONSdigital/dp-api-router/interceptor.getLink
    7.50MB  1.10% 90.56%    37.65MB  5.50%  regexp.(*Regexp).ReplaceAllString
       7MB  1.02% 91.58%        7MB  1.02%  io.MultiReader (inline)
    6.65MB  0.97% 92.55%     6.65MB  0.97%  regexp.(*bitState).reset
       6MB  0.88% 93.43%   659.28MB 96.28%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).RoundTrip
       5MB  0.73% 94.16%   207.55MB 30.31%  encoding/json.(*decodeState).arrayInterface
       5MB  0.73% 94.89%   209.16MB 30.54%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).updateSlice
       5MB  0.73% 95.62%        5MB  0.73%  encoding/json.unquote (inline)
    4.50MB  0.66% 96.28%    11.50MB  1.68%  github.com/ONSdigital/dp-api-router/interceptor.NewMultiReadCloser (inline)
       3MB  0.44% 96.71%   210.55MB 30.75%  encoding/json.(*decodeState).array
       2MB  0.29% 97.01%       39MB  5.70%  github.com/ONSdigital/dp-api-router/interceptor.updateMap
       2MB  0.29% 97.30%    30.15MB  4.40%  regexp.(*Regexp).replaceAll
    0.50MB 0.073% 97.37%   448.22MB 65.45%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).update

*/

// Conlusion: (5th Nov 2021)
//
// There is little opportunity to use sync.pool due to the functions used within the scope of RoundTrip()
// Where i could use sync.pool, in the particular test used, it reduced memory allocations by just under 3%
// (7265 reduced to 7061 bytes), which is not really enough to justify complicating the code from a maintenance
// perspective by using sync.pool
//
// Ultimately ALL App's that create work requiring the use of a RoundTrip function and those Apps that rely on
// that work need re-writting.
//
// The current re-write of the code up to and including the 4th of November should stop reading in large files in to
// memory which have no place being in memory and may well prove to be a sufficient approach.
// It needs testing in Develop and the logs need scrutinizing to ensure all endpoints that need processing are
// actually processed.
//
//
// One final thought:
// The unit tests work with modified code using sync.pool, and so does the one benchmark ...
//
// BUT this line:
// 		resp.Body = ioutil.NopCloser(bytes.NewReader(updatedB.Bytes()))
//
// is effectively postponing the use of the bytes in the buffer to after the function returns
// and that buffer is put back before the function returns with:
// 		defer eventBufPool.Put(updatedB)
//
// The compiler might track that buffer until it has gone out of scope, but it seems possible it wont.
//
// ... So, i don't think using sync.pool as used here is correct usage and may well fail in production
// under heavy load.

// in fact this article explains why my use of syncpool here is not going to work as required
// due to saving off a bointer to a butter that then goes out of scope when function returns:
// https://www.captaincodeman.com/golang-buffer-pool-gotcha

// I believe the correct usage of sync.pool is
//
/*
	// allocate a buffer pointer
	buf := eventBufPool.Get().(*bytes.Buffer) // with casting on the end
	buf.Reset()                               // Must reset before each block of usage

	- Build up stuff in buf

	- Write final complete buf to destination

	// release the buf' back to the pool
	eventBufPool.Put(buf)

*/

// -=-=-

var testJSON100 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
var testJSON1000 = testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100 + testJSON100
var testJSON10000 = testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000 + testJSON1000

func BenchmarkTest2(b *testing.B) {
	fmt.Println("Benchmarking: 'roundTrip'")

	fmt.Println("test interceptor correctly does not load in very large object, that might be a json object")
	transp := dummyRT{testJSON10000}

	t := NewRoundTripper(testDomain, "", transp)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {

		/*resp, err := */
		t.RoundTrip(&http.Request{RequestURI: "/datasets"})
	}
}

// results of above test show that large object that does not start with '{' or '[' is not loaed into memory:
//
//
//  106752	     11552 ns/op	    3148 B/op	      45 allocs/op
//
// and:
/*
File: interceptor.test
Type: alloc_space
Time: Nov 5, 2021 at 2:44pm (GMT)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top30
Showing nodes accounting for 336.56MB, 99.41% of 338.56MB total
Dropped 10 nodes (cum <= 1.69MB)
Showing top 30 nodes out of 36
      flat  flat%   sum%        cum   cum%
   52.53MB 15.51% 15.51%    52.53MB 15.51%  io.ReadAll
   42.01MB 12.41% 27.92%   310.06MB 91.58%  github.com/ONSdigital/dp-api-router/interceptor.(*Transport).RoundTrip
   28.51MB  8.42% 36.34%   338.56MB   100%  github.com/ONSdigital/dp-api-router/interceptor.BenchmarkTest2
   26.01MB  7.68% 44.02%    26.01MB  7.68%  github.com/ONSdigital/log.go/v2/log.printEvent
   20.50MB  6.06% 50.08%    64.51MB 19.05%  encoding/json.mapEncoder.encode
   19.50MB  5.76% 55.84%    91.51MB 27.03%  encoding/json.Marshal
   17.50MB  5.17% 61.01%   109.01MB 32.20%  github.com/ONSdigital/log.go/v2/log.styleForMachine
   16.50MB  4.87% 65.88%       34MB 10.04%  net/http/httptest.(*ResponseRecorder).Result
      14MB  4.14% 70.02%       14MB  4.14%  reflect.copyVal
   11.50MB  3.40% 73.42%    11.50MB  3.40%  internal/reflectlite.Swapper
      11MB  3.25% 76.66%    11.50MB  3.40%  github.com/ONSdigital/log.go/v2/log.createEvent
      10MB  2.95% 79.62%       25MB  7.38%  reflect.Value.MapKeys
       9MB  2.66% 82.28%        9MB  2.66%  net/http/httptest.NewRecorder (inline)
    8.50MB  2.51% 84.79%     8.50MB  2.51%  reflect.mapiterinit
       8MB  2.36% 87.15%        8MB  2.36%  bytes.NewReader (inline)
       8MB  2.36% 89.51%    14.50MB  4.28%  github.com/ONSdigital/dp-api-router/interceptor.NewMultiReadCloser (inline)
    7.50MB  2.22% 91.73%     7.50MB  2.22%  net/http.Header.Clone
       7MB  2.07% 93.80%        7MB  2.07%  time.Time.MarshalJSON
    6.50MB  1.92% 95.72%     6.50MB  1.92%  io.MultiReader (inline)
    5.50MB  1.62% 97.34%     5.50MB  1.62%  io.NopCloser (inline)
       5MB  1.48% 98.82%        5MB  1.48%  strings.NewReader (inline)
       2MB  0.59% 99.41%        2MB  0.59%  io.LimitReader (inline)
         0     0% 99.41%    72.01MB 21.27%  encoding/json.(*encodeState).marshal
         0     0% 99.41%    72.01MB 21.27%  encoding/json.(*encodeState).reflectValue
         0     0% 99.41%        7MB  2.07%  encoding/json.condAddrEncoder.encode
         0     0% 99.41%        7MB  2.07%  encoding/json.marshalerEncoder
         0     0% 99.41%    64.51MB 19.05%  encoding/json.ptrEncoder.encode
         0     0% 99.41%    71.51MB 21.12%  encoding/json.structEncoder.encode
         0     0% 99.41%       49MB 14.47%  github.com/ONSdigital/dp-api-router/interceptor.dummyRT.RoundTrip
         0     0% 99.41%   146.52MB 43.28%  github.com/ONSdigital/log.go/v2/log.Error
*/

// -=-=-

var testJSON2 = `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/1234"}}}`

// running test3 with:
//  go test -race -run=interceptor_roundtrip_benchmark_test.go -bench=Test3 -memprofile=mem0.out
// gives:
//  28	  36988328 ns/op	  162985 B/op	     884 allocs/op
//
// ... in the run with race checking, there is a lot of other things going on taking time away from benchmarking.
//
// running test3 with:
//  go test -run=interceptor_roundtrip_benchmark_test.go -bench=Test3 -memprofile=mem0.out
// gives:
//  172375	      8255 ns/op	   13088 B/op	     185 allocs/op

func BenchmarkTest3(b *testing.B) {
	fmt.Println("Benchmarking: 'roundTrip', using code from unit test that is known to work")

	fmt.Println("demonstrate sync.Pool failure when running parallel or OK without sync.Pool")
	//	testJSON := `[{"links":{"self":{"href":"/datasets/12345"}}}, {"links":{"self":{"href":"/datasets/12345"}}}]`
	transp := dummyRT{testJSON}
	t := NewRoundTripper(testDomain, "", transp)

	transp2 := dummyRT{testJSON2}
	t2 := NewRoundTripper(testDomain, "", transp2)

	b.ReportAllocs()

	var c int32

	trip := func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt32(&c, 1)
			resp, err := t.RoundTrip(&http.Request{RequestURI: "/datasets"})
			if err != nil {
				fmt.Printf("RoundTrip Error: %v\n", err)
				panic(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("ReadAll: %v\n", err)
				panic(err)
			}
			if string(body) != `[{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}},{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/12345"}}}]`+"\n" {
				panic(fmt.Errorf("wrong result: %d: %v\n", c, string(body)))
			}

			resp2, err := t2.RoundTrip(&http.Request{RequestURI: "/datasets"})
			if err != nil {
				fmt.Printf("RoundTrip Error: %v\n", err)
				panic(err)
			}
			b2, err := ioutil.ReadAll(resp2.Body)
			if err != nil {
				fmt.Printf("ReadAll: %v\n", err)
				panic(err)
			}
			if string(b2) != `{"links":{"self":{"href":"https://api.beta.ons.gov.uk/v1/datasets/1234"}}}`+"\n" {
				panic(fmt.Errorf("wrong result 2: %d: %v\n", c, string(b2)))
			}
		}
	}

	b.SetParallelism(500)
	b.RunParallel(trip)
}

/* Test results for benchmarks, bytes used on HEAP:

Code as of 16th Sep 2021

Test1 => 6,537
Test2 => 86,496
Test3 => 11,542

Code in Develop as updated on 2nd Nov 2021

Test1 => 6,539
Test2 => 53,802
Test3 => 11,576

Code on 'feature/undo-part-of-unmarshal-specific-things' as of 6th Nov 2021

Test1 => 7,261
Test2 => 3,149    <---- *** This is the important one that will avoid BIG files causing problems.
Test3 => 13,014

*/
