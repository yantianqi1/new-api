package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type collector struct {
	inboxDir   string
	workDir    string
	exportDir  string
	allowlist  map[string]struct{}
	offsets    *offsetStore
	keepInbox  int
	keepExport int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("sft-collector ")

	baseDir := envOr("SFT_COLLECT_DIR", "/data/sft-collector")
	c := &collector{
		inboxDir:   envOr("SFT_INBOX_DIR", filepath.Join(baseDir, "inbox")),
		workDir:    envOr("SFT_WORK_DIR", filepath.Join(baseDir, "work")),
		exportDir:  envOr("SFT_EXPORT_DIR", filepath.Join(baseDir, "export")),
		allowlist:  loadAllowlist(os.Getenv("SFT_MODEL_ALLOWLIST"), envOr("SFT_MODEL_ALLOWLIST_FILE", filepath.Join(baseDir, "models.txt"))),
		keepInbox:  envInt("SFT_INBOX_KEEP_DAYS", 7),
		keepExport: envInt("SFT_EXPORT_KEEP_DAYS", 30),
	}
	if err := ensureDir(c.inboxDir); err != nil {
		log.Fatalf("mkdir inbox: %v", err)
	}
	if err := ensureDir(c.workDir); err != nil {
		log.Fatalf("mkdir work: %v", err)
	}
	if err := ensureDir(c.exportDir); err != nil {
		log.Fatalf("mkdir export: %v", err)
	}
	c.offsets = loadOffsets(filepath.Join(baseDir, "inbox-offset.json"))

	pollEvery := envDuration("SFT_POLL_INTERVAL", 3*time.Second)
	packEvery := envDuration("SFT_PACK_INTERVAL", 60*time.Second)
	log.Printf("watching inbox=%s export=%s allowlist=%d", c.inboxDir, c.exportDir, len(c.allowlist))

	poll := time.NewTicker(pollEvery)
	pack := time.NewTicker(packEvery)
	defer poll.Stop()
	defer pack.Stop()

	c.ingest()
	c.packAll()
	for {
		select {
		case <-poll.C:
			c.ingest()
		case <-pack.C:
			c.packAll()
			c.cleanup()
		}
	}
}

func (c *collector) ingest() {
	entries, err := os.ReadDir(c.inboxDir)
	if err != nil {
		log.Printf("read inbox: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		if err := c.ingestFile(entry.Name()); err != nil {
			log.Printf("ingest %s: %v", entry.Name(), err)
		}
	}
}

func (c *collector) ingestFile(name string) error {
	path := filepath.Join(c.inboxDir, name)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	start := c.offsets.get(name)
	if start > 0 {
		if _, err := f.Seek(start, io.SeekStart); err != nil {
			return err
		}
	}

	day := strings.TrimSuffix(name, ".jsonl")
	workFile, err := os.OpenFile(workPath(c.workDir, day), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer workFile.Close()

	reader := bufio.NewReader(f)
	offset := start
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if rec, ok := c.convertLine(line); ok {
			if encErr := encodeShareGPT(workFile, rec); encErr != nil {
				return encErr
			}
		}
		offset += int64(len(line))
		if setErr := c.offsets.set(name, offset); setErr != nil {
			return setErr
		}
	}
}

func (c *collector) convertLine(line []byte) (ShareGPTRecord, bool) {
	var env Envelope
	if err := json.Unmarshal(line, &env); err != nil {
		return ShareGPTRecord{}, false
	}
	if !modelAllowed(env.Model, c.allowlist) {
		return ShareGPTRecord{}, false
	}
	rec, ok := toShareGPT(env)
	if !ok {
		return ShareGPTRecord{}, false
	}
	return redactRecord(rec), true
}

func (c *collector) packAll() {
	entries, err := os.ReadDir(c.workDir)
	if err != nil {
		log.Printf("read work: %v", err)
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sharegpt.jsonl") {
			continue
		}
		day := strings.TrimSuffix(name, ".sharegpt.jsonl")
		src := filepath.Join(c.workDir, name)
		dst := exportPath(c.exportDir, day)
		if err := packJSONArray(src, dst); err != nil {
			log.Printf("pack %s: %v", name, err)
		}
	}
}

func (c *collector) cleanup() {
	c.prune(c.inboxDir, ".jsonl", c.keepInbox)
	c.prune(c.exportDir, ".json", c.keepExport)
	c.prune(c.workDir, ".sharegpt.jsonl", c.keepExport)
}

func (c *collector) prune(dir, suffix string, keepDays int) {
	if keepDays <= 0 {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}
