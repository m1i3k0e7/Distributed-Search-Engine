package config

import (
	"path"
	"runtime"
)

var (
	RootPath string
)

func init() {
	RootPath = path.Dir(GetCurrentPath()+"..") + "/"
}

func GetCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1) // 0:GetCurrentPath, 1: GetCurrentPath caller, 2: GetCurrentPath caller's caller...
	return path.Dir(filename)
}
