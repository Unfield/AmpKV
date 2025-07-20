# ⚡ AmpKV: Your Flexible Go-Native Key-Value Store ⚡

[![Tests](https://github.com/Unfield/AmpKV/actions/workflows/go-tests-ci.yml/badge.svg)](https://github.com/Unfield/AmpKV/actions/workflows/go-ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Unfield/AmpKV)](https://goreportcard.com/report/github.com/Unfield/AmpKV)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Contributors](https://img.shields.io/github/contributors/Unfield/AmpKV.svg)](https://github.com/Unfield/AmpKV/graphs/contributors)
![GitHub stars](https://img.shields.io/github/stars/Unfield/AmpKV.svg?style=social&label=Stars)
![GitHub forks](https://img.shields.io/github/forks/Unfield/AmpKV.svg?style=social&label=Fork)

![AmpKV Banner](https://raw.githubusercontent.com/Unfield/AmpKV/main/public/initial_logo.svg)

AmpKV (pronounced "Amp-Key-Vee") is a supercharged, Go-native key-value store designed for **flexibility, performance, and seamless distribution**. Whether you need lightning-fast local access, smart client-side caching, or a robust central source of truth, AmpKV has a mode just for you!

## ✨ Why AmpKV?

In the world of distributed systems, choosing the right data access pattern can be tricky. AmpKV simplifies this by offering **three distinct, yet harmonious, operational modes**:

- **⚡ Embedded:** When you need a local, blazing-fast KV store directly within your application, perfect for caches, config, or local data.
- **🔗 Replication:** Get the best of both worlds! Local-first reads with smart caching, automatically falling back to a remote AmpKV server when data isn't found locally. Writes are synchronized for ultimate consistency.
- **🌐 Remote Only:** Your centralized, scalable source of truth. A robust AmpKV server that can be backed by your favorite battle-tested databases (PostgreSQL, MySQL, SQLite, and more coming!).

AmpKV empowers you to build resilient, high-performance Go applications with confidence.

## 🚀 Getting Started

To get AmpKV up and running, choose the mode that fits your needs!

### ⚡ Embedded Mode (Local Powerhouse)

For a simple, local-only key-value store that lives within your application.

```bash
go get github.com/Unfield/AmpKV/pkg/embedded
```

```go
package main

import (
	"fmt"
	"log"
	"github.com/Unfield/AmpKV/pkg/embedded"
)

func main() {
	// Initialize an embedded store (e.g., with SQLite for persistence)
	store, err := embedded.NewStore("ampkv_local.db")
	if err != nil {
		log.Fatalf("Failed to create embedded store: %v", err)
	}
	defer store.Close()

	// Put a key-value pair
	if err := store.Put("hello", []byte("world from embedded!")); err != nil {
		log.Fatalf("Failed to put: %v", err)
	}
	fmt.Println("Put 'hello': 'world from embedded!'")

	// Get a value
	val, err := store.Get("hello")
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}
	fmt.Printf("Get 'hello': %s\n", val)

	// Try a non-existent key
	_, err = store.Get("nonexistent")
	if err != nil {
		fmt.Printf("Get 'nonexistent' (expected error): %v\n", err)
	}
}
```

### 🌐 Remote Only Mode (Central Server)

This is your standalone AmpKV server, acting as the centralized source of truth.

1.  **Install the server:**
    ```bash
    go install github.com/Unfield/AmpKV/cmd/ampkv-server@latest path
    ```
2.  **Run the server:**
    By default, it might use SQLite, but you can configure it for PostgreSQL, MySQL, etc.
    ```bash
    ampkv-server --persistence-driver=sqlite --sqlite-dsn=ampkv_server.db
    # Or for PostgreSQL:
    # ampkv-server --persistence-driver=postgres --postgres-conn-string="host=localhost user=ampkv dbname=ampkv sslmode=disable password=ampkv"
    ```
3.  **Use a simple HTTP/gRPC client (e.g., curl or a test client from `pkg/client`):**
    ```bash
    # Assuming HTTP API exposed by default on :8080
    curl -X POST -d '{"key":"mykey","value":"myvalue"}' http://localhost:8080/put
    curl http://localhost:8080/get?key=mykey
    ```

### 🔗 Replication Mode (Smart Client with Remote Fallback)

(Coming Soon after initial Embedded and Remote Only modes are stable!)

This mode will offer a Go library that intelligently fetches from a local cache first, then calls out to a remote AmpKV server for misses, and ensures writes are propagated.

## ✨ Features (Current & Planned)

- **Multi-Mode Flexibility:** Seamlessly switch between Embedded, Replication, and Remote Only modes.
- **Pluggable Persistence:** For Remote Only mode, use SQLite, PostgreSQL, MySQL (with support for PlanetScale/CockroachDB), and more!
- **High Performance:** Optimized Go concurrency for speed.
- **Easy to Use:** Clean Go APIs and straightforward server configuration.
- **Observability:** Prometheus metrics and structured logging (coming soon!).
- **Robustness:** Designed for resilience in distributed environments.

## 🤝 Contributing

We ❤️ contributions! Whether it's fixing a bug, adding a new feature, improving documentation, or just sharing ideas, your input is incredibly valuable.

Please check out our [Contributing Guidelines](CONTRIBUTING.md) for more details.

## 📄 License

AmpKV is open-source and released under the [MIT License](LICENSE).
