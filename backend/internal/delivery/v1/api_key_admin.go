package v1

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// APIKeyAdminServer реализует emhv1.ApiKeyAdminServiceHandler.
type APIKeyAdminServer struct {
	apiKeyUC usecase.APIKeyUseCase
	logger   *slog.Logger
}

func NewAPIKeyAdminServer(uc usecase.APIKeyUseCase, logger *slog.Logger) *APIKeyAdminServer {
	return &APIKeyAdminServer{apiKeyUC: uc, logger: logger}
}

// CreateApiKey создаёт ключ. Полный ключ возвращается только в этом ответе.
func (s *APIKeyAdminServer) CreateApiKey(
	ctx context.Context,
	req *connect.Request[emhv1.CreateApiKeyRequest],
) (*connect.Response[emhv1.CreateApiKeyResponse], error) {
	// Идентифицируем создателя из контекста (JWT или имя другого ключа)
	createdBy := ""
	if claims, ok := interceptor.ClaimsFromContext(ctx); ok {
		createdBy = claims.Username
	}

	var expiresAt *time.Time
	if req.Msg.ExpiresAt != nil {
		t := req.Msg.ExpiresAt.AsTime()
		expiresAt = &t
	}

	params := domain.CreateAPIKeyParams{
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
		Role:        req.Msg.Role,
		CreatedBy:   createdBy,
		ExpiresAt:   expiresAt,
	}

	key, fullKey, err := s.apiKeyUC.Create(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create api key failed")
	}

	s.logger.InfoContext(ctx, "api key created", "name", key.Name, "created_by", createdBy)
	return connect.NewResponse(&emhv1.CreateApiKeyResponse{
		ApiKey:  mapAPIKeyToProto(key),
		FullKey: fullKey,
	}), nil
}

// ListApiKeys возвращает список ключей (без секретов).
func (s *APIKeyAdminServer) ListApiKeys(
	ctx context.Context,
	req *connect.Request[emhv1.ListApiKeysRequest],
) (*connect.Response[emhv1.ListApiKeysResponse], error) {
	keys, err := s.apiKeyUC.List(ctx, req.Msg.IncludeRevoked)
	if err != nil {
		s.logger.ErrorContext(ctx, "list api keys failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	protoKeys := make([]*emhv1.ApiKey, 0, len(keys))
	for _, k := range keys {
		protoKeys = append(protoKeys, mapAPIKeyToProto(k))
	}

	return connect.NewResponse(&emhv1.ListApiKeysResponse{ApiKeys: protoKeys}), nil
}

// RevokeApiKey отзывает ключ.
func (s *APIKeyAdminServer) RevokeApiKey(
	ctx context.Context,
	req *connect.Request[emhv1.RevokeApiKeyRequest],
) (*connect.Response[emhv1.RevokeApiKeyResponse], error) {
	s.logger.InfoContext(ctx, "revoking api key", "id", req.Msg.Id)

	if err := s.apiKeyUC.Revoke(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "revoke api key failed")
	}

	return connect.NewResponse(&emhv1.RevokeApiKeyResponse{Success: true}), nil
}

// mapAPIKeyToProto маппит доменную модель в proto. Секретный хеш НЕ передаётся.
func mapAPIKeyToProto(k *domain.APIKey) *emhv1.ApiKey {
	proto := &emhv1.ApiKey{
		Id:          k.ID,
		KeyId:       k.KeyID,
		Name:        k.Name,
		Description: k.Description,
		Role:        k.Role,
		CreatedBy:   k.CreatedBy,
		CreatedAt:   timestamppb.New(k.CreatedAt),
	}
	if k.RevokedAt != nil {
		proto.RevokedAt = timestamppb.New(*k.RevokedAt)
	}
	if k.ExpiresAt != nil {
		proto.ExpiresAt = timestamppb.New(*k.ExpiresAt)
	}
	if k.LastUsedAt != nil {
		proto.LastUsedAt = timestamppb.New(*k.LastUsedAt)
	}
	return proto
}
