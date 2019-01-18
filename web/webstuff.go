package web

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
	"raspi_readtemp/logging"
	"raspi_temperature_service/database"
	"raspi_temperature_service/model"
	"time"
)

var apiRequests uint64
var startupTime time.Time = time.Now()
var logger = logging.New("raspi_temperature_service_web", false)
var measurementRepo database.IMeasurementRepository

func SetupChi(chiRouter *chi.Mux) error {
	// initially, setup repo. ATTENTION: this requires an already initialized XORM engine
	measurementRepo = database.NewMeasurementRepository()

	chiRouter.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// count requests
			apiRequests++
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	chiRouter.Use(render.SetContentType(render.ContentTypeJSON))

	// HEALTH
	chiRouter.Get("/health", handleHealthRequest)

	// RESTy routes for "measurements" resource
	chiRouter.Route("/measurements", func(r chi.Router) {
		r.With(paginate).Get("/", ListMeasurements)
		r.Post("/", CreateMeasurement)      // POST /measurements
		//r.Get("/search", SearchMeasurement) //  GET /measurements/search

		r.Route("/{measurementID}", func(r chi.Router) {
			r.Use(MeasurementCtx)            // Load the *Article on the request context
			r.Get("/", GetMeasurement)       // GET /measurements/123
			//r.Put("/", UpdateMeasurement)    // PUT /measurements/123
			//r.Delete("/", DeleteMeasurement) // DELETE /measurements/123
		})

		// GET /measurements/whats-up
		//r.With(MeasurementCtx).Get("/{articleSlug:[a-z-]+}", GetMeasurement)
	})

	return nil
}

// === HEALTH ===
func handleHealthRequest(writer http.ResponseWriter, request *http.Request) {
	b, err := json.Marshal(
		Health{
			APIRequests:   apiRequests,
			UptimeSeconds: uint64(time.Since(startupTime).Seconds()),
		},
	)

	if err != nil {
		logger.Error("Cannot marshall data", zap.String("err", err.Error()))
		http.Error(writer, err.Error(), 500)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.Write(b)
}

// === /HEALTH ===

// MeasurementCtx middleware is used to load measurement object from
// the URL parameters passed through as the request. In case
// the measurement could not be found, we stop here and return a 404.
func MeasurementCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var measurement *model.Measurement
		var err error

		if measurementID := chi.URLParam(r, "measurementID"); measurementID != "" {
			err, measurement  = measurementRepo.FindMeasurementByID(measurementID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "measurement", measurement)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateArticle persists the posted Article and returns it
// back to the client as an acknowledgement.
func CreateMeasurement(w http.ResponseWriter, r *http.Request) {
	data := &MeasurementRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	measurement := data.Measurement
	measurementRepo.InsertMeasurement(measurement.Value, measurement.Unit, measurement.InstanceId)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewMeasurementResponse(measurement))
}

func ListMeasurements(w http.ResponseWriter, r *http.Request) {
	dbErr, measurements := measurementRepo.FindAllMeasurements()
	if dbErr != nil {
		render.Render(w, r, ErrRender(dbErr))
		return
	}

	if err := render.RenderList(w, r, NewMeasurementsListResponse(measurements)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetMeasurement returns the specific Measurement. You'll notice it just
// fetches the Measurement right off the context, as its understood that
// if we made it this far, the Measurement must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func GetMeasurement(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the ArticleCtx
	// middleware. The worst case, the recoverer middleware will save us.
	measurement := r.Context().Value("measurement").(*model.Measurement)

	if err := render.Render(w, r, NewMeasurementResponse(measurement)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}


// ArticleRequest is the request payload for Article data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type MeasurementRequest struct {
	*model.Measurement

	//User *UserPayload `json:"user,omitempty"`

	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *MeasurementRequest) Bind(r *http.Request) error {
	// a.Article is nil if no Article fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	if a.Measurement == nil {
		return errors.New("missing required Measurement fields.")
	}

	// a.User is nil if no Userpayload fields are sent in the request. In this app
	// this won't cause a panic, but checks in this Bind method may be required if
	// a.User or futher nested fields like a.User.Name are accessed elsewhere.

	// just a post-process after a decode..
	a.ProtectedID = ""                                 // unset the protected ID

	return nil
}

// ArticleResponse is the response payload for the Article data model.
// See NOTE above in ArticleRequest as well.
//
// In the ArticleResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type MeasurementResponse struct {
	*model.Measurement

	//User *UserPayload `json:"user,omitempty"`

	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func NewMeasurementResponse(measurement *model.Measurement) *MeasurementResponse {
	resp := &MeasurementResponse{Measurement: measurement}

	//if resp.User == nil {
	//	if user, _ := dbGetUser(resp.UserID); user != nil {
	//		resp.User = NewUserPayloadResponse(user)
	//	}
	//}

	return resp
}

func (resp *MeasurementResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	resp.Elapsed = 10
	return nil
}

type MeasurementsListResponse []*MeasurementResponse
func NewMeasurementsListResponse(measurements []*model.Measurement) []render.Renderer {
	list := []render.Renderer{}
	for _, measurement := range measurements {
		list = append(list, NewMeasurementResponse(measurement))
	}
	return list
}


//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

// paginate is a stub, but very possible to implement middleware logic
// to handle the request params for handling a paginated request.
func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

// ==== MODEL ===
type Health struct {
	UptimeSeconds uint64 `json:"uptimeSeconds"`
	APIRequests   uint64 `json:"apiRequests"`
}
