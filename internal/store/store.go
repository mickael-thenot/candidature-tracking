package store

import (
	"errors"
	"uuid"

	"github.com/mickael-thenot/candidature-tracker/internal/model"
)

var ErrNotFound = errors.New("candidature not found")

type Store interface {
	Save(c model.Candidature) error
	Get(id uuid.UUID) (model.Candidature, error)
	List() ([]model.Candidature, error)
	Delete(id uuid.UUID) error
}
