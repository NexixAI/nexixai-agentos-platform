    package main

    import (
        "flag"
        "fmt"
        "log"
        "os"
        "path/filepath"
        "runtime"
        "strings"

        "github.com/eyoshidagorgonia/nexixai-agentos-platform/federation"
        stacka "github.com/eyoshidagorgonia/nexixai-agentos-platform/stack-a"
        stackb "github.com/eyoshidagorgonia/nexixai-agentos-platform/stack-b"
        "github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/deploy"
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
        case "up":
            up(os.Args[2:])
        case "redeploy":
            redeploy(os.Args[2:])
        case "validate":
            validate(os.Args[2:])
        case "status":
            status(os.Args[2:])
        case "nuke":
            nuke(os.Args[2:])
        case "version":
            fmt.Println(version)
        default:
            usage()
            os.Exit(2)
        }
    }

    func usage() {
        fmt.Println(`agentos (Phase 6 deployment UX)
Usage:
  agentos serve <stack-a|stack-b|federation> [--addr :PORT]
  agentos up [--compose-file PATH] [--project NAME]
  agentos redeploy [--compose-file PATH] [--project NAME]
  agentos validate [--stack-a URL] [--stack-b URL] [--federation URL]
  agentos status [--compose-file PATH] [--project NAME]
  agentos nuke [--compose-file PATH] [--project NAME] [--hard]

Defaults:
  compose file: deploy/local/compose.yaml
  endpoints:
    stack-a:     http://localhost:8081
    stack-b:     http://localhost:8082
    federation:  http://localhost:8083
`)
    }

    func repoRoot() string {
        // best-effort: use current working directory as repo root for local dev
        cwd, _ := os.Getwd()
        return cwd
    }

    func defaultComposeFile() string {
        // path relative to repo root
        return filepath.Join(repoRoot(), "deploy", "local", "compose.yaml")
    }

    func newRunner(composeFile, project string) deploy.ComposeRunner {
        r := deploy.ComposeRunner{
            ComposeFile: composeFile,
            ProjectName: project,
            Stdout: func(s string) { log.Println(s) },
            Stderr: func(s string) { log.Println(s) },
        }
        return r
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

        switch target {
        case "stack-a":
            log.Printf("serving stack-a on %s", *addr)
            log.Fatal(stacka.ListenAndServe(*addr, version))
        case "stack-b":
            log.Printf("serving stack-b on %s", *addr)
            log.Fatal(stackb.ListenAndServe(*addr, version))
        case "federation":
            log.Printf("serving federation on %s", *addr)
            log.Fatal(federation.ListenAndServe(*addr, version))
        default:
            log.Fatalf("unknown target: %s", target)
        }
    }

    func up(args []string) {
        fs := flag.NewFlagSet("up", flag.ExitOnError)
        composeFile := fs.String("compose-file", defaultComposeFile(), "compose file path")
        project := fs.String("project", "agentos", "compose project name")
        _ = fs.Parse(args)

        log.Printf("platform: %s/%s", runtime.GOOS, runtime.GOARCH)
        r := newRunner(*composeFile, *project)

        reportsDir, err := deploy.EnsureReportsDir(repoRoot())
        if err != nil {
            log.Fatal(err)
        }

        log.Println("==> up: starting containers (build on first run)")
        if err := r.Up(true); err != nil {
            log.Fatalf("up failed: %v", err)
        }

        // Validate as part of up
        v := deploy.Validator{
            StackA: "http://localhost:8081",
            StackB: "http://localhost:8082",
            Fed:    "http://localhost:8083",
            RepoRoot: repoRoot(),
        }
        log.Println("==> up: validating")
        checks, verr := v.ValidateAll()

        sum := deploy.Summary{
            Mode: "up",
            Endpoints: map[string]string{
                "stack-a": "http://localhost:8081",
                "stack-b": "http://localhost:8082",
                "federation": "http://localhost:8083",
            },
            Checks: checks,
        }
        _ = deploy.WriteSummary(reportsDir, sum)

        if verr != nil {
            log.Printf("validation failed; summary written to %s", reportsDir)
            log.Fatal(verr)
        }

        log.Printf("✅ up complete. summary: %s", reportsDir)
        printAccess(sum.Endpoints)
    }

    func redeploy(args []string) {
        fs := flag.NewFlagSet("redeploy", flag.ExitOnError)
        composeFile := fs.String("compose-file", defaultComposeFile(), "compose file path")
        project := fs.String("project", "agentos", "compose project name")
        _ = fs.Parse(args)

        r := newRunner(*composeFile, *project)
        log.Println("==> redeploy: rebuilding + restarting")
        if err := r.Up(true); err != nil {
            log.Fatalf("redeploy failed: %v", err)
        }
        log.Println("✅ redeploy complete")
    }

    func validate(args []string) {
        fs := flag.NewFlagSet("validate", flag.ExitOnError)
        stackA := fs.String("stack-a", "http://localhost:8081", "Stack A base URL")
        stackB := fs.String("stack-b", "http://localhost:8082", "Stack B base URL")
        fed := fs.String("federation", "http://localhost:8083", "Federation base URL")
        _ = fs.Parse(args)

        reportsDir, err := deploy.EnsureReportsDir(repoRoot())
        if err != nil {
            log.Fatal(err)
        }

        v := deploy.Validator{StackA: *stackA, StackB: *stackB, Fed: *fed, RepoRoot: repoRoot()}
        checks, verr := v.ValidateAll()

        sum := deploy.Summary{
            Mode: "validate",
            Endpoints: map[string]string{
                "stack-a": *stackA,
                "stack-b": *stackB,
                "federation": *fed,
            },
            Checks: checks,
        }
        _ = deploy.WriteSummary(reportsDir, sum)

        if verr != nil {
            log.Printf("❌ validate failed. summary: %s", reportsDir)
            os.Exit(1)
        }
        log.Printf("✅ validate ok. summary: %s", reportsDir)
    }

    func status(args []string) {
        fs := flag.NewFlagSet("status", flag.ExitOnError)
        composeFile := fs.String("compose-file", defaultComposeFile(), "compose file path")
        project := fs.String("project", "agentos", "compose project name")
        _ = fs.Parse(args)

        r := newRunner(*composeFile, *project)
        _ = r.Ps()
        printAccess(map[string]string{
            "stack-a": "http://localhost:8081",
            "stack-b": "http://localhost:8082",
            "federation": "http://localhost:8083",
        })
    }

    func nuke(args []string) {
        fs := flag.NewFlagSet("nuke", flag.ExitOnError)
        composeFile := fs.String("compose-file", defaultComposeFile(), "compose file path")
        project := fs.String("project", "agentos", "compose project name")
        hard := fs.Bool("hard", false, "remove volumes as well (destructive)")
        _ = fs.Parse(args)

        r := newRunner(*composeFile, *project)
        log.Println("==> nuke: stopping")
        if err := r.Down(*hard); err != nil {
            log.Fatalf("nuke failed: %v", err)
        }
        log.Println("✅ nuke complete")
    }

    func printAccess(endpoints map[string]string) {
        log.Println("==> access")
        if v, ok := endpoints["stack-a"]; ok {
            log.Printf("Stack A health: %s/v1/health", v)
        }
        if v, ok := endpoints["stack-b"]; ok {
            log.Printf("Stack B health: %s/v1/health", v)
        }
        if v, ok := endpoints["federation"]; ok {
            log.Printf("Federation health: %s/v1/federation/health", v)
        }
    }
