---
title: "Installation"
type: docs
weight: 1
---

# Installation

There are various ways you can obtain the `terraform-provider-dcs` binary.

## Build and Install From Source

Make sure Go, `make` and [GoLangCI-Lint](https://github.com/golangci/golangci-lint#install) are installed.

1. On Linux call `PLATFORMS=linux make`. On macOS call `PLATFORMS=darwin make`
2. Copy the binary in `terraform.d/plugins` to `~/.terraform.d` with `cp -r terraform.d/plugin ~/.terraform.d`. 

