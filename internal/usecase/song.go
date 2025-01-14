package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/yourusername/yourproject/internal/domain"
)

type MusicInfoAPIClient interface {
	GetSongInfo(group, song string) (*SongDetail, error)
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongUseCase struct {
	repo    domain.SongRepository
	infoAPI MusicInfoAPIClient
}

func NewSongUseCase(repo domain.SongRepository, infoAPI MusicInfoAPIClient) *SongUseCase {
	return &SongUseCase{
		repo:    repo,
		infoAPI: infoAPI,
	}
}

// Обогащение через внешний API и сохранение в БД
func (uc *SongUseCase) CreateSong(ctx context.Context, group, title string) (*domain.Song, error) {
	// 1. Забираем доп.инфу из внешнего API
	songInfo, err := uc.infoAPI.GetSongInfo(group, title)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get song info from external API")
	}

	// 2. Преобразуем дату (пример: "16.07.2006")
	layout := "02.01.2006" // день.месяц.год
	parsedDate, err := time.Parse(layout, songInfo.ReleaseDate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse release date")
	}

	// 3. Создаём запись в БД
	s := &domain.Song{
		Group:       group,
		Title:       title,
		ReleaseDate: parsedDate,
		Text:        songInfo.Text,
		Link:        songInfo.Link,
	}

	err = uc.repo.Create(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create song in DB")
	}

	return s, nil
}

func (uc *SongUseCase) GetSongByID(ctx context.Context, id int64) (*domain.Song, error) {
	return uc.repo.GetByID(id)
}

func (uc *SongUseCase) GetAllSongs(ctx context.Context, filter *domain.SongFilter) ([]domain.Song, int64, error) {
	return uc.repo.GetAll(filter)
}

func (uc *SongUseCase) UpdateSong(ctx context.Context, s *domain.Song) error {
	return uc.repo.Update(s)
}

func (uc *SongUseCase) DeleteSong(ctx context.Context, id int64) error {
	return uc.repo.Delete(id)
}

type musicInfoClient struct {
	baseURL string
}

func NewMusicInfoClient(baseURL string) MusicInfoAPIClient {
	return &musicInfoClient{baseURL: baseURL}
}

func (c *musicInfoClient) GetSongInfo(group, song string) (*SongDetail, error) {
	url := fmt.Sprintf("%s/info?group=%s&song=%s", c.baseURL, group, song)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var detail SongDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}
