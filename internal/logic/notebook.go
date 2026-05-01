// Package logic provides business logic for SiYuan operations.
package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"siyuan/internal/siyuan"
)

// NotebookLogic handles notebook business logic.
type NotebookLogic struct {
	client *siyuan.Client

	// Cache for notebook lookups
	cache       []siyuan.Notebook
	cacheTime   time.Time
	cacheMu     sync.RWMutex
	cacheExpiry time.Duration
}

// NewNotebookLogic creates a new NotebookLogic.
func NewNotebookLogic() (*NotebookLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &NotebookLogic{
		client:      c,
		cacheExpiry: 30 * time.Second,
	}, nil
}

// SetCacheExpiry sets the cache expiry duration (for testing).
func (l *NotebookLogic) SetCacheExpiry(d time.Duration) {
	l.cacheExpiry = d
}

// List returns all notebooks.
func (l *NotebookLogic) List(ctx context.Context) ([]siyuan.Notebook, error) {
	// Check cache first
	l.cacheMu.RLock()
	if l.cache != nil && time.Since(l.cacheTime) < l.cacheExpiry {
		cached := l.cache
		l.cacheMu.RUnlock()
		return cached, nil
	}
	l.cacheMu.RUnlock()

	// Fetch from API
	resp, err := l.client.ListNotebooks(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	l.cacheMu.Lock()
	l.cache = resp.Notebooks
	l.cacheTime = time.Now()
	l.cacheMu.Unlock()

	return resp.Notebooks, nil
}

// Create creates a new notebook.
func (l *NotebookLogic) Create(ctx context.Context, name string) (*siyuan.Notebook, error) {
	notebook, err := l.client.CreateNotebook(ctx, name)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	l.cacheMu.Lock()
	l.cache = nil
	l.cacheMu.Unlock()

	return notebook, nil
}

// Rename renames a notebook.
func (l *NotebookLogic) Rename(ctx context.Context, notebookID, name string) error {
	err := l.client.RenameNotebook(ctx, notebookID, name)
	if err != nil {
		return err
	}

	// Invalidate cache
	l.cacheMu.Lock()
	l.cache = nil
	l.cacheMu.Unlock()

	return nil
}

// Remove removes a notebook.
func (l *NotebookLogic) Remove(ctx context.Context, notebookID string) error {
	err := l.client.RemoveNotebook(ctx, notebookID)
	if err != nil {
		return err
	}

	// Invalidate cache
	l.cacheMu.Lock()
	l.cache = nil
	l.cacheMu.Unlock()

	return nil
}

// Open opens a closed notebook.
func (l *NotebookLogic) Open(ctx context.Context, notebookID string) error {
	return l.client.OpenNotebook(ctx, notebookID)
}

// Close closes an open notebook.
func (l *NotebookLogic) Close(ctx context.Context, notebookID string) error {
	return l.client.CloseNotebook(ctx, notebookID)
}

// FindByName finds a notebook by name (case-insensitive).
func (l *NotebookLogic) FindByName(ctx context.Context, name string) (*siyuan.Notebook, error) {
	notebooks, err := l.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, nb := range notebooks {
		if nb.Name == name {
			return &nb, nil
		}
	}

	return nil, fmt.Errorf("notebook not found: %s", name)
}

// FindByIDOrName finds a notebook by ID or name.
func (l *NotebookLogic) FindByIDOrName(ctx context.Context, idOrName string) (*siyuan.Notebook, error) {
	notebooks, err := l.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, nb := range notebooks {
		if nb.ID == idOrName || nb.Name == idOrName {
			return &nb, nil
		}
	}

	return nil, fmt.Errorf("notebook not found: %s", idOrName)
}
