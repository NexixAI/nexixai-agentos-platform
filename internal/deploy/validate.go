package deploy

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

type Validator struct {
    StackA string
    StackB string
    Fed    string
    RepoRoot string
}

func (v Validator) ValidateAll() ([]CheckResult, error) {
    checks := []CheckResult{}
    // Health
    checks = append(checks, v.checkGET("stack-a health", v.StackA+"/v1/health"))
    checks = append(checks, v.checkGET("stack-b health", v.StackB+"/v1/health"))
    checks = append(checks, v.checkGET("federation health", v.Fed+"/v1/federation/health"))

    // Smoke: Stack A create run using canonical request
    reqPath := filepath.Join(v.RepoRoot, "docs", "api", "stack-a", "examples", "runs-create.request.json")
    b, err := os.ReadFile(reqPath)
    if err == nil {
        checks = append(checks, v.checkPOST("stack-a create run", v.StackA+"/v1/agents/agt_demo/runs", b))
    } else {
        checks = append(checks, CheckResult{Name: "stack-a create run", OK: false, Detail: "missing example file", URL: reqPath})
    }

    // Smoke: Stack B invoke using canonical request (chat.request.json)
    invPath := filepath.Join(v.RepoRoot, "docs", "api", "stack-b", "examples", "chat.request.json")
    b2, err2 := os.ReadFile(invPath)
    if err2 == nil {
        checks = append(checks, v.checkPOST("stack-b invoke", v.StackB+"/v1/models:invoke", b2))
    } else {
        checks = append(checks, CheckResult{Name: "stack-b invoke", OK: false, Detail: "missing example file", URL: invPath})
    }

    // Consider overall error if any failed
    failed := false
    for _, c := range checks {
        if !c.OK {
            failed = true
            break
        }
    }
    if failed {
        return checks, fmt.Errorf("validation failed (one or more checks failed)")
    }
    return checks, nil
}

func (v Validator) checkGET(name, url string) CheckResult {
    start := time.Now()
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    latency := int(time.Since(start).Milliseconds())
    if err != nil {
        return CheckResult{Name: name, OK: false, Detail: err.Error(), URL: url, LatencyMs: latency}
    }
    defer resp.Body.Close()
    _, _ = io.ReadAll(resp.Body)
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return CheckResult{Name: name, OK: true, Detail: resp.Status, URL: url, LatencyMs: latency}
    }
    return CheckResult{Name: name, OK: false, Detail: resp.Status, URL: url, LatencyMs: latency}
}

func (v Validator) checkPOST(name, url string, body []byte) CheckResult {
    start := time.Now()
    client := &http.Client{Timeout: 8 * time.Second}
    req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    latency := int(time.Since(start).Milliseconds())
    if err != nil {
        return CheckResult{Name: name, OK: false, Detail: err.Error(), URL: url, LatencyMs: latency}
    }
    defer resp.Body.Close()
    rb, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        // Ensure body is JSON for sanity
        var tmp any
        if json.Unmarshal(rb, &tmp) == nil {
            return CheckResult{Name: name, OK: true, Detail: resp.Status, URL: url, LatencyMs: latency}
        }
        return CheckResult{Name: name, OK: true, Detail: resp.Status + " (non-json body)", URL: url, LatencyMs: latency}
    }
    return CheckResult{Name: name, OK: false, Detail: resp.Status, URL: url, LatencyMs: latency}
}
