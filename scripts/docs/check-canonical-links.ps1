Param(
  [string]$BasePath = "."
)

$legacyPatterns = @(
  "docs/plan/INDEX.md",
  "docs/plan/phase-",
  "docs/plan/tracks/",
  "docs/product/agentos-prs/v1.02-prs.md",
  "docs/product/agentos-prs/v1.02-schemas-appendix.md",
  "docs/design/agentos-v1.02-design.md"
)

$ignorePatterns = @(
  "docs/plan/LEGACY.md",
  "docs/plan/INDEX.md",
  "docs/plan/phase-*.md",
  "docs/plan/tracks/INDEX.md",
  "docs/plan/tracks/track-*.md",
  "docs/product/agentos-prs/v1.02-prs.md",
  "docs/product/agentos-prs/v1.02-schemas-appendix.md",
  "docs/design/agentos-v1.02-design.md",
  "docs/product/agentos-prs/v1.02/*.md",
  "docs/design/v1.02/*.md",
  "docs/plan/agentos-v1.02-execution-plan.md",
  "docs/plan/progress-log.md",
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
