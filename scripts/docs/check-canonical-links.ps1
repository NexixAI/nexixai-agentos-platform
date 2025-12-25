Param(
  [string]$BasePath = "."
)

$legacyPatterns = @(
  # Legacy plan paths that should not appear in non-stub docs
  "docs/plan/INDEX.md",
  "docs/plan/phase-",
  "docs/plan/tracks/",
  "docs/product/agentos-prs/v1.02-prs.md",
  "docs/product/agentos-prs/v1.02-schemas-appendix.md",
  "docs/design/agentos-v1.02-design.md"
)

$ignorePatterns = @(
  # Legacy plan index stub is expected to mention legacy paths
  "docs/plan/LEGACY.md",
  # Legacy compatibility stub for plan index
  "docs/plan/INDEX.md",
  # Legacy phase stubs allowed to include old links
  "docs/plan/phase-*.md",
  # Legacy track index stub allowed to include old links
  "docs/plan/tracks/INDEX.md",
  # Legacy track stubs allowed to include old links
  "docs/plan/tracks/track-*.md",
  # Legacy PRS stub location (self-referential)
  "docs/product/agentos-prs/v1.02-prs.md",
  # Legacy Schemas Appendix stub location (self-referential)
  "docs/product/agentos-prs/v1.02-schemas-appendix.md",
  # Legacy design stub location (self-referential)
  "docs/design/agentos-v1.02-design.md",
  # Canonical frozen PRS folder may contain references by design
  "docs/product/agentos-prs/v1.02/*.md",
  # Canonical frozen design folder may contain references by design
  "docs/design/v1.02/*.md",
  # Process doc compatibility stub for execution plan path
  "docs/plan/agentos-v1.02-execution-plan.md",
  # Process doc compatibility stub for progress log path
  "docs/plan/progress-log.md",
  # Process doc compatibility stub for PR checklist path
  "docs/plan/pr-checklist.md"
)

function ShouldIgnore($path) {
  foreach ($pattern in $ignorePatterns) {
    if ($path -like $pattern) {
      return $true
    }
  }
  return $false
}

$matches = @()

Get-ChildItem -Path $BasePath -Recurse -Filter *.md -File | ForEach-Object {
  $relativePath = $_.FullName.Substring((Resolve-Path $BasePath).Path.Length).TrimStart('\','/').Replace('\','/')
  if (ShouldIgnore $relativePath) {
    return
  }

  foreach ($pattern in $legacyPatterns) {
    $hits = Select-String -Path $_.FullName -SimpleMatch $pattern
    foreach ($hit in $hits) {
      $matches += [PSCustomObject]@{
        File   = $relativePath
        String = $pattern
      }
    }
  }
}

if ($matches.Count -gt 0) {
  Write-Host "Legacy references detected:"
  foreach ($m in $matches) {
    Write-Host "  $($m.File) -> $($m.String)"
  }
  exit 1
}

Write-Host "No legacy doc references found."
exit 0
