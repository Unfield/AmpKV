package server

import (
	"fmt"
	"time"

	"github.com/Unfield/AmpKV/pkg/common"
	"github.com/Unfield/AmpKV/pkg/embedded"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

	"net/http"
)

type AmpKVHttpServer struct {
	e     *echo.Echo
	store *embedded.AmpKV
}

func NewAmpKVHttpServer(store *embedded.AmpKV) *AmpKVHttpServer {
	server := &AmpKVHttpServer{
		e:     echo.New(),
		store: store,
	}

	server.e.Use(middleware.Recover())
	server.e.Use(middleware.Logger())

	server.e.GET("/api/v1/:key", server.handleGet())
	server.e.POST("/api/v1/", server.handleSet())
	server.e.DELETE("/api/v1/:key", server.handleDelete())

	return server
}

func (s *AmpKVHttpServer) ListenAutoTLS(address string, port uint16) {
	s.e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	go s.e.Logger.Fatal(s.e.StartAutoTLS(fmt.Sprintf("%s:%d", address, port)))
}

func (s *AmpKVHttpServer) Listen(address string, port uint16) {
	go s.e.Logger.Fatal(s.e.Start(fmt.Sprintf("%s:%d", address, port)))
}

func (s *AmpKVHttpServer) Use(mw echo.MiddlewareFunc) {
	s.e.Use(mw)
}

type getSuccessResponse struct {
	Error bool                 `json:"error"`
	Type  common.AmpKVDataType `json:"type"`
	Value []byte               `json:"value"`
}

func (s *AmpKVHttpServer) handleGet() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		key := ctx.Param("key")

		if len(key) < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "key is required")
		}

		val, found := s.store.Get(key)
		if !found || val == nil {
			return echo.NewHTTPError(http.StatusNotFound, "key not found")
		}

		return ctx.JSON(http.StatusOK, getSuccessResponse{Error: false, Type: val.Type, Value: val.Data})
	}
}

type setRequest struct {
	Key   string         `json:"key"`
	Value any            `json:"value"`
	TTL   *time.Duration `json:"ttl"`
}

type setSuccessResponse struct {
	Error bool `json:"error"`
}

func (s *AmpKVHttpServer) handleSet() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var request setRequest
		err := ctx.Bind(&request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "body malformed")
		}

		if len(request.Key) < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "key is required")
		}

		if request.TTL != nil && *request.TTL > 0 {
			err := s.store.SetWithTTL(request.Key, request.Value, 1, *request.TTL)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to save data")
			}
			return ctx.JSON(http.StatusCreated, setSuccessResponse{Error: false})
		} else {
			err := s.store.Set(request.Key, request.Value, 1)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to save data")
			}
			return ctx.JSON(http.StatusCreated, setSuccessResponse{Error: false})
		}
	}
}

type deleteSuccessResponse struct {
	Error bool `json:"error"`
}

func (s *AmpKVHttpServer) handleDelete() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		key := ctx.Param("key")

		if len(key) < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "key is required")
		}

		s.store.Delete(key)

		return ctx.JSON(http.StatusOK, deleteSuccessResponse{Error: false})
	}
}
