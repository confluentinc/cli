package config

type Config interface {
	Load() error
}
