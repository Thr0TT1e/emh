package metrics

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	rpcRequestsTotal *prometheus.CounterVec
	rpcDuration      *prometheus.HistogramVec
	rpcRequestSize   *prometheus.HistogramVec
	rpcResponseSize  *prometheus.HistogramVec
)

func initRPCMetrics() {
	rpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "rpc",
			Name:      "requests_total",
			Help:      "Общее количество RPC-запросов по сервису, методу и коду ответа.",
		},
		[]string{"rpc_service", "rpc_method", "rpc_type", "code"},
	)

	rpcDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "rpc",
			Name:      "duration_seconds",
			Help:      "Длительность обработки RPC-запросов.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"rpc_service", "rpc_method", "rpc_type"},
	)

	rpcRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "rpc",
			Name:      "request_size_bytes",
			Help:      "Размер входящих RPC-запросов в байтах.",
			Buckets:   prometheus.ExponentialBuckets(256, 4, 8), // 256B → 4MB
		},
		[]string{"rpc_service", "rpc_method"},
	)

	rpcResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "rpc",
			Name:      "response_size_bytes",
			Help:      "Размер исходящих RPC-ответов в байтах.",
			Buckets:   prometheus.ExponentialBuckets(256, 4, 8),
		},
		[]string{"rpc_service", "rpc_method"},
	)

	Registry.MustRegister(rpcRequestsTotal, rpcDuration, rpcRequestSize, rpcResponseSize)
}

// RPCInterceptor — Connect-интерцептор для сбора RPC-метрик.
// Подключается первым в цепочке интерцепторов (до authInterceptor),
// чтобы замерять все запросы, включая отклонённые аутентификацией.
type RPCInterceptor struct{}

// NewRPCInterceptor создаёт интерцептор метрик.
func NewRPCInterceptor() *RPCInterceptor {
	return &RPCInterceptor{}
}

// WrapUnary реализует connect.Interceptor для unary-запросов.
func (i *RPCInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		service, method := splitProcedure(req.Spec().Procedure)
		rpcType := "unary"
		if req.Spec().IsClient {
			// Для серверного интерцептора IsClient=false, но на всякий случай
			rpcType = "client"
		}

		resp, err := next(ctx, req)
		elapsed := time.Since(start).Seconds()

		code := "ok"
		if err != nil {
			if connectErr, ok := err.(*connect.Error); ok {
				code = connectErr.Code().String()
			} else {
				code = "unknown"
			}
		}

		rpcRequestsTotal.WithLabelValues(service, method, rpcType, code).Inc()
		rpcDuration.WithLabelValues(service, method, rpcType).Observe(elapsed)

		// Размер запроса (если доступен)
		if req != nil && req.Any() != nil {
			if size := estimateSize(req.Any()); size > 0 {
				rpcRequestSize.WithLabelValues(service, method).Observe(float64(size))
			}
		}
		// Размер ответа
		if resp != nil && resp.Any() != nil {
			if size := estimateSize(resp.Any()); size > 0 {
				rpcResponseSize.WithLabelValues(service, method).Observe(float64(size))
			}
		}

		return resp, err
	})
}

// WrapStreamingClient passthrough — стриминг в проекте не используется.
func (i *RPCInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler passthrough — стриминг в проекте не используется.
func (i *RPCInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// splitProcedure разбирает "/emh.v1.HeroAdminService/CreateHero" → ("emh.v1.HeroAdminService", "CreateHero").
func splitProcedure(proc string) (service, method string) {
	// Формат: "/<package>.<Service>/<Method>"
	if len(proc) > 0 && proc[0] == '/' {
		proc = proc[1:]
	}
	for i := len(proc) - 1; i >= 0; i-- {
		if proc[i] == '/' {
			return proc[:i], proc[i+1:]
		}
	}
	return proc, ""
}

// estimateSize оценивает размер protobuf-сообщения через proto.Size.
// Возвращает 0 если тип неизвестен.
func estimateSize(msg any) int {
	// Для proto.Message можно использовать proto.Size().
	// Здесь — простая оценка через fmt.Sprintf (достаточно для метрик).
	// В продакшене лучше использовать protobuf reflection.
	type sizer interface{ Size() int }
	if s, ok := msg.(sizer); ok {
		return s.Size()
	}
	return 0
}
