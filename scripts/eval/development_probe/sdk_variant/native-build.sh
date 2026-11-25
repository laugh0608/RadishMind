#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
readonly script_dir
readonly manifest="$script_dir/pillow-native-sources.sha256"
readonly candidate_prefix=/opt/radishmind/pillow-native

unpack() {
    local component=$1 destination=$2 strip=$3
    mkdir -p "$destination"
    tar --extract --file "$archives/$component.tar" --directory "$destination" --strip-components="$strip"
}

collect_notices() {
    local notice
    for notice in xz/COPYING xz/COPYING.0BSD xz/COPYING.GPLv2 xz/COPYING.GPLv3 xz/COPYING.LGPLv2.1 \
        zlib-ng/LICENSE.md xcb-proto/COPYING libxcb/COPYING libjpeg-turbo/LICENSE.md libjpeg-turbo/README.ijg \
        zstd/LICENSE tiff/LICENSE.md libavif/LICENSE libavif/ext/aom/LICENSE libavif/ext/aom/PATENTS \
        libavif/ext/dav1d/COPYING libavif/ext/libyuv/LICENSE libavif/ext/libyuv/PATENTS libpng/LICENSE \
        lcms2/LICENSE openjpeg/LICENSE libwebp/COPYING libwebp/PATENTS brotli/LICENSE \
        freetype/LICENSE.TXT freetype/docs/FTL.TXT harfbuzz/COPYING bzip2/LICENSE; do
        test -s "$sources/$notice"
        printf '\n----\n\n%s\n\n' "$notice"
        cat "$sources/$notice"
    done
}

prepare_sources() {
    (cd "$archives" && sha256sum --check "$manifest")
    local name
    for name in xz zlib-ng xcb-proto libxcb libjpeg-turbo zstd tiff libavif libpng \
        lcms2 openjpeg libwebp brotli freetype harfbuzz bzip2; do
        unpack "$name" "$sources/$name" 1
    done
    unpack freetype "$sources/freetype-with-harfbuzz" 1
    unpack aom "$sources/libavif/ext/aom" 0
    unpack dav1d "$sources/libavif/ext/dav1d" 1
    unpack libyuv "$sources/libavif/ext/libyuv" 0
    unpack libwebp "$sources/libavif/ext/libwebp" 1
    patch --directory "$sources/openjpeg" --strip=0 --batch --forward --fuzz=0 < "$archives/openjpeg.patch"
    test -f "$sources/libavif/ext/aom/CMakeLists.txt"
    test -f "$sources/libavif/ext/dav1d/meson.build"
    test -f "$sources/libavif/ext/libyuv/CMakeLists.txt"
    test -f "$sources/libavif/ext/libwebp/CMakeLists.txt"
    collect_notices > "$sources/native-notices.txt"
}

