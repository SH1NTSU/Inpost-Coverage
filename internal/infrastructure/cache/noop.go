package cache

import "context"

type Noop struct{}

func (Noop) Get(context.Context, string, any) (bool, error) { return false, nil }
func (Noop) Set(context.Context, string, any) error         { return nil }
func (Noop) Delete(context.Context, string) error           { return nil }
