package shortener

import (
	"shortener/internal/generator"
	"shortener/internal/stat"
	"shortener/internal/storage/sqlite"
	"time"
)

type UseCase struct {
	db        *sqlite.Storage
	ClickStat []stat.ClickStat
}

func NewUseCase(db *sqlite.Storage) *UseCase {
	return &UseCase{db: db}
}

func (u *UseCase) CreateURL(size int) string {
	return generator.Generate(size)
}

func (u *UseCase) SaveURL(urlToSave, alias, userAgent string, timeClick time.Time) (int64, error) {
	return u.db.SaveURL(urlToSave, alias, userAgent, timeClick)
}

func (u *UseCase) UpdateURLStat(id int, userAgent string, timeClick time.Time) error {
	return u.db.UpdateURLStat(id, userAgent, timeClick)
}

func (u *UseCase) GetURL(alias string) (int, string, error) {
	return u.db.GetURL(alias)
}

func (u *UseCase) LoadURLStat(shortenerId int) (size int, err error) {
	u.ClickStat, err = u.db.GetURLStat(shortenerId)
	size = len(u.ClickStat)
	return size, err
}

func (u *UseCase) LoadURLStatByDate(shortenerId int, date time.Time, dateBy string) (size int, err error) {
	u.ClickStat, err = u.db.GetURLStatByDate(shortenerId, date, dateBy)
	size = len(u.ClickStat)
	return size, err
}

func (u *UseCase) LoadURLStatByUserAgent(shortenerId int, userAgent string) (size int, err error) {
	u.ClickStat, err = u.db.GetURLStatByUserAgent(shortenerId, userAgent)
	size = len(u.ClickStat)
	return size, err
}
