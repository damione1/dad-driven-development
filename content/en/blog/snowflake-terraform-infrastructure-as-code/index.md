---
title: "Snowflake + Terraform: Stop Managing Your Data Infrastructure in SQL"
date: 2025-12-05
draft: false
description: "How to move from SQL scripts and migrations to Snowflake infrastructure managed with Terraform - layered roles, future grants, masking and reproducibility."
tags: ["Snowflake", "Terraform", "Data Engineering", "Infrastructure as Code"]
categories: ["Data Engineering", "Infrastructure"]
---

There's a moment in every data engineer's life when you find yourself staring at a 300-line SQL file that creates roles, grants, warehouses, and you wonder how you got here. This is my story.

## Snowflake: All SQL, for Better and Worse

Snowflake has accomplished something few platforms manage: completely abstracting physical infrastructure while retaining granular control over everything that matters. No servers to maintain, no clusters to size. Just databases, schemas, warehouses, roles, and SQL commands to drive everything.

Snowsight, its UI, is capable. You can create objects, manage access, visualize data. But like any tool that exists in both UI and programmatic form, the real power is on the code side. The UI gives you access to features. SQL gives you control. Programmatic manipulation gives you reproducibility.

It's the classic analogy with command-line tools: they seem less accessible than a GUI, but they're scriptable, versionable, automatable. Every action becomes reproducible, auditable, integrable into a pipeline. A `GRANT` executed in Snowsight disappears the moment you close the tab. The same `GRANT` declared in Terraform code is tracked, reviewable, reversible.

The problem is that Snowflake is so accessible in ad hoc SQL that you end up doing everything that way. A GRANT here, a new role through the console, a warehouse created in an emergency on a Tuesday night. Each of these actions is harmless on its own. Together, they become infrastructure that's impossible to audit.

## The SQL Scripts Era

When I first started building a Snowflake infrastructure, I did what everyone does: `.sql` files. One script to create databases, another for schemas, another for roles. Simple, direct, it works.

Except it works up to a point.

First week: 5 SQL files, well organized. First month: 15 files, a few `IF NOT EXISTS` for idempotence. Third week: you realize you need to modify a role and you're scanning 4 files to make sure you don't miss a grant somewhere. And you end up wondering whether the current state of Snowflake actually matches what's in your scripts, or if someone ran a `GRANT` manually in the console on a Tuesday night.

## The Migrations Episode

Then I tried migrations. Like in application development: numbered files, each with an incremental change. `001_create_databases.sql`, `002_add_marketing_role.sql`, `003_fix_grant_on_silver.sql`...

On paper, it's better. You have a history. You can trace the evolution. But in practice:
- If someone makes a manual change, the migrations and reality silently diverge
- Rolling back a `REVOKE` or `DROP ROLE` isn't like an `ALTER TABLE`, the cascade effects are unpredictable
- You have no idea about the current state without running a full audit
- And above all, you end up with 50 migration files and nobody knows what the infrastructure is *supposed* to be, just what it's *become*

## The Obvious Answer: Terraform

Then one day I wondered if there was a Terraform provider for Snowflake. The answer: yes, and it's [maintained by Snowflake directly](https://github.com/snowflakedb/terraform-provider-snowflake).

It seemed like the obvious fit. Terraform is exactly the right tool for this problem:
- **Declarative**: you describe the desired state, not the steps to get there
- **Plan before apply**: `terraform plan` shows you exactly what will change before it changes
- **Single source of truth**: the Terraform code describes the desired state of the infrastructure, not an approximation, not a history of migrations
- **History via git**: every change is a commit, reviewable, reversible
- **Idempotent**: you can run it 10 times, it gives the same result

Compared to SQL scripts: if someone makes a manual change in the Snowflake console, the next `terraform plan` shows it as drift. You see the difference between the real state and the desired state. With SQL scripts, you see nothing.

(In reality, Terraform has its own irritants: state that gets corrupted, painful `import`s, breaking changes in the provider between versions. But these problems are manageable, and the gain in visibility more than compensates.)

## Roles: Thinking in Layers

The concept that helped me most is structuring roles in three levels. It's a practice recommended by Snowflake in their [access control documentation](https://docs.snowflake.com/en/user-guide/security-access-control-considerations), documented by several Snowflake architects ([example on the official blog](https://medium.com/snowflake/rbac-and-cloning-for-devops-950c800c594e)).

**Level 1: Snowflake system roles.** ACCOUNTADMIN, SYSADMIN, SECURITYADMIN. You don't create them, they already exist. But you configure Terraform to use the right role in the right place: SYSADMIN to create databases and warehouses, SECURITYADMIN for roles and grants. Principle of least privilege.

**Level 2: Access roles.** These are technical, granular roles that grant access to a specific schema at a specific level. Like `SILVER_RO` (read-only on the Silver schema), `GOLD_RW` (read-write on Gold), `STAGING_FULL` (full access to the staging schema). Naming convention matters: by reading the role name, you know exactly what it does.

