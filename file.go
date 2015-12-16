package log5

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
	"path/filepath"
)

type SplitType string

const (
	SPLIT_TYPE_BY_SIZE  SplitType = "size"
	SPLIT_TYPE_BY_DAILY SplitType = "daily"
)

const (
	DEFAULT_LEVEL     Level = Trace
	DEFAULT_FILE_SIZE       = 30
)

type FileLog struct {
	log       *log.Logger
	Level     Level     `json:"level"`
	FileName  string    `json:"filename"`
	MaxSize   int64     `json:"maxsize"` //MB
	SplitType SplitType `json:"split"`   //拆分方式,2种(1-天数,2-文件大小)
	suffix    string    //后缀,方便分割log

	date time.Time

	lock sync.Mutex

	mw *MuxWriter
}

type MuxWriter struct {
	mu sync.RWMutex
	fd *os.File
}

func (f *MuxWriter) Write(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fd.Write(b)
}

func NewFile() LogEngine {
	t, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	fl := &FileLog{
		Level:     DEFAULT_LEVEL,
		FileName:  "",
		MaxSize:   DEFAULT_FILE_SIZE,
		SplitType: SPLIT_TYPE_BY_SIZE,
		suffix:    "",
		date:      t,
	}

	fl.mw = new(MuxWriter)
	fl.log = log.New(fl.mw, "", log.Ldate|log.Ltime)

	return fl
}

func (l *FileLog) Init(conf string) error {

	if len(conf) != 0 {
		err := json.Unmarshal([]byte(conf), l)
		if err != nil {
			return err
		}
	}

	if len(l.FileName) == 0 {
		name := path.Join("log", "log"+".log")
		l.FileName = name
	}

	return l.initLog()
}

func (l *FileLog) initLog() error {
	fd, err := l.createFile()
	if err != nil {
		return err
	}
	if l.mw.fd != nil {
		l.mw.fd.Close()
	}
	l.mw.fd = fd

	// 判断log分割类型
	switch l.SplitType {
	case SPLIT_TYPE_BY_SIZE:
		return l.splitbysize()
	case SPLIT_TYPE_BY_DAILY:
		return l.splitbydaily()
	default:
		return l.splitbysize()
	}
}

// 按文件大小分割
func (l *FileLog) splitbysize() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 检查文件大小
	finfo, err := os.Stat(l.FileName)
	if err != nil {
		return fmt.Errorf("get %s stat err: %s\n", l.FileName, err)
	}

	if finfo.Size() >= l.MaxSize<<10<<10 {
		suffix, err := strconv.Atoi(l.suffix)
		if err != nil {
			suffix = 1
		}
		suffix += 1
		l.FileName += "." + strconv.Itoa(suffix)
		fd, err := l.createFile()
		if err != nil {
			return err
		}
		l.mw.fd = fd
	}
	return nil
}

// 按天数分割
func (l *FileLog) splitbydaily() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	t, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))

	if l.date.Before(t) {
		l.suffix = time.Now().Format(time.RFC3339)
		l.FileName += "." + l.suffix
		fd, err := l.createFile()
		if err != nil {
			return err
		}
		l.mw.fd = fd
	}

	return nil
}

// 创建文件
func (l *FileLog) createFile() (*os.File, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	dir := filepath.Dir(l.FileName)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(l.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
}

func (l *FileLog) Write(msg string, level Level) error {
	return nil
}

func (l *FileLog) Destroy() {
	l.mw.fd.Close()
}

func (l *FileLog) Flush() {
	l.mw.fd.Sync()
}

func init() {
	Register("file", NewFile)
}
