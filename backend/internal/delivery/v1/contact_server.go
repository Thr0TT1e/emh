package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/netutil"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// ContactServer обработчик Connect RPC для сервиса обратной связи.
type ContactServer struct {
	contactUC    *usecase.ContactUseCase
	proxyChecker *netutil.ProxyChecker
	logger       *slog.Logger
}

// NewContactServer создаёт сервер ContactService.
// trustedProxies — список доверенных прокси (Caddy) для корректного извлечения IP.
func NewContactServer(
	uc *usecase.ContactUseCase,
	trustedProxies []string,
	logger *slog.Logger,
) *ContactServer {
	return &ContactServer{
		contactUC:    uc,
		proxyChecker: netutil.NewProxyChecker(trustedProxies),
		logger:       logger,
	}
}

// CreateContactMessage принимает сообщение формы обратной связи,
// сохраняет его в БД и инициирует отправку email-уведомления.
func (s *ContactServer) CreateContactMessage(
	ctx context.Context,
	req *connect.Request[emhv1.CreateContactMessageRequest],
) (*connect.Response[emhv1.CreateContactMessageResponse], error) {
	clientIP := s.extractClientIP(req)
	userAgent := req.Header().Get("User-Agent")

	input := usecase.SubmitContactInput{
		Name:      req.Msg.Name,
		Email:     req.Msg.Email,
		Subject:   req.Msg.Subject,
		Message:   req.Msg.Message,
		PageURL:   req.Msg.PageUrl,
		Honeypot:  req.Msg.Honeypot,
		Consent:   req.Msg.Consent,
		ClientIP:  clientIP,
		UserAgent: userAgent,
	}

	id, err := s.contactUC.Submit(ctx, input)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create contact message failed")
	}

	return connect.NewResponse(&emhv1.CreateContactMessageResponse{
		Id: id,
	}), nil
}

// extractClientIP извлекает IP клиента с учётом trusted proxies
// (метод rightmost-untrusted, см. netutil.ClientIP).
func (s *ContactServer) extractClientIP(req *connect.Request[emhv1.CreateContactMessageRequest]) string {
	return netutil.ClientIP(s.proxyChecker, req.Header(), req.Peer().Addr)
}
