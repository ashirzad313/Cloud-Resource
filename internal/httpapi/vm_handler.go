package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/vm"
)

type vmHandler struct {
	service *vm.Service
}

func (h vmHandler) list(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	machines, err := h.service.List(ctx)
	if err != nil {
		log.Printf("list VMs: %v", err)
		writeError(w, http.StatusInternalServerError, "could not list virtual machines")
		return
	}

	writeJSON(w, http.StatusOK, machines)
}

func (h vmHandler) create(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var request struct {
		Name     string `json:"name"`
		CPUCores int32  `json:"cpu_cores"`
		MemoryMB int32  `json:"memory_mb"`
		Region   string `json:"region"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeDecodeError(w, err)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeDecodeError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	machine, err := h.service.Create(ctx, vm.CreateInput{
		Name: request.Name, CPUCores: request.CPUCores,
		MemoryMB: request.MemoryMB, Region: request.Region,
	})
	if errors.Is(err, vm.ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Printf("create VM: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create virtual machine")
		return
	}

	writeJSON(w, http.StatusCreated, machine)
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	writeError(w, http.StatusBadRequest, "expected one JSON object with name, cpu_cores, memory_mb and region")
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, struct {
		Error string `json:"error"`
	}{Error: message})
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
