package log5

import (
	"testing"
)

func Test_Console(t *testing.T) {
	// 初始化
	log := NewLog(1000)
	// 设置输出引擎
	log.SetEngine("file", `{"spilt":"size", "filename":"logs/test.log"}`)
	// 设置是否输出行号
	log.SetFuncCall(true)

	// 设置log级别
	log.SetLevel(Warning)

	log.Trac("Trac")
	log.Info("Info")
	log.Warn("Warning")
	log.Error("Error")
	log.Fatal("Fatal")
}
