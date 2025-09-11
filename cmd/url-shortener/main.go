// TODO: init config: cleanenv
package main

import (
	"fmt"

	"url-shortener/internal/config1"
)

func main() {
	cfg := config1.MustLoad()

	fmt.Printf("Env=%s\nStorage=%s\nHTTP=%+v\n", cfg.Env, cfg.StoragePath, cfg.HTTPServer)
}

// TODO: init logger: slog

// TODO: init storage: sqlite

// TODO: init router: chi, "chi render"

// TODO: run server
