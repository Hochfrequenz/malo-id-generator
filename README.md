# malo-id-generator

![Unittest status badge](https://github.com/hochfrequenz/go-template-repository/workflows/Unittests/badge.svg)
![Coverage status badge](https://github.com/hochfrequenz/go-template-repository/workflows/coverage/badge.svg)
![Linter status badge](https://github.com/hochfrequenz/go-template-repository/workflows/golangci-lint/badge.svg)

This repository contains
an [Azure Function with a Go Handler](https://docs.microsoft.com/en-us/azure/azure-functions/create-first-function-vs-code-other?tabs=go%2Cwindows).

Its purpose is to

- generate Marktlokations-IDs
- with a valid checksum
- on the fly.

The business logic is written in Go using [Gin Gonic](https://gin-gonic.com/) and can be found in [cmd/api.go](cmd/api.go).

It's a super basic website with three "pseudo files":

1. `/api/generate-malo-id` that returns an HTML site which refers to
2. `/api/favicon` (returns a favicon) and refers to
3. `/api/style` (returns a stylesheet)

The files are not really served as plain files as you would expect it from a usual web app setup but they are all
separate Azure Functions and hence have their own respective `function.json`.

The files are embedded into the go binary using `go:embed`.
This means you need to rebuild in order to change e.g. the stylesheet.

## Running it Locally

The setup is generally described quite well in [this article by Thorsten Hans](https://www.thorsten-hans.com/azure-functions-with-go/).

First install the [Azure Function Core Tools](https://docs.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Cwindows%2Ccsharp%2Cportal%2Cbash#v2).

Then, in the root directory of this repo, execute:

```bash
go build ./cmd/api.go
```

followed by

```bash
func start
```

## CI/CD
