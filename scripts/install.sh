#!/usr/bin/env sh
# Install Reddock from GitHub Releases to your system (Linux x86_64 only).
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/OWNER/reddock/BRANCH/scripts/install.sh | sh
#   curl -fsSL ... | sh -s -- --help
#
# Environment:
#   REDDOCK_REPO   default: jimed-rand/reddock
#   PREFIX         default: /usr/local  (binary goes to $PREFIX/bin)

set -eu

REDDOCK_REPO="${REDDOCK_REPO:-jimed-rand/reddock}"
PREFIX="${PREFIX:-/usr/local}"
TAG=""
DRY_RUN=0

usage() {
	printf '%s\n' "Usage: install.sh [options]
  --tag TAG     Release tag (e.g. v0.2.0). Default: latest GitHub release.
  --prefix DIR  Install prefix (default: $PREFIX → reddock at DIR/bin/reddock)
  --dry-run     Print actions only
  -h, --help    This help
"
}

while [ $# -gt 0 ]; do
	case "$1" in
		--tag) TAG="${2:-}"; shift 2 ;;
		--prefix) PREFIX="${2:-}"; shift 2 ;;
		--dry-run) DRY_RUN=1; shift ;;
		-h|--help) usage; exit 0 ;;
		*) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
	esac
done

if [ "$(uname -s)" != "Linux" ]; then
	echo "Reddock only supports Linux." >&2
	exit 1
fi

arch="$(uname -m)"
case "$arch" in
	x86_64|amd64) ;;
	*) echo "Unsupported architecture: $arch (need x86_64)." >&2; exit 1 ;;
esac

if ! command -v curl >/dev/null 2>&1; then
	echo "curl is required." >&2
	exit 1
fi

if [ -z "$TAG" ]; then
	TAG="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REDDOCK_REPO}/releases/latest")"
	TAG="${TAG##*/}"
	if [ -z "$TAG" ] || [ "$TAG" = "latest" ]; then
		echo "Could not resolve latest release tag." >&2
		exit 1
	fi
fi

case "$TAG" in
	v*) ;;
	*) TAG="v${TAG#v}" ;;
esac

URL="https://github.com/${REDDOCK_REPO}/releases/download/${TAG}/reddock"
DEST="${PREFIX}/bin/reddock"
TMP="$(mktemp "${TMPDIR:-/tmp}/reddock-install.XXXXXX")"
trap 'rm -f "$TMP"' EXIT INT HUP

echo "Repository: ${REDDOCK_REPO}"
echo "Release:    ${TAG}"
echo "URL:        ${URL}"
echo "Install to: ${DEST}"

if [ "$DRY_RUN" -eq 1 ]; then
	exit 0
fi

mkdir -p "${PREFIX}/bin"
curl -fsSL -o "$TMP" "$URL"
chmod +x "$TMP"

if [ ! -x "$TMP" ]; then
	echo "Downloaded file is not executable." >&2
	rm -f "$TMP"
	exit 1
fi

if [ "$(id -u)" -eq 0 ]; then
	mv -f "$TMP" "$DEST"
else
	echo "Installing to ${DEST} requires root; using sudo."
	sudo mv -f "$TMP" "$DEST"
fi

echo "Done. Try: reddock version"
command -v reddock >/dev/null 2>&1 && reddock version || true
