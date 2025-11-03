package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"shortener/internal/stat"
	"shortener/internal/storage"
	"shortener/pkg/logger"
	"time"

	"github.com/mattn/go-sqlite3"
)

const (
	day   = "D"
	month = "M"
	year  = "Y"
)

type Storage struct {
	db *sql.DB
	l  logger.Interface
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
	  CREATE TABLE IF NOT EXISTS shortener(
	   id        INTEGER PRIMARY KEY,
	   alias     TEXT NOT NULL UNIQUE,
	   shortener TEXT NOT NULL)
   `)
	if err != nil {
		return nil, fmt.Errorf("%s: create shortener table: %w", op, err)
	}

	_, err = db.Exec(`
	  CREATE TABLE IF NOT EXISTS clicks(
	   id INTEGER PRIMARY KEY AUTOINCREMENT,
	   shortener_id INTEGER NOT NULL,
	   userAgent TEXT NOT NULL,
	   timeClick DATETIME NOT NULL,
	   FOREIGN KEY(shortener_id) REFERENCES shortener(id) ON DELETE CASCADE)`)
	if err != nil {
		return nil, fmt.Errorf("%s: create clicks table: %w", op, err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_alias ON shortener(alias)`)
	if err != nil {
		return nil, fmt.Errorf("%s: create alias index: %w", op, err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_shortener_id ON clicks(shortener_id)`)
	if err != nil {
		return nil, fmt.Errorf("%s: create shortener_id index: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias, userAgent string, timeClick time.Time) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO shortener(shortener,alias) values(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: execute statement %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert ID: %w", op, err)
	}

	statmt, err := s.db.Prepare("INSERT INTO clicks(shortener_id,userAgent,timeClick) values(?,?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = statmt.Close()
		if err != nil {
			s.l.Error("Failed to close statmt", "error", err)
		}
	}()

	_, err = statmt.Exec(id, userAgent, timeClick)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement %w", op, err)
	}

	return id, nil
}

func (s *Storage) UpdateURLStat(id int, userAgent string, timeClick time.Time) error {
	const op = "storage.sqlite.UpdateURLStat"

	stmt, err := s.db.Prepare("INSERT INTO clicks(shortener_id, userAgent, timeClick) VALUES(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	_, err = stmt.Exec(id, userAgent, timeClick)
	if err != nil {
		return fmt.Errorf("%s: execute statement %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (int, string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT id,shortener FROM shortener WHERE alias = ?")
	if err != nil {
		return -1, "", fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	var id int
	var resURL string

	err = stmt.QueryRow(alias).Scan(&id, &resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, "", storage.ErrURLNotFound
	}
	if err != nil {
		return -1, "", fmt.Errorf("%s: execute statement %w", op, err)
	}

	return id, resURL, nil
}

func (s *Storage) GetURLStat(shortenerId int) ([]stat.ClickStat, error) {
	const op = "storage.sqlite.GetURLStat"

	stmt, err := s.db.Prepare("SELECT id, userAgent, timeClick FROM clicks WHERE shortener_id=?")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	rows, err := stmt.Query(shortenerId)
	if err != nil {
		return nil, fmt.Errorf("%s: execute query %w", op, err)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			s.l.Error("Failed to close rows", "error", err)
		}
	}()

	var stats []stat.ClickStat
	for rows.Next() {
		var st stat.ClickStat
		if err = rows.Scan(&st.ID, &st.UserAgent, &st.TimeClick); err != nil {
			return nil, fmt.Errorf("%s: scan row %w", op, err)
		}
		stats = append(stats, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration %w", op, err)
	}

	return stats, nil
}

func (s *Storage) GetURLStatByDate(shortenerId int, date time.Time, dateBy string) ([]stat.ClickStat, error) {
	const op = "storage.sqlite.GetURLStat"

	stmt, err := s.db.Prepare("SELECT id, userAgent, timeClick FROM clicks WHERE shortener_id=? AND timeClick>=? and timeClick<?")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()
	var endDate time.Time
	if dateBy == year {
		date = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
		endDate = date.AddDate(1, 0, 0)
	} else if dateBy == month {
		date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		endDate = date.AddDate(0, 1, 0)
	} else {
		endDate = date.AddDate(0, 0, 1)
	}

	rows, err := stmt.Query(shortenerId, date, endDate)
	if err != nil {
		return nil, fmt.Errorf("%s: execute query %w", op, err)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			s.l.Error("Failed to close rows", "error", err)
		}
	}()

	var stats []stat.ClickStat
	for rows.Next() {
		var st stat.ClickStat
		if err = rows.Scan(&st.ID, &st.UserAgent, &st.TimeClick); err != nil {
			return nil, fmt.Errorf("%s: scan row %w", op, err)
		}
		stats = append(stats, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration %w", op, err)
	}

	return stats, nil
}

func (s *Storage) GetURLStatByUserAgent(shortenerId int, userAgent string) ([]stat.ClickStat, error) {
	const op = "storage.sqlite.GetURLStat"

	stmt, err := s.db.Prepare("SELECT id, userAgent, timeClick FROM clicks WHERE shortener_id=? AND userAgent=?")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	rows, err := stmt.Query(shortenerId, userAgent)
	if err != nil {
		return nil, fmt.Errorf("%s: execute query %w", op, err)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			s.l.Error("Failed to close rows", "error", err)
		}
	}()

	var stats []stat.ClickStat
	for rows.Next() {
		var st stat.ClickStat
		if err = rows.Scan(&st.ID, &st.UserAgent, &st.TimeClick); err != nil {
			return nil, fmt.Errorf("%s: scan row %w", op, err)
		}
		stats = append(stats, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration %w", op, err)
	}

	return stats, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM shortener WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement %w", op, err)
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			s.l.Error("Failed to close stmt", "error", err)
		}
	}()

	_, err = stmt.Exec(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrURLNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: execute statement %w", op, err)
	}
	return nil
}
