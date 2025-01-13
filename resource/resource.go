package resource

import (
	"context"
	"errors"
	"net/http"

	"github.com/keithpaterson/resweave-utils/logging"
	"github.com/keithpaterson/resweave-utils/response"
	"github.com/keithpaterson/resweave-utils/utility/rw"
	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

// Some common errors that resource objects should be using
var (
	ErrNilServer      = errors.New("server is nil")
	ErrNoSuchResource = errors.New("no such resource")
)

// ToServiceError converts an error type to a [response.ServiceError] that can be
// returned in your HTTP response.
func ToServiceError(err error) response.ServiceError {
	// Brute-force but a lookup table and Unwrap wasn't working so ...
	if errors.Is(err, rw.ErrorNoData) {
		return response.SvcErrorReadRequestFailed.WithError(err)
	}
	if errors.Is(err, rw.ErrorJsonUnmarshalFailed) {
		return response.SvcErrorJsonUnmarshalFailed.WithError(err)
	}
	if errors.Is(err, ErrNoSuchResource) {
		return response.SvcErrorInvalidResourceId
	}
	// for lack of a better default ...
	return response.SvcErrorReadRequestFailed.WithError(err)
}

// tag::easyresource[]

// Easy Resources must implement a subset of these functions in order to be valid:
//
//	Create(context.Context, response.Writer, *http.Request)
//	List(context.Context, response.Writer, *http.Request)
//	Fetch(string, context.Context, response.Writer, *http.Request)
//	Delete(string, context.Context, response.Writer, *http.Request)
//	Update(string, context.Context, response.Writer, *http.Request)
//
// If any method is not implemented the resource's action handler will automatically
// return an error response indicating that the action is invalid.
//
// Create and List do not require an id.
//
// Fetch, Delete, and Update require an id (as a string) and are expected to validate the id.
type EasyResource interface {
	resweave.LogHolder
}

// end::easyresource[]

// Instantiates a resource handler that can be registered with the resweave.Service.
//
// When RESTful requests are received via the service, the handler takes care of calling
// your REST functions as appropriate.
func NewResource(name resweave.ResourceName, resource EasyResource) *EasyResourceHandler {
	lh := resweave.NewLogholder(name.String(), func(logger *zap.SugaredLogger) {
		if resource != nil {
			resource.SetLogger(logger, false)
		}
	})
	erh := &EasyResourceHandler{
		LogHolder:       lh,
		LogFactory:      logging.LogFactory{LogHolder: lh},
		api:             resweave.NewAPI(name),
		resource:        resource,
		acceptedMethods: make(acceptedMethodsMap),
		validations:     make(validationFuncMap),
	}
	for id := range acceptedMethods {
		erh.setAcceptedMethods(id, nil)
	}

	lh.SetLogger(erh.Logger(), false)

	erh.api.SetHandler(erh.handleResourceAction)

	return erh
}

type EasyResourceHandler struct {
	resweave.LogHolder
	logging.LogFactory
	api resweave.APIResource

	resource        EasyResource // the object implementing Create, List, etc.
	acceptedMethods acceptedMethodsMap
	validations     validationFuncMap
}

// resweave.APIResource functions that I am exposing as our own

func (erh EasyResourceHandler) Name() resweave.ResourceName {
	return erh.api.Name()
}

func (erh EasyResourceHandler) SetID(id resweave.ID) error {
	return erh.api.SetID(id)
}

func (erh EasyResourceHandler) GetIDValue(ctx context.Context) (string, error) {
	return erh.api.GetIDValue(ctx)
}

// end resweave.APIResource function equivalents

type easyCreator interface {
	Create(ctx context.Context, writer response.Writer, req *http.Request)
}

type easyLister interface {
	List(ctx context.Context, writer response.Writer, req *http.Request)
}

type easyFetcher interface {
	Fetch(id string, ctx context.Context, writer response.Writer, req *http.Request)
}

type easyDeleter interface {
	Delete(id string, ctx context.Context, writer response.Writer, req *http.Request)
}

type easyUpdater interface {
	Update(id string, ctx context.Context, writer response.Writer, req *http.Request)
}

// Registers a resource handler created using [NewResource] with the resweave server.
func (erh EasyResourceHandler) AddEasyResource(s resweave.Server) error {
	if s == nil {
		return ErrNilServer
	}
	return s.AddResource(erh.api)
}

// By default an Update can be either Put or Patch; both are supported.
//
// You can limit updates to one or the other using this method.
//
// If you do not want to support update at all, do not implement an Update function for your resource`.
//
// If you pass (false, false) here nothing will be changed
func (erh EasyResourceHandler) SetUpdateAcceptedMethods(acceptPut bool, acceptPatch bool) {
	if !acceptPut && !acceptPatch {
		// ignore this;
		return
	}
	accept := methodAcceptance{http.MethodPut: acceptPut, http.MethodPatch: acceptPatch}

	erh.setAcceptedMethods(resweave.Update, accept)
}

type methodAcceptance map[string]bool
type acceptedMethodsMap map[resweave.ActionType]methodAcceptance

// internal defaults
var (
	// default valid http methods for various resource actions
	acceptedMethods = acceptedMethodsMap{
		resweave.Create: {http.MethodPost: true},
		resweave.List:   {http.MethodGet: true},
		resweave.Fetch:  {http.MethodGet: true},
		resweave.Delete: {http.MethodDelete: true},
		resweave.Update: {http.MethodPatch: true, http.MethodPut: true},
	}
)

type validationFuncMap map[resweave.ActionType]easyValidateFunc

// used internally as a means to perform validations/processing specific to a particular
// resource action.
//
// returns an http status code and an optional service error if something went wrong
//
// status code won't be used except when an error is being reported.
type easyValidateFunc func(context.Context, response.Writer, *http.Request) (int, response.ServiceError)

func (erh EasyResourceHandler) setAcceptedMethods(at resweave.ActionType, accept methodAcceptance) {
	if accept == nil {
		// revert to defaults - allows caller to easily 'reset' things
		erh.acceptedMethods[at] = acceptedMethods[at]
		return
	}
	erh.acceptedMethods[at] = accept
}

func (erh EasyResourceHandler) handleResourceAction(at resweave.ActionType, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	funcName := "handleResourceAction"
	erh.NewInfo(funcName, "Starting").Log()
	defer erh.NewInfo(funcName, "Completed").Log()

	// make a writer
	writer := response.NewWriter(w)

	// validations
	if status, err := erh.standardValidations(at, r); err != nil {
		writer.WriteErrorResponse(status, err)
		return
	}
	if validateFn := erh.validations[at]; validateFn != nil {
		if status, err := validateFn(ctx, writer, r); err != nil {
			writer.WriteErrorResponse(status, err)
			return
		}
	}

	// Since we prevalidated whether the function is implemented we know we can call it based on the value of at
	erh.NewInfo(at.String(), "Starting").Log()
	defer erh.NewInfo(at.String(), "Completed").Log()

	// try to call non-id methods first:
	switch at {
	case resweave.Create:
		erh.resource.(easyCreator).Create(ctx, writer, r)
		return
	case resweave.List:
		erh.resource.(easyLister).List(ctx, writer, r)
		return
	}

	// try to call id methods with an error if id is missing
	id, err := erh.api.GetIDValue(ctx)
	if err != nil {
		writer.WriteErrorResponse(http.StatusBadRequest, response.SvcErrorInvalidResourceId)
		return
	}
	switch at {
	case resweave.Fetch:
		erh.resource.(easyFetcher).Fetch(id, ctx, writer, r)
		return
	case resweave.Delete:
		erh.resource.(easyDeleter).Delete(id, ctx, writer, r)
		return
	case resweave.Update:
		erh.resource.(easyUpdater).Update(id, ctx, writer, r)
		return
	}
}

func (erh EasyResourceHandler) standardValidations(at resweave.ActionType, r *http.Request) (int, response.ServiceError) {
	// for now we don't need context or writer, but as we add more common validations we can add them in
	funcName := "standardValidations"

	if !erh.validateActionImplemented(at) {
		erh.NewError(funcName, errors.New("not implemented")).With("action", erh.api.Name()).Log()
		return http.StatusMethodNotAllowed, response.SvcErrorNoRegisteredMethod
	}

	if !erh.validateAcceptedMethods(at, r.Method) {
		erh.NewErrorMessage(funcName, errors.New("bad method"), r.Method).WithResource(erh.api.Name()).With("Accepted Methods", acceptedMethods[at]).Log()
		return http.StatusMethodNotAllowed, response.SvcErrorInvalidMethod.WithDetail(r.Method)
	}

	// if everything is considered ok then just return no error; status code isn't important in the success case
	return 0, nil
}

func (erh EasyResourceHandler) validateAcceptedMethods(at resweave.ActionType, method string) bool {
	_, found := erh.acceptedMethods[at][method]
	return found
}

func (erh EasyResourceHandler) validateActionImplemented(at resweave.ActionType) bool {
	if erh.resource == nil {
		return false
	}

	found := false
	switch at {
	case resweave.Create:
		_, found = erh.resource.(easyCreator)
	case resweave.List:
		_, found = erh.resource.(easyLister)
	case resweave.Fetch:
		_, found = erh.resource.(easyFetcher)
	case resweave.Delete:
		_, found = erh.resource.(easyDeleter)
	case resweave.Update:
		_, found = erh.resource.(easyUpdater)
	}

	return found
}
