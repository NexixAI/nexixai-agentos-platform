package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	agentorchestrator "github.com/eyoshidagorgonia/nexixai-agentos-platform/agentorchestrator"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/federation"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/config"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/deploy"
	modelpolicy "github.com/eyoshidagorgonia/nexixai-agentos-platform/modelpolicy"
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
	case "tenants":
		tenantsCmd(os.Args[2:])
	case "version":
		fmt.Println(version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`agentos (Phase 7 multi-tenancy scaffold)
Usage:
  agentos serve <agent-orchestrator|model-policy|federation> [--addr :PORT]
  agentos up [--compose-file PATH] [--project NAME] [--tenant TENANT] [--principal PRINCIPAL]
  agentos redeploy [--compose-file PATH] [--project NAME]
  agentos validate [--agent-orchestrator URL] [--model-policy URL] [--federation URL] [--tenant TENANT] [--principal PRINCIPAL]
  agentos status [--compose-file PATH] [--project NAME]
  agentos nuke [--compose-file PATH] [--project NAME] [--hard]
  agentos tenants list [--agent-orchestrator URL]
  agentos tenants create --id TENANT_ID [--name NAME] [--plan PLAN] [--agent-orchestrator URL]

Defaults:
  compose file: deploy/local/compose.yaml
  endpoints:
    agent-orchestrator:     http://localhost:50081
    model-policy:           http://localhost:50082
    federation:             http://localhost:50083
`)
}

func repoRoot() string {
	cwd, _ := os.Getwd()
	return cwd
}

func defaultComposeFile() string {
	return filepath.Join(repoRoot(), "deploy", "local", "compose.yaml")
}

func newRunner(composeFile, project string) deploy.ComposeRunner {
	return deploy.ComposeRunner{
		ComposeFile: composeFile,
		ProjectName: project,
		Stdout:      func(s string) { log.Println(s) },
		Stderr:      func(s string) { log.Println(s) },
	}
}

func serve(args []string) {
	if err := config.EnsureSafeProfile(); err != nil {
		log.Fatalf("refusing to start in prod: %v", err)
	}

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8081", "listen address (host:port)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		usage()
		os.Exit(2)
	}
	target := strings.ToLower(rest[0])

	if err := config.ValidateServiceConfig(target); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	switch target {
	case "agent-orchestrator":
		log.Printf("serving agent-orchestrator on %s", *addr)
		log.Fatal(agentorchestrator.ListenAndServe(*addr, version))
	case "model-policy":
		log.Printf("serving model-policy on %s", *addr)
		log.Fatal(modelpolicy.ListenAndServe(*addr, version))
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
	tenant := fs.String("tenant", "tnt_demo", "tenant id used for validation")
	principal := fs.String("principal", "prn_local", "principal id used for validation")
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

	v := deploy.Validator{
		AgentOrchestrator: "http://localhost:50081",
		ModelPolicy:       "http://localhost:50082",
		Fed:               "http://localhost:50083",
		RepoRoot:          repoRoot(),
		TenantID:          *tenant,
		PrincipalID:       *principal,
	}
	log.Println("==> up: validating")
	checks, verr := v.ValidateAll()

	sum := deploy.Summary{
		Mode: "up",
		Endpoints: map[string]string{
			"agent-orchestrator": "http://localhost:50081",
			"model-policy":       "http://localhost:50082",
			"federation":         "http://localhost:50083",
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
	agentOrchestrator := fs.String("agent-orchestrator", "http://localhost:50081", "Agent Orchestrator base URL")
	modelPolicy := fs.String("model-policy", "http://localhost:50082", "Model Policy base URL")
	fed := fs.String("federation", "http://localhost:50083", "Federation base URL")
	tenant := fs.String("tenant", "tnt_demo", "tenant id")
	principal := fs.String("principal", "prn_local", "principal id")
	_ = fs.Parse(args)

	reportsDir, err := deploy.EnsureReportsDir(repoRoot())
	if err != nil {
		log.Fatal(err)
	}

	v := deploy.Validator{
		AgentOrchestrator: *agentOrchestrator,
		ModelPolicy:       *modelPolicy,
		Fed:               *fed,
		RepoRoot:          repoRoot(),
		TenantID:          *tenant,
		PrincipalID:       *principal,
	}
	checks, verr := v.ValidateAll()

	sum := deploy.Summary{
		Mode: "validate",
		Endpoints: map[string]string{
			"agent-orchestrator": *agentOrchestrator,
			"model-policy":       *modelPolicy,
			"federation":         *fed,
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
		"agent-orchestrator": "http://localhost:50081",
		"model-policy":       "http://localhost:50082",
		"federation":         "http://localhost:50083",
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

func tenantsCmd(args []string) {
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	switch strings.ToLower(args[0]) {
	case "list":
		tenantsList(args[1:])
	case "create":
		tenantsCreate(args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func tenantsList(args []string) {
	fs := flag.NewFlagSet("tenants list", flag.ExitOnError)
	agentOrchestrator := fs.String("agent-orchestrator", "http://localhost:50081", "Agent Orchestrator base URL")
	_ = fs.Parse(args)

	url := strings.TrimRight(*agentOrchestrator, "/") + "/v1/admin/tenants"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("list tenants failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func tenantsCreate(args []string) {
	fs := flag.NewFlagSet("tenants create", flag.ExitOnError)
	agentOrchestrator := fs.String("agent-orchestrator", "http://localhost:50081", "Agent Orchestrator base URL")
	id := fs.String("id", "", "tenant id (required)")
	name := fs.String("name", "", "tenant name")
	plan := fs.String("plan", "", "plan tier")
	_ = fs.Parse(args)

	if strings.TrimSpace(*id) == "" {
		log.Fatal("tenant id required")
	}

	payload := map[string]any{
		"tenant_id": *id,
	}
	if *name != "" {
		payload["name"] = *name
	}
	if *plan != "" {
		payload["plan_tier"] = *plan
	}
	b, _ := json.Marshal(payload)
	url := strings.TrimRight(*agentOrchestrator, "/") + "/v1/admin/tenants"
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("create tenant failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func printAccess(endpoints map[string]string) {
	log.Println("==> access")
	if v, ok := endpoints["agent-orchestrator"]; ok {
		log.Printf("Agent Orchestrator health: %s/v1/health", v)
	}
	if v, ok := endpoints["model-policy"]; ok {
		log.Printf("Model Policy health: %s/v1/health", v)
	}
	if v, ok := endpoints["federation"]; ok {
		log.Printf("Federation health: %s/v1/federation/health", v)
	}
}
