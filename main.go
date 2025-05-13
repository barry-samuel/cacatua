package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/containers/podman/v5/pkg/bindings"

	container "github.com/barry-samuel/cacatua/api"
)

func main() {
	sockDir := os.Getenv("XDG_RUNTIME_DIR")
	if sockDir == "" {
		sockDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	// socket := "unix:" + sockDir + "/podman/podman.sock"
	remote_socket := "ssh://noxus@128.199.92.238:411/run/user/1000/podman/podman.sock"
	identity := "/home/samuel/.ssh/id_ed25519"

	// conn, err := bindings.NewConnection(context.Background(), socket)
	conn, err := bindings.NewConnectionWithIdentity(context.Background(), remote_socket, identity, true)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	http.HandleFunc("/container/{cmd}", container.Handle(conn))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
