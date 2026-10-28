---
title: "dbt Data Tests in YAML: Stop Praying Your Data Is Correct"
translationKey: "dbt-tests-constraints"
date: 2026-02-27
lastmod: 2026-10-28
draft: false
description: "How I use dbt data tests in YAML, catch nulls and duplicate keys, and run checks in CI without wasting Snowflake credits."
tags: ["dbt", "Data Quality", "Data Engineering", "Testing"]
categories: ["Data Engineering", "Data Quality"]
images: ["dbt-tests-constraints-yml-featured.jpg"]
---

A report spits out weird numbers. An analyst tells you "the totals don't match," and you spend your day tracing back up the chain. The habit I keep coming back to is boring: **dbt data tests, declared in YAML beside the columns, and actually run in CI**. Tell the pipeline what you expect before a dashboard tells you what went wrong.

## What dbt Data Tests Actually Check

[Data tests in dbt](https://docs.getdbt.com/docs/build/data-tests) are SQL checks against data. Generic tests let you declare those checks in the same YAML files that document your models and sources. dbt generates the test SQL; you run it with `dbt test`, or build and test selected resources together with `dbt build`. `dbt run` alone doesn't execute the tests.

The four built-in generic tests are a useful starting point:

- **`not_null`**: no SQL `NULL` in this column. An empty string is a different value.
- **`unique`**: no duplicate non-null values. Pair it with `not_null` for a required identifier.
- **`accepted_values`**: non-null values belong to the list you supplied.
- **`relationships`**: non-null values exist in the referenced relation and column.

The [null-handling details](https://docs.getdbt.com/reference/resource-properties/data-tests) matter. `accepted_values` and `relationships` don't reject `NULL`; add `not_null` when absence is invalid.

These are assertions about the data you query. They aren't database constraints that prevent a bad row from being inserted. A passing test means that assertion held for the data it checked, at that moment. It doesn't prove the whole model is correct.

## A YAML Example I Can Read at Merge Time

For an `orders` model, I want the identifier present and unique, and the status both present and drawn from a known set:

```yaml
# models/orders.yml
version: 2

models:
  - name: orders
    description: "One row per order"
    columns:
      - name: order_id
        description: "Order identifier"
        data_tests:
          - not_null
          - unique
      - name: status
        description: "Order status"
        data_tests:
          - not_null
          - accepted_values:
              arguments:
                values: ['PENDING', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED']
```

The `arguments:` syntax is available from dbt 1.10.5 onward; older versions place test arguments directly under the test name. `tests:` remains an alias for `data_tests:`, but I use the explicit name here.

To check the existing model, or rebuild and check it:

```bash
dbt test --select orders
dbt build --select orders
```

That's also why I like keeping descriptions and tests together. You can read the intended contract without hunting through a separate test suite. I go further into that idea in [the article on dbt documentation and governance]({{< ref "/blog/dbt-documentation-governance-yml/" >}}).

When a new order status arrives, the test gives us something concrete to discuss. Was the source changed legitimately? Did we forget a case? Is the data wrong? Updating the YAML is the result of that conversation, not the automatic cure for a failing test.

## When Uniqueness Spans Several Columns

A daily order snapshot can repeat an `order_id` across dates. What must stay unique is the combination of `order_id` and `snapshot_date`. Testing either column alone would describe the wrong grain.

[`dbt_utils.unique_combination_of_columns`](https://github.com/dbt-labs/dbt-utils#unique_combination_of_columns-source) handles that, once a compatible `dbt-utils` package is installed. This example follows its documented argument format for dbt Core 1.10.6 and newer:

```yaml
models:
  - name: order_daily_snapshots
    data_tests:
      - dbt_utils.unique_combination_of_columns:
          arguments:
            combination_of_columns:
              - order_id
              - snapshot_date
    columns:
      - name: order_id
        data_tests:
          - not_null
      - name: snapshot_date
        data_tests:
          - not_null
```

The combination test checks duplicates; it doesn't make each component mandatory. I keep `not_null` on both because a snapshot without its order or date doesn't meet this model's contract. These are data checks, not a declaration of a composite primary key in Snowflake.

Beyond that, there are rules the four built-ins can't express. "This optional column shouldn't be mostly null." "Subtotal plus tax plus shipping should equal the order total." Packages and custom SQL tests can cover those. The hard part is choosing a rule that means something, then reading the SQL that claims to enforce it.

## Source Tests: A First Lead, Not a Diagnosis

A pattern I particularly appreciate is testing source data, not just transformed models.

When data arrives in Snowflake via Fivetran, Airbyte or another replication tool, I don't assume it is clean. A `unique` test on a source identifier or an `accepted_values` test on a status gives me an earlier place to look when something breaks downstream.

That's often how I discover a change made on the dev side without the "notify the data team" step. A renamed column, a new status, a type that changed in production. A source test can fail or error before I spend an afternoon debugging the wrong transformation. It gives me a lead; it doesn't prove which service caused the problem.

The four generic tests don't check a source's entire schema or assert every column's data type. For those guarantees, I need explicit schema checks. Merely listing columns in YAML isn't enough.

Freshness is another separate question. dbt's [freshness configuration](https://docs.getdbt.com/reference/resource-configs/freshness) compares a configured timestamp or supported warehouse metadata with your thresholds. `dbt source freshness` checks sources; dbt v2 also provides `dbt freshness` for configured sources and models. `dbt build` doesn't include these checks. A recent ingestion timestamp is useful, but doesn't by itself prove the underlying business events are current.

## CI Is Where the Tests Become a Gate

Tests are useless if nobody runs them. My rule is that the CI job must do the checking, and its result must decide whether the PR can merge.

The cycle I want is straightforward:

1. Check source freshness where the change depends on it.
2. Build and test the selected resources in an isolated CI target.
3. Validate documentation completeness.
4. Block the merge if a required check fails.

[`dbt build`](https://docs.getdbt.com/reference/commands/build) handles models and their tests in dependency order. For the data tests that must block a merge, I keep error severity and review any [thresholds or warning configuration](https://docs.getdbt.com/reference/resource-configs/severity). Turning an inconvenient failure into a warning isn't fixing the data.

### Select Changes, Then Defer Unbuilt Dependencies

For a larger DAG, rebuilding everything on every PR gets expensive. Here's an illustrative CI command using a saved production manifest in `artifacts/prod/`:

```bash
dbt build --target ci --select "state:modified+" \
  --state artifacts/prod --defer
```

The pieces do different jobs. `state:modified` compares with the saved manifest. The trailing [`+`](https://docs.getdbt.com/reference/node-selection/graph-operators) adds downstream descendants, not just the next model. [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) resolves eligible unselected dependencies against the saved environment instead of requiring you to rebuild them.

Deferral doesn't select the changed models. By default, an unselected relation already present in the CI target can still be used; `--favor-state` changes that preference. Tests with several parents can also read across CI and production, so I check what they will actually compare. A saved manifest isn't a guarantee of a small or self-contained run.

I keep that saved state outside the current run's `target/` directory. The [DAG and targets described in my dbt infrastructure article]({{< ref "/blog/dbt-data-infrastructure-as-code/" >}}) are what make this separation manageable.

### Keep Snowflake CI Resources Under Control

A few things deserve an explicit CI policy.

**Size the warehouse for the workload.** Start small and measure the build and its concurrency. An XS isn't a universal answer, especially when several PRs run together. Tests still execute SQL; a short YAML declaration doesn't make its warehouse cost disappear.

**Use transient storage when the data is disposable.** [Transient databases and tables](https://docs.snowflake.com/en/user-guide/tables-temp-transient) avoid Fail-safe storage, but that also means less recovery protection. It fits CI data I can reconstruct. It isn't a free optimization for data I need to keep.

**Control Dynamic Table refreshes explicitly.** [Target lag](https://docs.snowflake.com/en/user-guide/dynamic-tables/target-lag) is a freshness target, not a refresh interval or an off switch. Setting a very long lag doesn't guarantee no automatic refresh. If a CI check needs fixed contents, initialize the CI tables, wait for that load to finish, then explicitly [suspend scheduled refreshes](https://docs.snowflake.com/en/user-guide/dynamic-tables/manage#suspend-a-dynamic-table) before checking them. Suspension affects downstream Dynamic Tables too; the CI sequence needs to account for that graph.

**Clean up even when the job fails.** Drop the dedicated CI database in a cleanup step, and check for abandoned databases. A failed pipeline shouldn't leave its Dynamic Tables refreshing until someone notices the bill.

## The Errors That Convinced Me

**The unwanted null.** A boolean column contained `NULL` alongside `true` and `false`. That's allowed by a nullable SQL type, but it wasn't the contract we wanted. The source code treated the absence as false; the transformation didn't necessarily do the same. For a required boolean, `not_null` makes the requirement explicit. If I also use `accepted_values` for booleans, I set `quote: false` so they aren't treated as strings.

**The duplicate identifier.** A migration bug duplicated records in a source system. Without a uniqueness test, the duplicates could travel through the transformations without anything complaining. A small check on the identifier was more useful than another afternoon staring at the final report.

## What I Take Away

I still start with the boring checks. A required identifier gets `not_null` and `unique`. A closed set of statuses gets `accepted_values`, plus `not_null` if absence is invalid. A relationship gets checked where the model relies on it. Pick the tests that describe the column's actual job, then run them.

More complex tests come after that, when I can read what they do. That's where I learned the hard way: delegating test generation to an LLM without serious review can produce false confidence. I reviewed generated tests I'd let through at merge time and found inverted logic, the wrong table referenced, thresholds with no business meaning. They ran. They passed. They didn't test what I thought they tested.

The rule I apply now is simple: declarative checks I understand, enforced in CI. Custom tests when I have time to read them line by line. I'd rather have a smaller set of assertions I trust than a long list of green checks nobody can explain.
