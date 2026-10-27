---
title: "Qualité des données"
description: "Tests et contraintes dbt : définir les règles de qualité en YAML et distinguer validation et application par la base de données."
seoHub: true
translationKey: "hub-data-quality"
---

Une règle de qualité doit avoir un responsable et un point de contrôle explicites. Un test dbt détecte une violation ; une contrainte peut imposer une règle dans la base, selon les possibilités de la plateforme.

Le [guide des tests et contraintes dbt]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}) explique cette différence et présente leur déclaration en YAML. Pour l’outillage associé, voir [dbt-guard]({{< ref "/blog/dbt-guard-package-python/" >}}).
