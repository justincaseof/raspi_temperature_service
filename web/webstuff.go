package web

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
	"raspi_readtemp/logging"
	"time"
)

var apiRequests uint64
var startupTime time.Time = time.Now()
var logger = logging.New("raspi_temperature_service_web", false)

func SetupChi(chiRouter *chi.Mux) error {
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
	//chiRouter.Route("/measurements", func(r chi.Router) {
	//	r.With(paginate).Get("/", ListMeasurements)
	//	r.Post("/", CreateMeasurement)       // POST /measurements
	//	r.Get("/search", SearchMeasurement)  //  GET /measurements/search
	//
	//	r.Route("/{measurementID}", func(r chi.Router) {
	//		r.Use(MeasurementCtx)            		 // Load the *Article on the request context
	//		r.Get("/", GetMeasurement)       // GET /measurements/123
	//		r.Put("/", UpdateMeasurement)    // PUT /measurements/123
	//		r.Delete("/", DeleteMeasurement) // DELETE /measurements/123
	//	})
	//
	//	// GET /articles/whats-up
	//	r.With(ArticleCtx).Get("/{articleSlug:[a-z-]+}", GetMeasurement)
	//})

	return nil
}

// HEALTH
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
// / HEALTH

// MeasurementCtx middleware is used to load measurement object from
// the URL parameters passed through as the request. In case
// the measurement could not be found, we stop here and return a 404.
//func ArticleCtx(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		var article *Meas
//		var err error
//
//		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
//			article, err = dbGetArticle(articleID)
//		} else if articleSlug := chi.URLParam(r, "articleSlug"); articleSlug != "" {
//			article, err = dbGetArticleBySlug(articleSlug)
//		} else {
//			render.Render(w, r, ErrNotFound)
//			return
//		}
//		if err != nil {
//			render.Render(w, r, ErrNotFound)
//			return
//		}
//
//		ctx := context.WithValue(r.Context(), "article", article)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}

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
