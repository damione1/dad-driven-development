---
title: "Documenting a Source Database with Multi-Agent LLMs"
date: 2026-02-06
draft: false
description: "A multi-agent approach to automatically generate dbt source YAML documentation - what LLMs discover, their limits, and the complete workflow."
tags: ["dbt", "LLM", "Data Engineering", "AI"]
categories: ["Data Engineering", "AI"]
---

Documenting columns in a source database is the kind of task nobody wants to do. You have an operational system with hundreds of tables, thousands of columns, and documentation ranging from "nonexistent" to "a 2017 comment that says `TODO: document this`."

## The Context

When you work with dbt and [define your sources](https://docs.getdbt.com/docs/build/sources), you ideally want to document every column. Not just its name and type, but what it actually represents, its quirks, its possible values, its relationships with other tables.

The bronze layer (raw data as it arrives from source systems) is the hardest to document. Unlike silver and gold layers, where the transformation itself is a form of documentation (the SQL says what the data is supposed to be), the bronze layer inherits the conventions, bugs and design decisions of the system that feeds it. The knowledge doesn't live in dbt, it lives in the application codebase.

This is also where everything rests. If you don't know what a bronze column means, you can't correctly document its silver transformation, or the business metric it feeds at the gold layer. Documentation builds from the bottom up, and the bottom is the hardest part.

The problem is that this knowledge is often scattered. It's in the application code that writes to these tables. It's in the heads of backend developers. Sometimes it's in a wiki nobody has updated since 2019.

And nobody wants to spend 3 weeks sifting through legacy code to understand what `legacy_field_42` means.

## The Idea: Specialized LLM Agents

The approach I experimented with is using LLM agents to do the heavy lifting of investigation. Not a single giant prompt trying to understand everything at once, but a multi-agent approach where each agent has a specific role.

The principle:
1. **Explorer agent**: traverses the source database schema, identifies tables and columns, notes types, apparent FKs, naming patterns
2. **Code analyst agent**: takes the application code that interacts with each table and analyzes how each column is used: reads, writes, validations applied, transformations
3. **Documentarian agent**: synthesizes information from the two previous agents and produces structured documentation in dbt YAML format

Each agent works table by table, column by column. Methodical and systematic.

## What the Agents Discover

The most interesting part is what the agents find that nobody knew (or had forgotten):

**Repurposed columns.** A `notes` column that theoretically contains free text, but in practice stores serialized JSON with a specific structure that the frontend parses.

**Magic values.** A `status` that takes values 0, 1, 2, 3, 4. But nobody remembers that 3 means "pending manual validation" and 4 means "automatically cancelled by the system." The agent analyzing the code finds the constants and conditions.

**Implicit constraints.** A column with no NOT NULL constraint in the database, but which the application code never leaves empty. Or a column that should be unique but has duplicates due to a bug fixed 3 years ago.

**Serialized data.** JSON, XML, proprietary formats inside a text field. The agent identifies the format and documents the internal structure.

**Undocumented relationships.** FKs that don't exist in the database but the code uses systematically. Columns referencing other tables via a naming convention nobody formalized.

## Integration into dbt

Once the YAMLs are generated and validated, they integrate directly into the dbt project as source definitions. With `persist_docs` enabled, descriptions surface in Snowflake and classification metadata feeds governance policies. This mechanism is covered in detail in [the article on YAMLs as governance]({{< ref "/blog/dbt-documentation-governance-yml/" >}}).

What matters here: the agents transform a tedious documentation exercise into a concrete foundation for active governance, without it being a separate project.

## From Bronze to Upper Layers

Once the bronze layer is documented, something changes in how you document the rest.

In silver, each dbt model is an explicit transformation from known sources. The SQL itself says a lot: a `total_amount` column calculated from `unit_price * quantity` doesn't need a long description. What matters is documenting cleaning decisions, deduplication rules, edge cases. And an LLM can infer those by reading the SQL and the bronze documentation in parallel.

In gold, models are often business aggregations. Columns correspond to metrics whose meaning is in business logic, not in code. That's where documentation becomes more manual, but at least you start from a solid foundation. You know exactly what each upstream field represents, which makes documenting derived metrics much more precise.

The leverage effect is real: the bronze layer is the longest to document and the hardest to partially automate. Upper layers benefit directly from this foundation work. Every correctly described bronze column propagates through the lineage and reduces documentation work for the layers that depend on it.

This is also what makes bronze documentation so worthwhile to do first, despite the effort: it's the only place where knowledge is buried in an external codebase, and therefore the only place where LLM agents have a real advantage over a data engineer who doesn't know that code.

## The Limits

Let's be honest about what works less well:

**Business context.** An LLM can understand that `creation_date` is a creation date. It can't know that in your context, this date has a precise contractual meaning that affects downstream calculations. Fine-grained business context remains human.

**Unreadable legacy code.** When the code interacting with a table is a 3000-line file with no clear structure, even an LLM struggles to extract coherent documentation from it.

**Validation.** Everything an LLM produces must be validated by someone who knows the domain. Agents do the heavy lifting, but validation, correction and adding business context remain essential and irreplaceable. And as I often say, we're responsible for how we use AI, and that includes validating what it produces.

## The Complete Workflow

In practice:

1. You give your agents the source database schema dump and the application code
2. The agents produce documented YAML files, table by table
3. A human reviews, corrects errors, adds missing business context
4. The corrected YAMLs become the source definitions in dbt
5. Metadata is pushed to Snowflake via [`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs)
6. Classifications feed governance policies

Total time? For a database with a few hundred tables: a few days of agents plus a few days of human review. Without agents, that's weeks or months of manual work that nobody wants to do.

One last practical tip: if you're going to do this, go all in. Full reverse-engineering prompt, deep thinking mode enabled, systematic table-by-table review. This means burning a few million tokens with our friends at OpenAI, Anthropic or Google, but it's a one-time investment for a long-lasting asset. Run it overnight. In the morning, you have a first version of documentation for your entire bronze layer, and you didn't have to manually work through a single `legacy_field_42` yourself.

## The Lesson

Source documentation is one of the best use cases for LLMs in data engineering. It's not glamorous, it's not machine learning, it's not data science. It's foundational work, tedious but essential, that LLMs do well because it's systematic, the context is in the code, and the output is structured.

And unlike other LLM applications, here validation is straightforward: a data engineer or backend developer can verify the produced documentation in a few minutes per table. Errors are easy to spot and correct.

It's not magic. It's just a good tool applied to the right problem.
