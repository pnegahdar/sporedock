package service

type Config map[string]string

type SporeDockService interface {
	Init()
	Run()
}
