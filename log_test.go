package log5

import "testing"

func Test_Console(t *testing.T) {
	// 初始化
	log := NewLog(1000)
	// 设置输出引擎
	log.SetEngine("console")
	// 设置是否输出行号
	log.SetFuncCall(true)

	log.Trac("Trac")
	log.Info("Info")
	log.Warn("Warning")
	log.Error("Error")
	log.Fatal("Fatal")
}

func Benchmark_Console(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 初始化
		log := NewLog(100)
		// 设置输出引擎
		log.SetEngine("console")

		log.Trac("%d", i)

	}
}

