package main

import (
	"archive/zip"
	"bytes"
	_ "embed"

	"github.com/google/uuid"
)

//go:embed python_base.py
var PythonBase []byte

type Language uint

const (
	LanguagePython Language = iota
)

type WorkerStatus uint

const (
	StatusUnknown WorkerStatus = iota
	StatusReady
	StatusError
)

type WorkerState uint

const (
	StateDisabled WorkerState = iota
	StateEnabled
)

type UserWorker struct {
	ID           uuid.UUID    `json:"uuid"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	State        WorkerState  `json:"state"`
	Language     Language     `json:"language"`
	Organisation int64        `json:"organisation"`
	Major        uint         `json:"major"`
	Revision     uint         `json:"revision"`
	Status       WorkerStatus `json:"status"`
	StatusInfo   string       `json:"status_info"`
	Source       []byte       `json:"-"`
	Entrypoint   string       `json:"entrypoint"`
}

func CreateWorker(name, description string, source []byte) (*UserWorker, error) {
	sourceZip := bytes.NewBuffer(make([]byte, 0))
	zipWriter := zip.NewWriter(sourceZip)

	// Write "main.py"
	mainPY, err := zipWriter.Create("main.py")
	if err != nil {
		return nil, err
	}
	if _, err := mainPY.Write(PythonBase); err != nil {
		return nil, err
	}
	// Write "usercode.py"
	usercodePY, err := zipWriter.Create("usercode.py")
	if err != nil {
		return nil, err
	}
	if _, err := usercodePY.Write(source); err != nil {
		return nil, err
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	archive := sourceZip.Bytes()

	return &UserWorker{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		State:        StateEnabled,
		Language:     LanguagePython,
		Organisation: 0,
		Major:        1,
		Revision:     1,
		Status:       StatusUnknown,
		Source:       archive,
		Entrypoint:   "main.main",
	}, nil
}
