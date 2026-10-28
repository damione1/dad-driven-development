---
title: "dbt Documentation: YAML, persist_docs and Governance"
translationKey: "dbt-documentation-governance"
date: 2026-01-16
lastmod: 2026-10-28
draft: false
description: "How I use dbt YAML, persist_docs and CI checks to document Snowflake models, then connect their metadata to data governance."
tags: ["dbt", "Snowflake", "Data Governance", "Data Engineering"]
categories: ["Data Engineering", "Data Governance"]
images: ["dbt-documentation-governance-yml-featured.jpg"]
---

In my dbt projects, YAML is where descriptions, tests and ownership live together. I use `persist_docs` to carry model descriptions into Snowflake, then CI checks to stop the documentation quietly falling behind. The governance part comes from what we build around that metadata. It doesn't happen because a file happens to be YAML.

Documentation is still the thing nobody wants to do. Especially in data. You have hundreds of columns across dozens of tables, and someone asks what the `status` field in `orders` means. The honest answer is often “uh... an enum I think that probably means X.”

That's the problem I want the YAML to solve. A place to write down what we know, next to the code that uses it, and a build that notices when we stop doing it.

## dbt documentation starts with descriptions

A dbt project can have perfectly runnable SQL and almost no useful documentation. A column called `type` still doesn't tell an analyst whether it means a customer category, a payment method or something left over from a source system nobody remembers.

