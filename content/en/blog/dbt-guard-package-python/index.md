---
title: "dbt-guard: My First Python Package (and Why I Needed It)"
date: 2026-04-05
draft: false
description: "Building dbt-guard, an offline static analysis tool that detects breaking changes in a dbt DAG by comparing two manifests - no Snowflake connection required."
tags: ["dbt", "Python", "Data Engineering", "Open Source"]
categories: ["Data Engineering", "Open Source"]
---

Publishing a package on PyPI. It's one of those things that looks intimidating from the outside, but turns out to be a matter of good timing and a precise enough problem to solve.

## The Problem

When you work with dbt, you eventually have a DAG (directed acyclic graph) with dozens or hundreds of models depending on each other. The thing is, dbt doesn't tell you when you break something downstream. You rename a column in a silver model, push your PR, the tests pass... and three days later, someone realizes a gold dashboard is broken because it referenced that column.

It's the kind of silent problem. No compilation error. No failing test. Just a downstream consumer left with missing data.

## Doesn't Something Already Exist?

Before diving in, I did what everyone does: searched for whether someone had already solved the problem.

There are a few tools in the ecosystem that come close:

- **[dbt Core 1.5+ with model contracts](https://docs.getdbt.com/docs/mesh/govern/model-contracts)**: if you declare `contract: enforced: true` on a model, dbt detects breaking changes at run time. But that requires explicit configuration model by model. On an existing project, retrofitting everything isn't realistic. An LLM can refactor your project to add the contracts, but I don't love this use of contracts. I see them more like an OpenAPI spec (developer view) than a breaking change validation tool (ops view).
- **[Recce](https://github.com/DataRecce/recce)**: the most complete open source tool for dbt change validation. It uses [SQLGlot](https://github.com/tobymao/sqlglot) to analyze breaking changes, it's serious and active. But its workflow is designed to compare two connected environments (dev and prod) with database access.
- **dbt-manifest-differ**: compares two manifests, but to debug why dbt marked a node as `state:modified`. Not to detect column-level breaking changes.

None of them addressed my exact need: **compare two [manifests](https://docs.getdbt.com/reference/artifacts/manifest-json) offline, without a Snowflake connection, without preconfiguring models**.

The reason this gap exists, I think, is a combination of factors. Teams with DAGs large enough to suffer from this problem are often on dbt Cloud, which has breaking change detection behind its paid tier. And in the data engineering culture, tooling connects to the warehouse by reflex. The idea of static analysis on local JSON files is counter-intuitive in this ecosystem.

There's also an implicit prerequisite: for static analysis on the manifest to work, columns need to be documented in YAML files. Which may not be the case in many dbt projects.

The other motivation is cost. A classic dbt CI pipeline runs a full `dbt build`. That means Snowflake compute, warehouses spinning up, credits burning. Every PR, every push. On a data stack of reasonable size, that adds up fast. And for what? To validate a column rename that could have been detected without ever touching Snowflake.

## The Idea

I wanted something simple: a tool that compares two versions of a dbt manifest (the base branch vs the PR) and tells you "warning, you removed/renamed column X in model Y, and models Z1, Z2, Z3 depend on it."

No database connection required. Zero Snowflake compute. 100% offline. Just static analysis on the [manifest.json](https://docs.getdbt.com/reference/artifacts/manifest-json) files that dbt already generates. You run [`dbt parse`](https://docs.getdbt.com/reference/commands/parse) (which is near-instant and doesn't touch your database), compare the manifests, and you're done.

## How It Works

The principle is fairly direct:

1. You give it two folders: one with the manifest from the main branch, the other with the one from your PR
2. It parses both manifests and extracts columns from each model
3. It compares and identifies changes: deleted columns, renames, type changes
4. For breaking changes, it traverses the DAG in BFS (breadth-first search) to find all impacted downstream models
5. Optionally, it traces column-level lineage to eliminate false positives: if a downstream model doesn't reference the modified column, no alert

The distinction between breaking and non-breaking is simple:
- **Breaking**: deleted column, rename, type change
- **Non-breaking**: new column, new model

```bash
dbt-guard diff --base target/base --current target/current --dialect snowflake
```

It outputs a report in text, JSON, or as GitHub Actions annotations. The last format is handy: alerts appear directly on the code lines in the PR.

## Column-Level Lineage

This is the feature that gave me the most trouble, and also the one that makes the tool genuinely useful.

Without column lineage, if you modify a column in an upstream model, all downstream models get flagged as potentially impacted. With hundreds of models, that generates a lot of noise.

With lineage, dbt-guard traces which downstream columns actually reference the modified column. If a downstream model does `SELECT col_a, col_b` and you modified `col_c`, no alert. It's [SQLGlot](https://github.com/tobymao/sqlglot) doing this by parsing the SQL and building the dependency tree.

Obviously this has its limits. `SELECT *` is the classic case that complicates things. When a model does a `SELECT *`, you can't statically know which columns it actually consumes. And some complex SQL patterns can fool the parser. But for the majority of cases, it works and significantly reduces noise.

## Publishing on PyPI

First Python package published. The process itself isn't that mysterious: a properly configured `pyproject.toml`, some metadata, and `pip install build && python -m build && twine upload dist/*`. Still satisfying to see your package appear on pypi.org and be able to run `pip install dbt-guard`.

I put the standard quality gates in place: pytest for tests, mypy for typing, ruff for linting, 80% coverage threshold. Tests use synthetic fixtures rather than real databases, consistent with the "no connection required" philosophy of the tool.

## What I Learned

A few lessons in no particular order:

**Graceful degradation matters.** SQLGlot can't parse every imaginable SQL pattern. Rather than crashing, dbt-guard falls back to the columns documented in the manifest when SQL parsing fails. Not perfect, but better than a blocking error.

**Keep dependencies minimal.** Every dependency you add is a potential source of version conflicts in someone else's environment. With just SQLGlot and Click, the chances of conflicts are low.

**Synthetic fixtures for tests.** No need for a real Snowflake database to test a static analysis tool. Hand-crafted JSON manifests do the job and tests run in seconds.

## Why Documentation as a Prerequisite Isn't a Problem

One prerequisite for dbt-guard is that your columns are documented in YAML files. For someone coming from software development, this seems obvious: documentation is part of an API's contract, not a nice-to-have. But in data engineering, it's far from the norm.

My view: with LLMs, there's no longer an excuse. Manually documenting hundreds of columns was painful. Today, an agent can read your SQL, infer business context and generate a first draft of YAML documentation in seconds. I go into more detail on this in [the article on dbt documentation as governance]({{< ref "/blog/dbt-documentation-governance-yml/" >}}).

And in CI, enforcement is automatic via [`dbt_meta_testing`](https://github.com/tnightengale/dbt-meta-testing): if a PR adds a model or column without a description, it doesn't pass. Documentation isn't optional, it's enforced the same way tests are.

## A dbt Project Is Testable Infrastructure, Offline

What convinced me to go with the static approach is a simple analogy: a dbt project is a representation of your data stack's infrastructure. The DAG, columns, dependencies, all of this lives in JSON and YAML files. Like Terraform for your data.

And infrastructure can be validated offline. `terraform plan` doesn't touch your cloud. [`dbt parse`](https://docs.getdbt.com/reference/commands/parse) doesn't touch Snowflake. The resulting manifest is a complete description of what the project is supposed to do.

Testing that manifest means testing infrastructure before deploying it. It doesn't replace tests on real data (uniqueness constraints, null values, freshness), but it lets you be strict about pipeline structure without any compute cost. It's an upstream filter, fast and free, before you ever touch the database.

## The Cost Argument

I'll come back to this because it's an underestimated point. A "standard" dbt CI running `dbt build --target ci` involves:
- A Snowflake warehouse waking up
- Models materializing (even in CI mode)
- Tests running on real data
- Credits leaving on every PR

With dbt-guard, breaking change detection costs exactly zero Snowflake credits. It runs on the CI runner itself, in seconds, with local JSON files. It doesn't replace your dbt CI (you still need that to validate logic), but it catches an entire category of problems before ever touching your database. It's a fast, free upstream filter.

## What's Next?

The tool does what it's supposed to do. It's a guardrail that integrates into your CI and blocks (or warns) when a PR risks breaking downstream consumers. No more, no less.

It's my first published Python package. It's not revolutionary, it's not a framework that will change the world. It's a tool that solves a specific problem I had, that nobody else had solved in open source yet, and that might be useful to others.

The code is on [GitHub](https://github.com/damione1/dbt-guard), under the Apache 2.0 license. If you work with dbt and have ever had the pleasure of discovering a breaking change in production on a Friday night, it might be worth checking out.
