package ilog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type iLog struct {
	l    *log.Logger
	mtx  sync.Mutex
	f    *os.File
	t    time.Time
	name string
}

var logDir string
var debugL *iLog
var rinfoL *iLog
var errorL *iLog
var fatalL *iLog
var warningL *iLog

func Init(dir string) error {
	if dir == "" {
		return errors.New("dir is empty")
	}
	logDir = dir
	debugL = &iLog{name: "debug"}
	rinfoL = &iLog{name: "rinfo"}
	errorL = &iLog{name: "error"}
	fatalL = &iLog{name: "fatal"}
	warningL = &iLog{name: "warning"}
	return os.MkdirAll(logDir, 0755)
}
func Debug(format string, v ...any)   { debugL.write(format, v...) }
func Rinfo(format string, v ...any)   { rinfoL.write(format, v...) }
func Error(format string, v ...any)   { errorL.write(format, v...) }
func Fatal(format string, v ...any)   { fatalL.write(format, v...) }
func Warning(format string, v ...any) { warningL.write(format, v...) }

func (l *iLog) check() error {
	t := time.Now()
	if l.l == nil || t.Day() != l.t.Day() {
		l.close()
		if err := l.open(&t); err != nil {
			return err
		}
	}
	return nil
}
func (l *iLog) open(t *time.Time) (err error) {
	l.t = *t
	fname := fmt.Sprintf("%s-%d-%02d-%02d.log", l.name, l.t.Year(), l.t.Month(), l.t.Day())
	logFile := path.Join(logDir, fname)
	l.f, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.l = log.New(l.f, "", log.Ldate|log.Lmicroseconds)
	return nil
}
func (l *iLog) close() {
	if l.f != nil {
		l.f.Close()
	}
	l.f = nil
	l.l = nil
}
func (l *iLog) write(format string, v ...any) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if err := l.check(); err != nil {
		return
	}
	if len(v) == 0 {
		l.l.Print(format)
		return
	}
	l.l.Printf(format, v...)
}
