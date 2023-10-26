package userworkers

import (
	"archive/zip"
	"bytes"
	"embed"
	_ "embed"
	"io"
	"io/fs"
	"net/http"
	"regexp"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/web"
)

//go:embed python/*
var PythonFiles embed.FS

var ErrWorkerInvalidName = web.NewError(http.StatusBadRequest, "Worker name is invalid", "ERR_INVALID_WORKER_NAME")

type Language string

const (
	LanguagePython Language = "python"
)

type WorkerStatus string

const (
	StatusUnknown WorkerStatus = "unknown"
	StatusReady                = "ready"
	StatusError                = "error"
)

type WorkerState string

const (
	StateUnknown  WorkerState = "unknown"
	StateDisabled             = "disabled"
	StateEnabled              = "enabled"
)

type UserWorker struct {
	// if this is true, then the revision should increase on save
	dirty bool

	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	State        WorkerState  `json:"state"`
	Language     Language     `json:"language"`
	Organisation int64        `json:"organisation"`
	Revision     uint         `json:"revision"`
	Status       WorkerStatus `json:"status"`
	StatusInfo   string       `json:"status_info" db:"status_info"`
	ZipSource    []byte       `json:"-" db:"source"`
	Entrypoint   string       `json:"entrypoint"`
}

func CreateWorker(name, description string, userCode []byte) (*UserWorker, error) {
	worker := &UserWorker{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		State:        StateDisabled,
		Language:     LanguagePython,
		Organisation: 0,
		Revision:     1,
		Status:       StatusUnknown,
		Entrypoint:   "main.main",
	}
	if err := worker.SetUserCode(userCode); err != nil {
		return nil, err
	}
	return worker, nil
}

func (w *UserWorker) SetUserCode(usercode []byte) error {
	sourceZip := bytes.NewBuffer(make([]byte, 0))
	zipWriter := zip.NewWriter(sourceZip)

	if err := writeToZip(PythonFiles, "python/base.py", zipWriter, "main.py"); err != nil {
		return err
	}
	if err := writeToZip(PythonFiles, "python/requirements.txt", zipWriter, "requirements.txt"); err != nil {
		return err
	}
	if err := writeToZip(PythonFiles, "python/build.sh", zipWriter, "build.sh"); err != nil {
		return err
	}
	// Write "usercode.py"
	usercodePY, err := zipWriter.Create("usercode.py")
	if err != nil {
		return err
	}
	if _, err := usercodePY.Write(usercode); err != nil {
		return err
	}
	if err := zipWriter.Close(); err != nil {
		return err
	}
	archive := sourceZip.Bytes()
	w.ZipSource = archive
	w.dirty = true
	return nil
}

func (w *UserWorker) GetUserCode() ([]byte, error) {
	zipSource := bytes.NewReader(w.ZipSource)
	zipReader, err := zip.NewReader(zipSource, int64(len(w.ZipSource)))
	if err != nil {
		return nil, err
	}
	file, err := zipReader.Open("usercode.py")
	if err != nil {
		return nil, err
	}
	userCode, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return userCode, nil
}

var r_worker_name = regexp.MustCompile("")

func (w *UserWorker) SetName(name string) error {
	if !r_worker_name.MatchString(name) {
		return ErrWorkerInvalidName
	}
	w.Name = name
	return nil
}

func (w *UserWorker) Disable() {
	w.State = StateDisabled
}

func (w *UserWorker) Enable() {
	w.State = StateEnabled
}

func (w *UserWorker) Commit() {
	if !w.dirty {
		return
	}
	w.Revision++
	w.dirty = false
}

func writeToZip(f fs.FS, from string, writer *zip.Writer, to string) error {
	file, err := f.Open(from)
	if err != nil {
		return err
	}
	contents, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	file.Close()
	zipFile, err := writer.Create(to)
	if err != nil {
		return err
	}
	_, err = zipFile.Write(contents)
	if err != nil {
		return err
	}
	return nil
}
