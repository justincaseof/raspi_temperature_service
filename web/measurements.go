package web

import (
	"context"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"raspi_temperature_service/model"
)

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
