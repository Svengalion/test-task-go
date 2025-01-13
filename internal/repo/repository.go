package repo

type SongRepository interface {
	Create(song *Song) error
	GetByID(id int64) (*Song, error)
	GetAll(filter *SongFilter) ([]Song, int64, error) // Возвращаем список и общее кол-во для пагинации
	Update(song *Song) error
	Delete(id int64) error
}
