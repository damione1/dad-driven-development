---
title: "dbt: Tests in YAML, or How to Stop Praying Your Data Is Correct"
date: 2026-02-27
draft: false
description: "Declarative tests in dbt - from the four native tests to advanced tests, with concrete examples of problems caught before production."
tags: ["dbt", "Data Quality", "Data Engineering", "Testing"]
categories: ["Data Engineering", "Data Quality"]
---

You know the feeling: a report spitting out weird numbers, an analyst telling you "the totals don't match," and you spend your day tracing back up the chain to find where the data went wrong. Often, the problem could have been detected automatically if someone had put a test somewhere.

## Declarative Tests in dbt

dbt has a testing system built directly into the documentation YAMLs. It's the same file that documents your columns and declares your tests. The idea is simple: you describe your expectations about the data, and dbt verifies them at every execution.

The [four native tests](https://docs.getdbt.com/docs/build/data-tests):
- **not_null**: this column should never be empty
- **unique**: no duplicates on this column
- **accepted_values**: the only possible values are this list
- **relationships**: this column references another table (referential integrity)

It's declarative. You don't write test SQL, you declare constraints.

## Beyond Basic Tests

The four basic tests cover a good portion of needs, but not everything. For the rest, there are test packages and custom tests.

**Combination tests.** "This combination of columns must be unique." For example, an order should only appear once per date and per customer. It's not a simple `unique` on one column, it's a composite constraint. [`dbt-utils`](https://github.com/dbt-labs/dbt-utils) provides `unique_combination_of_columns` for this.

**Distribution tests.** "This column should not have more than X% null values." Useful for columns that *can* be null but shouldn't be null *too often*.

**Freshness tests.** "The most recent data in this source should not be more than 24 hours old." Technically this is a separate mechanism in dbt ([`dbt source freshness`](https://docs.getdbt.com/docs/build/sources#source-data-freshness)), but it's declared in the same place in the YAMLs. If your source stops sending data and nobody notices for a week, you have a problem.

**Consistency tests.** "Subtotal + taxes + shipping should equal the order total." This is the kind of test that catches rounding bugs and calculation inconsistencies before a customer or supplier reports them to you.

## Tests on Sources: The First Line of Defense

A pattern I particularly appreciate: testing source data, not just transformed models.

When your data arrives in Snowflake via a replication tool (like Fivetran, Airbyte), you have no guarantee about its quality. The source system can have bugs. Replication can have issues. Types can change without warning.

By putting tests directly on source definitions in dbt, you create a first line of defense:
- Are the columns you expect still there?
- Are the types correct?
- Are IDs properly unique?
- Is there recent data?

When a source test fails, it tells you "the problem comes from upstream, not from your transformation." That's valuable information for debugging.

In practice, this is often the best way to discover that a colleague on the dev side made a change to one of their services without going through the "notify the data team" step. A renamed column, a new status added, a type that silently changes in production. Without source tests, you discover it when a dashboard is broken. With them, you catch quickly why the dynamic table and the whole lineage are failing. You run your tests, they give you a first lead, and you can go ask the right question to the right team before debugging in the wrong direction.

## CI as a Safety Net

Tests are useless if nobody runs them. The CI pipeline is there for that.

Every PR touching dbt models triggers a complete cycle:
1. Build models in CI environment
2. Execute all tests
3. Validate documentation completeness
4. If everything passes, the PR can be merged

The key point: CI fails if a test fails. No ignored warnings, no "we'll fix it later." If your data doesn't pass the constraints you declared, the code doesn't go to production.

### CI on Snowflake: A Few Settings to Stop Bleeding Credits

If your CI builds a complete ephemeral stack on every PR, a few settings make a real difference on the bill.

**Size warehouses to CI volume.** An XS is enough for most CI builds, no need to oversize. The real parameter to tune is how many parallel runs you can have simultaneously on a busy day.

**Use transient databases for CI.** A [transient database in Snowflake](https://docs.snowflake.com/en/user-guide/tables-temp-transient) doesn't retain Fail-Safe (the data retention in case of corruption or accidental deletion that's enabled by default on standard tables). For CI data that gets recreated on every run anyway, paying for Fail-Safe makes no sense. Declaring the CI target database as transient cuts this cost with no functional impact.

**Clean up properly at the end.** Your CI cleanup step must drop the entire database, not just the tables created during the run. A pipeline that crashes midway without cleanup leaves orphaned objects running, especially Dynamic Tables which keep refreshing and consuming credits until someone manually deletes them. Periodically checking that old CI databases aren't hanging around is good hygiene.

**Override Dynamic Table lag in CI.** By default, a Dynamic Table deployed in CI will try to refresh according to its target lag: every hour, every 5 minutes, whatever is declared for production. In CI, you want the exact opposite: for them to never refresh automatically. The solution is to override `target_lag` to a long value (like `8760 hours`, meaning one year) in your CI profile. The table is created, tests run on the initial content, and no automatic refresh comes to disrupt or extend execution.

**Use `--defer` with a production manifest.** This is probably the most impactful optimization. dbt has a [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) option that, combined with the main branch manifest, lets you only build the models modified in the PR. For unmodified models, dbt "proxies" them to the existing production version instead of recreating them from scratch. A PR that modifies 3 models in a 200-model DAG builds only those 3 models and their direct dependents, not the entire stack. The time and credit savings are considerable on larger projects.

## Tests as Living Documentation

What's elegant about declarative tests in YAMLs is that they also serve as documentation. When you see:

```yaml
- name: status
  description: "Order status"
  tests:
    - not_null
    - accepted_values:
        values: ['PENDING', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED']
```

You immediately know three things:
1. What the column contains (description)
2. That it can't be empty (not_null)
3. What its possible values are (accepted_values)

It's documentation that verifies itself automatically. When a new order status appears in the data, the test fails, the documentation gets updated, and everyone knows about it.

## The Errors That Convinced Me

A few concrete examples of problems caught by tests:

**The phantom boolean type.** A column that should be boolean but contains `NULL` in addition to `true/false`. The source code treats `NULL` as `false`, but your dbt transformation doesn't necessarily do the same. An `accepted_values: [true, false]` test combined with `not_null` clarifies the intent.

**The duplicate ID.** A source system that, following a migration bug, duplicated a few thousand records. Without a `unique` test, these duplicates silently propagate through the entire transformation chain.

## What I Take Away

The approach that worked for me starts with a simple hierarchy.

**First: YAML constraints.** `not_null`, `unique`, `accepted_values`, `relationships`. These tests aren't separable from documentation: they *are* the documentation. Declaring that a `status` column accepts `['PENDING', 'SHIPPED', 'DELIVERED']` is both documenting the contract and verifying it at every run. The cost is near zero (one line of YAML), and it gives basic coverage across all columns with no particular effort. This is the non-negotiable minimum.

**Then and only then: complex tests.** Consistency tests, distribution tests, mapping tests: those requiring custom SQL or packages like `dbt-utils`. These are valuable, but they have a cost: you have to re-read them.

That's where I learned the hard way: delegating test generation to an LLM without serious review means ending up with false coverage. Tests that execute, that pass, and that don't actually test what they claim to test. That's worse than no tests at all, because it gives unearned confidence. I reviewed LLM-generated tests I hadn't checked at merge time, and some were simply off the mark: inverted logic, wrong table referenced, arbitrary threshold with no business meaning. They pass but serve no purpose.

The rule I apply now: YAML constraints, always, systematically, enforced in CI. Complex tests, only when I have time to read them line by line before merging. 30% coverage of well-understood tests is worth more than 80% coverage of tests nobody really knows what they're checking.
