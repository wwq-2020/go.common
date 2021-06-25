package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wwq-2020/go.common/errors"
)

// FileFormatter FileFormatter
type FileFormatter func(string) string

// RotateType RotateType
type RotateType int

// RotateTypes
const (
	PeriodRotateType RotateType = iota
	SizeRotateType
	defaultPeriodRotateArg         = "24h"
	defaultSizeRotateArg           = "1gb"
	defaultSizeRotateArgInt        = 1 << 30
	b                       uint64 = 1
	kb                      uint64 = 1 << 10
	mb                      uint64 = 1 << 20
)

var (
	reg = regexp.MustCompile(`(\d{1,})(m|mb|k|kb|b|g|gb)`)
)

// RotateConf RotateConf
type RotateConf struct {
	RotateType    RotateType
	RotateArg     string
	FileFormatter FileFormatter
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
	if rc.FileFormatter == nil {
		rc.FileFormatter = defaultFileFormatter
	}
}

// NewFileLog NewFileLog
func NewFileLog(file string) Logger {
	return NewFileLogEx(file, nil)
}

// NewFileLogEx NewFileLogEx
func NewFileLogEx(file string, rotateConf *RotateConf) Logger {
	output := BuildRotatedOutput(file, rotateConf)
	return New(WithOutput(output))
}

type periodRotatedOutput struct {
	file   string
	output *os.File
	d      time.Duration
	bytes  uint64
	sync.Mutex
	fileFormatter FileFormatter
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
	nextFile := p.fileFormatter(p.file)
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
	fileFormatter FileFormatter
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
	nextFile := s.fileFormatter(s.file)
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

// BuildRotatedOutput BuildRotatedOutput
func BuildRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
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
		return BuildPeriodRotatedOutput(file, rotateConf)
	case SizeRotateType:
		return BuildSizeRotatedOutput(file, rotateConf)
	default:
		return BuildPeriodRotatedOutput(file, rotateConf)
	}
}

// BuildPeriodRotatedOutput BuildPeriodRotatedOutput
func BuildPeriodRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
	d, err := time.ParseDuration(rotateConf.RotateArg)
	if err != nil {
		d, err = time.ParseDuration(defaultPeriodRotateArg)
		if err != nil {
			Fatalf("failed to ParseDuration:%s,err:%v", defaultPeriodRotateArg, err)
		}
	}
	nextFile := rotateConf.FileFormatter(file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Fatalf("failed to OpenFile:%s,err:%v", nextFile, err)
	}
	p := &periodRotatedOutput{file: file, output: f, d: d, fileFormatter: rotateConf.FileFormatter}
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

// BuildSizeRotatedOutput BuildSizeRotatedOutput
func BuildSizeRotatedOutput(file string, rotateConf *RotateConf) io.WriteCloser {
	nextFile := rotateConf.FileFormatter(file)
	f, err := os.OpenFile(nextFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		Fatalf("failed to OpenFile:%s,err:%v", nextFile, err)
	}
	p := &sizeRotatedOutput{file: nextFile, output: f, size: parseSize(rotateConf.RotateArg), fileFormatter: rotateConf.FileFormatter}
	return p
}

func defaultFileFormatter(file string) string {
	dir := path.Dir(file)
	filename := path.Base(file)
	filename = fmt.Sprintf("%s.%s", time.Now().Format("2006-01-02-15-04-05"), filename)
	return path.Join(dir, filename)
}
