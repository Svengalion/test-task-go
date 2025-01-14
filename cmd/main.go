package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/Svengalion/test-task-go/internal/config"
	"github.com/Svengalion/test-task-go/internal/delivery/http"
	"github.com/Svengalion/test-task-go/internal/repository/postgres"
	"github.com/Svengalion/test-task-go/internal/usecase"
)

func main() {
	// Загружаем .env
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error reading it, proceeding with environment variables")
	}

	cfg := config.New()

	// Логгер
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Подключение к БД
	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		logger.Fatal("failed to connect to postgres: ", err)
	}
	defer db.Close()

	// Запуск миграций (пример с golang-migrate)
	// migrateDB(cfg.PostgresDSN, logger)

	// Инициализируем репозиторий и usecase
	songRepo := postgres.NewSongRepository(db)
	infoClient := usecase.NewMusicInfoClient(cfg.ExternalAPIBaseURL)
	songUseCase := usecase.NewSongUseCase(songRepo, infoClient)

	// Настраиваем роутер gin
	r := gin.Default()

	// Регистрируем хендлеры
	songHandler := http.NewSongHandler(songUseCase, logger)

	// Роутинг
	v1 := r.Group("/songs")
	{
		v1.GET("", songHandler.GetAllSongs)
		v1.POST("", songHandler.CreateSong)
		v1.GET("/:id/text", songHandler.GetSongText)
		v1.PUT("/:id", songHandler.UpdateSong)
		v1.DELETE("/:id", songHandler.DeleteSong)
	}

	// Инициализация swagger-документации
	// import _ "github.com/yourusername/yourproject/docs"
	// swaggerURL := ginSwagger.URL("http://localhost:" + cfg.Port + "/swagger/doc.json")
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))

	// Запуск
	logger.Info("Starting server on port ", cfg.Port)
	r.Run(":" + cfg.Port)
}
