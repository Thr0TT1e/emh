// Реализация MediaService

package v1

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// MediaServer реализует emhv1.MediaServiceHandler.
type MediaServer struct {
	mediaUC usecase.MediaUseCase
	logger  *slog.Logger
}

func NewMediaServer(uc usecase.MediaUseCase, logger *slog.Logger) *MediaServer {
	return &MediaServer{mediaUC: uc, logger: logger}
}

// GetUploadUrl генерирует presigned URL для прямой загрузки файла в хранилище.
func (s *MediaServer) GetUploadUrl(
	ctx context.Context,
	req *connect.Request[emhv1.GetUploadUrlRequest],
) (*connect.Response[emhv1.GetUploadUrlResponse], error) {
	s.logger.InfoContext(ctx, "generating upload url",
		"type", req.Msg.Type.String(),
		"filename", req.Msg.Filename,
	)

	params := domain.UploadParams{
		Type:        domain.UploadType(req.Msg.Type),
		Filename:    req.Msg.Filename,
		ContentType: req.Msg.ContentType,
	}

	result, err := s.mediaUC.GetUploadURL(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get upload url failed")
	}

	return connect.NewResponse(&emhv1.GetUploadUrlResponse{
		UploadUrl: result.UploadURL,
		PublicUrl: result.PublicURL,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	}), nil
}

// BatchGetUploadUrls генерирует несколько presigned URL за один запрос.
func (s *MediaServer) BatchGetUploadUrls(
	ctx context.Context,
	req *connect.Request[emhv1.BatchGetUploadUrlsRequest],
) (*connect.Response[emhv1.BatchGetUploadUrlsResponse], error) {
	s.logger.InfoContext(ctx, "batch generating upload urls",
		"type", req.Msg.Type.String(),
		"count", len(req.Msg.Files),
	)

	files := make([]domain.FileToUpload, 0, len(req.Msg.Files))
	for _, f := range req.Msg.Files {
		files = append(files, domain.FileToUpload{
			Filename:    f.Filename,
			ContentType: f.ContentType,
		})
	}

	urls, err := s.mediaUC.BatchGetUploadURLs(ctx, domain.UploadType(req.Msg.Type), files)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "batch get upload urls failed")
	}

	respUrls := make([]*emhv1.UploadUrlInfo, 0, len(urls))
	for _, u := range urls {
		respUrls = append(respUrls, &emhv1.UploadUrlInfo{
			Filename:  u.Filename,
			UploadUrl: u.UploadURL,
			PublicUrl: u.PublicURL,
			ExpiresAt: u.ExpiresAt,
		})
	}

	return connect.NewResponse(&emhv1.BatchGetUploadUrlsResponse{
		Urls:  respUrls,
		Count: int32(len(respUrls)),
	}), nil
}
