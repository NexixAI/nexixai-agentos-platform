    package deploy

    import (
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "time"
    )

    type CheckResult struct {
        Name    string `json:"name"`
        OK      bool   `json:"ok"`
        Detail  string `json:"detail,omitempty"`
        URL     string `json:"url,omitempty"`
        LatencyMs int  `json:"latency_ms,omitempty"`
    }

    type Summary struct {
        Timestamp string        `json:"timestamp"`
        Mode      string        `json:"mode"`
        Endpoints map[string]string `json:"endpoints"`
        Checks    []CheckResult `json:"checks"`
    }

    func WriteSummary(dir string, s Summary) error {
        s.Timestamp = time.Now().UTC().Format(time.RFC3339)
        // JSON
        jb, err := json.MarshalIndent(s, "", "  ")
        if err != nil {
            return err
        }
        if err := os.WriteFile(filepath.Join(dir, "summary.json"), jb, 0o644); err != nil {
            return err
        }

        // Markdown
        md := "# AgentOS Deployment Summary

"
        md += fmt.Sprintf("- Timestamp: `%s`
", s.Timestamp)
        md += fmt.Sprintf("- Mode: `%s`

", s.Mode)

        md += "## Endpoints
"
        for k, v := range s.Endpoints {
            md += fmt.Sprintf("- **%s**: %s
", k, v)
        }
        md += "
## Checks
"
        for _, c := range s.Checks {
            status := "✅"
            if !c.OK {
                status = "❌"
            }
            md += fmt.Sprintf("- %s **%s** (%s) — %s (%dms)
", status, c.Name, c.URL, c.Detail, c.LatencyMs)
        }

        return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(md), 0o644)
    }
