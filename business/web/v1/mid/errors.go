package mid

import (
	"context"
	"net/http"

	"github.com/go-olive/olive/business/sys/validate"
	v1Web "github.com/go-olive/olive/business/web/v1"
	"github.com/go-olive/olive/foundation/web"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated error.
			if err := handler(ctx, w, r); err != nil {

				// Log the error.
				log.Errorw("ERROR", "traceid", v.TraceID, "message", err)

				// Build out the error response.
				var er v1Web.ErrorResponse
				var status int
				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = v1Web.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case v1Web.IsRequestError(err):
					reqErr := v1Web.GetRequestError(err)
					er = v1Web.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = v1Web.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Respond with the error back to the client.
				if err := Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}

			// The error has been handled so we can stop propagating it.
			return nil
		}

		return h
	}

	return m
}

const portalSuccessCode = "0000"

type portalResponse struct {
	Code    string      `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {

	// Set the status code for the request logger middleware.
	web.SetStatusCode(ctx, statusCode)

	var payload portalResponse
	switch statusCode {
	case http.StatusOK,
		http.StatusCreated:
		payload = portalResponse{
			Code: portalSuccessCode,
			Data: data,
		}
	default:
		errMsg, _ := jsoniter.MarshalToString(data)
		payload = portalResponse{
			Message: errMsg,
		}
	}

	// Convert the response value to JSON.
	jsonData, err := jsoniter.Marshal(payload)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
