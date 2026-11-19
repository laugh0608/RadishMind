#!/usr/bin/env bash
set -euo pipefail

readonly candidate_python=/opt/python/cp312-cp312/bin/python
readonly input_root=/opt/radishmind/pillow-inputs
readonly native_prefix=/opt/radishmind/pillow-native
readonly tools_env=/opt/radishmind/pillow-tools
readonly native_recipe=/project/.radishmind/pillow/native-build.sh
readonly gcc_dir=/opt/rh/gcc-toolset-14/root/usr
readonly fribidi_dir="$input_root/fribidi/usr/lib64"
readonly -a build_environment=(
    "BUILD_PREFIX=$native_prefix"
    "PKG_CONFIG=/usr/bin/pkg-config"
    "PKG_CONFIG_PATH=$native_prefix/lib/pkgconfig:$native_prefix/share/pkgconfig"
    "PKG_CONFIG_LIBDIR=$native_prefix/system-pc"
    "PKG_CONFIG_SYSROOT_DIR="
    "CPPFLAGS=-I$native_prefix/include"
    "LDFLAGS=-L$native_prefix/lib"
    "LIBRARY_PATH=$native_prefix/lib"
    "CMAKE_PREFIX_PATH=$native_prefix"
    "LD_LIBRARY_PATH=$native_prefix/lib:$gcc_dir/lib64:$gcc_dir/lib:$gcc_dir/lib64/dyninst:$gcc_dir/lib/dyninst"
    "PIP_CONFIG_FILE=/dev/null"
    "PIP_NO_INDEX=1"
    "PIP_FIND_LINKS=$input_root/wheels"
)

fail() {
    printf '%s\n' "$*" >&2
    exit 1
}

[[ $# -le 1 ]] || fail 'Expected at most one mode argument.'
case "${1-}" in
    --cibw-environment)
        printf '%s\n' "${build_environment[*]}"
        exit 0
        ;;
    --cibw-test-environment)
        printf '%s\n' "LD_LIBRARY_PATH=$fribidi_dir"
        exit 0
        ;;
    '') ;;
    *) fail 'Usage: bash pillow-before-all.sh [--cibw-environment|--cibw-test-environment]' ;;
esac

[[ -f "$native_recipe" ]] || fail "Missing staged native recipe $native_recipe"
[[ -d "$input_root/wheels" ]] || fail 'Missing verified wheelhouse.'
[[ ! -e "$tools_env" ]] || fail "Expected a fresh tool environment at $tools_env"
[[ ! -e "$native_prefix" ]] || fail "Expected a fresh native prefix at $native_prefix"
[[ -f /usr/lib64/pkgconfig/xau.pc ]] || fail 'Missing the builder LibXau pkg-config file.'
[[ -f /usr/share/pkgconfig/xproto.pc ]] || fail 'Missing the builder xproto pkg-config file required by LibXau.'
[[ "$("$candidate_python" -I -c 'import sys; print(".".join(map(str, sys.version_info[:3])))')" == 3.12.15 ]] ||
    fail 'The pinned build interpreter must be CPython 3.12.15.'

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
export "${build_environment[@]}"
mkdir -p "$native_prefix/lib/pkgconfig" "$native_prefix/share/pkgconfig" "$native_prefix/system-pc"
ln -s /usr/lib64/pkgconfig/xau.pc "$native_prefix/system-pc/xau.pc"
ln -s /usr/share/pkgconfig/xproto.pc "$native_prefix/system-pc/xproto.pc"
"$PKG_CONFIG" --exists 'xau >= 0.99.2'

"$candidate_python" -I -m venv "$tools_env"
"$tools_env/bin/python" -I -m pip install --no-index --find-links "$input_root/wheels" \
    --require-hashes --only-binary=:all: -r "$script_dir/pillow-native-tools.lock"
[[ -x "$tools_env/bin/meson" && -x "$tools_env/bin/ninja" ]] || fail 'Missing native build tool entry points.'
export PATH="$tools_env/bin:${candidate_python%/*}:$PATH"
/bin/bash "$native_recipe"

"$PKG_CONFIG" --exact-version=1.17.0 xcb-proto
"$PKG_CONFIG" --exact-version=1.17.0 xcb
for pkg in xcb zlib; do
    [[ "$("$PKG_CONFIG" --variable=libdir "$pkg")" == "$native_prefix/lib" ]] ||
        fail "$pkg must select a library in $native_prefix/lib"
done
for lib in xcb z; do
    target=$(realpath -e -- "$native_prefix/lib/lib$lib.so")
    [[ "$target" == "$native_prefix/lib/"* ]] || fail "Unexpected link target $target"
    dynamic=$(LC_ALL=C "$gcc_dir/bin/readelf" -d "$target")
    [[ "$dynamic" == *"Library soname: [lib$lib.so.1]"* ]] || fail "Unexpected SONAME in $target"
done
[[ "$("$PKG_CONFIG" --variable=libdir xau)" == /usr/lib64 ]] || fail 'LibXau must come from the fixed builder.'
[[ "$(realpath -e -- /usr/lib64/libXau.so)" == /usr/lib64/libXau.so.6.0.0 ]] ||
    fail 'Unexpected LibXau link target.'
