package log5

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"
	"sync"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

type Level uint

const (
	Trace Level = iota
	Info
	Warning
	Error
	Fatal
)

// log输出接口
type LogEngine interface {
	Init() error                         //初始化
	Write(msg string, level Level) error //写入
	Destroy()
}

// log结构体
type Log struct {
	level         Level
	msg           chan *logMsg
	trackFuncCall bool //是否追踪调用函数
	funcCallDepth int
	output        map[string]LogEngine
	lock          sync.Mutex
}

// log内容
type logMsg struct {
	level Level
	msg   string
}

// 定义输出引擎字典
type engineType func() LogEngine

var engines = make(map[string]engineType)

// 注册引擎
func Register(name string, log engineType) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := engines[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	engines[name] = log
}

// 初始化log
// adaptername -- 适配名称 为空(默认)console
// chanlen -- 缓存大小
func NewLog(chanlen uint64) *Log {
	l := &Log{
		level:         Trace,
		trackFuncCall: false,
		funcCallDepth: 2,
		msg:           make(chan *logMsg, chanlen),
		output:        make(map[string]LogEngine),
	}

	l.SetEngine("console")

	return l
}

// 设置是否输出行号
func (l *Log) SetFuncCall(bool) {
	l.trackFuncCall = true
}

// 设置是否输出行号
func (l *Log) SetFuncCallDepth(depth int) {
	l.funcCallDepth = depth
}

// 设置输出引擎
func (l *Log) SetEngine(engname string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	//获取引擎
	if log, ok := engines[engname]; ok {
		lg := log()
		err := lg.Init()
		if err != nil {
			errmsg := fmt.Errorf("SetEngine error: %s", err)
			return errmsg
		}

		l.output[engname] = lg
	} else {
		return fmt.Errorf("unknown Enginee %q ", engname)
	}

	return nil
}

// 初始化logMsg
func (l *Log) newMsg(msg string ,level Level ) {
	l.lock.Lock()
	defer l.lock.Unlock()

	lm := new(logMsg)
	lm.level = level

	if l.trackFuncCall {
		_, file, line, ok := runtime.Caller(l.funcCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		lm.msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	} else {
		lm.msg = msg
	}

	l.msg <- lm

}

// 写入
func (l *Log) write() {
	lm := <-l.msg
	for name, e := range l.output {
		err := e.Write(lm.msg, lm.level)
		if err != nil {
			fmt.Println("ERROR, unable to WriteMsg:", name, err)
		}
	}
}


// 获取调用的位置
func (l *Log) getInvokerLocation() string {
	//runtime.Caller(skip) skip=0 返回当前调用Caller函数的函数名、文件、程序指针PC，1是上一层函数，以此类推
	pc, file, line, ok := runtime.Caller(l.funcCallDepth)
	if !ok {
		return ""
	}
	fname := ""
	if index := strings.LastIndex(file, "/"); index > 0 {
		fname = file[index+1 : len(file)]
	}
	funcPath := ""
	funcPtr := runtime.FuncForPC(pc)
	if funcPtr != nil {
		funcPath = funcPtr.Name()
	}
	return fmt.Sprintf("%s : [%s:%d]", funcPath, fname, line)
}

// Trace
func (l *Log) Trac(format string, v ...interface{}) {
	// 如果设置的级别比 trace 级别高,不输出
	if l.level > Trace {
		return
	}
	msg := fmt.Sprintf("[TRAC] "+format, v...)
	l.newMsg(msg, Trace)
	l.write()
}

// INFO
func (l *Log) Info(format string, v ...interface{}) {
	if l.level > Info {
		return
	}
	msg := fmt.Sprintf("[INFO] "+format, v...)
	l.newMsg(msg, Info)
	l.write()
}

//WARNING
func (l *Log) Warn(format string, v ...interface{}) {
	if l.level > Warning {
		return
	}
	msg := fmt.Sprintf("[WARN] "+format, v...)
	l.newMsg(msg, Warning)
	l.write()
}

// ERROR
func (l *Log) Error(format string, v ...interface{}) {
	if l.level > Error {
		return
	}
	msg := fmt.Sprintf("[ERRO] "+format, v...)
	l.newMsg(msg, Error)
	l.write()
}

// FATAL
func (l *Log) Fatal(format string, v ...interface{}) {
	if l.level > Fatal {
		return
	}
	msg := fmt.Sprintf("[FATA] "+format, v...)
	l.newMsg(msg, Fatal)
	l.write()
}
