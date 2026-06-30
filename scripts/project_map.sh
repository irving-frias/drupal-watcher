#!/usr/bin/env bash
# Regenerate project_map.md
# Usage: ./scripts/project_map.sh

set -euo pipefail
cd "$(dirname "$0")/.."

OUT="project_map.md"

cat > "$OUT" << 'HEADER'
# Drupal Watcher — Project Map

> Auto-generated dependency and structure map.
> Regenerate with: `./scripts/project_map.sh`

## Entry Point

```
cmd/drupal-watcher/main.go
  └─ main() → crea App con 5 módulos:
       config → watcher → executor → orchestrator → ui
     Escribe PID, corre health check, arranca App
```
HEADER

echo "" >> "$OUT"
echo "## File Tree" >> "$OUT"
echo "\`\`\`" >> "$OUT"
find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' -not -name '*_test.go' | sort >> "$OUT"
echo "\`\`\`" >> "$OUT"

echo "" >> "$OUT"
echo "## Package Dependencies" >> "$OUT"
echo "\`\`\`" >> "$OUT"
go list -json ./... 2>/dev/null | jq -r '
  select(.Name != "main") |
  {pkg: .ImportPath, imports: .Imports} |
  "\(.pkg | sub("github.com/irving-frias/drupal-watcher/"; "")): \(.imports | map(select(startswith("github.com/irving-frias/drupal-watcher")) | sub("github.com/irving-frias/drupal-watcher/"; "")) | join(", "))"
' | grep -v '^[^:]*: $' | sort >> "$OUT"
echo "\`\`\`" >> "$OUT"

echo "" >> "$OUT"
echo "## Interfaces" >> "$OUT"
echo "\`\`\`" >> "$OUT"
grep -rn '^type .* interface' --include='*.go' internal/ pkg/ cmd/ | grep -v '_test.go' | sed 's/:.*type/ → type/' >> "$OUT"
echo "\`\`\`" >> "$OUT"

echo "" >> "$OUT"
echo "## Key Types" >> "$OUT"
echo "\`\`\`" >> "$OUT"
grep -rn '^type [A-Z]' --include='*.go' internal/ pkg/ cmd/ | grep -v '_test.go' | grep -v ' interface' | sed 's/:.*type/ → type/' >> "$OUT"
echo "\`\`\`" >> "$OUT"

echo "" >> "$OUT"
echo "## Exported Functions" >> "$OUT"
echo "\`\`\`" >> "$OUT"
grep -rn '^func [A-Z]' --include='*.go' internal/ pkg/ cmd/ | grep -v '_test.go' | grep -v '\.String\b' | sed 's/:.*func/ → func/' | sort >> "$OUT"
echo "\`\`\`" >> "$OUT"

echo "Regenerated $OUT"
