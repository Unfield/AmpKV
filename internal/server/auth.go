package server

import (
	"context"
	"errors"

	"github.com/Unfield/AmpKV/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	apiKeyMetadataKey = "api-key"
)

func AuthUnaryServerInterceptor(manager auth.ApiKeyManager) grpc.UnaryServerInterceptor {
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
			return nil, status.Errorf(codes.Unauthenticated, "authorization failed: insufficient permissions for method %s", info.FullMethod)
		}

		return handler(ctx, req)
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
