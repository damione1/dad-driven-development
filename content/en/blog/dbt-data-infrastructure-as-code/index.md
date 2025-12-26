---
title: "dbt: Treating Your Data Transformations Like Infrastructure"
date: 2025-12-26
draft: false
description: "dbt applies the same rigor to your data pipelines that Terraform applies to infrastructure - declarative approach, DAG, Dynamic Tables, manifest and comparison with SQLMesh."
tags: ["dbt", "Snowflake", "Data Engineering", "SQL"]
categories: ["Data Engineering"]
---

Snowflake is fundamentally SQL-first. That's its strength: everything is driven by SQL, from grants to object creation to transformations. Infrastructure, we've seen how to tame it with Terraform in [the previous article]({{< ref "/blog/snowflake-terraform-infrastructure-as-code/" >}}). But data transformations fall into a blind spot. SQL scripts scattered everywhere, no tests, no serious versioning, one colleague who knows what order to run things in.

That's where dbt comes in. And it's no coincidence that both tools work so naturally together: Snowflake abstracts physical infrastructure, dbt abstracts transformation orchestration. Both are declarative, SQL-first, designed so that code is the source of truth. dbt isn't a generic data transformation framework, it's the tool that completes Snowflake where Snowflake doesn't complete itself: the organization, tests, documentation and reproducible deployment of all that SQL.

## The Problem with SQL Transformation Scripts

Before dbt, the classic data pipeline looked like this: numbered SQL scripts, some scheduler, and a README explaining what order to run things. Or worse, stored procedures in Snowflake with business logic buried inside.

The problem is the same as with infrastructure SQL scripts: DDLs accumulate. You have `create_orders_table.sql`, then `add_status_column.sql`, then `alter_orders_add_shipping_address.sql`. After a while, nobody knows what the table is *supposed* to be. You have to run all the scripts in order to reconstruct the real state, and you hope none of them are missing.

And when you want to add a column to a 5-stage pipeline (staging, enrichment, aggregation, export, report), you have to open 5 files, add the column in each one, then coordinate the deployment in the right order in production. Without a framework, you manage this by hand. And you almost always forget a model somewhere.

As for tests: writing a non-null test on a column without a framework means writing a `SELECT COUNT(*) WHERE col IS NULL` and checking the result is 0. Doable once. Painful to maintain across 200 columns.

## What dbt Changes Fundamentally

