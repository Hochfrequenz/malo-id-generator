# Hosting: cost analysis and constraints

**Status:** analysis, decided 2026-09-01 by @hf-kklein — *stay on Azure Functions for now, automate
the deployment.* See "What was decided instead", and "Open actions" plus "Revisit this decision when"
at the bottom. **Re-verify before relying on any figure here after 2027-09-01.**

This records why the "obviously over-engineered" architecture (one Azure Function App per ID type —
five deployed today, six once #268 lands — for six tiny websites) was **not** replaced, so that the next person to have the same idea does not have to redo
the research. All prices are West Europe list prices from the
[Azure Retail Prices API](https://prices.azure.com/api/retail/prices), collected 2026-09-01. The
public pricing pages render their numbers via JavaScript and cannot be scraped, which is why the API
was used. EUR figures are Microsoft's own published EUR list prices, not an FX conversion. Some rate-card lines
are published only in USD; those are marked with $ and are left unconverted.

## Traffic model

Everything below assumes what the sites actually get: **~20 page views per day per site, six sites**.
A page view is ~8 HTTP requests (HTML + CSS + 3 TTF fonts + 2 PNGs + favicon), so roughly
**30,000 requests and ~2 GB egress per month in total**. Rescale accordingly if that ever changes;
most conclusions below flip only at 30x this traffic.

Scope: Azure only, and only compute that can bind these six custom domains directly. Deliberately not
evaluated — **Azure Static Web Apps** (the pages are nearly static, but `/json` has to generate a new
ID with a valid checksum per request, which needs the Go code server-side), **one app behind Azure
Front Door** (adds a component and a bill to solve a routing problem that host-based routing already
solves), and **non-Azure hosting** (the domains, the subscription and the org's deployment tooling are
all here). Reopening one of those is new research, not a repeat of this.

## The numbers

| Option | Fixed €/month at zero traffic | Realistic total | Cold start |
|---|---|---|---|
| **Six Consumption (Y1) Function Apps** — status quo | €0.20–1 (storage only, no compute) | **€0.20–1** | 2–5 s, not mitigable without changing plan |
| Six Flex Consumption apps | €0.20–1 (storage only, same as status quo) | €0.20–1 | still present without always-ready instances |
| Six Flex apps with 1 always-ready instance each | €4.46/app | €27 | none |
| **One Container App**, min replicas 0, image from ghcr.io | €0 | **€0, worst case €2.60** | ~1–3 s for a small Go image |
| **One App Service for Containers**, Linux B1 + Always On | €11.32 | **€11.32** | none |

Zero in every option at this scale: egress (first 100 GB/month free), TLS certificates, and
Application Insights / Log Analytics (5 GB/month free ingestion).

The €/month figures are list prices from the Retail Prices API. The cold-start figures and the €2.60
worst case are estimates.

## Is the status quo actually nearly free? Yes.

This is the crux, and the answer holds up. The Functions Consumption free grant is
**1,000,000 executions + 400,000 GB-s per subscription per calendar month** — *per subscription, not
per app*, so all six share one grant:

- Executions: 30,000/month against 1,000,000 free — **~3% of the grant**.
- GB-s: even assuming a fat 512 MB bucket and a generous 300 ms per request,
  `30,000 x 0.5 GB x 0.3 s ~= 4,500 GB-s` against 400,000 free — **~1% of the grant**.

Traffic would have to grow roughly **30x** before a cent of compute is billed.

The only genuinely non-zero line is the **storage accounts** that every Function App requires for
`AzureWebJobsStorage`. Six apps do *not* need six storage accounts — the docs state that multiple
function apps can share one, and the only hazard is host-ID collision between app names longer than
32 characters sharing their first 32, which ours are not. The idle transaction count cannot be
derived from public documentation, but the rate card bounds it: even a million "other" operations a
month is $0.43. Expect **cents**. Confirm with Cost Analysis on the resource group filtered to
Storage if an exact number is ever needed.

**Conclusion: cost cannot justify a migration.** Any argument for moving has to rest on something
else.

## What was decided instead

Keep one Function App per ID type and remove the manual step, which is the part that has actually
bitten: deploying means running `func azure functionapp publish` once per app from a developer
machine, and a site that nobody performed those steps for is a site that is not live (#268).
Concretely, a GitHub Actions workflow that builds once for `GOOS=linux` and publishes the same
artefact to each app in a matrix, logging in with a federated credential the way
`fristenkalender-functions` does. That costs nothing, touches neither domains nor certificates nor the
runtime, and forecloses no later migration.

Accepted knowingly in exchange: **the 2–5 s cold starts stay.** At roughly 20 page views per day per
site of test-data traffic nobody is paying much for that latency, but it is a real cost of the "stay"
decision, and it is the one thing a Container App would have fixed at the same price.

## The constraints that actually shape this

### The domain constraint says less than it looks like

The README's justification for six apps is that a site behind a registered top level domain must be
reachable at that domain's root, `foo.azurewebsites.net` rather than `foo.azurewebsites.net/malo`.

That is a requirement for **each site to answer at `/` of its own hostname**. It is *not* a
requirement for one app per hostname — that is an inference, not the constraint. Both Azure App
Service and Container Apps bind many custom domains to a single app, and the incoming `Host` header
survives. Consolidating would therefore mean **host-based routing**: branch on the `Host` header,
with `ID_TYPE_TO_GENERATE` demoted to a local-dev and fallback override. Small, but not free, and
easy to overlook when comparing "six resources versus one".

### Deadlines on the current plan

- **Linux Consumption retires 30 September 2028** and receives no new features or language versions.
- **Apps on the v3 runtime on Linux Consumption stop running after 30 September 2026.** Check
  `FUNCTIONS_EXTENSION_VERSION` on the function apps — this one is close.

These, not cost, are the real reason a migration will eventually be necessary.

### Certificates on Consumption are on shaky ground

Microsoft's prerequisites for a free App Service Managed Certificate list the Basic, Standard,
Premium and Isolated tiers — **Consumption is not among them**. There is also a
[Microsoft Q&A report](https://learn.microsoft.com/en-us/answers/questions/5529074/deployment-ssl-certificate-for-custom-domain-stop)
of ASMC issuance failing on Consumption and Flex Consumption plans since 2025-08-15 — a third-party
report, not documentation. Function apps
on Consumption additionally support **only CNAME** domain mapping, not A records, which rules out a
true apex domain.

Azure Container Apps, by contrast, documents free managed certificates for **multiple custom domains
on one app**, apex (A + `asuid` TXT) and subdomain (CNAME + `asuid.` TXT) alike. Flex Consumption's
managed certificates were still in preview as of 2026-09-01.

**Before touching anything, establish how the existing certificates are issued and when they renew.**

### Cold start is the honest UX cost

The Consumption plan's documented cold-start behaviour is that *"apps can scale to zero when idle,
meaning some requests might have more latencies at startup"*, while noting that the plan *"does have
some optimizations to help decrease cold start time, including pulling from prewarmed placeholder
functions that already have the host and language processes running"*. Those placeholders do not help
a **custom handler**, though: the placeholder has the host and a *language* process warm, and our Go
binary is neither. So expect **2–5 s** on an idle site. Consumption scales to zero after "a few
minutes" (no exact figure published; ~20 min in ad-hoc observation, not a measurement campaign).

What Azure does document as flatly unavailable on Consumption is *dedicated compute* to mitigate cold
starts — the "Dedicated compute (mitigate cold starts)" row of the plan comparison table is "None"
for Consumption. Buying the cold start away means changing plan.

Container Apps scales to zero after **exactly 300 seconds** (KEDA cool-down, documented). What
matters there is therefore not the request count but **how many separate 5-minute windows requests
fall into** — the free grant covers ~200 hours of warm replica time per month, i.e. **~80 wake-ups
per day**. At ~120 page views/day across the six sites the average gap between views is ~4 minutes,
the same order as the 300 s window, so this rests entirely on clustering: an eight-hour office day
holds at most 96 such windows against ~80 free. Bursty human traffic lands under it; evenly spread
traffic would not. The €2.60 worst case in the table is that upper bound, and it is still negligible.

## Traps worth remembering

- **ACR Basic is avoidable.** $5.07/month + $0.10/GB would be the largest line item in a Container
  Apps setup. Container Apps pulls from any registry, and `ghcr.io` is free for public images — and
  is already the convention in `fristenkalender-functions` and `ahbicht-functions`.
- **Log Analytics is not mandatory** for a Container Apps environment. `--logs-destination` accepts
  `none`, and real-time log streaming still works. Do not let the portal silently create a workspace.
- **Idle billing does not apply at min replicas 0.** The cheap idle rate requires a minimum replica
  count greater than zero; at zero, every running second bills at the active rate.
- **An unexplained meter.** The Retail Prices API exposes an `Environment Management Hour` meter for
  West Europe repricing from $0.01/hr to **$0.143/hr** with an effective date of 2026-09-01. The
  docs say Consumption environments incur no plan-management charge unless Dedicated profiles,
  private endpoints or planned maintenance are used, and the pricing calculator lists no such fee —
  but no Azure Update confirms the scope. At $0.143/hr it would be **$104/month**, which inverts the
  entire recommendation. **Check the first bill against this meter name** if a Container App is ever
  provisioned.
- **F1 / D1 / Shared cannot host these sites.** Free F1 has no custom domains; Shared/D1 has no
  custom TLS bindings and is Windows-only. **B1 is the hard floor** for App Service with custom
  domains and TLS.
- **Container Apps "express"** has subsecond scale-from-zero at the same price and was, as of
  2026-09-01, in public preview in ~41 regions including West Europe — so availability is not the
  obstacle. It does not support custom domains yet, which for these sites is disqualifying on its own.
  That single gap is the most likely thing in this document to have changed since it was written, and
  the first thing to re-check.

## If a migration does happen

`fristenkalender-functions` is the blueprint and it is already proven in this organisation: `azure.yaml`
(Azure Developer CLI) plus `infra/bicep`, and a `deploy-azure.yml` that logs in with a federated
credential and runs `azd up` on release. Note that repo, despite its name, contains no Azure Functions
at all — it is a plain container on Azure Container Apps. `ahbicht-functions` is the same shape on App
Service for Containers, and is the one still deploying by hand.

The delta for this repo would be the six custom domains on one app plus the host-based routing change
described above.

Three things that are easy to underestimate about such a cutover: certificates have to be issued on
the new host *before* DNS moves (bind the domain and its `asuid` TXT record first, and lower TTLs a
day ahead), the old Function Apps have to stay provisioned until the new certificates have renewed
once so that there is a rollback, and the six `*.azurewebsites.net` hostnames that the README
documents as entry points do not survive the move — they would have to be dropped or redirected.

Also worth weighing: consolidating **flips the blast radius**. Today a bad deploy breaks one site;
with one app it breaks all six. In exchange, the whole class of "one of the six was missed" failures
disappears — per-app setup steps that nobody performed is exactly why `lokations.buendel.id` is still
not live (#268), and the six-way manual publish list in the README is the same hazard for deploys.

## Verification

Every numeric claim above was independently re-checked against primary sources on 2026-09-01 by a
second pass that did not see the first one's work. Thirteen of fifteen load-bearing claims confirmed
verbatim; two were corrected in the process:

- the Consumption cold-start wording (an earlier draft quoted a "None - cold starts are expected in
  this plan" line that does not exist in the current documentation), and
- the Container Apps express region count (an earlier draft said two regions; it is ~41).

The `Environment Management Hour` repricing was confirmed exactly: two records share one meter ID,
the $0.01 one ending 2026-08-31 and the $0.143 one starting 2026-09-01, and the same 0.01 -> 0.143
change also hit `Environment Private Endpoint Hour` and `Environment Planned Maintenance Hour`. So it
is a broad environment-meter repricing rather than an anomaly - and still not documented as applying
to a plain Consumption environment.

## Sources

- [Azure Retail Prices API](https://prices.azure.com/api/retail/prices) (`armRegionName eq 'westeurope'`)
- [Billing in Azure Container Apps](https://learn.microsoft.com/en-us/azure/container-apps/billing)
- [Scaling in Azure Container Apps](https://learn.microsoft.com/en-us/azure/container-apps/scale-app)
- [Custom domains and free managed certificates in Container Apps](https://learn.microsoft.com/en-us/azure/container-apps/custom-domains-managed-certificates)
- [Log storage and monitoring options in Container Apps](https://learn.microsoft.com/en-us/azure/container-apps/log-options)
- [Estimating consumption-based costs in Azure Functions](https://learn.microsoft.com/en-us/azure/azure-functions/functions-consumption-costs)
- [Azure Functions scale and hosting](https://learn.microsoft.com/en-us/azure/azure-functions/functions-scale) (cold-start behaviour, custom domain CNAME-only footnote)
- [Azure Container Apps express overview](https://learn.microsoft.com/en-us/azure/container-apps/express-overview)
- [Storage considerations for Azure Functions](https://learn.microsoft.com/en-us/azure/azure-functions/storage-considerations)
- [Azure Functions Flex Consumption plan](https://learn.microsoft.com/en-us/azure/azure-functions/flex-consumption-plan)
- [Install a TLS/SSL certificate for your app](https://learn.microsoft.com/en-us/azure/app-service/configure-ssl-certificate)
- [Azure App Service plans](https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans)

## Open actions

None of these were resolved while this document was written. Ordered by urgency.

1. **Check `FUNCTIONS_EXTENSION_VERSION` on all five apps before 30 September 2026.** Any app still on
   `~3` stops running. This is the only item here that can take a site offline.
2. Establish how the existing custom-domain certificates were issued and when they renew, before
   touching any domain binding.
3. Get the exact storage figure from Cost Analysis on the `malo-id-generator` resource group, filtered
   to Storage, if anyone ever wants a number instead of "cents".
4. If a Container App is ever provisioned: check the first bill against the `Environment Management
   Hour` meter.

## Revisit this decision when

- **Traffic grows ~30x** (roughly 600 page views/day/site). Consumption compute starts being billed
  and the whole comparison has to be redone.
- **A certificate renewal fails, or a domain has to be re-bound.** The ASMC prerequisites do not list
  Consumption, so renewals are on borrowed time.
- **The Linux Consumption retirement (30 September 2028) comes inside the planning horizon.** A
  migration is mandatory by then regardless of anything else here.
- **Container Apps "express" gains custom-domain support.** That is the single gap that would make
  consolidation strictly better than the status quo on both cost and cold start.
- **A seventh site is added, or the sites need to diverge** beyond the `ID_TYPE_TO_GENERATE` setting.

Nothing else in this document should reopen the question on its own.
