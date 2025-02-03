/*
Pcakge resource simplifies the implementation of a [mortecai/reswave] compatible resource.

Example:

Implement a resource that handles Create and List and automatically returns HTTP error codes for
FETCH, UPDATE, and DELETE

	// file: foo_resource.go

	type Foo struct {
	    ID: int      `json:"id"`
	    Name: string `json:"name"`
	}

	type FooResource struct {
	    resweave.LogHolder

	    foos   []MyData
	    nextID int
	    mtx    sync.Mutex
	}

	func AddResource(server resweave.Server) {
	    res := resource.NewResource("foo", &FooResource {
	        LogHolder: resweave.NewLogholder("foo", nil),
	        foos: make([]Foo, 0),
	    })
	    res.SetID(resweave.NumericID)
	    return res.AddEasyResource(server)
	}

	func (fr *FooResource) List(_ context.Context, writer resource.ResponseWriter, _ *http.Request) {
	    fr.mtx.Lock()
	    defer fr.mtx.Unlock()

	    if err := writer.WriteJsonResponse(http.StatusOK, fr.foos); err != nil {
	        fr.NewError("List", err).Log()
	    }
	}

	func (fr *FooResource) Create(_context.Context, writer resource.ResponseWriter, req *http.Request) {
	    var input Foo
	    if err := rw.UnmarshalJson(req.Body, &input); err != nil {
	        fr.NewError("Create", err).Log()
	        writer.WriteErrorResponse(http.StatusBadRequest, response.SvrErrorReadRequestFailed.WithError(err))
	    }

	    fr.mtx.Lock()
	    defer fr.mtx.Unlock()

	    input.ID = fr.nextID
	    fr.nextID++

	    fr.foos = append(fr.foos, foo)
	    if err := writerWriteJsonResponse(http.StatusOK, foo); err != nil {
	        fr.NewError("Create", err).Log()
	    }
	}
*/
package resource