[dbt](https://www.getdbt.com/) (data build tool) takes a simple idea: data transformations are code. And code is managed with the same practices as everything else: version control, tests, documentation, CI/CD.

In practice, you write models: `.sql` files that each define a table or view. A model references other models with `{{ ref('model_name') }}` syntax. dbt automatically resolves dependencies and builds a [DAG](https://docs.getdbt.com/terms/dag) (directed acyclic graph) of your transformations.

```sql
-- models/silver/orders_enriched.sql
select
    o.order_id,
    o.created_at,
    o.status,
    c.email as customer_email,
    p.name as product_name
from {{ ref('stg_orders') }} o
left join {{ ref('stg_customers') }} c on o.customer_id = c.customer_id
left join {{ ref('stg_products') }} p on o.product_id = p.product_id
```

dbt knows that `orders_enriched` depends on `stg_orders`, `stg_customers` and `stg_products`. It builds them in the right order, automatically. When you add a column to `stg_orders`, it's available in all downstream models without you having to coordinate anything.

## The Terraform Analogy

The parallel with Terraform isn't superficial. Both tools share the same fundamental philosophy:

**Desired state, not steps.** You don't say "run this transformation after that one." You describe what each table should contain, and dbt figures out how to get there.

**Declarative.** Your code describes the result, not the process. `SELECT ... FROM ref('source')` says "this table contains these columns calculated from this source," not "take this table, join it with that, filter this."

**Reproducible.** `dbt build` rebuilds everything from scratch, in the right order, every time. Like `terraform apply` rebuilds your infrastructure from HCL files.

**Versioned.** Every change to a transformation is a commit. You can see who changed what, why, and roll back.

The analogy has one important limit though: dbt has no remote state. If someone deletes a table in Snowflake manually, dbt doesn't know about it. The next `dbt run` will simply recreate it without flagging the drift. Terraform, on the other hand, would detect the gap between the real state and the desired state and propose correcting it. This isn't a critical flaw, but worth keeping in mind: dbt is declarative about what it creates, not about what exists.

## Extensibility: What More Than Compensates

That lack of remote state is one of the few areas where Terraform does better. But dbt compensates with something different: remarkable extensibility.

The [dbt Hub](https://hub.getdbt.com/) aggregates hundreds of community packages. Additional tests, utility macros, integrations with specific sources. [dbt-utils](https://github.com/dbt-labs/dbt-utils) is probably in every serious dbt project: it adds dozens of macros and tests you wouldn't want to write yourself.

But the real power is the [Jinja macro system](https://docs.getdbt.com/docs/build/jinja-macros). Everything dbt does internally, you can do too. Which means features reserved for dbt Cloud are often reproducible in dbt Core with a few macros.

Breaking change detection in schemas? We cover that in [a later article]({{< ref "/blog/dbt-guard-package-python/" >}}), and it's exactly this kind of feature you can implement yourself. Alerting on source freshness? Macros. Automatic documentation generation? Macros. dbt Core isn't a stripped-down version of dbt Cloud, it's a foundation on which you build what you need.

## dbt as Infrastructure, Not Just an ETL Runner

Many teams use dbt as an ETL runner: they materialize tables and run `dbt run` every hour in Airflow or Prefect. That's a valid use case. But it's far from exhausting what dbt can do.

Snowflake has a concept that changes the equation: [Dynamic Tables](https://docs.snowflake.com/en/user-guide/dynamic-tables-intro). Instead of materializing a table and refreshing it manually, you declare a Dynamic Table with a target lag ("this table must be no more than 1 hour stale"). Snowflake manages the refresh automatically, propagating changes through the DAG.

Combined with dbt, this means you no longer need an external orchestrator to manage refreshes. You declare your models as Dynamic Tables in dbt, deploy, and Snowflake handles the rest. dbt becomes what it truly is: a tool for declaring data infrastructure, not a scheduler.

```yaml
# dbt_project.yml
models:
  my_project:
    silver:
      +materialized: dynamic_table
      +target_lag: '1 hour'
      +snowflake_warehouse: TRANSFORMING_L
```

## Targets: One Environment per Context

One of dbt's most practical concepts is the [targets system](https://docs.getdbt.com/docs/core/connect-data-platform/connection-profiles). A target is a deployment configuration: which database, which schema, which warehouse to use.

In practice, each developer has a default `dev` target pointing to their own isolated database, with a small warehouse to avoid burning credits unnecessarily:

```yaml
# profiles.yml
my_project:
  target: dev
  outputs:
    dev:
      database: DEV_JEAN
      schema: silver
      warehouse: DEV_XS
    prod:
      database: PROD
      schema: silver
      warehouse: TRANSFORMING_L
```

When Jean runs `dbt run`, he deploys to `DEV_JEAN.silver`. He can't accidentally overwrite production. Deploying to prod requires being explicit: `dbt run --target prod`. It's a deliberate opt-in, not the default behavior.

What's elegant is that dbt lets you [override the database, schema, and even the warehouse](https://docs.getdbt.com/docs/build/custom-schemas) at multiple levels: project-wide, folder-level, individual model. You can have a specific model that always runs on a dedicated warehouse, regardless of the target. Configuration composes.

## Precise Selection in the DAG

Once you have a DAG with 200 models, you don't want to systematically rebuild everything. dbt has an expressive [selection system](https://docs.getdbt.com/reference/node-selection/syntax) for targeting exactly what you need.

The `+` syntax controls direction in the DAG:

```bash
# Just the model
dbt run --select orders_enriched

# The model and everything upstream (its dependencies)
dbt run --select +orders_enriched

# The model and everything downstream (what depends on it)
dbt run --select orders_enriched+

# The model, its dependencies AND its dependents
dbt run --select +orders_enriched+
```

You can also select by tag. If you tag your retail models:

```yaml
-- models/silver/orders_enriched.sql
{{ config(tags=['retail', 'orders']) }}
```

You can then target the entire retail domain in one command:

```bash
dbt run --select tag:retail
```

Combined with `state:modified`, this gives very precise control over what runs in CI: only the modified models and their direct dependents.

## Sources and Exposures: The Two Ends of the DAG

dbt manages not only intermediate transformations but also both ends of the pipeline.

**Sources** are tables you don't control: your raw data arriving via Fivetran, Airbyte, or another replication tool. You [declare them in YAML](https://docs.getdbt.com/docs/build/sources):

```yaml
sources:
  - name: shopify
    database: RAW
    schema: shopify
    tables:
      - name: orders
      - name: customers
      - name: products
```

This lets you reference them with `{{ source('shopify', 'orders') }}` in your models, associate tests with them, and monitor their freshness. If Shopify data hasn't been updated in 6 hours, dbt can alert you before your reports go stale.

These source descriptions also integrate with Snowflake Horizon: the metadata you declare in dbt surfaces in the [Snowflake data catalog](https://docs.snowflake.com/en/user-guide/snowflake-horizon), visible to everyone with access to the instance.

**Exposures** are the other end: consumers of your data that dbt doesn't manage. A Tableau dashboard, an API, a file exported to a partner. You [declare them too](https://docs.getdbt.com/docs/build/exposures):

```yaml
exposures:
  - name: tableau_revenue_dashboard
    type: dashboard
    owner:
      name: BI Team
    depends_on:
      - ref('daily_revenue')
      - ref('top_products')
```

This completes the lineage: you can see in the DAG not only how data transforms, but also where it ends up. When you modify `daily_revenue`, dbt knows the Tableau dashboard depends on it and can alert you.

## Complementary Assets: Seeds and UDFs

Two elements worth mentioning to complete the picture.

**[Seeds](https://docs.getdbt.com/docs/build/seeds)** are CSV files that dbt manages as tables. Reference tables, mappings, static configurations. Instead of having a CSV file somewhere on a server or in a Google Sheet, you version it in the dbt repo and dbt creates the corresponding table in Snowflake.

```
seeds/
  product_categories.csv
  country_codes.csv
  shipping_carriers.csv
```

**Snowflake UDFs** (User Defined Functions) aren't natively managed by dbt, but you can deploy them via [pre-hooks or macros](https://docs.getdbt.com/docs/build/jinja-macros). This is the limit of the infrastructure-as-code metaphor: certain Snowflake objects remain in Terraform territory.

## The Manifest: The Compiled Plan of Your Stack

When you run [`dbt parse`](https://docs.getdbt.com/reference/commands/parse), dbt compiles your entire project and produces a [`manifest.json`](https://docs.getdbt.com/reference/artifacts/manifest-json). This is the central artifact: a complete, machine-readable representation of everything your project knows how to do.

The manifest contains models, their declared columns, their dependencies, tests, sources, exposures. Everything.

It's the equivalent of the Terraform state for your data. And like the Terraform state, it can be used for comparisons: what changed between the previous version and the current version? That's exactly what we'll leverage in [a later article]({{< ref "/blog/dbt-guard-package-python/" >}}) to detect breaking changes before they reach production.

## Why This Changes Team Dynamics

Before dbt, transformations were often owned by one person. Someone knew what order to run the scripts, which tables depended on what, where the deduplication logic lived. The knowledge was tribal.

With dbt, any data engineer can open the repo, see the DAG, understand the dependencies, modify a model, and deploy to their dev environment without risk. The knowledge is in the code.

It's the same gain we had with Terraform for infrastructure. We go from "ask the person who knows" to "read the code."

## Why dbt and Not SQLMesh?

When I started evaluating tools to manage my transformations declaratively, [SQLMesh](https://sqlmesh.com/) was on my list. It's a serious tool: open source, Apache 2.0, backwards-compatible with existing dbt projects, with interesting concepts like automatic breaking change detection and native incremental evaluation.

On paper, SQLMesh is technically more rigorous than dbt on several points. But I chose dbt for one simple reason: network effects.

dbt is by far the most widely used tool in the data ecosystem. Which means:
- Hundreds of packages on the [dbt Hub](https://hub.getdbt.com/): tests, macros, utilities already written
- An active community with answers to almost every problem you'll encounter
- Native integrations in all stack tools: Fivetran, Hightouch, Metabase, Tableau, and essentially all data catalogs
- Data engineers who already know dbt when they join a team

And since 2024, Snowflake has formalized this partnership with an even deeper integration: the [dbt Snowflake Native App](https://docs.getdbt.com/docs/cloud-integrations/snowflake-native-app). You can now orchestrate and monitor your dbt jobs directly from Snowflake, without external infrastructure. dbt runs inside your Snowflake instance, not alongside it.

SQLMesh would probably have done the job just as well. But when you choose infrastructure, you also choose an ecosystem. And the dbt ecosystem is unbeatable right now.

## dbt + Snowflake: The Duo for a Dev-Minded Data Engineer

What makes this combination particularly solid is that both tools share the same philosophy and complement each other without overlapping.

Terraform declares infrastructure: databases, schemas, roles, permissions, masking policies. dbt declares transformations: models, tests, documentation, lineage. Both are code, versioned in git, deployed via CI pipelines, with environments isolated per developer.

But beyond the shared philosophy, the integration is concrete. Snowflake Dynamic Tables materialize natively in dbt. Column descriptions in your YAML surface in Snowflake Horizon via `persist_docs`. And since 2024, dbt can run directly in your Snowflake instance via the [dbt Snowflake Native App](https://docs.getdbt.com/docs/cloud-integrations/snowflake-native-app), without external infrastructure to manage.

The complete stack looks like this: Terraform manages what's above the data, dbt manages what's inside it, and Snowflake is the foundation both rely on. Each layer has its responsibility, each layer is code. For someone coming from software development, this is exactly what a data stack should look like. Not SQL scripts in a shared folder, not tribal knowledge, not manual deployment on a Friday night.

The rigor we apply to application code has every reason to apply to data too. dbt is the tool that makes this possible, and Snowflake is the platform where that rigor makes complete sense.
