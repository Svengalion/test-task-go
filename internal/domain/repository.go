package domain

type SongRepository interface {
	Create(song *Song) error
	GetByID(id int64) (*Song, error)
	GetAll(filter *SongFilter) ([]Song, int64, error)
	Update(song *Song) error
	Delete(id int64) error
}
