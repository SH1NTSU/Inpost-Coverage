package cache

import (
	"context"
	"encoding/json"
)

type Cache interface {
	Get(ctx context.Context, key string, dst any) (bool, error)
	Set(ctx context.Context, key string, val any) error
	Delete(ctx context.Context, key string) error
}

func marshal(val any) ([]byte, error) { return json.Marshal(val) }
func unmarshal(buf []byte, dst any) error {
	return json.Unmarshal(buf, dst)
}
