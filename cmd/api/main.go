package main

import (
	"context"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/config"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/controller"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/database"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/repository"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/service"
)

func main() {
	// Set TimeZone
	os.Setenv("TZ", "Asia/Makassar")

	// Logger Init
	config.InitLogger()
	defer config.CloseLogger()

	// Validator Init
	config.ValidatorInit()

	// Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		appError.LogError(err, "failed to load config")
		os.Exit(1)
	}

	// DB Init
	db, err := database.ConnectDB(cfg)
	if err != nil {
		appError.LogError(err, "failed to initialize database")
		os.Exit(1)
	}
	defer db.Close()

	// User Routes
	userRepository := repository.NewUserRepository()
	votingAccessRepository := repository.NewVotingAccessRepository()
	userService := service.NewUserService(userRepository, votingAccessRepository, cfg, db, config.Validate)
	userController := controller.NewUserController(userService)

	// Auth Routes
	authRepository := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepository, userService, db, config.Validate)
	authController := controller.NewAuthController(authService)

	// Upload Routes
	uploadRepository := repository.NewUploadRepository()
	uploadService := service.NewUploadService(uploadRepository, cfg, db, config.Validate)
	uploadController := controller.NewUploadController(uploadService)

	// Download Routes
	downloadRepository := repository.NewDownloadRepository()
	downloadService := service.NewDownloadService(downloadRepository, cfg, db, config.Validate)
	downloadController := controller.NewDownloadController(downloadService)

	// Log Routes

	// Vote Routes
	voteRepository := repository.NewVoteRepository()
	voteService := service.NewVoteService(voteRepository, votingAccessRepository, authRepository, userService, db, config.Validate)
	voteController := controller.NewVoteController(voteService, db)

	// Candidate Routes
	candidateRepository := repository.NewCandidateRepository()
	candidateService := service.NewCandidateService(candidateRepository, cfg, voteService, db, config.Validate)

	// Inject candidateService to voteService to solve Circular Dependency
	voteService.SetCandidateService(candidateService)

	// Candidate Controller
	candidateController := controller.NewCandidateController(candidateService)

	router := httprouter.New()

	// User Path
	router.POST("/api/users", middleware.AdminMiddleware(userController.Create, authService))
	router.GET("/api/users", middleware.AdminMiddleware(userController.GetAll, authService))
	// router.PATCH("/api/users/current", middleware.UserMiddleware(userController.UpdateCurrent, authService))
	router.GET("/api/users/current", middleware.UserMiddleware(userController.GetCurrent, authService))
	router.PATCH("/api/users/:userId", middleware.AdminMiddleware(userController.UpdateById, authService))
	router.DELETE("/api/users/:userId", middleware.AdminMiddleware(userController.DeleteById, authService))
	router.POST("/api/users/bulk", middleware.AdminMiddleware(userController.BulkCreate, authService))
	router.POST("/api/users/generate-passwords", middleware.AdminMiddleware(userController.GeneratePassword, authService))

	// Auth Path
	router.POST("/api/auth/login", authController.Login)
	router.POST("/api/auth/logout", middleware.UserMiddleware(authController.Logout, authService))

	// Upload Path
	router.GET("/api/upload/candidates/presigned-url", middleware.AdminMiddleware(uploadController.GetPresignedUrl, authService))

	// Download Path
	router.GET("/api/download/logs/vote", middleware.AdminMiddleware(downloadController.GetPresignedUrl, authService))

	// Candidate Path
	router.POST("/api/candidates", middleware.AdminMiddleware(candidateController.Create, authService))
	router.GET("/api/candidates", middleware.UserMiddleware(candidateController.GetCandidates, authService))
	router.GET("/api/candidates/:candidateId", middleware.AdminMiddleware(candidateController.GetCandidateById, authService))
	router.PATCH("/api/candidates/:candidateId", middleware.AdminMiddleware(candidateController.UpdateCandidateById, authService))
	router.DELETE("/api/candidates/:candidateId", middleware.AdminMiddleware(candidateController.DeleteCandidateById, authService))

	// Vote Path
	router.POST("/api/votes", middleware.UserMiddleware(voteController.Save, authService))
	router.GET("/api/votes/:candidateId", middleware.AdminMiddleware(voteController.GetTotalVotesByCandidateId, authService))
	router.GET("/ws/votes", middleware.AdminMiddleware(voteController.VotesLiveResult, authService))
	router.GET("/api/user/vote-status", middleware.UserMiddleware(voteController.CheckIfUserHasVoted, authService))

	go voteController.ListenToDB(context.Background())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: middleware.CORSMiddleware(middleware.LoggingMiddleware(router)),
	}

	err = server.ListenAndServe()
	appError.LogError(err, "server failed to start")
}
