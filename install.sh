#!/usr/bin/env bash
#
# install.sh — build kickdesk and register it so it runs from any directory.
#
#   ./install.sh              build, symlink ~/.local/bin/kickdesk, patch ~/.bashrc
#   ./install.sh --uninstall  remove the symlink and the bashrc block
#   ./install.sh --self-test  run installer checks with a fake HOME (no real files)
#
# Matches the fleet pattern (symlink into ~/.local/bin) and writes a marked
# ~/.bashrc block so KICKDESK_ROOT and PATH survive new shells.
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NAME="kickdesk"
BEGIN_MARK="# BEGIN kickdesk"
END_MARK="# END kickdesk"
MIN_GO_MINOR=25
GO_BOOTSTRAP_VERSION="${KICKDESK_GO_VERSION:-1.25.14}"

usage() {
  cat <<'EOF'
usage: ./install.sh [--uninstall | --self-test | --help]

  Build the kickdesk binary, symlink it to ~/.local/bin/kickdesk, and register
  a marked block in ~/.bashrc (KICKDESK_ROOT + ~/.local/bin on PATH).

  Re-run after pulling to rebuild. Idempotent: the bashrc block is replaced,
  not duplicated.

environment:
  KICKDESK_BINDIR     install dir (default: ~/.local/bin)
  KICKDESK_BASHRC     bashrc path (default: ~/.bashrc)
  KICKDESK_BUILD_DIR  binary output dir (default: <repo>/bin)
  KICKDESK_GO_DIR     user-local Go toolchain if PATH has none
                      (default: ~/.local/share/kickdesk/go)
  KICKDESK_GO_VERSION Go tarball version when bootstrapping (default: 1.25.14)
  KICKDESK_INSTALL_GO 1 (default) bootstrap Go if missing; 0 fail instead
EOF
}

die() {
  echo >&2 "install.sh: $*"
  exit 1
}

bin_dir() {
  echo "${KICKDESK_BINDIR:-$HOME/.local/bin}"
}

bashrc_path() {
  echo "${KICKDESK_BASHRC:-$HOME/.bashrc}"
}

go_dir() {
  echo "${KICKDESK_GO_DIR:-$HOME/.local/share/kickdesk/go}"
}

repo_binary() {
  echo "${KICKDESK_BUILD_DIR:-$ROOT/bin}/${BIN_NAME}"
}

link_path() {
  echo "$(bin_dir)/${BIN_NAME}"
}

go_minor() {
  # stdin: "go1.25.14" or "go version go1.25.14 linux/amd64"
  local raw rest
  raw="$(cat)"
  raw="${raw##*go}"
  raw="${raw%% *}"
  rest="${raw#*.}"
  echo "${rest%%.*}"
}

go_usable() {
  local exe="$1"
  [ -x "$exe" ] || return 1
  local ver minor
  ver="$("$exe" env GOVERSION 2>/dev/null || true)"
  if [ -z "$ver" ]; then
    ver="$("$exe" version 2>/dev/null || true)"
  fi
  [ -n "$ver" ] || return 1
  minor="$(printf '%s\n' "$ver" | go_minor)"
  case "$minor" in
    ''|*[!0-9]*) return 1 ;;
  esac
  [ "$minor" -ge "$MIN_GO_MINOR" ]
}

resolve_go() {
  local candidate
  if candidate="$(command -v go 2>/dev/null || true)" && go_usable "$candidate"; then
    echo "$candidate"
    return 0
  fi
  candidate="$(go_dir)/bin/go"
  if go_usable "$candidate"; then
    echo "$candidate"
    return 0
  fi
  return 1
}

