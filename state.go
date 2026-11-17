package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// A draw holds one text per Santa. It is the only record of who gives a gift
// to whom, so it must never be printed: the giftee's name appears only inside
// each message.
type draw []drawnText

type drawnText struct {
	Santa   string     `json:"santa"`
	Message string     `json:"message"`
	SentAt  *time.Time `json:"sentAt"`
}

// createDrawFile writes d to a new file. It fails if the file exists, so that
// a second draw never replaces one whose texts may already be sent.
func createDrawFile(path string, d draw) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if err := writeDraw(f, d); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// replaceDrawFile overwrites the file at path with d. It writes to a temporary
// file first, so that a crash never leaves a half-written draw.
func replaceDrawFile(path string, d draw) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if err := writeDraw(f, d); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeDraw(f *os.File, d draw) error {
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		return fmt.Errorf("encoding draw: %w", err)
	}
	return f.Sync()
}

func readDrawFile(path string) (draw, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d draw
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return d, nil
}
