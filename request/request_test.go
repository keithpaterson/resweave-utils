package request

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/keithpaterson/resweave-utils/header"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type testData struct {
	Name string  `json:"name"`
	Cost float64 `json:"cost,omitempty"`
}

func WithTestDataProvider(input []byte) BodyDataProvider {
	return func() ([]byte, string, error) {
		// json unmarshal the input string,  if that works return the string as []byte
		var object testData
		err := json.Unmarshal(input, &object)
		if err != nil {
			return nil, header.MimeTypeJson, err
		}
		return input, header.MimeTypeJson, nil
	}
}

var _ = Describe("Request Helpers", func() {
	Describe("Requests with no Body data", func() {
		// (for now) we can test all the data requests in a loop
		testdata := []struct {
			method    string
			requestFn func(string) (*http.Request, error)
		}{
			{http.MethodGet, NewGetRequest}, {http.MethodDelete, NewDeleteRequest},
		}
		for _, test := range testdata {
			Context(fmt.Sprintf("New%sRequest", cases.Title(language.Und).String(test.method)), func() {
				It("should return an error when uri is malformed", func() {
					req, err := test.requestFn(string([]byte{0x7f}))
					Expect(req).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("net/url: invalid control character")))
				})
				It("should return an error when uri is empty", func() {
					req, err := test.requestFn("")
					Expect(req).To(BeNil())
					Expect(err).To(Equal(ErrorMissingUri))
				})
				It("should return a request when uri is valid", func() {
					req, err := test.requestFn("foo.com")
					Expect(req).ToNot(BeNil())
					Expect(req.Method).To(Equal(test.method))
					Expect(req.URL.Path).To(Equal("foo.com"))
					Expect(err).To(BeNil())
				})
				It("should return another request when uri is valid", func() {
					req, err := test.requestFn("http://www.foo.com/bar")
					Expect(req).ToNot(BeNil())
					Expect(req.Method).To(Equal(test.method))
					Expect(req.URL.Scheme).To(Equal("http"))
					Expect(req.URL.Host).To(Equal("www.foo.com"))
					Expect(req.URL.Path).To(Equal("/bar"))
					Expect(err).To(BeNil())
				})
			})
		}
	})

	Describe("Requests with Body Data", func() {
		// (for now) we can test all the data requests in a loop
		testdata := []struct {
			method    string
			requestFn func(string, BodyDataProvider) (*http.Request, error)
		}{
			{http.MethodPut, NewPutRequest}, {http.MethodPost, NewPostRequest}, {http.MethodPatch, NewPatchRequest},
		}
		for _, test := range testdata {
			Context(fmt.Sprintf("New%sRequest", cases.Title(language.Und).String(test.method)), func() {
				It("should return an error when uri is malformed", func() {
					req, err := test.requestFn(string([]byte{0x7f}), WithNoBody())
					Expect(req).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("net/url: invalid control character")))
				})
				It("should return an error when uri is empty", func() {
					req, err := test.requestFn("", WithNoBody())
					Expect(req).To(BeNil())
					Expect(err).To(Equal(ErrorMissingUri))
				})
				It("should return a request when uri is valid", func() {
					req, err := test.requestFn("foo.com", WithNoBody())
					Expect(req).ToNot(BeNil())
					Expect(req.Method).To(Equal(test.method))
					Expect(req.URL.Path).To(Equal("foo.com"))
					Expect(err).To(BeNil())
				})
				It("should return another request when uri is valid", func() {
					req, err := test.requestFn("http://www.foo.com/bar", WithNoBody())
					Expect(req).ToNot(BeNil())
					Expect(req.Method).To(Equal(test.method))
					Expect(req.URL.Scheme).To(Equal("http"))
					Expect(req.URL.Host).To(Equal("www.foo.com"))
					Expect(req.URL.Path).To(Equal("/bar"))
					Expect(err).To(BeNil())
				})
				It("should return an error when data cannot be provided", func() {
					req, err := test.requestFn("foo.com", WithTestDataProvider([]byte("not json")))
					Expect(req).To(BeNil())
					Expect(err).To(MatchError(HavePrefix("invalid character")))
				})
				It("should return an error when json data cannot be marshaled", func() {
					data := testData{Name: "valid name", Cost: math.NaN()}
					req, err := test.requestFn("foo.com", WithJsonBody(data))
					Expect(req).To(BeNil())
					Expect(err).To(MatchError(ErrorMarshalingBody))
				})
				It("should return a request when json data is valid", func() {
					data := testData{Name: "valid request"}
					req, err := test.requestFn("foo.com", WithJsonBody(data))
					Expect(req.Method).To(Equal(test.method))
					Expect(err).To(BeNil())

					// read the body data to see that it is valid
					raw, err := io.ReadAll(req.Body)
					Expect(err).To(BeNil())
					var actual testData
					err = json.Unmarshal(raw, &actual)
					Expect(err).To(BeNil())
					Expect(actual).To(Equal(data))
				})
				It("should return a request when binary data is valid", func() {
					req, err := test.requestFn("foo.com", WithBinaryBody([]byte("test data")))
					Expect(req.Method).To(Equal(test.method))
					Expect(err).To(BeNil())

					// read the body data to see that it is valid
					raw, err := io.ReadAll(req.Body)
					Expect(err).To(BeNil())
					Expect(raw).To(Equal([]byte("test data")))
				})
			})
		}
	})
})
