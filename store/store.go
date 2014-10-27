package store

type Config map[string]string

type SporeDockStore interface {
	Get(string) string
	Set(string) string
	Exists(string) bool
}
