package service

type actionService struct {
	Executor ActionExecutor
}

type ActionService interface {
	Execute(name string) ([]byte, error)
}

type ActionExecutor interface {
	Do(name string) ([]byte, error)
}

func NewActionService(executor ActionExecutor) ActionService {
	return &actionService{
		Executor: executor,
	}
}

func (as *actionService) Execute(name string) ([]byte, error) {
	return as.Executor.Do(name)
}
