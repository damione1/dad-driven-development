---
title: "dbt and Terraform: My Snowflake Infrastructure as Code"
date: 2025-12-26
lastmod: 2026-10-28
draft: false
description: "How I split Snowflake between Terraform and dbt: models, Dynamic Tables, isolated targets, CI selection and the limits of manifest.json."
tags: ["dbt", "Snowflake", "Data Engineering", "SQL"]
categories: ["Data Engineering"]
images: ["dbt-data-infrastructure-as-code-featured.jpg"]
---

I use Terraform to provision my Snowflake environment and dbt to describe the transformations inside it. Databases, roles and warehouses on one side; models, tests and documentation on the other. Both are code, reviewed in git and deployed through a pipeline.

Snowflake is fundamentally SQL-first. That's its strength: grants, objects, transformations, all accessible through SQL. In [the Terraform article]({{< ref "/blog/snowflake-terraform-infrastructure-as-code/" >}}), I explained how I stopped managing the infrastructure with a pile of SQL scripts. The transformations deserved the same treatment.

Otherwise, you get SQL files scattered everywhere, a README explaining the order and one colleague who knows which script really has to run first. I don't want the deployment plan living in that colleague's head.

## dbt vs Terraform: the boundary I choose

The two tools complement each other in this stack. Here's where I put the responsibility:

| Concern | What I use |
|---|---|
| Databases, warehouses, roles and their privileges | Terraform |
| Transformation SQL, model dependencies and materializations | dbt |
| Data tests, descriptions and project lineage | dbt |
| Continuous refresh of deployed Dynamic Tables | Snowflake's scheduler, when configured |
| Reviewing and deploying changes | CI around those tools |

That's an ownership choice, not a claim that their capabilities never overlap. dbt can manage some grants; Terraform can manage database objects. The important rule is that the same object doesn't have two tools both trying to be its authority.

## From SQL scripts to dbt models and a DAG

Before that structure, DDL accumulates. `create_orders_table.sql`, then `add_status_column.sql`, then `alter_orders_add_shipping_address.sql`. After a while, reconstructing what the table is supposed to be starts to look like archaeology.

