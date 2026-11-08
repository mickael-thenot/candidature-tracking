package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"uuid"

	"github.com/mickael-thenot/candidature-tracker/internal/model"
)

// JSONStore implémente Store en persistant les données dans un fichier JSON.
type JSONStore struct {
	mu       sync.RWMutex
	filePath string
	data     map[uuid.UUID]model.Candidature // map pour un accès rapide par ID
}

// NewJSONStore initialise le store et charge les données existantes.
func NewJSONStore(path string) (*JSONStore, error) {
	s := &JSONStore{
		filePath: path,
		data:     make(map[uuid.UUID]model.Candidature),
	}

	// Charge le fichier s'il existe
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load lit le fichier et remplit la map.
func (s *JSONStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Le fichier n'existe pas encore, c'est normal au premier lancement.
			return nil
		}
		return fmt.Errorf("read store file: %w", err)
	}

	// Si le fichier est vide, on évite une erreur de unmarshal
	if len(data) == 0 {
		return nil
	}

	var candidatures []model.Candidature
	if err := json.Unmarshal(data, &candidatures); err != nil {
		return fmt.Errorf("decode store file: %w", err)
	}

	for _, c := range candidatures {
		// On valide les données chargées pour éviter de propager des données corrompues
		if err := c.Validate(); err != nil {
			return fmt.Errorf("invalid candidature in store: %w", err)
		}
		s.data[c.ID] = c
	}
	return nil
}

// saveLocked écrit l'état actuel de la map dans le fichier.
// Le verrou (s.mu) doit être acquis avant d'appeler cette méthode.
func (s *JSONStore) saveLocked() error {
	// Convertir la map en slice
	candidatures := make([]model.Candidature, 0, len(s.data))
	for _, c := range s.data {
		candidatures = append(candidatures, c)
	}

	// Sérialiser avec indentation pour la lisibilité
	data, err := json.MarshalIndent(candidatures, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store data: %w", err)
	}

	// S'assurer que le dossier de destination existe
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}

	// Définir le chemin du fichier temporaire
	tmpPath := s.filePath + ".tmp"

	// Écrire dans le fichier temporaire
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temporary store file: %w", err)
	}

	// Renommer le fichier temporaire (opération atomique)
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		// Nettoyage si le renommage échoue
		os.Remove(tmpPath)
		return fmt.Errorf("rename temporary store file: %w", err)
	}

	return nil
}

// Save implémente l'interface Store.
func (s *JSONStore) Save(c model.Candidature) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[c.ID] = c
	return s.saveLocked()
}

// Get implémente l'interface Store.
func (s *JSONStore) Get(id uuid.UUID) (model.Candidature, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.data[id]
	if !ok {
		return model.Candidature{}, ErrNotFound
	}
	return c, nil
}

// List implémente l'interface Store.
func (s *JSONStore) List() ([]model.Candidature, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]model.Candidature, 0, len(s.data))
	for _, c := range s.data {
		res = append(res, c)
	}

	// Trier par date de candidature, de la plus récente à la plus ancienne.
	// En cas d'égalité, on trie par ID (qui est ordonné dans le temps avec UUID v7).
	slices.SortFunc(res, func(a, b model.Candidature) int {
		if !a.AppliedAt.Equal(b.AppliedAt) {
			return b.AppliedAt.Compare(a.AppliedAt) // b avant a = tri décroissant
		}
		return bytes.Compare(b.ID[:], a.ID[:])
	})

	return res, nil
}

// Delete implémente l'interface Store.
func (s *JSONStore) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return s.saveLocked()
}
