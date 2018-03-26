package firedb

import (
	"context"
	"sync"

	"cloud.google.com/go/firestore"
)

var mu sync.RWMutex
var db *firestore.Client

// SetUp is SetUp
func SetUp(ctx context.Context, projectID string) error {
	return createWithSetClient(ctx, projectID)
}

func createWithSetClient(ctx context.Context, projectID string) error {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	db = client

	return nil
}
