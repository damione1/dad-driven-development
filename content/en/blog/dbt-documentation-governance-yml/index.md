---
title: "dbt: When Your YAML Files Become Your Data Governance"
date: 2026-01-16
draft: false
description: "Turning dbt YAML files from an optional formality into a source of truth for governance - PII classifications, persist_docs, automatic masking and CI enforcement."
tags: ["dbt", "Snowflake", "Data Governance", "Data Engineering"]
categories: ["Data Engineering", "Data Governance"]
---

Documentation is the thing nobody wants to do. Especially in data. You have hundreds of columns across dozens of tables, and someone asks "what's the `status` field in the `orders` table?" And the honest answer is often "uh... an enum I think that probably means X."

## The Data Documentation Problem

In a typical dbt project, documentation is optional. You can write your SQL models, deploy them, and never document a single column. dbt doesn't force you to do anything.

The result is predictable: Snowflake schemas with hundreds of columns that nobody knows the exact meaning of. Column names inherited from a 10-year-old source system. Columns called `type` or `status` with no indication of what they mean.

And when a marketing analyst wants to understand the data, they either have to bother someone who has the tribal knowledge, or guess. Both are problematic.

## dbt YAMLs: More Than a Formality

dbt has a built-in documentation mechanism: `schema.yml` files (or whatever you call them). You can describe each model, each column, with free text. Most teams use them little or not at all.

But if you take the time to structure these files properly, they become much more than passive documentation. They become the source of truth for your data governance.

The idea is to use the `meta` field of each column to store structured metadata:

- **Classification**: does this column contain personal (PII), financial, or confidential information?
- **Semantic category**: is it an email, an amount, an address, an identifier?
- **Sensitivity**: high, medium, low?
- **Regulatory obligations**: PIPEDA, GDPR, consumer data, e-commerce?
- **Retention**: how long to keep this data?

When every column has its metadata, you move from passive documentation to active governance.

## persist_docs: From YAML to Snowflake

The thing that makes the difference is [`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs). It's a dbt option that takes your YAML descriptions and pushes them as comments on Snowflake objects. When you enable it:

```yaml
models:
  my_project:
    +persist_docs:
      relation: true
      columns: true
```

Every model and column description in your YAMLs becomes a comment visible in Snowflake. No external tool needed. Someone browsing data in Snowsight sees the descriptions you wrote in dbt directly.

And if you use [Snowflake Horizon](https://docs.snowflake.com/en/user-guide/snowflake-horizon) (their governance platform), these descriptions feed directly into the data catalog. Your dbt documentation IS your Snowflake documentation. A single source of truth.

## Enforcing Documentation in CI

My view on this is pretty firm: documentation is as mandatory as the code itself. Not a nice-to-have, not something we'll do "when we have time." If you deploy untested code to production, that's a problem. Deploying an undocumented model should be just as much of one.

And if you automate everything else (deployments, tests, validations), why would documentation escape this logic? CI is the obvious answer. It's the first check I put there: before even validating the transformation logic, we verify that the documentation exists.

In practice, this means a step that validates completeness:
- Every Silver and Gold model must have a description
- Every column in the SQL must be present in the YAML
- Every documented column must have a non-empty description

If a PR adds a model without documentation, CI fails. Full stop. It's the only way to maintain discipline over the long term. The [dbt-meta-testing](https://github.com/tnightengale/dbt-meta-testing) package does exactly this: it exposes `required_docs` and `required_tests` macros that you wire into your pipeline.

## Governance Tags: From YAML to Masking

This is where it gets powerful: combining YAML metadata with Snowflake's governance features.

You declare in your YAML that a column contains PII. dbt can apply a corresponding Snowflake tag when it creates the model, via a post-hook you write yourself (not built-in, this is custom work). And Snowflake, through masking policies tied to tags, automatically masks the value for users who don't have the right role.

The result: you document your columns in YAML, and data masking happens on its own. No masking logic in SQL, no special views, no maintenance. Governance flows directly from documentation.

## Tests: The Documentation That Verifies Itself

dbt YAMLs aren't just for documentation, they're for tests too. And that's where it loops back: your documentation becomes verifiable.

You document that an `order_id` is a primary key? Add a `unique` and `not_null` test. You document that a status can only have certain values? Add an `accepted_values` test. You document that a column references another table? Add a relationship test.

Tests are declared in the same place as documentation. A single YAML file that says: "this column is called X, it contains Y, it can't be null, and its possible values are Z." Documentation and tests are the same thing. We go into this in detail in [the article on dbt tests]({{< ref "/blog/dbt-tests-constraints-yml/" >}}).

When tests pass, your documentation is proven correct. When a test fails, your documentation or your data is wrong. Either way, you need to investigate.

## Governance Contacts

An often overlooked aspect: who is responsible for what? In each model's metadata, you can declare an owner, a steward, an approver. These metadata are pushed to Snowflake via a custom post-hook (again, this is handcrafted, not native dbt).

When someone finds a data problem, they know exactly who to contact. No need to search a wiki or ask around. The information is attached directly to the data itself.

## The Real Cost: Less Than You Think

The classic objection: "it takes time to document every column." True. But compare it to the alternative:
- Hours lost by analysts guessing what columns mean
- Errors in reports because someone misinterpreted a field
- Audits that take weeks because nobody knows which data is sensitive
- Security incidents because a PII column wasn't identified as such

Documenting a column takes 30 seconds. Not documenting it can cost hours, days, or worse. And with LLMs, the startup cost has dropped further: [documenting an entire source database]({{< ref "/blog/dbt-document-sources-llm-multi-agent/" >}}) in a few days is no longer a pipe dream.

## The Bonus: Documentation as the Key to Talking to Your Data

Documentation serves two audiences: your human colleagues, and LLMs. And that's where it gets interesting.

[Snowflake Cortex Analyst](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst) is Snowflake's answer to "talk to my data": ask a question in natural language and get back the correct SQL query. Ask for a specific KPI, filter by criteria, compare with another metric, without writing a line of SQL. It sounds like magic. And people instinctively think it takes months of engineering work to get there.

It doesn't, if the documentation is clean.

Cortex Analyst works from a semantic model: a YAML file describing tables, columns, metrics, relationships between entities, and associated business vocabulary. The structure of this file is very close to what dbt already produces in its documentation files. The infrastructure exists. The models exist. If column descriptions are there, if relationships are documented, if key metrics are defined, the gap to Cortex Analyst is small.

Snowflake Labs has even published [`dbt_semantic_view`](https://github.com/Snowflake-Labs/dbt_semantic_view), a package that generates Snowflake Semantic Views directly from the dbt semantic model. These Semantic Views are natively usable by Cortex Analyst. The pipeline becomes: dbt documents, dbt_semantic_view publishes the semantic views, Cortex Analyst answers questions in natural language.

A well-documented dbt project is 2 to 4 weeks away from a working talk-to-your-data experience, not 6 months. The remaining work (formalizing a few metrics, distinguishing dimensions from measures, adding business synonyms) is marginal compared to what's already in place. It's a quick win that only unlocks under one condition: treating documentation as a constraint from the start, not a task to do later.

## What I Take Away

In the end, what worked for me is treating data documentation with the same seriousness as transformation code. Not as a nice-to-have we'll do "when we have time." As an integral part of the pipeline, verified in CI, propagated automatically.

dbt YAMLs are the ideal place for this. A single file serving human documentation, automated tests, and data governance. That sounds fancy when you say it that way, but in practice it's just YAML files doing their job.
