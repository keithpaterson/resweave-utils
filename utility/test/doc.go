// contains tools that are only available during testing.
// add '--tags testutil' to your (test) builds to access these tools.
//
// These tools expect that you are using gomega for your assertions/matching.
//
// # Test Service
//
// # func HttpService() *httpService
//
// Provides a service implementation that can be configured to assert/expect
// specific conditions on incoming requests.
//
// # func (s *httpService) Start() (string, func())
//
// Starts the service and waits for requests.
//
// Returns the hostname (usually localhost and some port) and a shutdown function that should be called at the end of the test.
//
//	host, stopFn := HttpService().WithMethod(http.MethodPost).WithPath("/foo").Start()
//	defer stopFn()
//
// # Conditions
//
// If a condition fails, the test will fail.
//
// # func (s *httpService) WithMethod(method string) *httpService
//
// # func (s *httpService) WithPath(path string) *httpService
//
// # func (s *httpService) WithBody(body interface{}) *httpService
//
// Expect body data and that it matches the data provided.
//
// If body's type is []byte, will act as WithBinaryBody(body), otherwise
// will act as WithJsonBody(body).
//
// If you don't expect body data don't add this condition.
//
// # func (s *httpService) WithBinaryBody(body []byte) *httpService
//
// Expect body data, and that it matches the data provided.
//
// # func (s *httpService) WithJsonBody(object interface{}) *httpService
//
// Expect body data, and that it can be Unmarshaled into json, and that the json data matches.
//
// # Controls and Responses
//
// # func (s *httpService) ReturnStatusCode(status int) *httpService
//
// The response will contain the specified status code.
//
// # func (s *httpService) ReturnBody(body interface{}) *httpService
//
// Add the body data to the response.
//
// if body's type is []byte it will act as ReturnBinaryBody(object), otherwise
// it will act as ReturnJsonBody(object).
//
// # func (s *httpService) ReturnBinaryBody(body []byte) *httpService
//
// Add binary data to the response (Mime type will be "application/octet-stream")
//
// # func (s *httpService) ReturnJsonBody(object interface{}) *httpService
//
// Add the object as json data to the response (Mime type will be "application/json")
//
// # func (s *httpService) WithTimeouts(count int) *httpService
//
// Induce (count) timeouts before returning the response.
//
// When a timeout occurs, it must be cleared using ReleaseTimeoutHold().
//
// # func (s *httpService) ReleaseTimeoutHold()
//
// Used when testing timeouts to clear an induced timeout so that the next request can be processed.
//
// # func (s *httpService) GetCallCount()
//
// Allows a test to determine how many times the endpoint was called.
//
// # Json Marshaling
//
// # func MustMarshalJson(object interface{}) []byte
//
// Marshals the object into raw (json) data and returns it.
//
// This simplifies generating table-test data by ensuring that any json
// required during your test can be read
//
// If marshaling fails the test will automatically fail, using:
//
//	Expect(err).ToNot(HaveOccurred())
package test

// NOTE:
// We're resorting to documenting things in here because godoc doesn't generate using
// build tags.
