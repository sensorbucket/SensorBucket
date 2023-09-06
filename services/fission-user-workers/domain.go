package main

import "github.com/google/uuid"

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

type UserWorker struct {
	ID           uuid.UUID    `json:"uuid"`
	Language     Language     `json:"language"`
	Organisation int64        `json:"organisation"`
	Major        uint         `json:"major"`
	Revision     uint         `json:"revision"`
	Status       WorkerStatus `json:"status"`
	StatusInfo   string       `json:"status_info"`
}
