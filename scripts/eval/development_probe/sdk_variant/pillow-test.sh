#!/usr/bin/env bash
set -euo pipefail

readonly input_root=/opt/radishmind/pillow-inputs
test -r "$input_root/fribidi/usr/lib64/libfribidi.so.0"
test -d "$input_root/test-images"
test -d Tests/images
cp -R "$input_root/test-images/." Tests/images/

python3 selftest.py
python3 -m pytest -vv -x checks/check_wheel.py
python3 -m pytest -vv -x --numprocesses=logical --dist=worksteal
