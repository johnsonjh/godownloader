package main

import (
	"fmt"
)

func processGodownloader(repo string, filename string) (string, error) {
	cfg, err := Load(repo, filename)
	if err != nil {
		return "", fmt.Errorf("unable to parse: %s", err)
	}
	// get name template
	name, err := makeName(cfg.Archive.NameTemplate)
	cfg.Archive.NameTemplate = name
	if err != nil {
		return "", fmt.Errorf("unable generate name: %s", err)
	}

	return makeShell(shellGodownloader, cfg)
}

var shellGodownloader = `#!/bin/sh
set -e
#  Code generated by godownloader. DO NOT EDIT.
#

usage() {
  this=$1
  cat <<EOF
$this: download go binaries for {{ $.Release.GitHub.Owner }}/{{ $.Release.GitHub.Name }}

Usage: $this [-b] bindir [version]
  -b sets bindir or installation directory, default "./bin"
   [version] is a version number from
   https://github.com/{{ $.Release.GitHub.Owner }}/{{ $.Release.GitHub.Name }}/releases
   If version is missing, then an attempt to find the latest will be found.

Generated by godownloader
 https://github.com/goreleaser/godownloader

EOF
  exit 2
}

parse_args() {
  #BINDIR is ./bin unless set be ENV
  # over-ridden by flag below

  BINDIR=${BINDIR:-./bin}
  while getopts "b:h?" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      h | \?) usage "$0" ;;
    esac
  done
  shift $((OPTIND - 1))
  VERSION=$1
}
# this function wraps all the destructive operations
# if a curl|bash cuts off the end of the script due to
# network, either nothing will happen or will syntax error
# out preventing half-done work
execute() {
  TMPDIR=$(mktmpdir)
  echo "$PREFIX: downloading ${TARBALL_URL}"
  http_download "${TMPDIR}/${TARBALL}" "${TARBALL_URL}"

  echo "$PREFIX: verifying checksums"
  http_download "${TMPDIR}/${CHECKSUM}" "${CHECKSUM_URL}"
  hash_sha256_verify "${TMPDIR}/${TARBALL}" "${TMPDIR}/${CHECKSUM}"

  (cd "${TMPDIR}" && untar "${TARBALL}")
  install -d "${BINDIR}"
  install "${TMPDIR}/${BINARY}" "${BINDIR}/"
  echo "$PREFIX: installed as ${BINDIR}/${BINARY}"
}
is_supported_platform() {
  platform=$1
  found=1
  case "$platform" in
  {{- range $goos := $.Build.Goos }}{{ range $goarch := $.Build.Goarch }}
{{ if not (eq $goarch "arm") }}    {{ $goos }}/{{ $goarch }}) found=0 ;;{{ end }}
  {{- end }}{{ end }}
  {{- if $.Build.Goarm }}
  {{- range $goos := $.Build.Goos }}{{ range $goarch := $.Build.Goarch }}{{ range $goarm := $.Build.Goarm }}
{{- if eq $goarch "arm" }}  {{ $goos }}/armv{{ $goarm }}) found=0 ;;
{{ end }}
  {{- end }}{{ end }}{{ end }}
  {{- end }}
  esac
  {{- if $.Build.Ignore }}
  case "$platform" in 
    {{- range $ignore := $.Build.Ignore }}
    {{ $ignore.Goos }}/{{ $ignore.Goarch }}{{ if $ignore.Goarm }}v{{ $ignore.Goarm }}{{ end }}) found=1 ;;{{ end }}
  esac
  {{- end }}
  return $found
}
check_platform() {
  if is_supported_platform "$PLATFORM"; then
    # optional logging goes here
    true
  else
    echo "${PREFIX}: platform $PLATFORM is not supported.  Make sure this script is up-to-date and file request at https://github.com/${PREFIX}/issues/new"
    exit 1
  fi
}
adjust_version() {
  if [ -z "${VERSION}" ]; then
    echo "$PREFIX: checking GitHub for latest version"
    VERSION=$(github_last_release "$OWNER/$REPO")
  fi
  # if version starts with 'v', remove it
  VERSION=${VERSION#v}
}
adjust_format() {
  # change format (tar.gz or zip) based on ARCH
  {{- with .Archive.FormatOverrides }}
  case ${ARCH} in
  {{- range . }}
    {{ .Goos }}) FORMAT={{ .Format }} ;;
  esac
  {{- end }}
  {{- end }}
  true
}
adjust_os() {
  # adjust archive name based on OS
  {{- with .Archive.Replacements }}
  case ${OS} in
  {{- range $k, $v := . }}
    {{ $k }}) OS={{ $v }} ;;
  {{- end }}
  esac
  true
}
adjust_arch() {
  # adjust archive name based on ARCH
  case ${ARCH} in
  {{- range $k, $v := . }}
    {{ $k }}) ARCH={{ $v }} ;;
  {{- end }}
  esac
  {{- end }}
  true
}
` + shellfn + `
OWNER={{ $.Release.GitHub.Owner }}
REPO={{ $.Release.GitHub.Name }}
BINARY={{ .Build.Binary }}
FORMAT={{ .Archive.Format }}
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="$OWNER/$REPO"
PLATFORM="${OS}/${ARCH}"
GITHUB_DOWNLOAD=https://github.com/${OWNER}/${REPO}/releases/download

uname_os_check "$OS"
uname_arch_check "$ARCH"

parse_args "$@"

check_platform

adjust_version

adjust_format

adjust_os

adjust_arch

echo "$PREFIX: found version ${VERSION} for ${OS}/${ARCH}"

{{ .Archive.NameTemplate }}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/v${VERSION}/${TARBALL}
CHECKSUM=${REPO}_checksums.txt
CHECKSUM_URL=${GITHUB_DOWNLOAD}/v${VERSION}/${CHECKSUM}

# Adjust binary name if windows
if [ "$OS" = "windows" ]; then
  BINARY="${BINARY}.exe"
fi

execute
`
