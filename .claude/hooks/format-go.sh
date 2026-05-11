#!/usr/bin/env bash
file=$(jq -r '.tool_input.file_path // empty')
[[ "$file" == *.go ]] || exit 0
gofmt -w "$file" && goimports -w "$file"
