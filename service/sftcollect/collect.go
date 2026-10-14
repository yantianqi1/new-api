package sftcollect

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type Job struct {
	Model    string
	Messages []Message
	Tools    []byte
	Response []byte
	IsStream bool
}

var (
	enabled     atomic.Bool
	queue       chan Job
	dropped     atomic.Uint64
	inboxDir    string
	locShanghai *time.Location
)

func Enabled() bool {
	return enabled.Load()
}

func Start() {
	if !common.GetEnvOrDefaultBool("SFT_COLLECT_ENABLED", false) {
		return
	}
	baseDir := common.GetEnvOrDefaultString("SFT_COLLECT_DIR", "/data/sft-collector")
	inboxDir = filepath.Join(baseDir, "inbox")
	if err := os.MkdirAll(inboxDir, 0755); err != nil {
		common.SysError("sft collector disabled, mkdir failed: " + err.Error())
		return
	}
	queueSize := common.GetEnvOrDefault("SFT_COLLECT_QUEUE_SIZE", 256)
	if queueSize < 32 {
		queueSize = 32
	}
	queue = make(chan Job, queueSize)
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		locShanghai = loc
	} else {
		locShanghai = time.Local
	}
	enabled.Store(true)
	go worker()
	common.SysLog("sft collector enabled, inbox=" + inboxDir)
}

func TryEnqueue(job Job) {
	if !enabled.Load() || queue == nil {
		return
	}
	if job.Model == "" || len(job.Messages) == 0 || len(job.Response) == 0 {
		return
	}
	select {
	case queue <- job:
	default:
		dropped.Add(1)
		if dropped.Load()%100 == 1 {
			common.SysLog("sft collector queue full, dropping samples")
		}
	}
}

func worker() {
	writer := &jsonlWriter{dir: inboxDir}
	defer writer.close()
	for job := range queue {
		writeJob(writer, job)
	}
}

func writeJob(writer *jsonlWriter, job Job) {
	defer func() {
		if r := recover(); r != nil {
			common.SysError("sft collector panic recovered")
		}
	}()
	assistant, finish, ok := ParseAssistant(job.Response, job.IsStream)
	if !ok || !AcceptableFinish(finish) {
		return
	}
	env := Envelope{
		Model:        job.Model,
		FinishReason: finish,
		Messages:     append(job.Messages, assistant),
		Tools:        job.Tools,
	}
	line, err := common.Marshal(env)
	if err != nil {
		return
	}
	if err := writer.write(line); err != nil {
		common.SysError("sft collector write failed: " + err.Error())
	}
}

type jsonlWriter struct {
	dir string
	day string
	f   *os.File
}

func (w *jsonlWriter) write(line []byte) error {
	day := time.Now().In(locShanghai).Format("2006-01-02")
	if w.f == nil || w.day != day {
		if err := w.rotate(day); err != nil {
			return err
		}
	}
	if _, err := w.f.Write(line); err != nil {
		return err
	}
	_, err := w.f.Write([]byte("\n"))
	return err
}

func (w *jsonlWriter) rotate(day string) error {
	w.close()
	path := filepath.Join(w.dir, day+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.f = f
	w.day = day
	return nil
}

func (w *jsonlWriter) close() {
	if w.f != nil {
		_ = w.f.Close()
		w.f = nil
	}
}