bootstrap_go() {
  local dest os arch goarch url tmp
  dest="$(go_dir)"
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64) goarch=amd64 ;;
    aarch64|arm64) goarch=arm64 ;;
    *) die "unsupported architecture ${arch} (need amd64 or arm64)" ;;
  esac
  case "$os" in
    linux|darwin) ;;
    *) die "unsupported OS ${os}" ;;
  esac

  url="https://go.dev/dl/go${GO_BOOTSTRAP_VERSION}.${os}-${goarch}.tar.gz"
  echo >&2 "Go 1.${MIN_GO_MINOR}+ not on PATH; installing go${GO_BOOTSTRAP_VERSION} to ${dest}"
  command -v curl >/dev/null 2>&1 || die "curl is required to download Go"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN
  curl -fsSL "$url" -o "${tmp}/go.tgz"
  tar -C "$tmp" -xzf "${tmp}/go.tgz"
  [ -x "${tmp}/go/bin/go" ] || die "Go tarball did not contain bin/go"
  rm -rf "$dest"
  mkdir -p "$(dirname "$dest")"
  mv "${tmp}/go" "$dest"
  trap - RETURN
  rm -rf "$tmp"
  echo >&2 "Go toolchain: ${dest}/bin/go"
}

ensure_go() {
  local exe
  if exe="$(resolve_go)"; then
    echo "$exe"
    return 0
  fi
  if [ "${KICKDESK_INSTALL_GO:-1}" != "1" ]; then
    die "go 1.${MIN_GO_MINOR}+ is required (https://go.dev/doc/install)"
  fi
  bootstrap_go
  exe="$(go_dir)/bin/go"
  go_usable "$exe" || die "bootstrapped Go is not usable at ${exe}"
  echo "$exe"
}

remove_bashrc_block() {
  local bashrc="$1"
  [ -f "$bashrc" ] || return 0
  if grep -q "^${BEGIN_MARK}$" "$bashrc"; then
    sed -i "/^${BEGIN_MARK}\$/,/^${END_MARK}\$/d" "$bashrc"
  fi
}

write_bashrc_block() {
  local bashrc="$1"
  local bindir="$2"
  local path_lit case_lit
  mkdir -p "$(dirname "$bashrc")"
  touch "$bashrc"
  remove_bashrc_block "$bashrc"
  if [ -s "$bashrc" ] && [ "$(tail -c1 "$bashrc" | wc -l)" -ne 1 ]; then
    printf '\n' >>"$bashrc"
  fi
  if [ "$bindir" = "${HOME}/.local/bin" ]; then
    path_lit='$HOME/.local/bin'
    case_lit='$HOME/.local/bin'
  else
    path_lit="$bindir"
    case_lit="$bindir"
  fi
  cat >>"$bashrc" <<EOF

${BEGIN_MARK}
# Added by ${ROOT}/install.sh — kickdesk available from any directory.
export KICKDESK_ROOT="${ROOT}"
if [ -d "${path_lit}" ]; then
  case ":\${PATH}:" in
    *":${case_lit}:"*) ;;
    *) export PATH="${path_lit}:\${PATH}" ;;
  esac
fi
${END_MARK}
EOF
}

build_binary() {
  local goexe="$1"
  local out
  out="$(repo_binary)"
  mkdir -p "$(dirname "$out")"
  echo "Building ${out}"
  (
    cd "$ROOT"
    CGO_ENABLED=0 "$goexe" build -o "$out" .
  )
  chmod +x "$out"
}

install_link() {
  local bindir src link
  bindir="$(bin_dir)"
  src="$(repo_binary)"
  link="$(link_path)"
  mkdir -p "$bindir"
  ln -sfn "$src" "$link"
  echo "Symlink: ${link} -> ${src}"
}

do_install() {
  local goexe bindir bashrc
  goexe="$(ensure_go)"
  echo "Using $($goexe version)"
  build_binary "$goexe"
  install_link
  bindir="$(bin_dir)"
  bashrc="$(bashrc_path)"
  write_bashrc_block "$bashrc" "$bindir"
  echo "Registered in ${bashrc} (${BEGIN_MARK} … ${END_MARK})"
  echo
  echo "Installed. New shells can run: ${BIN_NAME}"
  echo "This shell:"
  echo "  hash -r"
  echo "  export PATH=\"${bindir}:\$PATH\""
  echo "  ${BIN_NAME} config validate"
}

