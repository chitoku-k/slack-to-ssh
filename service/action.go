package service

import (
	"context"
)

type actionService struct {
	Executor ActionExecutor
}

type ActionService interface {
	Execute(ctx context.Context, name string) ([]byte, error)
}

type ActionExecutor interface {
	Do(ctx context.Context, name string) ([]byte, error)
}

func NewActionService(executor ActionExecutor) ActionService {
	return &actionService{
		Executor: executor,
	}
}

func (as *actionService) Execute(ctx context.Context, name string) ([]byte, error) {
	return as.Executor.Do(ctx, name)
}