if [[ $# == 3 && $1 == --prepare-only ]]; then
    archives=$(cd -- "$2" && pwd)
    mkdir -- "$3"
    sources=$(cd -- "$3" && pwd)
    readonly archives sources
    prepare_sources
    exit 0
fi

if [[ $# != 0 || ${BUILD_PREFIX-} != "$candidate_prefix" || $(uname -m) != aarch64 || $(uname -s) != Linux ]]; then
    printf '%s\n' 'Use the fixed AArch64 builder through before-all, or --prepare-only ARCHIVES NEW_DIRECTORY.' >&2
    exit 1
fi

readonly archives=/opt/radishmind/pillow-inputs/native
sources=$(mktemp -d)
readonly sources
trap 'rm -rf -- "$sources"' EXIT
prepare_sources

export GIT_ALLOW_PROTOCOL=
readonly jobs=4
readonly libdir="$BUILD_PREFIX/lib"

configure_install() (
    cd "$sources/$1"
    shift
    ./configure --prefix="$BUILD_PREFIX" --libdir="$libdir" "$@"
    make -j "$jobs"
    make install
)

cmake_install() {
    local name=$1
    shift
    cmake -G 'Unix Makefiles' -S "$sources/$name" -B "$sources/$name/build-offline" \
        -DCMAKE_INSTALL_PREFIX="$BUILD_PREFIX" -DCMAKE_INSTALL_LIBDIR="$libdir" \
        -DCMAKE_INSTALL_NAME_DIR="$libdir" -DCMAKE_BUILD_TYPE=Release \
        -DFETCHCONTENT_FULLY_DISCONNECTED=ON "$@"
    cmake --build "$sources/$name/build-offline" --parallel "$jobs"
    cmake --install "$sources/$name/build-offline"
}

configure_install xz
configure_install zlib-ng --zlib-compat
configure_install xcb-proto
# libxcb consumes the protocol directory from pkg-config on this Linux image.
sed 's/\${pc_sysrootdir}//g' "$BUILD_PREFIX/share/pkgconfig/xcb-proto.pc" > "$libdir/pkgconfig/xcb-proto.pc"
configure_install libxcb
cmake_install libjpeg-turbo -DCMAKE_POSITION_INDEPENDENT_CODE=ON
make -C "$sources/zstd" -j "$jobs" install PREFIX="$BUILD_PREFIX" LIBDIR="$libdir"
configure_install tiff

readonly avif="$sources/libavif"
meson setup "$avif/ext/dav1d/build" "$avif/ext/dav1d" --wrap-mode=nodownload --buildtype=release \
    --default-library=static -Db_staticpic=true -Denable_tools=false -Denable_tests=false
meson compile -C "$avif/ext/dav1d/build" -j "$jobs"
test -f "$avif/ext/dav1d/build/src/libdav1d.a"
cmake_install libavif -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=MinSizeRel \
    -DAVIF_BUILD_APPS=OFF -DAVIF_BUILD_TESTS=OFF -DAVIF_LIBSHARPYUV=LOCAL -DAVIF_LIBYUV=LOCAL \
    -DAVIF_CODEC_AOM=LOCAL -DAVIF_CODEC_AOM_DECODE=OFF -DAVIF_CODEC_DAV1D=LOCAL \
    -DCONFIG_AV1_HIGHBITDEPTH=0 -DCMAKE_INTERPROCEDURAL_OPTIMIZATION=ON \
    -DCMAKE_C_VISIBILITY_PRESET=hidden -DCMAKE_CXX_VISIBILITY_PRESET=hidden \
    -DCMAKE_SHARED_LINKER_FLAGS=-Wl,--strip-all,-z,relro,-z,now \
    -DFETCHCONTENT_SOURCE_DIR_LIBAOM="$avif/ext/aom" \
    -DFETCHCONTENT_SOURCE_DIR_LIBYUV="$avif/ext/libyuv" \
    -DFETCHCONTENT_SOURCE_DIR_LIBWEBP="$avif/ext/libwebp"

configure_install libpng
configure_install lcms2
cmake_install openjpeg
CFLAGS='-O3 -DNDEBUG' configure_install libwebp --enable-libwebpmux --enable-libwebpdemux
cmake_install brotli -DBUILD_SHARED_LIBS=ON
make -C "$sources/bzip2" -f Makefile-libbz2_so
make -C "$sources/bzip2" install PREFIX="$BUILD_PREFIX"
configure_install freetype --with-harfbuzz=no
meson setup "$sources/harfbuzz/build-offline" "$sources/harfbuzz" --wrap-mode=nodownload \
    --prefix "$BUILD_PREFIX" --libdir "$libdir" --buildtype=minsize -Dfreetype=enabled -Dglib=disabled -Dtests=disabled
meson compile -C "$sources/harfbuzz/build-offline" -j "$jobs"
meson install -C "$sources/harfbuzz/build-offline"
CFLAGS="${CFLAGS-} -DFT_CONFIG_OPTION_USE_HARFBUZZ" configure_install freetype-with-harfbuzz --with-harfbuzz=yes
cat "$sources/native-notices.txt" >> /project/LICENSE
