package goorm

type GoormConfig struct {
	Driver Driver
	Logger Logger
	DSN    string
}
