#!/usr/bin/env bash
set -euo pipefail

CHANGE_CRITERIA="${1:-}"
VERSION_FILE="${VERSION_FILE:-VERSION}"

if [[ ! -f "${VERSION_FILE}" ]]; then
  echo "0.1.0" > "${VERSION_FILE}"
fi

CURRENT_VERSION="$(cat "${VERSION_FILE}")"
NEW_VERSION="${CURRENT_VERSION}"

if [[ -n "${CHANGE_CRITERIA}" ]]; then
  IFS='.' read -r major minor patch <<< "${CURRENT_VERSION}"
  case "${CHANGE_CRITERIA}" in
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    cosmetics|patch)
      patch=$((patch + 1))
      ;;
    *)
      echo "Unknown change criteria: ${CHANGE_CRITERIA} (use major|minor|cosmetics|patch)" >&2
      exit 1
      ;;
  esac
  NEW_VERSION="${major}.${minor}.${patch}"
  echo "${NEW_VERSION}" > "${VERSION_FILE}"
fi

echo "Using version: ${NEW_VERSION}"

if git rev-parse --git-dir >/dev/null 2>&1; then
  git checkout -b "release/v${NEW_VERSION}" 2>/dev/null || git checkout "release/v${NEW_VERSION}"
  git add "${VERSION_FILE}"
  git commit -m "chore: release v${NEW_VERSION}" || true
  git tag -a "release" -m "release v${NEW_VERSION}"
  echo "Checked out release/v${NEW_VERSION} and tagged release"
else
  echo "Not a git repo; skipping tag."
fi
