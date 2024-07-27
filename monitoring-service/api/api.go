package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type Router struct {
	Mux    *mux.Router
	Logger kitlog.Logger
}

type Route struct {
	Method   string
	Path     string
	Endpoint endpoint.Endpoint
	Decoder  kithttp.DecodeRequestFunc
	Encoder  kithttp.EncodeResponseFunc
}

func NewRouter(logger kitlog.Logger) Router {
	r := mux.NewRouter()
	r.Use(commonMiddleware)

	return Router{
		Mux:    r,
		Logger: logger,
	}
}

func (router *Router) AddRoutes(routes []Route) {
	for _, r := range routes {
		router.AddRoute(r)
	}
}

func (router *Router) AddRoute(rt Route) {
	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(EncodeError),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(router.Logger)),
	}

	router.Mux.Handle(rt.Path, kithttp.NewServer(
		rt.Endpoint,
		rt.Decoder,
		rt.Encoder,
		options...,
	)).Methods(rt.Method)

	router.Logger.Log("Route", fmt.Sprintf("%s %s", rt.Method, rt.Path))
}

func DecodeEmptyRequest(_ context.Context, _ *http.Request) (request interface{}, err error) {
	return nil, nil
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", fmt.Sprintf("%s, %s, %s", http.MethodGet, http.MethodPost, http.MethodOptions))
		next.ServeHTTP(w, r)
	})
}

type errorWrapperResponse struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	httpCode := http.StatusInternalServerError

	resp := errorWrapperResponse{
		Error: errorResponse{
			Code:    httpCode,
			Message: err.Error(),
		},
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	type body struct {
		Data interface{} `json:"data"`
	}

	wrapped := body{
		Data: response,
	}

	resp, err := json.Marshal(wrapped)
	if err != nil {
		EncodeError(ctx, err, w)
	}

	_, err = w.Write(resp)
	return err
}
