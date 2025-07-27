package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Unfield/AmpKV/internal/auth"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	apiKeyMetadataKey = "api-key"
)

func AuthUnaryServerInterceptor(manager *auth.ApiKeyManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "authentication required: missing metadata")
		}

		apiKeys := md.Get(apiKeyMetadataKey)
		if len(apiKeys) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: api-key missing")
		}

		clientApiKey := apiKeys[0]

		apiKeyRecord, err := manager.GetApiKey(clientApiKey)
		if err != nil {
			if errors.Is(err, auth.ErrKeyExpired) || errors.Is(err, auth.ErrKeyDisabled) {
				return nil, status.Errorf(codes.Unauthenticated, "authentication failed: api-key expired or disabled")
			}
			return nil, status.Errorf(codes.Unauthenticated, "authentication error: %v", err)
		}

		if apiKeyRecord == nil || !apiKeyRecord.IsValid() {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: api-key invalid or expired")
		}

		requiredPerm := methodToPermission(info.FullMethod)
		if requiredPerm != "" && !apiKeyRecord.HasPermission(requiredPerm) {
			return nil, status.Errorf(codes.Unauthenticated, "authorization failed: insufficient permissions for method: %s", info.FullMethod)
		}

		return handler(ctx, req)
	}
}

func HttpAuthMiddleware(manager *auth.ApiKeyManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			authHeader := ctx.Request().Header.Get("authorization")
			authHeaderParts := strings.Fields(authHeader)
			if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer:" {
				return echo.NewHTTPError(http.StatusUnauthorized, "bearer token not found or malformed")
			}

			apiKeyRecord, err := manager.GetApiKey(authHeaderParts[1])
			if err != nil {
				if errors.Is(err, auth.ErrKeyExpired) || errors.Is(err, auth.ErrKeyDisabled) {
					return echo.NewHTTPError(http.StatusUnauthorized, "apikey expired or disabled")
				}
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify apikey")
			}

			if apiKeyRecord == nil || !apiKeyRecord.IsValid() {
				return echo.NewHTTPError(http.StatusUnauthorized, "apikey invalid or expired")
			}

			requiredPerm := httpMethodToPermission(ctx.Request().Method)
			if requiredPerm != "" && !apiKeyRecord.HasPermission(requiredPerm) {
				return echo.NewHTTPError(http.StatusUnauthorized, "insufficent permissions for method: %s", ctx.Request().Method)
			}

			return next(ctx)
		}
	}
}

func methodToPermission(fullMethod string) auth.Permission {
	switch fullMethod {
	case "/ampkv.AmpKVService/Get":
		return auth.PermRead
	case "/ampkv.AmpKVService/Set", "/ampkv.AmpKVService/SetWithTTL":
		return auth.PermRead
	case "/ampkv.AmpKVService/Delete":
		return auth.PermRead
	default:
		return ""
	}
}

func httpMethodToPermission(method string) auth.Permission {
	switch method {
	case "GET":
		return auth.PermRead
	case "POST":
		return auth.PermRead
	case "DELETE":
		return auth.PermRead
	default:
		return ""
	}
}
