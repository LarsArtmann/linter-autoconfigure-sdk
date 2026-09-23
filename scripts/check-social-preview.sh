#!/usr/bin/env bash
# Social-preview asset guard (TODO T29): the committed PNG must be
# 1280x640 and under 1 MB (GitHub card slot limits). When rsvg-convert is
# available, the SVG is re-rendered and must reproduce the committed PNG
# byte-identically, so the two cannot drift apart silently.
set -euo pipefail

png="assets/branding/social-preview.png"
svg="assets/branding/social-preview.svg"

fail() {
	echo "ERROR: $1" >&2
	exit 1
}

[ -f "$png" ] || fail "missing $png"
[ -f "$svg" ] || fail "missing $svg"

size=$(stat -c%s "$png")
[ "$size" -lt 1000000 ] || fail "$png is $size bytes (>= 1 MB GitHub limit)"

# PNG IHDR: bytes 16..24 hold big-endian width and height.
read -r w1 w2 w3 w4 h1 h2 h3 h4 < <(od -An -tu1 -j16 -N8 "$png" | tr -s ' ')
width=$((w1 * 16777216 + w2 * 65536 + w3 * 256 + w4))
height=$((h1 * 16777216 + h2 * 65536 + h3 * 256 + h4))
[ "$width" = "1280" ] || fail "$png width is $width, want 1280"
[ "$height" = "640" ] || fail "$png height is $height, want 640"

if command -v rsvg-convert >/dev/null 2>&1; then
	rendered=$(mktemp)
	trap 'rm -f "${rendered}"' EXIT
	rsvg-convert --width=1280 --height=640 "$svg" >"$rendered"
	cmp -s "$rendered" "$png" || fail "$png differs from a fresh render of $svg (regenerate via assets/branding/generate.sh)"
	echo "OK: $png is 1280x640, $size bytes, byte-identical to the SVG render"
else
	echo "OK: $png is 1280x640, $size bytes (rsvg-convert absent: SVG parity check skipped)"
fi
