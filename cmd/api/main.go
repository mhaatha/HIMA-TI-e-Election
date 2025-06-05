package main

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mhaatha/HIMA-TI-e-Election/config"
	"github.com/mhaatha/HIMA-TI-e-Election/controller"
	"github.com/mhaatha/HIMA-TI-e-Election/database"
	myErrors "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
	"github.com/mhaatha/HIMA-TI-e-Election/service"
)

func main() {
	// Validator Init
	config.ValidatorInit()

	// Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		myErrors.LogError(err, "failed to load config")
	}

	// Logger Init
	config.InitLogger()
	defer config.CloseLogger()

	// DB Init
	err = database.ConnectDB(cfg)
	if err != nil {
		myErrors.LogError(err, "unable to connect to database")
	}
	defer database.DB.Close()

	// User Routes
	userRepository := repository.NewUserRepository()
	votingAccessRepository := repository.NewVotingAccessRepository()
	userService := service.NewUserService(userRepository, votingAccessRepository, cfg, database.DB, config.Validate)
	userController := controller.NewUserController(userService)

	// Auth Routes
	authRepository := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepository, userService, database.DB, config.Validate)
	authController := controller.NewAuthController(authService)

	// Upload Routes
	uploadRepository := repository.NewUploadRepository()
	uploadService := service.NewUploadService(uploadRepository, cfg, database.DB, config.Validate)
	uploadController := controller.NewUploadController(uploadService)

	// Download Routes
	downloadRepository := repository.NewDownloadRepository()
	downloadService := service.NewDownloadService(downloadRepository, cfg, database.DB, config.Validate)
	downloadController := controller.NewDownloadController(downloadService)

	// Log Routes

	// Vote Routes
	voteRepository := repository.NewVoteRepository()
	voteService := service.NewVoteService(voteRepository, votingAccessRepository, userService, database.DB, config.Validate)
	voteController := controller.NewVoteController(voteService, database.DB)

	// Candidate Routes
	candidateRepository := repository.NewCandidateRepository()
	candidateService := service.NewCandidateService(candidateRepository, cfg, voteService, database.DB, config.Validate)

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

	go voteController.ListenToDB(context.Background())

	server := http.Server{
		Addr:    "localhost:5410",
		Handler: middleware.CORSMiddleware(middleware.LoggingMiddleware(router)),
	}

	err = server.ListenAndServe()
	myErrors.LogError(err, "server failed to start")
}
