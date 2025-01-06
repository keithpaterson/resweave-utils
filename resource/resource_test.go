package resource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/keithpaterson/resweave-utils/mocks"
	"github.com/keithpaterson/resweave-utils/response"
	"github.com/keithpaterson/resweave-utils/utility/rw"
	"github.com/mortedecai/resweave"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

type callRecord struct {
	at     resweave.ActionType
	method string
	id     string
}
type testEasyResource struct {
	resweave.LogHolder

	// updated when methods are called.
	calls []callRecord
}

func newTestEasyResource() *EasyResourceHandler {
	ter := &testEasyResource{
		LogHolder: resweave.NewLogholder("test", nil),
		calls:     make([]callRecord, 0),
	}

	return NewResource("test", ter)
}

// make action handlers that don't really do anything
func (ter *testEasyResource) Create(_ context.Context, writer response.Writer, req *http.Request) {
	ter.calls = append(ter.calls, callRecord{at: resweave.Create, method: req.Method})
	writer.WriteResponse(http.StatusOK)
}
func (ter *testEasyResource) List(_ context.Context, writer response.Writer, req *http.Request) {
	ter.calls = append(ter.calls, callRecord{at: resweave.List, method: req.Method})
	writer.WriteResponse(http.StatusOK)
}
func (ter *testEasyResource) Fetch(id string, _ context.Context, writer response.Writer, req *http.Request) {
	ter.calls = append(ter.calls, callRecord{at: resweave.Fetch, method: req.Method, id: id})
	writer.WriteResponse(http.StatusOK)
}
func (ter *testEasyResource) Delete(id string, _ context.Context, writer response.Writer, req *http.Request) {
	ter.calls = append(ter.calls, callRecord{at: resweave.Delete, method: req.Method, id: id})
	writer.WriteResponse(http.StatusOK)
}
func (ter *testEasyResource) Update(id string, _ context.Context, writer response.Writer, req *http.Request) {
	ter.calls = append(ter.calls, callRecord{at: resweave.Update, method: req.Method, id: id})
	writer.WriteResponse(http.StatusOK)
}