[dbt models](https://docs.getdbt.com/docs/build/models) give the transformation a home. A SQL model describes a result; its materialization determines whether dbt builds a view, a table or another supported object. References to other models establish the dependencies.

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

[`ref`](https://docs.getdbt.com/reference/dbt-jinja-functions/ref) resolves the relation name and gives dbt the dependency. It knows that this model depends on those three staging models, and can order selected builds accordingly.

It doesn't propagate every new column through the SQL for us. If I add `shipping_country` upstream, this explicit `SELECT` still needs to select it downstream. What disappears is the handwritten ordering of the scripts, not the need to decide what each model contains.

The tests and descriptions live next to this code. I cover those separately in [the YAML testing article]({{< ref "/blog/dbt-tests-constraints-yml/" >}}) and [the documentation and governance article]({{< ref "/blog/dbt-documentation-governance-yml/" >}}).

## Where the Terraform analogy holds

The parallel is useful: describe the intended result, version it, review the change and let a tool carry it through the dependency graph.

But neither command means “destroy everything and rebuild it.” [`dbt build`](https://docs.getdbt.com/reference/commands/build) runs the selected resources and their tests in DAG order, according to their materializations. An incremental model can update incrementally. A Dynamic Table has its own deployment and refresh behavior. Selection and `--full-refresh` change what happens.

Similarly, Terraform plans changes against the infrastructure it manages, rather than recreating the entire stack on every `apply`.

The bigger distinction is [Terraform's state](https://developer.hashicorp.com/terraform/language/state/purpose): it records the association between configured resources and real objects. A normal [`terraform plan`](https://developer.hashicorp.com/terraform/cli/commands/plan) refreshes information about those objects and proposes changes.

dbt's project artifacts don't provide that same general infrastructure reconciliation. They are useful for describing the project and comparing versions. They aren't proof that nobody changed a table by hand yesterday.

I still think of dbt as infrastructure as code. I just don't ask the analogy to do work the tool doesn't do.

## Dynamic Tables: declare the model, choose who refreshes it

[Snowflake Dynamic Tables](https://docs.snowflake.com/en/user-guide/dynamic-tables-intro) are where this way of thinking becomes particularly useful. We declare a query, and Snowflake can maintain its result through refreshes instead of us scheduling every transformation as a separate batch.

The [dbt Snowflake adapter](https://docs.getdbt.com/reference/resource-configs/snowflake-configs) supports a `dynamic_table` materialization. This example uses the adapter's dbt 1.12+ scheduling configuration and deliberately selects Snowflake-managed refresh:

```yaml
# dbt_project.yml
models:
  my_project:
    silver:
      +materialized: dynamic_table
      +snowflake_warehouse: TRANSFORMING_L
      +scheduler: ENABLE
      +target_lag: '1 hour'
```

The model's SQL still defines the data. The configuration defines how that result is maintained. A [target lag](https://docs.snowflake.com/en/user-guide/dynamic-tables/target-lag) is a freshness target relative to the underlying data, not “run exactly once an hour” or a guarantee that the target will always be met.

Scheduling and refresh mode are different decisions. Snowflake can use full or incremental refresh where the query and configuration support it. The adapter version matters too: current configurations distinguish Snowflake-managed scheduling from dbt-managed refresh. I would check the installed adapter before copying a configuration from a different release.

For this part of the DAG, Snowflake can take over the recurring refresh work. We still need to deploy changes, run the appropriate tests, check source freshness and monitor failures. It doesn't make the rest of the pipeline disappear. I go into the CI side in [the article on dbt tests]({{< ref "/blog/dbt-tests-constraints-yml/" >}}).

## Targets: one environment per context

A [dbt target](https://docs.getdbt.com/docs/local/profiles.yml) chooses connection settings for a run. I want a developer's default to point at an isolated development database, with a warehouse sized for that work.

Here's the environment part of a `profiles.yml`. This is an excerpt: account, user and authentication settings still have to be supplied using the Snowflake adapter's supported configuration.

```yaml
my_project:
  target: dev
  outputs:
    dev:
      type: snowflake
      database: DEV_JEAN
      schema: silver
      warehouse: DEV_XS
      role: DBT_DEV
    prod:
      type: snowflake
      database: PROD
      schema: silver
      warehouse: TRANSFORMING_L
      role: DBT_PROD
```

`dbt run` uses the default target; `dbt run --target prod` explicitly selects production. That's a useful guardrail. The protection against writing to production comes from Snowflake privileges and the credentials used by the job, not from naming a target `dev`.

The project can also override database and schema settings. I want those overrides reviewed with the same care as the SQL. An isolated development target is only useful if the models actually stay inside the boundary we've given it.

## Select the part of the DAG that needs work

Once a project has hundreds of models, rebuilding all of them for each change is rarely what I want. The [graph operators](https://docs.getdbt.com/reference/node-selection/graph-operators) make the scope explicit:

```bash
# Just the model
dbt run --select orders_enriched

# The model and its upstream dependencies
dbt run --select +orders_enriched

# The model and all downstream descendants
dbt run --select orders_enriched+

# Both directions
dbt run --select +orders_enriched+
```

Tags give another entry point. A model with `{{ config(tags=['retail', 'orders']) }}` can be selected with `dbt run --select tag:retail`. We still have to choose whether to include its dependencies.

In CI, a previous production manifest lets us select changed resources. With a separately configured, clean `ci` target, a command can look like this:

```bash
dbt build --select "state:modified+" --state artifacts/prod --defer --target ci
```

`state:modified+` selects modified resources and all downstream descendants, not just direct children. [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) is the separate mechanism that can resolve unselected upstream references to existing production relations. Existing relations in the current target can affect that resolution, so keeping CI isolated matters. The saved production artifacts live outside the current run's `target/`, where they won't be overwritten.

This is the part I like: the scope is written down. We can inspect it, review it and change it. “Only run what changed” stops being an assumption buried in the scheduler.

## Sources and exposures: both ends of the lineage

A [source declaration](https://docs.getdbt.com/docs/build/sources) gives a name to data that already arrives through another system. It doesn't load that data for us.

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

Models can reference `{{ source('shopify', 'orders') }}`. We can declare tests and configure freshness checks for those sources, then actually run `dbt source freshness` in the pipeline. A source declaration alone isn't an alerting service.

The same boundary applies to documentation. Descriptions in dbt help its own documentation; pushing comments onto source tables needs a separate implementation. [`persist_docs`]({{< ref "/blog/dbt-documentation-governance-yml/" >}}) handles supported model descriptions, not source comments or custom governance tags.

[Exposures](https://docs.getdbt.com/docs/build/exposures) describe the consumers at the other end: a dashboard, an API, an export.

```yaml
exposures:
  - name: tableau_revenue_dashboard
    type: dashboard
    owner:
      name: BI Team
      email: analytics@example.com
    depends_on:
      - ref('daily_revenue')
      - ref('top_products')
```

Now the project records that the dashboard depends on those models. That makes an impact review possible. Sending a notification when they change is something the workflow or an integration still has to do. The YAML doesn't subscribe the BI team to emails by itself.

## Seeds, functions and macros

[Seeds](https://docs.getdbt.com/docs/build/seeds) are CSV files dbt can load as tables: small reference lists, mappings, static configuration. I prefer seeing those changes in the repo over wondering which copy of a spreadsheet the job picked up.

[User-defined functions](https://docs.getdbt.com/docs/build/udfs) also have native support from dbt 1.11 onward, with support depending on the adapter and function language. They no longer belong in a blanket “dbt can't manage this” category. Older projects may still deploy functions through hooks or macros; the version we run determines the options.

The [Jinja macro system](https://docs.getdbt.com/docs/build/jinja-macros) and packages such as [dbt-utils](https://github.com/dbt-labs/dbt-utils) are another reason the tool suits me. I can reuse a transformation or a test, and add project-specific checks when the built-in behavior doesn't cover the requirement. That's a useful extension point, not a promise to reproduce every hosted-platform feature with a few lines of Jinja.

## The manifest: project state, not warehouse state

[`dbt parse`](https://docs.getdbt.com/reference/commands/parse) parses and validates the project, and writes `manifest.json`. It doesn't connect to the warehouse or compile the model SQL.

The [manifest](https://docs.getdbt.com/reference/artifacts/manifest-json) describes resources, configurations, declared columns and dependencies. Commands that execute nodes can add compiled SQL for those nodes. It is an artifact of what dbt understands about the project, not a live inventory of everything in Snowflake.

That's still extremely useful. We can compare versions, select modified resources and inspect contracts. It's what I use in [dbt-guard]({{< ref "/blog/dbt-guard-package-python/" >}}) to look for breaking changes before deployment. But a manifest comparison and an infrastructure drift check answer different questions.

## Why I chose dbt rather than SQLMesh

[SQLMesh](https://sqlmesh.readthedocs.io/en/stable/) was on my list when I evaluated this approach. Its [plans and change categories](https://sqlmesh.readthedocs.io/en/stable/concepts/plans/) make changes and backfills explicit. It also has a [dbt compatibility mode](https://sqlmesh.readthedocs.io/en/stable/integrations/dbt/), with limitations that need checking against the actual project.

I chose dbt for the ecosystem I wanted to work in: packages on [dbt Hub](https://hub.getdbt.com/), the existing integrations in my stack and the familiarity of the workflow. That was a practical choice, not a benchmark proving one tool wins every category.

There are also several places to run dbt. [dbt Projects on Snowflake](https://docs.snowflake.com/en/user-guide/data-engineering/dbt-projects-on-snowflake) provides managed runtimes inside Snowflake, with its own development and scheduling options. That's a specific execution product. Using the dbt Snowflake adapter from a laptop, CI runner or hosted dbt service doesn't by itself mean the dbt runtime is inside Snowflake.

SQLMesh could also have been a reasonable choice. Choosing infrastructure includes choosing who will understand it, how it will be deployed and which parts we'll have to maintain ourselves.

## What this changes for the team

The part I carry over from software development is simple: someone else should be able to open the repo and understand how the system is meant to work.

Terraform declares the environment. dbt declares the transformations and their checks. Snowflake executes the SQL and, where we've chosen Dynamic Tables, manages their refreshes. Each layer has a responsibility we can point to, and changes pass through review instead of a manual deployment on Friday night.

The knowledge doesn't become complete just because it's in code. But the dependency graph, the configuration and the tests give the next person something much better than “ask the colleague who knows.”

The rigor we apply to application code has every reason to apply to data too. That's what I want from this stack.
