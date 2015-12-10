package log5
import (
	"log"
	"os"
)

// 设置颜色刷
type Brush func(string) string

func NewBrush(color string) Brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = map[Level]Brush{
	Trace: NewBrush("1;32"), // Trace      cyan
	Info: NewBrush("1;34"), // Info		blue
	Warning: NewBrush("1;33"), // Warning    yellow
	Error: NewBrush("1;31"), // Error      red
	Fatal: NewBrush("1;37"), // Fatal		white

}

type ConsoleLog struct {
	log *log.Logger
	level Level
}

// 初始化控制台输出引擎
func NewConsole() LogEngine {
	return &ConsoleLog{
		log:    log.New(os.Stdout, "", log.Ldate|log.Ltime),
		level: Trace,
	}
}

func (c *ConsoleLog) Init() error {
	return nil
}

func (c *ConsoleLog) Write(msg string, level Level) error {
	if level < c.level {
		return nil
	}
	c.log.Println(colors[level](msg))
	return nil
}

func (c *ConsoleLog) Destroy() {}

func init() {
	Register("console", NewConsole)
}
