package task

type Task interface {
	Name() string
	Run() error
	Rollback() error
}
