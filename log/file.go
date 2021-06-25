package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wwq-2020/go.common/errors"
)

// RotateType RotateType
type RotateType int

// RotateTypes
const (
	PeriodRotateType RotateType = iota
	SizeRotateType
	defaultPeriodRotateArg         = "24h"
	defaultSizeRotateArg           = "1gb"
	defaultSizeRotateArgInt        = 1 << 30
	logIndexFile                   = ".logindex"
	b                       uint64 = 1
	kb                      uint64 = 1 << 10
	mb                      uint64 = 1 << 20
)

var (
	reg = regexp.MustCompile(`(\d{1,})(m|mb|k|kb|b|g|gb)`)
)

// RotateConf RotateConf
type RotateConf struct {
	RotateType RotateType
	RotateArg  string
}

func (rc *RotateConf) fill() {
	if rc.RotateType != PeriodRotateType &&
		rc.RotateType != SizeRotateType {
		rc.RotateType = PeriodRotateType
	}
	if rc.RotateType == PeriodRotateType &&
		rc.RotateArg == "" {
		rc.RotateArg = defaultPeriodRotateArg
	}
	if rc.RotateType == SizeRotateType &&
		rc.RotateArg == "" {
		rc.RotateArg = defaultSizeRotateArg
	}
}

// NewFileLog NewFileLog
func NewFileLog(file string) Logger {
	return NewFileLogEx(file, nil)
}

// NewFileLogEx NewFileLogEx
func NewFileLogEx(file string, rotateConf *RotateConf) Logger {
	output := buildRotatedOutput(file, rotateConf)
	return New(WithOutput(output))
}

type periodRotatedOutput struct {
	file   string
	output *os.File
	d      time.Duration
	bytes  uint64
	sync.Mutex
}

func (p *periodRotatedOutput) run() {
	ticker := time.NewTicker(p.d)
	for range ticker.C {
		p.swapOutpout()
	}
}

func (p *periodRotatedOutput) swapOutpout() {
	p.Lock()
	defer p.Unlock()
	nextFile := getNextFile(p.file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Error(err)
		return
	}
	if err := p.output.Sync(); err != nil {
		Error(err)
	}
	if err := p.output.Close(); err != nil {
		Error(err)
	}
	p.output = f
}

func (p *periodRotatedOutput) Write(b []byte) (int, error) {
	p.Lock()
	defer p.Unlock()
	n, err := p.output.Write(b)
	if err != nil {
		return n, errors.Trace(err)
	}
	return n, nil
}

func (p *periodRotatedOutput) Close() error {
	p.Lock()
	defer p.Unlock()
	if err := p.output.Close(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

type sizeRotatedOutput struct {
	file   string
	output *os.File
	size   uint64
	bytes  uint64
	sync.Mutex
}

func (s *sizeRotatedOutput) Close() error {
	s.Lock()
	defer s.Unlock()
	if err := s.output.Close(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *sizeRotatedOutput) Write(b []byte) (int, error) {
	s.Lock()
	defer s.Unlock()
	if s.bytes >= s.size {
		s.swapOutpout()
	}
	n, err := s.output.Write(b)
	s.bytes += uint64(n)
	if err != nil {
		return n, errors.Trace(err)
	}
	return n, nil
}

func (s *sizeRotatedOutput) swapOutpout() {
	nextFile := getNextFile(s.file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Error(err)
		return
	}
	if err := s.output.Sync(); err != nil {
		Error(err)
	}
	if err := s.output.Close(); err != nil {
		Error(err)
	}
	s.output = f
	s.bytes = 0
}

func buildRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
	if rotateConf == nil {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			Fatalf("failed to OpenFile:%s,err:%v", file, err)
		}
		return f
	}
	rotateConf.fill()
	switch rotateConf.RotateType {
	case PeriodRotateType:
		return buildPeriodRotatedOutput(file, rotateConf)
	case SizeRotateType:
		return buildSizeRotatedOutput(file, rotateConf)
	default:
		return buildPeriodRotatedOutput(file, rotateConf)
	}
}

func buildPeriodRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
	d, err := time.ParseDuration(rotateConf.RotateArg)
	if err != nil {
		d, err = time.ParseDuration(defaultPeriodRotateArg)
		if err != nil {
			Fatalf("failed to ParseDuration:%s,err:%v", defaultPeriodRotateArg, err)
		}
	}
	nextFile := getNextFile(file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Fatalf("failed to OpenFile:%s,err:%v", nextFile, err)
	}
	p := &periodRotatedOutput{file: file, output: f, d: d}
	go p.run()
	return p
}

func parseSize(arg string) uint64 {
	arg = strings.ToLower(arg)
	result := reg.FindStringSubmatch(arg)
	if len(result) != 3 {
		return defaultSizeRotateArgInt
	}
	num, err := strconv.ParseUint(result[1], 10, 64)
	if err != nil {
		return defaultSizeRotateArgInt
	}
	var unit uint64
	switch result[2] {
	case "b":
		unit = b
	case "kb", "k":
		unit = kb
	case "mb", "m":
		unit = mb
	default:
		return defaultSizeRotateArgInt
	}
	return num * unit
}

func buildSizeRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
	nextFile := getNextFile(file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Fatalf("failed to OpenFile:%s,err:%v", nextFile, err)
	}
	p := &sizeRotatedOutput{file: nextFile, output: f, size: parseSize(rotateConf.RotateArg)}
	return p
}

func getNextFile(file string) string {
	dir := path.Dir(file)
	id := getNextID(dir)
	return fmt.Sprintf("%s.%d", file, id)
}

func getNextID(dir string) int64 {
	file := path.Join(dir, logIndexFile)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		if err := ioutil.WriteFile(file, []byte(strconv.FormatInt(0, 10)), 0666); err != nil {
			return 0
		}
		return 0
	}
	id, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0
	}
	nextID := id + 1
	if err := ioutil.WriteFile(file, []byte(strconv.FormatInt(id+1, 10)), 0666); err != nil {
		return id
	}
	return nextID
}
