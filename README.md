# malo-id-generator / nelo-id-generator

![Unittest status badge](https://github.com/hochfrequenz/go-template-repository/workflows/Unittests/badge.svg)
![Coverage status badge](https://github.com/hochfrequenz/go-template-repository/workflows/coverage/badge.svg)
![Linter status badge](https://github.com/hochfrequenz/go-template-repository/workflows/golangci-lint/badge.svg)

🇩🇪 Dieses Repository enthält den Source Code hinter den Websites:
* [markt.lokations.id](https://markt.lokations.id), einem Generator für Marktlokations-IDs ("MaLo-ID") zu Testzwecken
* [netz.lokations.id](https://netz.lokations.id), einem Generator für Netzlokations-IDs ("NeLo-ID") zu Testzwecken
* [mess.lokations.id](https://mess.lokations.id), einem Generator für Messlokations-IDs ("MeLo-ID") zu Testzwecken
* [technische.ressource.id](https://technische.ressource.id), einem Generator für Technische Ressourcen-IDs ("TR-ID") zu Testzwecken
* [steuerbare.ressource.id](https://steuerbare.ressource.id), einem Generator für Steuerbare Ressourcen-IDs ("SR-ID") zu Testzwecken
* lokations.buendel.id, einem Generator für Lokationsbündel-IDs ("LoBü-ID") zu Testzwecken (⚠️ noch nicht deployed, siehe unten)

🇬🇧 This repository contains
an [Azure Function with a Go Handler](https://docs.microsoft.com/en-us/azure/azure-functions/create-first-function-vs-code-other?tabs=go%2Cwindows) which is deployed to [netz.lokations.id](https://netz.lokations.id) and [markt.lokations.id](https://markt.lokations.id).
The tech stack is historically grown and not a good choice - but it works 😉

Its purpose is to

- generate:
  1. Marktlokations-IDs (MaLo-IDs)
  2. Netzlokations-IDs (NeLo-IDs)
  3. Messlokations-IDs (MeLo-IDs)
  4. Technische Ressourcen-IDs (TR-IDs)
  5. Steuerbare Ressourcen-IDs (SR-IDs)
  6. Lokationsbündel-IDs (LoBü-IDs)
- with a valid checksum
- on the fly

The business logic is written in Go using [Gin Gonic](https://gin-gonic.com/) and can be found in [cmd/api.go](cmd/api.go).

It's a super basic website with three "pseudo files":

1. [`/` (root)](https://malo-id-generator.azurewebsites.net/) that returns a basic HTML site which refers to (this is the main entry point for users)
2. `/api/favicon` (returns a favicon) and refers to
3. `/api/style` (returns a stylesheet)
4. `/json` returns a JSON payload with the generated ID

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
Both Function apps are assigned to the [malo-id-generator resource group on Azure](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourceGroups/malo-id-generator/overview).
There is one function app instance per supported ID type.
This is because to use the function app directly behind top level domain registered in Azure, its respective entry point must be a top level domain itself without any further, relative path (e.g. `foobarsomerandomstring.azurewebsites.net` and _not_ `foobarsomerandomstring.azurewebsites.net/malo`).

| Purpose           | `ID_TYPE_TO_GENERATE` env var value | Deployed to (URL)                                                                                                                                 | Settings                                                                                                                                                                                                                  |
|-------------------|-------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Generate MaLo-IDs | `"MALO"`                            | [`malo-id-generator.azurewebsites.net/`](https://malo-id-generator.azurewebsites.net/) and [markt.lokations.id](https://markt.lokations.id)       | [malo-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourceGroups/malo-id-generator/providers/Microsoft.Web/sites/malo-id-generator/appServices) |                                                                                                                                                                                                  |.
| Generate NeLo-IDs | `"NELO"`                            | [`nelo-id-generator.azurewebsites.net/`](https://nelo-id-generator.azurewebsites.net/) and [netz.lokations.id](https://netz.lokations.id)         | [nelo-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/nelo-id-generator/appServices) |
| Generate MeLo-IDs | `"MELO"`                            | [`melo-id-generator.azurewebsites.net/`](https://melo-id-generator.azurewebsites.net/) and [mess.lokations.id](https://mess.lokations.id)         | [melo-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourceGroups/malo-id-generator/providers/Microsoft.Web/sites/melo-id-generator/appServices) |
| Generate TR-IDs   | `"TRID"`                            | [`tr-id-generator.azurewebsites.net/`](https://tr-id-generator.azurewebsites.net/) and [technische.ressource.id](https://technische.ressource.id) | [tr-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/tr-id-generator/appServices)     |
| Generate SR-IDs   | `"SRID"`                            | [`sr-id-generator.azurewebsites.net/`](https://sr-id-generator.azurewebsites.net/) and [steuerbare.ressource.id](https://steuerbare.ressource.id) | [sr-id-generator](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/sr-id-generator/appServices)     |
| Generate LoBü-IDs | `"LOBUE"`                           | ⚠️ **not created yet**; planned: `lobue-id-generator.azurewebsites.net` and [lokations.buendel.id](https://lokations.buendel.id) | the function app still has to be created |

The function apps are all

- code based (instead of dockerized (todo @kevin))
- linux based (instead of windows)

There is an environment variable named `ID_TYPE_TO_GENERATE` which you can modify in the [function app settings](https://portal.azure.com/#@hochfrequenz.net/resource/subscriptions/1cdc65f0-62d2-4770-be11-9ec1da950c81/resourcegroups/malo-id-generator/providers/Microsoft.Web/sites/malo-id-generator/configuration).
Its value can be `"MALO"` or `"NELO"` or `"MELO"` or `"TRID"` or `"SRID"` or `"LOBUE"` at the moment.
If its value is not set or set to an invalid value, the function app will return a HTTP 501 error.
For your local tests you can modify the value in the `local.settings.json` file.

### How To Deploy

Deployment runs in Github Actions: [`deploy.yml`](.github/workflows/deploy.yml) builds the custom
handler for linux, assembles the same package that `func azure functionapp publish` would upload
(the `api` binary, `host.json` and one directory per function) and pushes it to every function app
that exists in one go.

Under the hood the action uploads the zip to the function app's own storage account and points
`WEBSITE_RUN_FROM_PACKAGE` at it with a SAS that is valid for one year - exactly what
`func azure functionapp publish` does. So redeploy at least annually: an app that is not
redeployed within a year stops starting, and the reason is not obvious.

It is **not** triggered by pushes to `main` - a merge should not deploy to production on its own.
Start it manually from the [Actions tab](https://github.com/Hochfrequenz/malo-id-generator/actions/workflows/deploy.yml)
("Run workflow"), or publish a Github release.

The workflow authenticates with a [federated credential](https://learn.microsoft.com/en-us/azure/developer/github/connect-from-azure-openid-connect)
instead of a stored password. This is not a preference: the Azure/functions-action documentation
states that [publish profile authentication is unsupported](https://github.com/Azure/functions-action#authentication-methods)
when the app runs on Linux in a Consumption plan and the project contains an executable file - which
is exactly this repo, with its `api` custom handler. OIDC is the only supported option here.

Three repository secrets have to exist:

| Secret | What it is |
|--------|------------|
| `AZURE_CLIENT_ID` | client ID of the identity that is allowed to deploy |
| `AZURE_TENANT_ID` | directory (tenant) ID |
| `AZURE_SUBSCRIPTION_ID` | the subscription that holds the `malo-id-generator` resource group |

The identity needs a federated credential whose subject matches this repository and the `Production`
environment, plus a role that includes `Microsoft.Web/sites/config/list/action` - the action reads
the app settings and the SCM credentials through ARM before it uploads. Microsoft's documented
minimum is [`Website Contributor`](https://github.com/Azure/functions-action#use-oidc-recommended),
which is narrower than `Contributor`.

Deployments run in the `Production`
[environment](https://github.com/Hochfrequenz/malo-id-generator/settings/environments), so required
reviewers can be configured there. Note that the environment gate applies per function app, so a
`Production` environment with required reviewers asks for one approval per app, not one per run.
The exact `az` commands are written up in
[#271](https://github.com/Hochfrequenz/malo-id-generator/issues/271).

`lobue-id-generator` is not in the workflow's list of function apps yet, because the function app
does not exist yet - see [#268](https://github.com/Hochfrequenz/malo-id-generator/issues/268).

#### Deploying by hand

Should the workflow be unavailable, the manual route still works.

First **build** locally for linux (note that the build is the same for all ID types, only the env var is different)

```bash
set GOOS=linux
go build -o api ./cmd
```
or
```ps
$env:GOOS = "linux"; go build -o api ./cmd
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
respectively (and similar for `lobue-id-generator`, `melo-id-generator`, `sr-id-generator` and `tr-id-generator`).

You have to be logged in (`az login`) using the [Azure CLI Tools](https://docs.microsoft.com/de-de/cli/azure/install-azure-cli-windows?tabs=azure-cli).

#### Deploy them all
```bash
# lobue-id-generator does not exist yet, see #268 - this line fails until it is created
func azure functionapp publish lobue-id-generator
func azure functionapp publish malo-id-generator
func azure functionapp publish melo-id-generator
func azure functionapp publish nelo-id-generator
func azure functionapp publish sr-id-generator
func azure functionapp publish tr-id-generator

```
