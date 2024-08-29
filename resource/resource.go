package resource

import (
	"context"
	"errors"
	"net/http"

	"github.com/keithpaterson/resweave-utils/response"
	"github.com/keithpaterson/resweave-utils/utility/rw"
	"github.com/mortedecai/resweave"
)

// Some common errors that resource objects should be using
var (
	ErrNilServer      = errors.New("server is nil")
	ErrNoSuchResource = errors.New("no such resource")
)

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
// Create and List do not require an id
// Fetch, Delete, and Update require an id (as a string) and are expected to validate the id
type EasyResource interface {
	resweave.LogHolder

	// These behave the same as the resweave.APIResource functions.
	Name() resweave.ResourceName
	SetID(id resweave.ID) error
	GetIDValue(ctx context.Context) (string, error)

	// Registers an Easy resource with the resweave server.
	AddEasyResource(s resweave.Server) error
}

type EasyResourceHandler struct {
	resweave.LogHolder
	api resweave.APIResource

	resource        interface{} // the object implementing Create, List, etc.
	acceptedMethods acceptedMethodsMap
	validations     validationFuncMap
}

func NewResource(name resweave.ResourceName, resource interface{}) *EasyResourceHandler {
	arh := &EasyResourceHandler{
		LogHolder:       resweave.NewLogholder(name.String(), nil),
		api:             resweave.NewAPI(name),
		resource:        resource,
		acceptedMethods: make(acceptedMethodsMap),
		validations:     make(validationFuncMap),
	}
	for id := range acceptedMethods {
		arh.setAcceptedMethods(id, nil)
	}

	if lh, ok := resource.(resweave.LogHolder); ok {
		lh.SetLogger(arh.Logger(), false)
	}

	arh.api.SetHandler(arh.handleResourceAction)

	return arh
}

// resweave.APIResource functions that I am exposing as our own

func (arh EasyResourceHandler) Name() resweave.ResourceName {
	return arh.api.Name()
}

func (arh EasyResourceHandler) SetID(id resweave.ID) error {
	return arh.api.SetID(id)
}

func (arh EasyResourceHandler) GetIDValue(ctx context.Context) (string, error) {
	return arh.api.GetIDValue(ctx)
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

func (arh EasyResourceHandler) AddEasyResource(s resweave.Server) error {
	if s == nil {
		return ErrNilServer
	}
	return s.AddResource(arh.api)
}

// By default an Update can be either Put or Patch; both are supported.
//
// You can limit updates to one or the other using this method.
//
// If you do not want to support update at all , use `arh.SetUpdate(nil)`.
//
// If you pass (false, false) here nothing will be changed
func (arh EasyResourceHandler) SetUpdateAcceptedMethods(acceptPut bool, acceptPatch bool) {
	if !acceptPut && !acceptPatch {
		// ignore this;
		return
	}
	accept := methodAcceptance{http.MethodPut: acceptPut, http.MethodPatch: acceptPatch}

	arh.setAcceptedMethods(resweave.Update, accept)
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

func (arh EasyResourceHandler) setAcceptedMethods(at resweave.ActionType, accept methodAcceptance) {
	if accept == nil {
		// revert to defaults - allows caller to easily 'reset' things
		arh.acceptedMethods[at] = acceptedMethods[at]
		return
	}
	arh.acceptedMethods[at] = accept
}

func (arh EasyResourceHandler) handleResourceAction(at resweave.ActionType, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	funcName := "handleResourceAction"
	arh.Infow(funcName, "Status", "Starting")
	defer arh.Infow(funcName, "Status", "Completed")

	// make a writer
	writer := response.NewWriter(w)

	// validations
	if status, err := arh.standardValidations(at, r); err != nil {
		writer.WriteErrorResponse(status, err)
		return
	}
	if validateFn := arh.validations[at]; validateFn != nil {
		if status, err := validateFn(ctx, writer, r); err != nil {
			writer.WriteErrorResponse(status, err)
			return
		}
	}

	// Since we prevalidated whether the function is implemented we know we can call it based on the value of at
	arh.Infow(at.String(), "Status", "Starting")
	defer arh.Infow(at.String(), "Status", "Completed")

	// try to call non-id methods first:
	switch at {
	case resweave.Create:
		arh.resource.(easyCreator).Create(ctx, writer, r)
		return
	case resweave.List:
		arh.resource.(easyLister).List(ctx, writer, r)
		return
	}

	// try to call id methods with an error if id is missing
	id, err := arh.api.GetIDValue(ctx)
	if err != nil {
		writer.WriteErrorResponse(http.StatusBadRequest, response.SvcErrorInvalidResourceId)
		return
	}
	switch at {
	case resweave.Fetch:
		arh.resource.(easyFetcher).Fetch(id, ctx, writer, r)
		return
	case resweave.Delete:
		arh.resource.(easyDeleter).Delete(id, ctx, writer, r)
		return
	case resweave.Update:
		arh.resource.(easyUpdater).Update(id, ctx, writer, r)
		return
	}
}

func (arh EasyResourceHandler) standardValidations(at resweave.ActionType, r *http.Request) (int, response.ServiceError) {
	// for now we don't need context or writer, but as we add more common validations we can add them in

	if !arh.validateActionImplemented(at) {
		arh.Errorw("standardValidations", "action", arh.api.Name().String(), "reason", "not-implemented")
		return http.StatusMethodNotAllowed, response.SvcErrorNoRegisteredMethod
	}

	if !arh.validateAcceptedMethods(at, r.Method) {
		arh.Errorw("standardValidations", "resource", arh.api.Name().String(), "bad-method", r.Method, "accepted-methods", acceptedMethods[at])
		return http.StatusMethodNotAllowed, response.SvcErrorInvalidMethod.WithDetail(r.Method)
	}

	// if everything is considered ok then just return no error; status code isn't important in the success case
	return 0, nil
}

func (arh EasyResourceHandler) validateAcceptedMethods(at resweave.ActionType, method string) bool {
	_, found := arh.acceptedMethods[at][method]
	return found
}

func (arh EasyResourceHandler) validateActionImplemented(at resweave.ActionType) bool {
	if arh.resource == nil {
		return false
	}

	found := false
	switch at {
	case resweave.Create:
		_, found = arh.resource.(easyCreator)
	case resweave.List:
		_, found = arh.resource.(easyLister)
	case resweave.Fetch:
		_, found = arh.resource.(easyFetcher)
	case resweave.Delete:
		_, found = arh.resource.(easyDeleter)
	case resweave.Update:
		_, found = arh.resource.(easyUpdater)
	}

	return found
}
