#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

is_text_file() {
  case "$1" in
    *.md|*.txt|*.json|*.yml|*.yaml|*.ps1|*.sh|.gitignore|.gitattributes|.editorconfig)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

violations=0

while IFS= read -r -d '' file; do
  is_text_file "$file" || continue

  full_path="${repo_root}/${file}"
  [[ -f "$full_path" ]] || continue

  if [[ "$(head -c 3 "$full_path" | od -An -tx1 | tr -d ' \n')" == "efbbbf" ]]; then
    echo "violation: ${file}: contains UTF-8 BOM" >&2
    violations=1
  fi

  if ! iconv -f UTF-8 -t UTF-8 "$full_path" >/dev/null 2>&1; then
    echo "violation: ${file}: is not valid UTF-8" >&2
    violations=1
    continue
  fi

  if LC_ALL=C grep -n $'\r' "$full_path" >/dev/null; then
    echo "violation: ${file}: contains CRLF or carriage returns" >&2
    violations=1
  fi

  if [[ -s "$full_path" ]] && [[ "$(tail -c 1 "$full_path" | wc -l)" -eq 0 ]]; then
    echo "violation: ${file}: is missing a final newline" >&2
    violations=1
  fi

  case "$file" in
    *.md)
      ;;
    *)
      if grep -nE '[[:blank:]]+$' "$full_path" >/dev/null; then
        echo "violation: ${file}: contains trailing whitespace" >&2
        grep -nE '[[:blank:]]+$' "$full_path" >&2 || true
        violations=1
      fi
      ;;
  esac
done < <(git -C "$repo_root" ls-files --cached --others --exclude-standard -z)

if [[ "$violations" -ne 0 ]]; then
  exit 1
fi

echo "text file hygiene checks passed."
