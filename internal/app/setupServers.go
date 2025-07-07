package setupservers

import (
	"context"
	"flag"
	"fmt"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natefinch/lumberjack"
	"github.com/ne4chelovek/auth-service/internal/api/access"
	"github.com/ne4chelovek/auth-service/internal/api/auth"
	"github.com/ne4chelovek/auth-service/internal/api/users"
	"github.com/ne4chelovek/auth-service/internal/cache/blackList"
	"github.com/ne4chelovek/auth-service/internal/interceptor"
	k "github.com/ne4chelovek/auth-service/internal/kafkaProducer"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"github.com/ne4chelovek/auth-service/internal/metrics"
	accessRepository "github.com/ne4chelovek/auth-service/internal/repository/access"
	logRepository "github.com/ne4chelovek/auth-service/internal/repository/log"
	usersRepository "github.com/ne4chelovek/auth-service/internal/repository/users"
	"github.com/ne4chelovek/auth-service/internal/service"
	accessService "github.com/ne4chelovek/auth-service/internal/service/access"
	authService "github.com/ne4chelovek/auth-service/internal/service/auth"
	usersService "github.com/ne4chelovek/auth-service/internal/service/users"
	"github.com/ne4chelovek/auth-service/internal/tracing"
	"github.com/ne4chelovek/auth-service/internal/utils"
	descAccess "github.com/ne4chelovek/auth-service/pkg/access_v1"
	descAuth "github.com/ne4chelovek/auth-service/pkg/auth_v1"
	descUsers "github.com/ne4chelovek/auth-service/pkg/users_v1"
	_ "github.com/ne4chelovek/auth-service/statik"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rakyll/statik/fs"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var logLevel = flag.String("1", "info", "log level")

var kafkaAddresses = []string{
	"localhost:9091", // Для доступа с хоста
	"localhost:9092",
	"localhost:9093",
}

const (
	dbDSN        = "host=localhost port=5434 dbname=auth user=auth-user password=auth-password sslmode=disable"
	grpcPort     = 9000
	httpPort     = 8000
	swaggerPort  = 8005
	grpcAddress  = "localhost:9000"
	serviceName  = "test-service"
	redisAddress = "localhost:6379"
)

type Servers struct {
	GRPC          *grpc.Server
	HTTP          *http.Server
	Swagger       *http.Server
	Kafka         *k.Producer
	Redis         *redis.Client
	Prometheus    *http.Server
	TracingClient *grpc.ClientConn
	DB            *pgxpool.Pool
	Listener      net.Listener
}

func SetupServers(ctx context.Context) (*Servers, error) {
	logger.Init(getCore(getAtomicLevel()))
	tracing.Init(logger.Logger(), serviceName)

	// Инициализация базы данных
	pool, err := initDB(ctx)
	if err != nil {
		logger.Info("failed to connect to database: ", zap.Error(err))
		return nil, err
	}

	kafkaProducer, err := kafkaProducer()
	if err != nil {
		logger.Info("failed to create kafka producer", zap.Error(err))
	}

	redisConn, err := newRedisClient(ctx)
	if err != nil {
		logger.Info("failed to create redis client", zap.Error(err))
	}

	tokenUtils := createTokenUtils(redisConn)

	//Создание слоёв приложения
	usersSrv := createUsersService(pool)
	authSrv := createAuthService(pool, kafkaProducer, redisConn, tokenUtils)
	accessSrv := createAccessService(pool, redisConn, tokenUtils)

	//Инициализация gRPC сервера
	grpcServer, lis, err := setupGRPCServer(usersSrv, authSrv, accessSrv)
	if err != nil {
		logger.Info("GRPC server setup failed: ", zap.Error(err))
		return nil, err
	}

	//Иниацилизация HTTP Gateway
	httpHandler, err := setupHttpGateway(ctx)
	if err != nil {
		logger.Info("HTTP gateWay setup failed:", zap.Error(err))
		return nil, err
	}

	//Инициализация Swagger UI
	swaggerHandler, err := setupSwagger()
	if err != nil {
		logger.Info("Swagger UI setup failed")
		return nil, err
	}

	prometheusHandler, err := setupPrometheus(ctx)
	if err != nil {
		logger.Info("Prometheus setup failed")
		return nil, err
	}

	tracingClient, err := initTracing()
	if err != nil {
		logger.Info("tracing setup failed")
		return nil, err
	}

	return &Servers{
		GRPC: grpcServer,
		HTTP: &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: httpHandler,
		},
		Swagger: &http.Server{
			Addr:    fmt.Sprintf(":%d", swaggerPort),
			Handler: swaggerHandler,
		},
		Kafka: kafkaProducer,
		Redis: redisConn,
		Prometheus: &http.Server{
			Addr:        "localhost:2112",
			Handler:     prometheusHandler,
			ReadTimeout: 15 * time.Second,
		},
		TracingClient: tracingClient,
		DB:            pool,
		Listener:      lis,
	}, nil
}

