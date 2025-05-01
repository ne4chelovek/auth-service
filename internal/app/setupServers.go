package setupservers

import (
	"github.com/ne4chelovek/auth-service/internal/api/access"
	"github.com/ne4chelovek/auth-service/internal/api/auth"
	"github.com/ne4chelovek/auth-service/internal/api/users"
	"github.com/ne4chelovek/auth-service/internal/interceptor"
	accessRepository "github.com/ne4chelovek/auth-service/internal/repository/access"
	logRepository "github.com/ne4chelovek/auth-service/internal/repository/log"
	usersRepository "github.com/ne4chelovek/auth-service/internal/repository/users"
	"github.com/ne4chelovek/auth-service/internal/service"
	accessService "github.com/ne4chelovek/auth-service/internal/service/access"
	authService "github.com/ne4chelovek/auth-service/internal/service/auth"
	usersService "github.com/ne4chelovek/auth-service/internal/service/users"
	descAccess "github.com/ne4chelovek/auth-service/pkg/access_v1"
	descAuth "github.com/ne4chelovek/auth-service/pkg/auth_v1"
	descUsers "github.com/ne4chelovek/auth-service/pkg/users_v1"

	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/ne4chelovek/auth-service/statik"
)

const (
	dbDSN       = "host=localhost port=54322 dbname=auth user=auth-user password=auth-password sslmode=disable"
	grpcPort    = 9000
	httpPort    = 8000
	swaggerPort = 8005
	grpcAddress = "localhost:9000"
)

type Servers struct {
	GRPC     *grpc.Server
	HTTP     *http.Server
	Swagger  *http.Server
	DB       *pgxpool.Pool
	Listener net.Listener
}

func SetupServers(ctx context.Context) (*Servers, error) {
	// Инициализация базы данных
	pool, err := initDB(ctx)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return nil, err
	}

	//Создание слоёв приложения
	usersSrv := createUsersService(pool)
	authSrv := createAuthService(pool)
	accessSrv := createAccessService(pool)

	//Инициализация gRPC сервера
	grpcServer, lis, err := setupGRPCServer(usersSrv, authSrv, accessSrv)
	if err != nil {
		log.Printf("GRPC server setup failed: %v", err)
		return nil, err
	}

	//Иниацилизация HTTP Gateway
	httpHandler, err := setupHttpGateway(ctx)
	if err != nil {
		log.Printf("HTTP gateWay setup failed: %v", err)
		return nil, err
	}

	//Инициализация Swagger UI
	swaggerHandler, err := setupSwagger()
	if err != nil {
		pool.Close()
		grpcServer.Stop()
		log.Printf("Swagger UI setup failed")
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
		DB:       pool,
		Listener: lis,
	}, nil
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
		log.Printf("failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("database ping failed: %v", err)
		pool.Close()
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return pool, nil
}

func createUsersService(pool *pgxpool.Pool) service.UsersService {
	return usersService.NewUsersService(
		usersRepository.NewRepository(pool),
		logRepository.NewLogRepository(pool),
		pool,
	)
}

func createAuthService(pool *pgxpool.Pool) service.AuthService {
	return authService.NewAuthService(
		usersRepository.NewRepository(pool),
		pool,
	)
}

func createAccessService(pool *pgxpool.Pool) service.AccessService {
	return accessService.NewAccessService(
		accessRepository.NewAccessRepository(pool),
		pool,
	)
}

// setupGRPCServer настраивает и запускает gRPC сервер:
// 1. Слушает TCP-порт (grpcPort)
// 2. Создает сервер с credentials и интерцептором валидации
// 3. Регестрирует слои сервисов
func setupGRPCServer(usersSrv service.UsersService, authSrv service.AuthService, accessToken service.AccessService) (*grpc.Server, net.Listener, error) {
	creds, err := credentials.NewServerTLSFromFile("certs/service.pem", "certs/service.key")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return nil, nil, err
	}

	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(interceptor.ValidateInterceptor),
	)
	reflection.Register(server)
	descUsers.RegisterUsersV1Server(server, users.NewUsersImplementation(usersSrv))
	descAuth.RegisterAuthV1Server(server, auth.NewAuthImplementation(authSrv))
	descAccess.RegisterAccessV1Server(server, access.NewAccessTokenImplementation(accessToken))

	return server, lis, nil
}

// setupHttpGateway создает HTTP-шлюз для gRPC сервиса (REST -> gRPC преобразование).
func setupHttpGateway(ctx context.Context) (http.Handler, error) {
	creds, err := credentials.NewClientTLSFromFile("certs/service.pem", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials transport: %w", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	mux := runtime.NewServeMux()
	if err := descUsers.RegisterUsersV1HandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		log.Printf("failed to register gateway: %v", err)
		return nil, err
	}

	return enableCORS(mux), nil
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
		io.Copy(w, file)
	})

	return mux, nil
}
