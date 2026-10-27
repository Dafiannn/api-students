package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students-v10/app/repository"
	"api-students-v10/app/service"
	"api-students-v10/config"
	"api-students-v10/database"
	"api-students-v10/helper"
	"api-students-v10/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "supersecretjwtkeypanjangminimal32karakter!123")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	// 1. Inisialisasi PostgreSQL Connection Pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 2. Inisialisasi MongoDB Client (Atlas / Lokal)
	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	txManager := repository.NewTxManager(pool)

	// Repositories PostgreSQL
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	supervisorRepo := repository.NewSupervisorRepository(pool)

	// Repositories MongoDB
	reviewRepo := repository.NewReviewRepository(mongoDB)
	if err := reviewRepo.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index course_reviews MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	portfolioRepo := repository.NewPortfolioRepository(mongoDB)
	if err := portfolioRepo.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index student_portfolios MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission RBAC dimuat", slog.Any("roles", permissions.KnownRoles()))

	// Services
	studentService := service.NewStudentService(studentRepo, permissions)
	authService := service.NewAuthService(
		userRepo, tokenRepo, jwtManager, permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	courseService := service.NewCourseService(courseRepo, txManager)
	supervisorService := service.NewSupervisorService(supervisorRepo, txManager)
	reviewService := service.NewReviewService(reviewRepo, courseRepo, permissions)
	portfolioService := service.NewPortfolioService(portfolioRepo, studentRepo, permissions)

	app := config.NewApp(logger, route.Dependencies{
		Pool:              pool,
		Mongo:             mongoClient,
		JWT:               jwtManager,
		Permissions:       permissions,
		StudentService:    studentService,
		AuthService:       authService,
		CourseService:     courseService,
		SupervisorService: supervisorService,
		ReviewService:     reviewService,
		PortfolioService:  portfolioService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan dengan Modul 10 (Polyglot Persistence: PostgreSQL + MongoDB Atlas)", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("gagal graceful shutdown HTTP", slog.String("error", err.Error()))
	}
	logger.Info("server berhasil ditutup")
}
