#!/bin/sh
# Copyright Lightstep Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "version must be set."
    exit 1
fi

echo $VERSION > ./VERSION

sed -i '' "s/const version.*/const version = \"$VERSION\"/" ./launcher/version.go

(cd pipelines && go get github.com/lightstep/otel-launcher-go/lightstep/sdk/metric@v$VERSION)
(cd pipelines && go get github.com/lightstep/otel-launcher-go/lightstep/instrumentation@v$VERSION)
(cd lightstep/sdk/metric && go get github.com/lightstep/otel-launcher-go/lightstep/sdk/internal@v$VERSION)
(cd lightstep/sdk/trace && go get github.com/lightstep/otel-launcher-go/lightstep/sdk/internal@v$VERSION)
(cd lightstep/sdk/metric/example && go get github.com/lightstep/otel-launcher-go/lightstep/sdk/metric@v$VERSION)

go get github.com/lightstep/otel-launcher-go/pipelines@v$VERSION
