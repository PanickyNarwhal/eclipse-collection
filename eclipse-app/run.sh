#!/usr/bin/env bash
# Quick launcher script for Castlevania Eclipse Collection desktop app
cd "$(dirname "$0")"

echo "Starting Castlevania: Eclipse Collection..."
export CGO_LDFLAGS="-framework UniformTypeIdentifiers"
go run -tags dev -buildvcs=false .
