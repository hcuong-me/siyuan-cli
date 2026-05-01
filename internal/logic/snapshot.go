// Package logic provides snapshot business logic.
package logic

import (
	"context"
	"fmt"

	"siyuan/internal/siyuan"
)

// SnapshotLogic handles snapshot business logic.
type SnapshotLogic struct {
	client *siyuan.Client
}

// Snapshot represents a repository snapshot.
type Snapshot struct {
	ID      string `json:"id"`
	Memo    string `json:"memo"`
	Created int64  `json:"created"`
	Count   int    `json:"count"`
	Size    int64  `json:"size"`
}

// NewSnapshotLogic creates a new SnapshotLogic.
func NewSnapshotLogic() (*SnapshotLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &SnapshotLogic{client: c}, nil
}

// List returns all snapshots.
func (l *SnapshotLogic) List(ctx context.Context) ([]Snapshot, error) {
	infos, err := l.client.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	snapshots := make([]Snapshot, len(infos))
	for i, info := range infos {
		snapshots[i] = Snapshot{
			ID:      info.ID,
			Memo:    info.Memo,
			Created: info.Created,
			Count:   info.Count,
			Size:    info.Size,
		}
	}

	return snapshots, nil
}

// Current returns the current snapshot.
func (l *SnapshotLogic) Current(ctx context.Context) (*Snapshot, error) {
	info, err := l.client.GetCurrentSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		ID:      info.ID,
		Memo:    info.Memo,
		Created: info.Created,
		Count:   info.Count,
		Size:    info.Size,
	}, nil
}

// Create creates a new snapshot.
func (l *SnapshotLogic) Create(ctx context.Context, memo string) (*Snapshot, error) {
	info, err := l.client.CreateSnapshot(ctx, memo)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		ID:      info.ID,
		Memo:    info.Memo,
		Created: info.Created,
		Count:   info.Count,
		Size:    info.Size,
	}, nil
}

// Restore restores to a specific snapshot.
func (l *SnapshotLogic) Restore(ctx context.Context, id string) error {
	return l.client.RestoreSnapshot(ctx, id)
}

// Remove removes a snapshot.
func (l *SnapshotLogic) Remove(ctx context.Context, id string) error {
	return l.client.RemoveSnapshot(ctx, id)
}

// ValidateSnapshotID checks if the ID is valid.
func ValidateSnapshotID(id string) error {
	if id == "" {
		return fmt.Errorf("snapshot ID is required")
	}
	return nil
}
