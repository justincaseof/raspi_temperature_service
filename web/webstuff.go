package web

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"raspi_readtemp/logging"
	"raspi_temperature_service/database"
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

	// MEASUREMENTS
	chiRouter.Route("/measurements", func(r chi.Router) {
		r.With(paginate).Get("/", ListMeasurements)
		r.Post("/", CreateMeasurement) // POST /measurements
		//r.Get("/search", SearchMeasurement) //  GET /measurements/search

		r.Route("/{measurementID}", func(r chi.Router) {
			r.Use(MeasurementCtx)      // Load the *Measurement on the request context
			r.Get("/", GetMeasurement) // GET /measurements/123
			//r.Put("/", UpdateMeasurement)    // PUT /measurements/123
			//r.Delete("/", DeleteMeasurement) // DELETE /measurements/123
		})

		// GET /measurements/whats-up
		//r.With(MeasurementCtx).Get("/{measurementSlug:[a-z-]+}", GetMeasurement)
	})

	return nil
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
