package stores

import (
	"errors"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/toboshii/hajimari/internal/models"
)

var startpages []*models.Startpage

type memoryStore struct {
	StartpageStore
}

func NewMemoryStore() StartpageStore {
	return &memoryStore{}
}

func (s *memoryStore) NewStartpage(startpage *models.Startpage) (string, error) {
	id, err := gonanoid.New()
	if err != nil {
		return "", err
	}
	startpage.ID = id
	startpages = append(startpages, startpage)
	return startpage.ID, nil
}

func (s *memoryStore) GetStartpage(id string) (*models.Startpage, error) {
	for _, s := range startpages {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, errors.New("startpage not found")
}

func (s *memoryStore) UpdateStartpage(id string, startpage *models.Startpage) (*models.Startpage, error) {
	for i, s := range startpages {
		if s.ID == id {
			startpages[i] = startpage
			return startpage, nil
		}
	}
	return nil, errors.New("startpage not found")
}

func (s *memoryStore) RemoveStartpage(id string) (*models.Startpage, error) {
	for i, s := range startpages {
		if s.ID == id {
			startpages = append((startpages)[:i], (startpages)[i+1:]...)
			return s, nil
		}
	}
	return nil, errors.New("startpage not found")
}
