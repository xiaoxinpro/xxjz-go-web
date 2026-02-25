package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/handler"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/importsql"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/service"
	"github.com/xiaoxinpro/xxjz-go-web/backend/pkg/db"
)

func main() {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = "../config.yaml"
		}
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	// SQLite：启动时补齐可能由旧版 MySQL 导入缺失的 sort 列，避免 no such column: sort
	if cfg.Database.Driver == "sqlite" || cfg.Database.Driver == "sqlite3" {
		if err := importsql.EnsureImportSchema(database); err != nil {
			log.Printf("EnsureImportSchema: %v", err)
		}
	}

	userRepo := repository.NewUserRepo(database)
	accountRepo := repository.NewAccountRepo(database)
	transferRepo := repository.NewTransferRepo(database)
	fundsRepo := repository.NewFundsRepo(database)
	classRepo := repository.NewClassRepo(database)
	userSvc := service.NewUserService(cfg, userRepo)
	statSvc := service.NewStatisticService(accountRepo)
	findSvc := service.NewFindService(accountRepo, transferRepo, cfg.User.PageSize)
	fundsSvc := service.NewFundsService(cfg, fundsRepo, accountRepo, transferRepo)
	classSvc := service.NewClassService(cfg, classRepo)
	accountSvc := service.NewAccountService(cfg, accountRepo, classRepo, fundsRepo)
	transferSvc := service.NewTransferService(cfg, transferRepo, fundsRepo)
	chartSvc := service.NewChartService(accountRepo, classRepo)
	imageRepo := repository.NewImageRepo(database)
	imageSvc := service.NewImageService(cfg, imageRepo, handler.UploadDir)
	apiHandler := handler.NewAPIHandler(cfg, userSvc, statSvc, fundsSvc, classSvc, accountSvc, transferSvc, findSvc, chartSvc, imageSvc, database)

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "xxjz-default-secret-change-in-production"
	}
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{Path: "/", MaxAge: 86400 * 7, HttpOnly: true})

	r := gin.Default()
	r.Use(sessions.Sessions("xxjz_session", store))

	// Init (no auth): status + setup + import (import requires admin when already initialized)
	apiInit := r.Group("/api")
	apiInit.GET("/init/status", apiHandler.InitStatus)
	apiInit.POST("/init/setup", apiHandler.InitSetup)
	apiInit.POST("/init/import", apiHandler.InitImport)
	compatInit := r.Group("/Home/Api")
	compatInit.GET("/init/status", apiHandler.InitStatus)
	compatInit.POST("/init/setup", apiHandler.InitSetup)
	compatInit.POST("/init/import", apiHandler.InitImport)

	// API routes (compatible with old path: /Home/Api/xxx and new /api/xxx)
	api := r.Group("/api")
	{
		api.GET("/login", apiHandler.Login)
		api.POST("/login", apiHandler.Login)
		api.GET("/version", apiHandler.Version)
		api.POST("/version", apiHandler.Version)
		api.GET("/user", apiHandler.User)
		api.POST("/user", apiHandler.User)
		api.GET("/statistic", apiHandler.Statistic)
		api.POST("/statistic", apiHandler.Statistic)
		api.GET("/funds", apiHandler.Funds)
		api.POST("/funds", apiHandler.Funds)
		api.GET("/aclass", apiHandler.Aclass)
		api.POST("/aclass", apiHandler.Aclass)
		api.GET("/account", apiHandler.Account)
		api.POST("/account", apiHandler.Account)
		api.POST("/account/upload", apiHandler.AccountUpload)
		api.GET("/transfer", apiHandler.Transfer)
		api.POST("/transfer", apiHandler.Transfer)
		api.GET("/find", apiHandler.Find)
		api.POST("/find", apiHandler.Find)
		api.GET("/chart", apiHandler.Chart)
		api.POST("/chart", apiHandler.Chart)
		api.GET("/autocopy", apiHandler.Autocopy)
		api.POST("/autocopy", apiHandler.Autocopy)
	}
	compat := r.Group("/Home/Api")
	{
		compat.GET("/login", apiHandler.Login)
		compat.POST("/login", apiHandler.Login)
		compat.GET("/version", apiHandler.Version)
		compat.POST("/version", apiHandler.Version)
		compat.GET("/user", apiHandler.User)
		compat.POST("/user", apiHandler.User)
		compat.GET("/statistic", apiHandler.Statistic)
		compat.POST("/statistic", apiHandler.Statistic)
		compat.GET("/funds", apiHandler.Funds)
		compat.POST("/funds", apiHandler.Funds)
		compat.GET("/aclass", apiHandler.Aclass)
		compat.POST("/aclass", apiHandler.Aclass)
		compat.GET("/account", apiHandler.Account)
		compat.POST("/account", apiHandler.Account)
		compat.POST("/account/upload", apiHandler.AccountUpload)
		compat.GET("/transfer", apiHandler.Transfer)
		compat.POST("/transfer", apiHandler.Transfer)
		compat.GET("/find", apiHandler.Find)
		compat.POST("/find", apiHandler.Find)
		compat.GET("/chart", apiHandler.Chart)
		compat.POST("/chart", apiHandler.Chart)
		compat.GET("/autocopy", apiHandler.Autocopy)
		compat.POST("/autocopy", apiHandler.Autocopy)
	}
	api.POST("/admin/import", apiHandler.AdminImport)
	compat.POST("/admin/import", apiHandler.AdminImport)

	// Uploaded images static files (ensure dir exists for uploads)
	if err := os.MkdirAll(handler.UploadDir, 0755); err != nil {
		log.Printf("mkdir %s: %v", handler.UploadDir, err)
	}
	r.Static("/uploads", handler.UploadDir)

	// SPA static files (when ./static exists, e.g. in Docker)
	if info, err := os.Stat("static"); err == nil && info.IsDir() {
		r.Static("/assets", "static/assets")
		r.NoRoute(func(c *gin.Context) {
			c.File("static/index.html")
		})
	}

	addr := ":" + strconv.Itoa(cfg.Server.Port)
	if cfg.Server.Port == 0 {
		addr = ":8080"
	}
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
