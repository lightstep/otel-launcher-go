#!/bin/sh

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "version must be set."
    exit 1
fi

echo $VERSION > ./VERSION

cat > locl/version.go <<EOF
package locl

const version = "$VERSION"
EOF

