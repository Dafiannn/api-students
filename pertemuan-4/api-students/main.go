package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
)

// main hanya bertugas sebagai titik masuk dan mengorkestrasikan perakitan (Composition Root).
// Tidak ada logika bisnis, tidak ada query, dan tidak ada handler di file ini.
func main() {
	// 1. Konfigurasi dan Logger Terstruktur
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Koneksi Basis Data PostgreSQL via Connection Pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Perakitan Ketergantungan dari dalam ke luar: repository -> service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 4. Inisialisasi Aplikasi Web
	app := config.NewApp(logger, pool, studentService)

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	// 5. Graceful Shutdown: Menangkap sinyal interupsi (Ctrl+C / SIGTERM)
	// dan memberikan tenggat waktu bagi request yang sedang berjalan untuk tuntas.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server dengan rapi")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
