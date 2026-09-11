package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// LLMAdminServer реализует emhv1connect.LlmAdminServiceHandler.
type LLMAdminServer struct {
	emhv1connect.UnimplementedLlmAdminServiceHandler
	uc     usecase.LLMAdminUseCase
	logger *slog.Logger
}

// NewLLMAdminServer создаёт сервер управления LLM-провайдерами.
func NewLLMAdminServer(uc usecase.LLMAdminUseCase, logger *slog.Logger) *LLMAdminServer {
	return &LLMAdminServer{
		uc:     uc,
		logger: logger,
	}
}

// ListLlmProviders возвращает список всех настроенных провайдеров.
func (s *LLMAdminServer) ListLlmProviders(
	ctx context.Context,
	req *connect.Request[emhv1.ListLlmProvidersRequest],
) (*connect.Response[emhv1.ListLlmProvidersResponse], error) {
	providers, err := s.uc.ListProviders(ctx)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list_llm_providers")
	}

	resp := &emhv1.ListLlmProvidersResponse{
		Count: int32(len(providers)),
	}

	for _, p := range providers {
		resp.Providers = append(resp.Providers, &emhv1.LlmProviderInfo{
			Id:        p.ID,
			Name:      p.Name,
			Type:      p.Type,
			Model:     p.Model,
			IsActive:  p.IsActive,
			Priority:  int32(p.Priority),
			Notes:     p.Notes,
			CreatedAt: timestamppb.New(p.CreatedAt),
			UpdatedAt: timestamppb.New(p.UpdatedAt),
		})
	}

	return connect.NewResponse(resp), nil
}

// SetActiveLlmProvider активирует указанный провайдер.
func (s *LLMAdminServer) SetActiveLlmProvider(
	ctx context.Context,
	req *connect.Request[emhv1.SetActiveLlmProviderRequest],
) (*connect.Response[emhv1.SetActiveLlmProviderResponse], error) {
	if err := s.uc.SetActiveProvider(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "set_active_llm_provider")
	}

	// Получаем обновлённый список для возврата
	providers, _ := s.uc.ListProviders(ctx)
	var activeProvider *emhv1.LlmProviderInfo
	for _, p := range providers {
		if p.ID == req.Msg.Id {
			activeProvider = &emhv1.LlmProviderInfo{
				Id:        p.ID,
				Name:      p.Name,
				Type:      p.Type,
				Model:     p.Model,
				IsActive:  p.IsActive,
				Priority:  int32(p.Priority),
				Notes:     p.Notes,
				CreatedAt: timestamppb.New(p.CreatedAt),
				UpdatedAt: timestamppb.New(p.UpdatedAt),
			}
			break
		}
	}

	return connect.NewResponse(&emhv1.SetActiveLlmProviderResponse{
		Success:  true,
		Provider: activeProvider,
	}), nil
}

// TestLlmProvider отправляет тестовый запрос к провайдеру.
func (s *LLMAdminServer) TestLlmProvider(
	ctx context.Context,
	req *connect.Request[emhv1.TestLlmProviderRequest],
) (*connect.Response[emhv1.TestLlmProviderResponse], error) {
	result, err := s.uc.TestProvider(ctx, req.Msg.Id, req.Msg.TestPrompt)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "test_llm_provider")
	}

	return connect.NewResponse(&emhv1.TestLlmProviderResponse{
		Success:          result.Success,
		RawResponse:      result.RawResponse,
		ProcessingTimeMs: result.ProcessingTimeMs,
		ErrorMessage:     result.ErrorMessage,
	}), nil
}

// UpdateLlmProvider обновляет приоритет и заметки провайдера.
func (s *LLMAdminServer) UpdateLlmProvider(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateLlmProviderRequest],
) (*connect.Response[emhv1.UpdateLlmProviderResponse], error) {
	if err := s.uc.UpdateProvider(ctx, req.Msg.Id, int(req.Msg.Priority), req.Msg.Notes); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "update_llm_provider")
	}

	// Получаем обновлённый провайдер
	providers, _ := s.uc.ListProviders(ctx)
	var updatedProvider *emhv1.LlmProviderInfo
	for _, p := range providers {
		if p.ID == req.Msg.Id {
			updatedProvider = &emhv1.LlmProviderInfo{
				Id:        p.ID,
				Name:      p.Name,
				Type:      p.Type,
				Model:     p.Model,
				IsActive:  p.IsActive,
				Priority:  int32(p.Priority),
				Notes:     p.Notes,
				CreatedAt: timestamppb.New(p.CreatedAt),
				UpdatedAt: timestamppb.New(p.UpdatedAt),
			}
			break
		}
	}

	return connect.NewResponse(&emhv1.UpdateLlmProviderResponse{
		Provider: updatedProvider,
	}), nil
}
