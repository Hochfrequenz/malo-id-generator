# malo-id-generator / nelo-id-generator

![Unittest status badge](https://github.com/hochfrequenz/go-template-repository/workflows/Unittests/badge.svg)
![Coverage status badge](https://github.com/hochfrequenz/go-template-repository/workflows/coverage/badge.svg)
![Linter status badge](https://github.com/hochfrequenz/go-template-repository/workflows/golangci-lint/badge.svg)

This repository contains
an [Azure Function with a Go Handler](https://docs.microsoft.com/en-us/azure/azure-functions/create-first-function-vs-code-other?tabs=go%2Cwindows) which is deployed to [netz.lokations.id](https://netz.lokations.id) and [markt.lokations.id](https://markt.lokations.id).

Its purpose is to

- generate:
  1. Marktlokations-IDs (MaLo-IDs)
  2. Netzlokations-IDs (NeLo-IDs)
- with a valid checksum
- on the fly

The business logic is written in Go using [Gin Gonic](https://gin-gonic.com/) and can be found in [cmd/api.go](cmd/api.go).

It's a super basic website with three "pseudo files":

1. [`/` (root)](https://malo-id-generator.azurewebsites.net/) that returns a basic HTML site which refers to (this is the main entry point for users)
2. `/api/favicon` (returns a favicon) and refers to
3. `/api/style` (returns a stylesheet)

The files are not really served as plain files as you would expect it from a usual web app setup, but they are all separate Azure Functions and hence have their own respective `function.json`.

The files are embedded into the go binary using `go:embed`.
This means you need to rebuild in order to change e.g. the stylesheet.

## Running it Locally

The setup is generally described quite well in [this article by Thorsten Hans](https://www.thorsten-hans.com/azure-functions-with-go/).

First install the [Azure Function Core Tools](https://docs.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Cwindows%2Ccsharp%2Cportal%2Cbash#v2).

Then, in the root directory of this repo, execute:

```bash
go build -o api ./cmd/
```

followed by (also in the repo root)

```bash
func start
```

## CI/CD

This function app is managed in two separate Azure Function Apps.
Both Function apps are assigned to the [malo-id-generator resource group on Azure](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/malo-id-generator/appServices).
There is one function app instance per supported ID type.
This is because to use the function app directly behind top level domain registered in Azure, its respective entry point must be a top level domain itself without any further, relative path (e.g. `foobarsomerandomstring.azurewebsites.net` and _not_ `foobarsomerandomstring.azurewebsites.net/malo`).

| Purpose           | `ID_TYPE_TO_GENERATE` env var value  | Deployed to (URL)                                                                      | Settings                                                                                                                                                                                                                  |
|-------------------|--------------------------------------|----------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Generate MaLo-IDs | `"MALO"`                             | [`malo-id-generator.azurewebsites.net/`](https://malo-id-generator.azurewebsites.net/) | [malo-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourceGroups/malo-id-generator/providers/Microsoft.Web/sites/malo-id-generator/appServices) and [markt.lokations.id](https://markt.lokations.id) |                                                                                                                                                                                                  |.
| Generate NeLo-IDs | `"NELO"`                             | [`nelo-id-generator.azurewebsites.net/`](https://nelo-id-generator.azurewebsites.net/) | [nelo-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/nelo-id-generator/appServices) and [netz.lokations.id](https://netz.lokations.id) |

The function apps are all

- code based (instead of dockerized (todo @kevin))
- linux based (instead of windows)

There is an environment variable named `ID_TYPE_TO_GENERATE` which you can modify in the [function app settings](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/malo-id-generator/configuration).
Its value can only be `"MALO"` or `"NELO"` at the moment.
If its value is not set or set to an invalid value, the function app will return a HTTP 501 error.
For your local tests you can modify the value in the `local.settings.json` file.

### How To Deploy

There is _no_ automatic deployment yet (fixable with docker).

To deploy:

First **build** locally for linux (note that the build is the same for all ID types, only the env var is different)

```bash
set GOOS=linux
go build -o api cmd/
```

The GOOS env var can be set in the build configuration in Goland.
The build should create an `api` (no file ending) file on root level.

Then **upload**

```bash
func azure functionapp publish malo-id-generator
```
or
```bash
func azure functionapp publish nelo-id-generator
```
respectively.

You have to be logged in (`az login`) using the [Azure CLI Tools](https://docs.microsoft.com/de-de/cli/azure/install-azure-cli-windows?tabs=azure-cli).
