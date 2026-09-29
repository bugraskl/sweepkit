#!/usr/bin/env bash
# Builds the release archives, Linux packages and checksums in dist/ from
# the binaries the release workflow collected in out/<os>_<arch>/.
#
# Usage: scripts/package.sh v1.2.3
set -euo pipefail

tag="$1"
version="${tag#v}"
docs=(LICENSE README.md README.tr.md)

rm -rf dist stage
mkdir -p dist stage

for arch in amd64 arm64; do
  name="sweepkit_${version}_windows_${arch}"
  mkdir -p "stage/$name"
  cp "out/windows_${arch}/sweepkit.exe" "out/windows_${arch}/sweepkit-gui.exe" "${docs[@]}" "stage/$name/"
  (cd stage && zip -qr "../dist/$name.zip" "$name")

  name="sweepkit_${version}_linux_${arch}"
  mkdir -p "stage/$name"
  cp "out/linux_${arch}/sweepkit" "out/linux_${arch}/sweepkit-gui" \
    packaging/linux/sweepkit.desktop packaging/linux/sweepkit.png "${docs[@]}" "stage/$name/"
  chmod 0755 "stage/$name/sweepkit" "stage/$name/sweepkit-gui" "out/linux_${arch}/"*
  tar -C stage -czf "dist/$name.tar.gz" "$name"

  # nfpm only expands variables in a few fields, so the binaries it packs
  # always come from the same place.
  rm -rf stage/pkg && mkdir -p stage/pkg
  cp "stage/$name/sweepkit" "stage/$name/sweepkit-gui" stage/pkg/
  for packager in deb rpm archlinux; do
    VERSION="$version" ARCH="$arch" nfpm package --config packaging/linux/nfpm.yaml --packager "$packager" --target dist/
  done
done

# Release notes: this version's CHANGELOG section plus how to verify downloads.
{
  awk -v v="$version" '
    $0 ~ "^## \\[" v "\\]" { on = 1; next }
    /^## \[/ && on { exit }
    on { print }
  ' CHANGELOG.md
  cat <<EOF

### Verify your download

Every file below was built by GitHub Actions from the tagged source. Check it before running it:

\`\`\`sh
sha256sum --check --ignore-missing SHA256SUMS.txt
gh attestation verify <file> --repo bugraskl/sweepkit
\`\`\`
EOF
} > stage/NOTES.md

(cd dist && sha256sum -- * > SHA256SUMS.txt)
ls -l dist