**Level 3: Functional roles.** These are the business roles, the ones you assign to humans. An `ANALYST` role, an `ENGINEER` role, a `REPORTING` role. Each functional role aggregates multiple access roles. The Analyst role gets `SILVER_RO` + `GOLD_RO`. The engineer gets broader access.

The flow is simple: **User → Functional role → Access roles → Permissions on schemas.**

The advantage of this approach: when a new analyst joins, you assign them to the ANALYST functional role and they automatically have access to everything they need. No grant list to maintain manually.

## Cascading Grants

One of the most satisfying aspects of this approach is grant cascading. In Snowflake, you can create a role hierarchy: one role can "contain" another.

Concretely, if you have three access levels for a schema (read, write, create), you structure it as a cascade:
- The Create role inherits from the Read-Write role
- The Read-Write role inherits from the Read-Only role
- You only need to grant privileges once at each level

When you assign the Create role to someone, they automatically get write and read permissions through inheritance. No grant duplication, no extra maintenance.

## Future Grants: Anticipating the Future

A critical pattern in Snowflake: [future grants](https://docs.snowflake.com/en/sql-reference/sql/grant-privilege.html). When you create a role with `SELECT` on a schema, it applies to tables that exist *now*. But what about when dbt creates a new model tomorrow? Without future grants, nobody has access to it.

Terraform lets you declare future grants: "all future objects in this schema will automatically inherit these permissions." It's the kind of detail that makes the difference between infrastructure that works on day 1 and infrastructure that still works 6 months later when the data team has added 50 models.

## Users: Configuration, Not Code

Adding a user, in this approach, doesn't require writing Terraform code. You fill in a configuration:

- Their identity (name, email)
- Their team (marketing, finance, data engineering...)
- Are they an admin?
- Do they need a development sandbox?

From this configuration, Terraform automatically calculates and generates:
- The user account
- Assignment to their team's functional role
- Default warehouse
- Sandbox access if applicable

Adding a new member means modifying a few lines of configuration, running `terraform plan` to verify, and `terraform apply`. No SQL to write, no grants to hunt for.

## The Terraform / dbt Boundary

An important point: where does Terraform stop and where does dbt begin?

My view: **Terraform manages everything above the schema. dbt manages everything below.**

Terraform handles:
- Creating databases and schemas
- Creating warehouses
- Managing roles and permissions
- Defining masking policies
- Configuring monitoring

dbt handles:
- Creating tables and views within schemas
- Transforming data
- Applying governance tags to columns
- Documenting models
- Testing data quality

Infrastructure (Terraform) evolves slowly: a new schema per month, a new role per quarter. Transformations (dbt) evolve every day: new models, business logic, corrections.

This separation means two teams can work in parallel without stepping on each other. The platform engineer manages Terraform, the data engineer manages dbt. Each in their own repo, with their own deployment cycle.

## Data Masking: A Good Integration Example

A concrete example of how Terraform and dbt collaborate: data masking.

Terraform creates [tag-based masking policies](https://docs.snowflake.com/en/user-guide/tag-based-masking-policies.html): "if a column is tagged PII, mask the value except for roles that have the right to see it." It also creates the tags and unmasking roles.

dbt, on its side, applies tags to columns when it creates models: "this email column is PII, this amount column is FINANCIAL."

The result: when a marketing user runs `SELECT * FROM clients`, emails are automatically masked. The user with the right role sees the real values. Nobody had to write masking logic in SQL. It's handled by the combination of infrastructure + metadata.

## Audit and Iterate

One of the most underrated benefits of this approach: auditability.

When a security audit asks "who has access to what?", the answer is in the code. Not in a Snowflake console where you have to navigate 15 pages. Not in a manually maintained Excel document. In the code, versioned, with the complete history of who changed what and when.

And when you want to add a new team, a new schema, or modify permissions, it's a standard process: modify the config, open a PR, review, merge, apply. The same workflow as for application code.

## LLM-Assisted

One last point worth mentioning: once your infrastructure is declared as structured code, an LLM becomes a remarkably effective assistant. You can ask it to create a new functional role, add a user, modify permissions, and it produces valid Terraform that follows your existing conventions.

With ad hoc SQL scripts, this is much riskier. The AI doesn't know the current state of the infrastructure. With Terraform, the desired state is in the code, and `plan` validates that the result is correct before any application. The AI proposes, Terraform verifies.

## The Real Gain

In the end, what changed isn't the technology, it's the confidence. You know that the state of your infrastructure matches the code. You know the permissions are correct. You know nobody made an undocumented change. And when someone asks you to add access or create a new schema, it's 5 minutes of configuration instead of half an hour of SQL scripts with the fear of breaking something.

SQL scripts are craftsmanship. Migrations are better-organized craftsmanship. Terraform isn't perfect either: state can be fragile, plan errors are sometimes cryptic, and the Snowflake provider has its own bugs. But at least you know where you stand. And when something breaks, you know why.
