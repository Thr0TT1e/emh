package main

import (
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/middleware"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// buildRouter создаёт chi-роутер со всеми middleware и сервисами.
// rateLimitMW создаётся и запускается в main.go (нужен контекст для cleanup loop).
func buildRouter(
	cfg *config.Config,
	logger *slog.Logger,
	deps *Dependencies,
	authInterceptor *interceptor.AuthInterceptor,
	rateLimitMW *middleware.RateLimitMiddleware,
) *chi.Mux {
	mux := chi.NewMux()

	// ─── CORS ───────────────────────────────────────────────────
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.App.CORSOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type",
			"Connect-Protocol-Version", "Connect-Timeout-Ms",
			"Connect-Content-Encoding", "Grpc-Timeout",
			"X-Grpc-Web", "X-User-Agent",
		},
		ExposedHeaders:   []string{"Connect-Content-Encoding"},
		AllowCredentials: true,
	}))

	// ─── Rate Limiting (создан в main.go, handler подключаем здесь) ───
	mux.Use(rateLimitMW.Handler)

	// ─── Connect Interceptors ───────────────────────────────
	validateInterceptor := validate.NewInterceptor()
	rpcMetricsInterceptor := metrics.NewRPCInterceptor()

	const maxRequestBytes = 4 << 20 // 4 MB

	adminOpts := []connect.HandlerOption{
		connect.WithInterceptors(rpcMetricsInterceptor, authInterceptor, validateInterceptor),
		connect.WithReadMaxBytes(maxRequestBytes),
		connect.WithCompression("gzip", nil, nil),
	}
	publicOpts := []connect.HandlerOption{
		connect.WithInterceptors(rpcMetricsInterceptor, validateInterceptor),
		connect.WithReadMaxBytes(maxRequestBytes),
		connect.WithCompression("gzip", nil, nil),
	}

	// Extraction-specific: tighter body limit.
	// 150K chars × 4 bytes (worst UTF-8) + JSON overhead ≈ 700KB.
	// 1 MB — безопасный запас; точную проверку на 150К символов делает usecase.
	const extractionMaxBytes = 1 << 20 // 1 MB
	extractionOpts := []connect.HandlerOption{
		connect.WithInterceptors(rpcMetricsInterceptor, authInterceptor, validateInterceptor),
		connect.WithReadMaxBytes(extractionMaxBytes),
		connect.WithCompression("gzip", nil, nil),
	}

	// ─── Mount сервисов ─────────────────────────────────────────
	mountServices(mux, deps, adminOpts, publicOpts, extractionOpts)

	// ─── Health check ───────────────────────────────────────────
	mux.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Prometheus metrics endpoint
	// Не требует аутентификации в dev, в production защищается Caddy (IP whitelist).
	mux.Handle("/metrics", metrics.Handler)

	return mux
}

// mountServices регистрирует все Connect RPC сервисы.
func mountServices(
	mux *chi.Mux,
	deps *Dependencies,
	adminOpts,
	publicOpts,
	extractionOpts []connect.HandlerOption,
) {
	// Hero
	path, handler := emhv1connect.NewHeroAdminServiceHandler(deps.HeroAdminServer, adminOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewHeroServiceHandler(deps.HeroPublicServer, publicOpts...)
	mux.Mount(path, handler)

	// Award
	path, handler = emhv1connect.NewAwardServiceHandler(deps.AwardServer, publicOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewAwardAdminServiceHandler(deps.AwardAdminServer, adminOpts...)
	mux.Mount(path, handler)

	// Media
	path, handler = emhv1connect.NewMediaServiceHandler(deps.MediaServer, adminOpts...)
	mux.Mount(path, handler)

	// Conflict
	path, handler = emhv1connect.NewConflictServiceHandler(deps.ConflictServer, publicOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewConflictAdminServiceHandler(deps.ConflictAdminServer, adminOpts...)
	mux.Mount(path, handler)

	// Location
	path, handler = emhv1connect.NewLocationServiceHandler(deps.LocationServer, publicOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewLocationAdminServiceHandler(deps.LocationAdminServer, adminOpts...)
	mux.Mount(path, handler)

	// Submission
	path, handler = emhv1connect.NewSubmissionServiceHandler(deps.SubmissionServer, publicOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewSubmissionAdminServiceHandler(deps.SubmissionAdminServer, adminOpts...)
	mux.Mount(path, handler)

	// Auth & API Keys
	path, handler = emhv1connect.NewAuthServiceHandler(deps.AuthServer, publicOpts...)
	mux.Mount(path, handler)

	path, handler = emhv1connect.NewApiKeyAdminServiceHandler(deps.APIKeyAdminServer, adminOpts...)
	mux.Mount(path, handler)

	// Contact
	path, handler = emhv1connect.NewContactServiceHandler(deps.ContactServer, publicOpts...)
	mux.Mount(path, handler)

	// Extraction (LLM)
	path, handler = emhv1connect.NewExtractionServiceHandler(deps.ExtractionServer, extractionOpts...)
	mux.Mount(path, handler)

	// LLM Admin
	path, handler = emhv1connect.NewLlmAdminServiceHandler(deps.LLMAdminServer, adminOpts...)
	mux.Mount(path, handler)
}
