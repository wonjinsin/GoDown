package service

// FileUsecase ...
type FileUsecase interface {
	Do(c chan int) (err error)
}
