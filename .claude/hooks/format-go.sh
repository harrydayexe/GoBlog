#!/usr/bin/env bash
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

file=$(jq -r '.tool_input.file_path // empty')
[[ "$file" == *.go ]] || exit 0
gofmt -w "$file" && goimports -w "$file"
