---
title: "Data quality"
description: "dbt tests and constraints: define data quality rules in YAML and distinguish validation from database enforcement."
seoHub: true
translationKey: "hub-data-quality"
---

Data quality rules need an explicit owner and a clear place to run. A dbt test detects a violation; a database constraint can enforce a rule, with support depending on the platform.

The [guide to dbt tests and constraints]({{< ref "/blog/dbt-tests-constraints-yml/" >}}) explains the difference and shows how to express these checks in YAML. For related tooling, see [dbt-guard]({{< ref "/blog/dbt-guard-package-python/" >}}).
