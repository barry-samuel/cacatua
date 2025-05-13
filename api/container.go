package container

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/barry-samuel/cacatua/utils"
	"github.com/containers/podman/v5/pkg/bindings/containers"
)

type DefaultResponse struct {
	Msg string `json:"msg"`
}

func Handle(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd := r.PathValue("cmd")

		switch cmd {
		case "list":
			getList(ctx, w, r)
		case "logs":
			getLogs(ctx, w, r)
		default:
			utils.SendResponseJSON(w, r, DefaultResponse{
				Msg: "command not found!",
			})
		}
	}
}

type ListOptionsJSON struct {
	All       *bool               `json:"all,omitempty"`
	External  *bool               `json:"external,omitempty"`
	Filters   map[string][]string `json:"filters,omitempty"`
	Last      *int                `json:"last,omitempty"`
	Namespace *bool               `json:"namespace,omitempty"`
	Size      *bool               `json:"size,omitempty"`
	Sync      *bool               `json:"sync,omitempty"`
}

func getList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	var listOpts ListOptionsJSON
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&listOpts); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	podmanOpts := new(containers.ListOptions)
	if err := utils.MapStruct(&listOpts, podmanOpts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctrs, err := containers.List(ctx, podmanOpts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var result []any
	for _, ctr := range ctrs {
		result = append(result, ctr)
	}

	utils.SendResponseJSON(w, r, result)
}

type LogOptionsJSON struct {
	ContainerID   string  `json:"container_id,omitempty"`
	ContainerName string  `json:"container_name,omitempty"`
	Follow        *bool   `json:"follow,omitempty"`
	Since         *string `json:"since,omitempty"`
	Stderr        *bool   `json:"stderr,omitempty"`
	Stdout        *bool   `json:"stdout,omitempty"`
	Tail          *string `json:"tail,omitempty"`
	Timestamps    *bool   `json:"timestamps,omitempty"`
	Until         *string `json:"until,omitempty"`
}

type LogResponse struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}

func getLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	var logOpts LogOptionsJSON
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logOpts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if logOpts.ContainerID == "" && logOpts.ContainerName == "" {
		http.Error(w, "container id/name must be provided!", http.StatusBadRequest)
		return
	}

	var ctrIndentifier string
	if logOpts.ContainerID != "" {
		ctrIndentifier = logOpts.ContainerID
	} else {
		ctrIndentifier = logOpts.ContainerName
	}

	isExist, err := containers.Exists(ctx, ctrIndentifier, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isExist {
		http.Error(w, "container not found!", http.StatusNotFound)
		return
	}

	var stdoutChan = make(chan string)
	var stderrChan = make(chan string)

	go func() {
		podmanOpts := new(containers.LogOptions)
		if err := utils.MapStruct(&logOpts, podmanOpts); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err := containers.Logs(ctx, ctrIndentifier, podmanOpts, stdoutChan, stderrChan)
		close(stdoutChan)
		close(stderrChan)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()

	var result LogResponse
	for stdoutChan != nil || stderrChan != nil {
		select {
		case line, ok := <-stdoutChan:
			if !ok {
				stdoutChan = nil
			} else {
				result.Stdout += hex.EncodeToString([]byte(line))
			}
		case line, ok := <-stderrChan:
			if !ok {
				stderrChan = nil
			} else if *logOpts.Stdout == true && *logOpts.Stderr == true {
				result.Stdout += hex.EncodeToString([]byte(line))
			} else {
				result.Stderr += hex.EncodeToString([]byte(line))
			}
		}
	}

	utils.SendResponseJSON(w, r, result)
}
