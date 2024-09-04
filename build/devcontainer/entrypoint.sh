#!/bin/sh

go mod download

exec "$@"
