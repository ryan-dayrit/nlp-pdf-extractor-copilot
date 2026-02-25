package store

import (
	"sync"
	"time"
)

type Document struct {
	ID         string
	Filename   string
	Content    []byte
	Status     string
	CreatedAt  time.Time
	DataPoints []DataPoint
	Results    []ExtractionResult
}

type DataPoint struct {
	Name        string
	Description string
}

type ExtractionResult struct {
	Name       string
	Value      string
	Confidence float64
}

type Store struct {
	mu   sync.RWMutex
	docs map[string]*Document
}

func New() *Store {
	return &Store{docs: make(map[string]*Document)}
}

func (s *Store) SaveDocument(doc *Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs[doc.ID] = doc
}

func (s *Store) GetDocument(id string) (*Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.docs[id]
	return doc, ok
}

func (s *Store) UpdateStatus(id, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, ok := s.docs[id]
	if !ok {
		return false
	}
	doc.Status = status
	return true
}

func (s *Store) SaveDataPoints(id string, dps []DataPoint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, ok := s.docs[id]
	if !ok {
		return false
	}
	doc.DataPoints = dps
	doc.Status = "processing"
	return true
}

func (s *Store) SaveResults(id string, results []ExtractionResult, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, ok := s.docs[id]
	if !ok {
		return false
	}
	doc.Results = results
	doc.Status = status
	return true
}

func (s *Store) ListDocuments() []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	docs := make([]*Document, 0, len(s.docs))
	for _, d := range s.docs {
		docs = append(docs, d)
	}
	return docs
}
