package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/yourusername/yourproject/internal/domain"
	"github.com/yourusername/yourproject/internal/usecase"
)

// SongHandler содержит зависимости для REST эндпоинтов
type SongHandler struct {
	uc  *usecase.SongUseCase
	log *logrus.Logger
}

func NewSongHandler(uc *usecase.SongUseCase, logger *logrus.Logger) *SongHandler {
	return &SongHandler{
		uc:  uc,
		log: logger,
	}
}

// @Summary     Получение списка песен
// @Description Возвращает список песен с фильтрацией и пагинацией
// @Tags        songs
// @Accept      json
// @Produce     json
// @Param       group     query string false "filter by group"
// @Param       song      query string false "filter by title"
// @Param       page      query int    false "page number"
// @Param       pageSize  query int    false "page size"
// @Success     200 {object} GetAllSongsResponse
// @Failure     500 {string} string "Internal error"
// @Router      /songs [get]
func (h *SongHandler) GetAllSongs(c *gin.Context) {
	group := c.Query("group")
	title := c.Query("song")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	filter := &domain.SongFilter{
		Group:    &group,
		Title:    &title,
		Page:     page,
		PageSize: pageSize,
	}

	songs, total, err := h.uc.GetAllSongs(c.Request.Context(), filter)
	if err != nil {
		h.log.WithError(err).Error("Error getting songs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Формируем ответ
	c.JSON(http.StatusOK, gin.H{
		"items": songs,
		"total": total,
	})
}

// @Summary     Получение текста песни
// @Description Возвращает текст песни с пагинацией по куплетам
// @Tags        songs
// @Param       id   path     int true "song id"
// @Param       page query    int false "page number (куплеты)"
// @Param       pageSize query int false "page size (кол-во куплетов на странице)"
// @Success     200 {object} GetSongTextResponse
// @Failure     404 {string} string "Not found"
// @Failure     500 {string} string "Internal error"
// @Router      /songs/{id}/text [get]
func (h *SongHandler) GetSongText(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	song, err := h.uc.GetSongByID(c.Request.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Error getting song")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if song == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Разбиваем текст на куплеты (по двум переносам строки, например)
	verses := splitVerses(song.Text)

	// Пагинация по куплетам
	pagedVerses := paginateVerses(verses, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"id":     song.ID,
		"group":  song.Group,
		"song":   song.Title,
		"verses": pagedVerses,
	})
}

// splitVerses - вспомогательная функция, которая разбивает текст на куплеты
func splitVerses(text string) []string {
	// Для простоты предположим, что куплеты разделены двойным переносом строки "\n\n"
	// Реальной логике может потребоваться более гибкий подход
	// import "strings"
	return strings.Split(text, "\n\n")
}

func paginateVerses(verses []string, page, pageSize int) []string {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 2
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(verses) {
		return []string{}
	}
	if end > len(verses) {
		end = len(verses)
	}
	return verses[start:end]
}

// @Summary     Удаление песни
// @Description Удаляет песню по id
// @Tags        songs
// @Param       id   path int true "song id"
// @Success     200
// @Failure     404 {string} string "Not found"
// @Failure     500 {string} string "Internal error"
// @Router      /songs/{id} [delete]
func (h *SongHandler) DeleteSong(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.uc.DeleteSong(c.Request.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Error deleting song")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.Status(http.StatusOK)
}

// @Summary     Изменение данных песни
// @Description Изменяет поля песни
// @Tags        songs
// @Accept      json
// @Produce     json
// @Param       id   path int true "song id"
// @Param       song body UpdateSongRequest true "Update Song"
// @Success     200 {object} domain.Song
// @Failure     404 {string} string "Not found"
// @Failure     500 {string} string "Internal error"
// @Router      /songs/{id} [put]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Сначала получим из БД, чтобы проверить, что песня существует
	song, err := h.uc.GetSongByID(c.Request.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Error getting song")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if song == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Обновляем поля
	if req.Group != nil {
		song.Group = *req.Group
	}
	if req.Song != nil {
		song.Title = *req.Song
	}
	if req.ReleaseDate != nil {
		// Преобразование формата даты (например, "2006-01-02")
		t, err := time.Parse("2006-01-02", *req.ReleaseDate)
		if err == nil {
			song.ReleaseDate = t
		}
	}
	if req.Text != nil {
		song.Text = *req.Text
	}
	if req.Link != nil {
		song.Link = *req.Link
	}

	err = h.uc.UpdateSong(c.Request.Context(), song)
	if err != nil {
		h.log.WithError(err).Error("Error updating song")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, song)
}

// @Summary     Добавление новой песни
// @Description Добавляет новую песню с запросом к внешнему API для получения releaseDate, текста и ссылки
// @Tags        songs
// @Accept      json
// @Produce     json
// @Param       input body CreateSongRequest true "Create Song"
// @Success     200 {object} domain.Song
// @Failure     400 {string} string "Bad request"
// @Failure     500 {string} string "Internal error"
// @Router      /songs [post]
func (h *SongHandler) CreateSong(c *gin.Context) {
	var req CreateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if req.Group == "" || req.Song == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group and song fields are required"})
		return
	}

	// Вызываем usecase для создания
	newSong, err := h.uc.CreateSong(c.Request.Context(), req.Group, req.Song)
	if err != nil {
		h.log.WithError(err).Error("Error creating song")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, newSong)
}

// DTO-шки
type CreateSongRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type UpdateSongRequest struct {
	Group       *string `json:"group"`
	Song        *string `json:"song"`
	ReleaseDate *string `json:"releaseDate"`
	Text        *string `json:"text"`
	Link        *string `json:"link"`
}

// Ответы для swagger
type GetAllSongsResponse struct {
	Items []domain.Song `json:"items"`
	Total int64         `json:"total"`
}

type GetSongTextResponse struct {
	ID     int64    `json:"id"`
	Group  string   `json:"group"`
	Song   string   `json:"song"`
	Verses []string `json:"verses"`
}
