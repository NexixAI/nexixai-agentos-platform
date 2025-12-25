#!/usr/bin/env bash
set -euo pipefail

BASE_PATH="${1:-.}"

legacy_patterns=(
  # Legacy plan paths that should not appear in non-stub docs
  "docs/plan/INDEX.md"
  "docs/plan/phase-"
  "docs/plan/tracks/"
  "docs/product/agentos-prs/v1.02-prs.md"
  "docs/product/agentos-prs/v1.02-schemas-appendix.md"
  "docs/design/agentos-v1.02-design.md"
)

ignore_patterns=(
  # Legacy plan index stub is expected to mention legacy paths
  "docs/plan/LEGACY.md"
  # Legacy compatibility stub for plan index
  "docs/plan/INDEX.md"
  # Legacy phase stubs allowed to include old links
  "docs/plan/phase-*.md"
  # Legacy track index stub allowed to include old links
  "docs/plan/tracks/INDEX.md"
  # Legacy track stubs allowed to include old links
  "docs/plan/tracks/track-*.md"
  # Legacy PRS stub location (self-referential)
  "docs/product/agentos-prs/v1.02-prs.md"
  # Legacy Schemas Appendix stub location (self-referential)
  "docs/product/agentos-prs/v1.02-schemas-appendix.md"
  # Legacy design stub location (self-referential)
  "docs/design/agentos-v1.02-design.md"
  # Canonical frozen PRS folder may contain references by design
  "docs/product/agentos-prs/v1.02/*.md"
  # Canonical frozen design folder may contain references by design
  "docs/design/v1.02/*.md"
  # Process doc compatibility stub for execution plan path
  "docs/plan/agentos-v1.02-execution-plan.md"
  # Process doc compatibility stub for progress log path
  "docs/plan/progress-log.md"
  # Process doc compatibility stub for PR checklist path
  "docs/plan/pr-checklist.md"
)

should_ignore() {
  local path="$1"
  for pattern in "${ignore_patterns[@]}"; do
    if [[ "$path" == $pattern ]]; then
      return 0
    fi
  done
  return 1
}

matches=()

while IFS= read -r -d '' file; do
  rel_path="${file#"$BASE_PATH"/}"
  rel_path="${rel_path#./}"

  if should_ignore "$rel_path"; then
    continue
  fi

  for pattern in "${legacy_patterns[@]}"; do
    if grep -Fq -- "$pattern" "$file"; then
      matches+=("$rel_path -> $pattern")
    fi
  done
done < <(find "$BASE_PATH" -type f -name '*.md' -print0)

if ((${#matches[@]} > 0)); then
  echo "Legacy references detected:"
  for m in "${matches[@]}"; do
    echo "  $m"
  done
  exit 1
fi

echo "No legacy doc references found."
exit 0
