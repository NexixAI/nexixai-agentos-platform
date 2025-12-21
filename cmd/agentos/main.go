    package main

    import (
        "flag"
        "fmt"
        "log"
        "net/http"
        "os"
        "strings"

        "github.com/eyoshidagorgonia/nexixai-agentos-platform/federation"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/stack-a"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/stack-b"
    )

    const version = "0.0.1-dev"

    func main() {
        if len(os.Args) < 2 {
            usage()
            os.Exit(2)
        }

        switch os.Args[1] {
        case "serve":
            serve(os.Args[2:])
        case "version":
            fmt.Println(version)
        default:
            usage()
            os.Exit(2)
        }
    }

    func usage() {
        fmt.Println(`agentos (Phase 5 scaffold)
Usage:
  agentos serve <stack-a|stack-b|federation> [--addr :PORT]

Examples:
  agentos serve stack-a --addr :8081
  agentos serve stack-b --addr :8082
  agentos serve federation --addr :8083
`)
    }

    func serve(args []string) {
        fs := flag.NewFlagSet("serve", flag.ExitOnError)
        addr := fs.String("addr", ":8081", "listen address (host:port)")
        _ = fs.Parse(args)

        rest := fs.Args()
        if len(rest) < 1 {
            usage()
            os.Exit(2)
        }
        target := strings.ToLower(rest[0])

        var h http.Handler
        switch target {
        case "stack-a":
            h = stacka.New(version).Handler()
        case "stack-b":
            h = stackb.New(version).Handler()
        case "federation":
            h = federation.New(version).Handler()
        default:
            log.Fatalf("unknown target: %s", target)
        }

        log.Printf("serving %s on %s", target, *addr)
        if err := http.ListenAndServe(*addr, h); err != nil {
            log.Fatal(err)
        }
    }