The basic mechanism is simple: [model and column descriptions](https://docs.getdbt.com/reference/resource-properties/description) in a properties file such as `schema.yml`. The filename isn't the contract. The content is.

Here's a small example, with the ownership and classification conventions I want alongside the prose:

```yaml
version: 2
models:
  - name: dim_customers
    description: "One row per customer, including inactive accounts."
    config:
      meta:
        owner: data-platform
    columns:
      - name: customer_id
        description: "Stable customer identifier from the source system."
      - name: email
        description: "Customer contact email; may be absent."
        config:
          meta:
            classification: pii
            semantic_type: email
            steward: customer-operations
```

Those `meta` keys are our conventions. dbt accepts [custom metadata](https://docs.getdbt.com/reference/resource-configs/meta); it doesn't decide what `pii` means or enforce a privacy policy for us. The current syntax places column metadata under `config.meta` (changed in dbt 1.10 and backported to 1.9).

I can add sensitivity, a retention rule or a business owner there too. The useful part is having a vocabulary the team agrees on, rather than five spellings of “confidential” and a wiki page nobody updates.

## Generate dbt docs, then keep them useful

For a local documentation site, the [dbt docs commands](https://docs.getdbt.com/reference/commands/cmd-docs) give the descriptions somewhere to live:

```bash
dbt docs generate
dbt docs serve
```

That makes the project browsable. It doesn't make the descriptions correct. “Customer identifier” is technically a description; whether it explains the difference between an account and a person is another question. Generating a site and reviewing its contents are different jobs.

## persist_docs: from YAML to Snowflake comments

[`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs) is the option that stops the YAML descriptions from staying inside dbt. In `dbt_project.yml`:

```yaml
models:
  my_project:
    +persist_docs:
      relation: true
      columns: true
```

For supported model materializations, dbt persists relation and column descriptions as warehouse comments when it builds the objects. Someone browsing those objects in Snowsight can read the descriptions there. We stop maintaining a second set by hand.

There are boundaries. The setting above applies to model relations and columns. dbt does **not** implement `persist_docs` for sources. To write comments on raw source tables, we need a separate script or macro using [Snowflake `COMMENT`](https://docs.snowflake.com/en/sql-reference/sql/comment), with the right object privileges. And comments don't turn custom `meta` fields into Snowflake tags or masking policies.

That's the division I keep explicit: descriptions become comments through a supported dbt feature; the rest of the metadata needs its own integration.

## Enforcing documentation in CI

My view hasn't changed: documentation is part of the model. If a PR adds an unexplained column, “we'll document it later” isn't a particularly convincing deployment plan.

The checks I want are concrete:

- Every Silver and Gold model has a description.
- Every column in the built relation is accounted for in the YAML.
- Every documented column has a non-empty description.

A static check can inspect descriptions in the manifest before building. Comparing the declared columns with the actual relation is a different check: the relation has to exist.

The [`dbt-meta-testing`](https://github.com/tnightengale/dbt-meta-testing) package is one way to do that second job. Its `required_docs` macro checks configured models after they've been built; ephemeral models are excluded. `required_tests` checks the configured test coverage. These are operations we have to put in the pipeline, not checks dbt silently runs for every project. I would pin the package version and verify it against the dbt version we run.

A CI sequence using that package can build the relevant models, run `dbt run-operation required_docs`, then fail the job on incomplete documentation. The separate [data tests]({{< ref "/blog/dbt-tests-constraints-yml/" >}}) check the assertions about the rows. Both matter. A beautiful description doesn't catch a duplicate ID.

## From metadata to tags and masking

A YAML field saying `classification: pii` doesn't mask anything on its own.

The implementation has a few separate pieces. Snowflake tags and masking policies are created and maintained. A custom hook or deployment step maps our agreed metadata values to those existing tags. Snowflake then applies the associated policy at query time. [Tag-based masking](https://docs.snowflake.com/en/user-guide/tag-based-masking-policies) requires Enterprise Edition or higher and the appropriate privileges.

There is still policy SQL to write, role behavior to check and an integration to maintain. I want the YAML to provide the classification, rather than copying it into every view. But I'd test the result with the roles that should see masked data. The presence of a tag isn't the same as proof that the whole setup does what we intended.

Ownership works the same way. An owner or steward in `meta` helps a reader of dbt docs. Making that contact visible elsewhere in Snowflake requires an explicit mapping to comments, tags or another catalog integration. The YAML is the input to that work, not a transport mechanism by itself.

## Tests verify assertions, not every sentence

The nice loop is still there: document that an `order_id` is a key, add `unique` and `not_null`. Document the permitted statuses, add `accepted_values`. Document a reference to another table, add `relationships`.

The same YAML carries the explanation and the check. That's what I like about it.

But passing tests only supports the assertions those tests actually check, against the data they queried. It doesn't prove that a business definition is right, that an owner is current or that a classification meets a legal obligation. Someone still has to review those decisions. I go further into the testing side in [the article on dbt tests and CI]({{< ref "/blog/dbt-tests-constraints-yml/" >}}).

## Documentation helps the semantic layer; it doesn't replace it

There's a second audience for the descriptions: tools that help people ask questions about their data.

[Snowflake Cortex Analyst](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst) generates SQL from natural-language questions using semantic context. Snowflake's current documentation recommends Semantic Views for that context, and points new implementations toward Cortex Agents. The prerequisite I care about stays the same: somebody has to define what the data means.

Good column descriptions are useful input. They aren't a complete semantic model. We still need relationships, dimensions, facts, business metrics and examples of the questions we expect the system to answer.

The Snowflake Labs [`dbt_semantic_view`](https://github.com/Snowflake-Labs/dbt_semantic_view) package lets us define a `semantic_view` model with Snowflake's semantic-view SQL. It doesn't automatically convert ordinary dbt descriptions, or a MetricFlow semantic model, into a working conversational layer. The definitions still have to be written.

And the generated answers need evaluation. [Cortex Analyst evaluations](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst-evaluations) compare the results of generated SQL with those of verified queries. I wouldn't promise a delivery date just because the YAML is clean. I would start with a small business domain, questions we can check and definitions somebody is willing to own.

The documentation work isn't wasted. It's the part we'd otherwise have to reconstruct later, while also trying to debug an answer that sounds plausible.

## What I take away

Documenting every column takes time. So does asking the same person what a field means every time somebody needs a new report. LLMs can help with a first draft, as I described in [the source-database documentation project]({{< ref "/blog/dbt-document-sources-llm-multi-agent/" >}}). Reviewing the meaning is still our job.

What worked for me is treating documentation with the same seriousness as the [transformation code]({{< ref "/blog/dbt-data-infrastructure-as-code/" >}}): versioned, reviewed and checked in CI. Descriptions stay next to the models; supported features carry them into the warehouse; explicit integrations carry the governance metadata further.

A single file doesn't give us data governance. It gives us a place to make the rules visible, and a pipeline in which to enforce the ones we've actually implemented. That's already a lot more useful than optional prose nobody reads.
