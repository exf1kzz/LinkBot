package files

import (
	"LinkBot/lib/e"
	"LinkBot/storage"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0774

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(ctx context.Context, page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	fPath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := FileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(ctx context.Context, userName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.DecodePage(filepath.Join(path, file.Name()))
}

func (s Storage) List(ctx context.Context, userName string) ([]storage.Page, error) {
	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, e.Wrap("can't list pages", err)
	}

	pages := make([]storage.Page, 0, len(files))
	for _, file := range files {
		page, err := s.DecodePage(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		pages = append(pages, *page)
	}

	if len(pages) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	return pages, nil
}

func (s Storage) Remove(ctx context.Context, p *storage.Page) error {
	fileName, err := FileName(p)
	if err != nil {
		return e.Wrap("can't remove page", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file %s", path)
		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) Exists(ctx context.Context, p *storage.Page) (bool, error) {
	fileName, err := FileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file %s exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", path)
		return false, e.Wrap(msg, err)
	}

	return true, nil
}

func (s Storage) DecodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	defer func() { _ = f.Close() }()

	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func FileName(p *storage.Page) (string, error) {
	return p.Hash()
}