var _ = Describe("Test EasyResource", func() {
	Context("Accepted Methods", func() {
		DescribeTable("Defaults",
			func(action resweave.ActionType, accept methodAcceptance) {
				// Arrange
				erh := NewResource("test", nil)

				// Act && Assert
				for key := range accept {
					actual := erh.validateAcceptedMethods(action, key)
					Expect(actual).To(Equal(accept[key]))
				}
			},
			Entry(resweave.Create.String(), resweave.Create,
				methodAcceptance{http.MethodGet: false, http.MethodPut: false, http.MethodPost: true, http.MethodPatch: false, http.MethodDelete: false}),
			Entry(resweave.List.String(), resweave.List,
				methodAcceptance{http.MethodGet: true, http.MethodPut: false, http.MethodPost: false, http.MethodPatch: false, http.MethodDelete: false}),
			Entry(resweave.Fetch.String(), resweave.Fetch,
				methodAcceptance{http.MethodGet: true, http.MethodPut: false, http.MethodPost: false, http.MethodPatch: false, http.MethodDelete: false}),
			Entry(resweave.Delete.String(), resweave.Delete,
				methodAcceptance{http.MethodGet: false, http.MethodPut: false, http.MethodPost: false, http.MethodPatch: false, http.MethodDelete: true}),
			Entry(resweave.Update.String(), resweave.Update,
				methodAcceptance{http.MethodGet: false, http.MethodPut: true, http.MethodPost: false, http.MethodPatch: true, http.MethodDelete: false}),
		)
		DescribeTable("Update Methods",
			func(acceptPut bool, acceptPatch bool) {
				// Arrange
				erh := NewResource("test", nil)

				// Act
				erh.SetUpdateAcceptedMethods(acceptPut, acceptPatch)

				// Assert
				// expect no change if both put and patch are false.  Assumes 'true' is the default
				expectPut := (!acceptPut && !acceptPatch) || acceptPut
				expectPatch := (!acceptPut && !acceptPatch) || acceptPatch
				Expect(erh.acceptedMethods[resweave.Update][http.MethodPut]).To(Equal(expectPut))
				Expect(erh.acceptedMethods[resweave.Update][http.MethodPatch]).To(Equal(expectPatch))
			},
			Entry(nil, true, true),
			Entry(nil, true, false),
			Entry(nil, false, true),
			Entry(nil, false, false),
		)
		It("should reset acceptance to defaults if set to nil", func() {
			// Arrange
			erh := NewResource("test", nil)
			erh.SetUpdateAcceptedMethods(true, false)

			Expect(erh.acceptedMethods[resweave.Update]).To(Equal(methodAcceptance{http.MethodPatch: false, http.MethodPut: true}))

			// Act
			erh.setAcceptedMethods(resweave.Update, nil)

			// Assert
			Expect(erh.acceptedMethods[resweave.Update]).To(Equal(methodAcceptance{http.MethodPatch: true, http.MethodPut: true}))
		})
	})

	Context("EasyResourceHandler", func() {
		var (
			erh *EasyResourceHandler
		)
		BeforeEach(func() {
			erh = NewResource("test", nil)
		})

		It("should return the correct name", func() {
			Expect(erh.Name()).To(Equal(resweave.ResourceName("test")))
			Expect(erh.Name()).To(Equal(erh.api.Name()))
		})
		DescribeTable("GetIDValue works properly",
			func(key string, id string, expectID string, expectErr error) {
				// Arrange
				erh.SetID(resweave.NumericID)
				ctx := context.WithValue(context.TODO(), resweave.Key(key), id)

				// Act
				value, err := erh.GetIDValue(ctx)

				// Assert
				if expectErr != nil {
					Expect(err).To(MatchError(expectErr))
					Expect(value).To(BeEmpty())
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(value).To(Equal(expectID))
				}
			},
			Entry("with valid id returns id", "id_test", "123", "123", nil),
			Entry("with invalid id returns error", "id_mattress", "123", "", resweave.ErrIDNotFound),
		)
	})

	Context("AddEasyResource", func() {
		It("should return error if server is nil", func() {
			// Arrange
			handler := NewResource("test", newTestEasyResource())

			// Act
			err := handler.AddEasyResource(nil)

			// Assert
			Expect(err).To(Equal(ErrNilServer))
		})
		It("should return nil if server is not nil", func() {
			// Arrange
			ctrl := gomock.NewController(GinkgoT())
			server := mocks.NewMockServer(ctrl)
			handler := NewResource("test", newTestEasyResource())

			server.EXPECT().AddResource(handler.api).Return(nil)

			// Act
			err := handler.AddEasyResource(server)

			// Assert
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Request Handling", func() {
		var (
			recorder *httptest.ResponseRecorder
			ctx      context.Context
			erh      *EasyResourceHandler // has no resource
			res      *EasyResourceHandler // is the test resource
		)
		BeforeEach(func() {
			recorder = httptest.NewRecorder()
			ctx = context.TODO()

			erh = NewResource("test", nil)
			res = newTestEasyResource()
		})

		DescribeTable("unset handlers should return error",
			func(at resweave.ActionType, method string) {
				// Arrange &&  Act
				req, err := http.NewRequest(method, "test/", nil)
				Expect(err).ToNot(HaveOccurred())
				erh.handleResourceAction(at, ctx, recorder, req)

				// Assert
				resp := recorder.Result()
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))

				var svcErr response.SvcError
				err = rw.UnmarshalJson(resp.Body, &svcErr)
				Expect(err).ToNot(HaveOccurred())
				Expect(&svcErr).To(MatchError(response.SvcErrorNoRegisteredMethod))
			},
			Entry(resweave.Create.String(), resweave.Create, http.MethodPost),
			Entry(resweave.List.String(), resweave.List, http.MethodGet),
			Entry(resweave.Fetch.String(), resweave.Fetch, http.MethodGet),
			Entry(resweave.Delete.String(), resweave.Delete, http.MethodDelete),
			Entry(resweave.Update.String(), resweave.Update, http.MethodPut),
		)

		DescribeTable("should invoke action handlers when they are set",
			func(at resweave.ActionType, method string, id string) {
				// Arrange && Act
				req, err := http.NewRequest(method, "/test", nil)
				Expect(err).ToNot(HaveOccurred())

				ctx := context.WithValue(ctx, resweave.Key("id_test"), id)

				// Act
				res.handleResourceAction(at, ctx, recorder, req)

				// Assert
				resp := recorder.Result()
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(res.resource.(*testEasyResource).calls).To(Equal([]callRecord{{at: at, method: method, id: id}}))
			},
			Entry(resweave.Create.String(), resweave.Create, http.MethodPost, ""),
			Entry(resweave.List.String(), resweave.List, http.MethodGet, ""),
			Entry(resweave.Fetch.String(), resweave.Fetch, http.MethodGet, "1"),
			Entry(resweave.Delete.String(), resweave.Delete, http.MethodDelete, "1"),
			Entry(resweave.Update.String(), resweave.Update, http.MethodPut, "1"),
		)
		It("should respond with an error if ID is invalid", func() {
			// Arrange
			req, err := http.NewRequest(http.MethodDelete, "/test/1", nil)
			Expect(err).ToNot(HaveOccurred())

			ctx := context.WithValue(ctx, resweave.Key("id_incorrect"), "1")

			// Act
			res.handleResourceAction(resweave.Delete, ctx, recorder, req)

			// Assert
			resp := recorder.Result()
			defer resp.Body.Close()

			err = response.ParseResponse(resp, http.StatusNoContent)

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(res.resource.(*testEasyResource).calls).To(BeEmpty())
			Expect(err).To(MatchError(response.SvcErrorInvalidResourceId))
		})
		It("should respond with an error if standard validations fail", func() {
			ctx := context.WithValue(ctx, resweave.Key("id_test"), "1")

			req, err := http.NewRequest(http.MethodDelete, "/test/1", nil)
			Expect(err).ToNot(HaveOccurred())

			// Act
			res.handleResourceAction(resweave.Create, ctx, recorder, req)

			// Assert
			resp := recorder.Result()
			defer resp.Body.Close()

			err = response.ParseResponse(resp, http.StatusNoContent)

			Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
			Expect(res.resource.(*testEasyResource).calls).To(BeEmpty())
			Expect(err).To(MatchError(response.SvcErrorInvalidMethod))
		})

		type validatorResponse struct {
			status int
			err    response.ServiceError
		}
		DescribeTable("should call custom validator if not nil",
			func(val validatorResponse, expectHandled bool) {
				// Arrange
				validated := false
				validator := func(_ context.Context, writer response.Writer, req *http.Request) (int, response.ServiceError) {
					validated = true
					return val.status, val.err
				}
				res.validations[resweave.List] = validator

				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				Expect(err).ToNot(HaveOccurred())

				// Act
				res.handleResourceAction(resweave.List, ctx, recorder, req)

				// Assert
				Expect(validated).To(BeTrue())

				resp := recorder.Result()
				resource := res.resource.(*testEasyResource)
				if expectHandled {
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(resource.calls).To(Equal([]callRecord{{at: resweave.List, method: http.MethodGet}}))
				} else {
					Expect(resp.StatusCode).To(Equal(val.status))
					Expect(resource.calls).To(BeEmpty())

					var svcErr response.SvcError
					err := rw.UnmarshalJson(resp.Body, &svcErr)
					Expect(err).ToNot(HaveOccurred())
					Expect(&svcErr).To(MatchError(val.err))
				}
			},
			Entry("validator succeeds", validatorResponse{0, nil}, true),
			Entry("validator returns error", validatorResponse{http.StatusBadRequest, response.SvcErrorInvalidResourceId}, false),
		)
	})

	Context("resweave Service Integration", func() {
		var (
			srv      resweave.Server
			recorder *httptest.ResponseRecorder
			res      *EasyResourceHandler

			port = 8080
		)
		BeforeEach(func() {
			srv = resweave.NewServer(port)
			recorder = httptest.NewRecorder()

			res = newTestEasyResource()
			res.SetID(resweave.NumericID)
		})
		DescribeTable("should invoke the correct handler",
			func(method string, action resweave.ActionType, id string) {
				// Arrange
				path := "/test"
				if id != "" {
					path = fmt.Sprintf("%s/%s", path, id)
				}
				req, err := http.NewRequest(method, path, nil)
				Expect(err).ToNot(HaveOccurred())

				// Act
				res.AddEasyResource(srv)
				srv.Serve(recorder, req)

				// Assert
				Expect(res.resource.(*testEasyResource).calls).To(Equal([]callRecord{{at: action, method: method, id: id}}))
			},
			Entry(nil, http.MethodPost, resweave.Create, ""),
			Entry(nil, http.MethodGet, resweave.List, ""),
			Entry(nil, http.MethodGet, resweave.Fetch, "1"),
			Entry(nil, http.MethodDelete, resweave.Delete, "1"),
			Entry(nil, http.MethodPut, resweave.Update, "1"),
			Entry(nil, http.MethodPatch, resweave.Update, "2"),
		)
	})

	DescribeTable("Test ToServiceError",
		func(err error, expect response.ServiceError) {
			// Arrange & Act
			actual := ToServiceError(err)

			// Assert
			Expect(actual).To(Equal(expect))
		},
		Entry(nil, rw.ErrorNoData, response.SvcErrorReadRequestFailed.WithError(rw.ErrorNoData)),
		Entry(nil, rw.ErrorJsonUnmarshalFailed, response.SvcErrorJsonUnmarshalFailed.WithError(rw.ErrorNoData)),
		Entry(nil, errors.New("default"), response.SvcErrorReadRequestFailed.WithError(errors.New("default"))),
	)
})