do_uninstall() {
  local link bashrc
  link="$(link_path)"
  bashrc="$(bashrc_path)"
  if [ -L "$link" ] || [ -f "$link" ]; then
    rm -f "$link"
    echo "Removed ${link}"
  else
    echo "No ${link} to remove"
  fi
  if [ -f "$bashrc" ] && grep -q "^${BEGIN_MARK}$" "$bashrc"; then
    remove_bashrc_block "$bashrc"
    echo "Removed bashrc block from ${bashrc}"
  else
    echo "No kickdesk bashrc block in ${bashrc}"
  fi
}

self_test() {
  local go_stub tmp
  tmp="$(mktemp -d)"
  # Trap must see a global; locals vanish before EXIT runs.
  _KICKDESK_SELF_TEST_TMP="$tmp"
  trap 'rm -rf "${_KICKDESK_SELF_TEST_TMP:-}"' EXIT
  go_stub="${tmp}/mockgo/go"
  mkdir -p "$(dirname "$go_stub")"
  cat >"$go_stub" <<'STUB'
#!/bin/sh
if [ "$1" = "env" ] && [ "${2:-}" = "GOVERSION" ]; then
  echo "go1.25.14"
  exit 0
fi
if [ "$1" = "version" ]; then
  echo "go version go1.25.14 linux/amd64"
  exit 0
fi
out=""
while [ $# -gt 0 ]; do
  case "$1" in
    -o) out="$2"; shift 2 ;;
    *) shift ;;
  esac
done
if [ -n "$out" ]; then
  mkdir -p "$(dirname "$out")"
  printf '#!/bin/sh\necho kickdesk-stub\n' >"$out"
  chmod +x "$out"
fi
STUB
  chmod +x "$go_stub"

  export HOME="${tmp}/home"
  export KICKDESK_INSTALL_GO=0
  export KICKDESK_BUILD_DIR="${tmp}/build"
  mkdir -p "$HOME" "$KICKDESK_BUILD_DIR"
  PATH="$(dirname "$go_stub"):/usr/bin:/bin"
  unset KICKDESK_BINDIR KICKDESK_BASHRC KICKDESK_GO_DIR

  do_install >/dev/null

  [ -x "$(repo_binary)" ] || die "self-test: repo binary missing"
  [ -L "$(link_path)" ] || die "self-test: symlink missing"
  grep -q "^${BEGIN_MARK}$" "$(bashrc_path)" || die "self-test: bashrc begin mark missing"
  grep -q "^${END_MARK}$" "$(bashrc_path)" || die "self-test: bashrc end mark missing"
  grep -q "KICKDESK_ROOT=\"${ROOT}\"" "$(bashrc_path)" || die "self-test: KICKDESK_ROOT missing"
  grep -c "^${BEGIN_MARK}$" "$(bashrc_path)" | grep -qx 1 || die "self-test: duplicated bashrc block"

  do_install >/dev/null
  grep -c "^${BEGIN_MARK}$" "$(bashrc_path)" | grep -qx 1 || die "self-test: reinstall duplicated bashrc block"

  do_uninstall >/dev/null
  [ ! -e "$(link_path)" ] || die "self-test: symlink still present"
  if grep -q "^${BEGIN_MARK}$" "$(bashrc_path)"; then
    die "self-test: bashrc block still present"
  fi

  echo "self-test: ok"
  rm -rf "$tmp"
  trap - EXIT
  unset _KICKDESK_SELF_TEST_TMP
}

case "${1:-}" in
  "" ) do_install ;;
  --uninstall) do_uninstall ;;
  --self-test) self_test ;;
  -h|--help) usage ;;
  *)
    echo >&2 "install.sh: unknown argument: $1"
    usage >&2
    exit 1
    ;;
esac