func kafkaProducer() (*k.Producer, error) {
	producer, err := k.NewProducer(kafkaAddresses)
	if err != nil {
		logger.Fatal("failed to create kafka producer: %w", zap.Error(err))
	}
	return producer, nil
}

// initDB инициализирует подключение к PostgreSQL через пул соединений.
// Принимает контекст для управления таймаутами, возвращает пул соединений или ошибку.
// Выполняет:
// 1. Создание пула с заданной строкой подключения (dbDSN)
// 2. Проверку соединения через Ping
// 3. При ошибках - закрытие пула и возврат описательной ошибки
func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		logger.Info("failed to connect to database: %v", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Info("database ping failed:", zap.Error(err))
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func newRedisClient(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	// Проверяем подключение
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

func createTokenUtils(redisConn *redis.Client) utils.TokenUtils {
	return utils.NewTokenService(blackList.NewBlackList(redisConn))
}

func createUsersService(pool *pgxpool.Pool) service.UsersService {
	return usersService.NewUsersService(
		usersRepository.NewRepository(pool),
		logRepository.NewLogRepository(pool),
		pool,
	)
}

func createAuthService(pool *pgxpool.Pool, kafkaProducer *k.Producer, redisConn *redis.Client, tokenUtils utils.TokenUtils) service.AuthService {
	return authService.NewAuthService(
		usersRepository.NewRepository(pool),
		pool,
		blackList.NewBlackList(redisConn),
		kafkaProducer,
		tokenUtils,
	)
}

func createAccessService(pool *pgxpool.Pool, redisConn *redis.Client, tokenUtils utils.TokenUtils) service.AccessService {
	return accessService.NewAccessService(
		accessRepository.NewAccessRepository(pool),
		blackList.NewBlackList(redisConn),
		pool,
		tokenUtils,
	)
}

func setupGRPCServer(usersSrv service.UsersService, authSrv service.AuthService, accessToken service.AccessService) (*grpc.Server, net.Listener, error) {
	//creds, err := credentials.NewServerTLSFromFile("certs/service.pem", "certs/service.key")
	//if err != nil {
	//	return nil, nil, fmt.Errorf("failed to create credentials: %w", err)
	//}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Info("failed to listen: ", zap.Error(err))
		return nil, nil, err
	}

	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				interceptor.ValidateInterceptor,
				interceptor.NewCircuitBreakerInterceptor(setupCircuitBreaker()).Unary,
				interceptor.LogInterceptor,
				interceptor.MetricsInterceptor,
				interceptor.ServerTracingInterceptor,
			),
		),
	)
	reflection.Register(server)
	descUsers.RegisterUsersV1Server(server, users.NewUsersImplementation(usersSrv))
	descAuth.RegisterAuthV1Server(server, auth.NewAuthImplementation(authSrv))
	descAccess.RegisterAccessV1Server(server, access.NewAccessTokenImplementation(accessToken))

	return server, lis, nil
}

// setupHttpGateway создает HTTP-шлюз для gRPC сервиса (REST -> gRPC преобразование).
func setupHttpGateway(ctx context.Context) (http.Handler, error) {
	//creds, err := credentials.NewClientTLSFromFile("ce rts/service.pem", "")
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create credentials transport: %w", err)
	//}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	mux := runtime.NewServeMux()
	if err := descUsers.RegisterUsersV1HandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		logger.Info("failed to register gateway: ", zap.Error(err))
		return nil, err
	}

	return enableCORS(mux), nil
}

// createSwaggerHandler создает файловый сервер для Swagger UI с особенностями:
// - Использует встроенные ресурсы через statik
// - Добавляет обработчик для /api.swagger.json
// - Возвращает 404 при отсутствии файла документации
func setupSwagger() (http.Handler, error) {
	statikFS, err := fs.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init statik: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(statikFS))

	mux.HandleFunc("/api.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		file, err := statikFS.Open("/api.swagger.json")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "application/json")
		_, err = io.Copy(w, file)
		if err != nil {
			logger.Error("failed copy file")
		}
	})

	return mux, nil

}

func setupPrometheus(ctx context.Context) (http.Handler, error) {
	err := metrics.Init(ctx)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return mux, nil
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Grpc-Web")
		w.Header().Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func getCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	})

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)
}

func getAtomicLevel() zap.AtomicLevel {
	var level zapcore.Level
	if err := level.Set(*logLevel); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}
	return zap.NewAtomicLevelAt(level)
}

func setupCircuitBreaker() *gobreaker.CircuitBreaker {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "auth_service",
		MaxRequests: 3,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit Breaker: %s, changed from %v, to %v\n", name, from, to)
		},
	})
	return cb
}

func initTracing() (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
