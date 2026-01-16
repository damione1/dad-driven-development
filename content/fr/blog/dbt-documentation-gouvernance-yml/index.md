---
title: "dbt : Quand tes fichiers YAML deviennent ta gouvernance de données"
date: 2026-01-16
draft: false
description: "Transformer les YAML dbt d'une formalité optionnelle en source de vérité pour la gouvernance - classifications PII, persist_docs, masquage automatique et enforcement en CI."
tags: ["dbt", "Snowflake", "Data Governance", "Data Engineering"]
categories: ["Data Engineering", "Data Governance"]
---

La documentation, c'est le truc que personne ne veut faire. Surtout en data. T'as des centaines de colonnes dans des dizaines de tables, et quelqu'un te demande "c'est quoi le champ `status` dans la table `orders` ?" Et la réponse honnête, c'est souvent "euh... un enum je pense qui veut probablement dire X."

## Le problème de la documentation data

Dans un projet dbt classique, la documentation est optionnelle. Tu peux écrire tes modèles SQL, les déployer, et ne jamais documenter une seule colonne. dbt ne t'oblige à rien.

Le résultat, c'est prévisible : des schémas Snowflake avec des centaines de colonnes dont personne ne connaît la signification exacte. Des noms de colonnes hérités d'un système source vieux de 10 ans. Des colonnes qui s'appellent `type` ou `status` sans aucune indication de ce que ça veut dire.

Et quand un analyste marketing veut comprendre les données, il doit soit déranger quelqu'un qui a la connaissance tribale, soit deviner. Les deux sont problématiques.

## Les YAML de dbt : plus qu'une formalité

dbt a un mécanisme de documentation intégré : les fichiers `schema.yml` (ou peu importe comment tu les nommes). Tu peux y décrire chaque modèle, chaque colonne, avec du texte libre. La plupart des équipes s'en servent peu ou pas.

Mais si tu prends le temps de bien structurer ces fichiers, ils deviennent bien plus qu'une documentation passive. Ils deviennent la source de vérité pour la gouvernance de tes données.

L'idée, c'est d'utiliser le champ `meta` de chaque colonne pour stocker des métadonnées structurées :

- **Classification** : est-ce que cette colonne contient de l'information personnelle (PII), financière, confidentielle ?
- **Catégorie sémantique** : est-ce un email, un montant, une adresse, un identifiant ?
- **Sensibilité** : haute, moyenne, basse ?
- **Obligations réglementaires** : LPRPDE, GDPR, données consommateurs, e-commerce ?
- **Rétention** : combien de temps garder ces données ?

Quand chaque colonne a ses métadonnées, tu passes d'une documentation passive à une gouvernance active.

## persist_docs : du YAML à Snowflake

Le truc qui fait la différence, c'est [`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs). C'est une option dbt qui prend tes descriptions YAML et les pousse comme commentaires sur les objets Snowflake. Quand tu actives ça :

```yaml
models:
  mon_projet:
    +persist_docs:
      relation: true
      columns: true
```

Chaque description de modèle et de colonne dans tes YAML devient un commentaire visible dans Snowflake. Pas besoin d'un outil externe. Quelqu'un qui navigue les données dans Snowsight voit directement les descriptions que t'as écrites dans dbt.

Et si tu utilises [Snowflake Horizon](https://docs.snowflake.com/en/user-guide/snowflake-horizon) (leur plateforme de gouvernance), ces descriptions alimentent directement le catalogue de données. Ta documentation dbt EST ta documentation Snowflake. Une seule source de vérité.

## Forcer la documentation en CI

Mon point de vue là-dessus est assez tranché : la documentation, c'est aussi obligatoire que le code lui-même. Pas un nice-to-have, pas quelque chose qu'on fera "quand on aura le temps". Si tu déploies du code non testé en production, c'est un problème. Déployer un modèle non documenté devrait l'être tout autant.

Et si tu automatises tout le reste (déploiements, tests, validations), pourquoi la documentation échapperait-elle à cette logique ? La CI est la réponse évidente. C'est le premier check que j'y mets : avant même de valider la logique des transformations, on vérifie que la documentation est là.

Concrètement, ça donne une étape qui valide la complétude :
- Chaque modèle Silver et Gold doit avoir une description
- Chaque colonne dans le SQL doit être présente dans le YAML
- Chaque colonne documentée doit avoir une description non vide

Si une PR ajoute un modèle sans documentation, la CI fail. Point final. C'est la seule façon de maintenir la discipline sur le long terme. Le package [dbt-meta-testing](https://github.com/tnightengale/dbt-meta-testing) fait exactement ça : il expose des macros `required_docs` et `required_tests` que tu branches dans ta pipeline.

## Les tags de gouvernance : du YAML au masquage

Là où ça devient puissant, c'est quand tu combines les métadonnées YAML avec les fonctionnalités de gouvernance de Snowflake.

Tu déclares dans ton YAML qu'une colonne contient du PII. dbt peut appliquer un tag Snowflake correspondant quand il crée le modèle, via un post-hook qu'on écrit nous-mêmes, c'est pas du built-in. Et Snowflake, grâce à des politiques de masquage liées aux tags, masque automatiquement la valeur pour les utilisateurs qui n'ont pas le bon rôle.

Le résultat : tu documentes tes colonnes dans le YAML, et le masquage de données se fait tout seul. Pas de logique de masquage dans le SQL, pas de vues spéciales, pas de maintenance. La gouvernance découle directement de la documentation.

## Les tests : la documentation qui se vérifie

Les YAML de dbt ne servent pas qu'à la documentation, ils servent aussi aux tests. Et c'est là que ça boucle : ta documentation devient vérifiable.

Tu documentes qu'un `order_id` est une clé primaire ? Mets un test `unique` et `not_null`. Tu documentes qu'un statut ne peut avoir que certaines valeurs ? Mets un test `accepted_values`. Tu documentes qu'une colonne référence une autre table ? Mets un test de relation.

Les tests sont déclarés au même endroit que la documentation. Un seul fichier YAML qui dit : "cette colonne s'appelle X, elle contient Y, elle ne peut pas être nulle, et ses valeurs possibles sont Z." La documentation et les tests sont la même chose. On revient là-dessus en détail dans [l'article sur les tests dbt]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}).

Quand les tests passent, ta documentation est prouvée correcte. Quand un test fail, ta documentation ou tes données sont fausses. Dans les deux cas, tu dois investiguer.

## Les contacts de gouvernance

Un aspect souvent négligé : qui est responsable de quoi ? Dans les métadonnées de chaque modèle, tu peux déclarer un propriétaire, un steward, un approbateur. Ces métadonnées sont poussées vers Snowflake via un post-hook custom (encore une fois, c'est du bricolage maison, pas du dbt natif).

Quand quelqu'un trouve un problème de données, il sait exactement qui contacter. Pas besoin de chercher dans un wiki ou de demander à la cantonade. L'information est attachée directement aux données elles-mêmes.

## Le coût réel : c'est moins que tu penses

Le reproche classique : "ça prend du temps de documenter chaque colonne." C'est vrai. Mais compare avec l'alternative :
- Des heures perdues par les analystes à deviner ce que les colonnes veulent dire
- Des erreurs dans les rapports parce que quelqu'un a interprété un champ de travers
- Des audits qui prennent des semaines parce que personne sait quelles données sont sensibles
- Des incidents de sécurité parce qu'une colonne PII n'était pas identifiée comme telle

Documenter une colonne prend 30 secondes. Ne pas la documenter peut coûter des heures, des jours, ou pire. Et avec les LLMs, le coût de départ a encore baissé : [documenter une base source entière]({{< ref "/blog/dbt-documenter-source-llm-multi-agent/" >}}) en quelques jours n'est plus une utopie.

## Le bonus : la documentation comme clé du talk to my data

La documentation sert deux audiences : tes collègues humains, et les LLMs. Et c'est là que ça devient intéressant.

[Snowflake Cortex Analyst](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst) est la réponse de Snowflake au "talk to my data" : poser une question en langage naturel et obtenir la requête SQL correcte en retour. Demander une KPI précise, la filtrer selon des critères, la comparer avec une autre métrique, sans écrire une ligne de SQL. Ça paraît magique. Et les gens pensent spontanément que c'est des mois de travail d'ingénierie pour y arriver.

Ce n'est pas le cas, si la documentation est propre.

Cortex Analyst fonctionne à partir d'un modèle sémantique : un fichier YAML qui décrit les tables, les colonnes, les métriques, les relations entre entités, et le vocabulaire métier associé. La structure de ce fichier est très proche de ce que dbt produit déjà dans ses fichiers de documentation. L'infrastructure existe. Les modèles existent. Si les descriptions de colonnes sont là, si les relations sont documentées, si les métriques clés sont définies, le gap pour Cortex Analyst est faible.

Snowflake Labs a même publié [`dbt_semantic_view`](https://github.com/Snowflake-Labs/dbt_semantic_view), un package qui génère des Semantic Views Snowflake directement depuis le modèle sémantique dbt. Ces Semantic Views sont nativement exploitables par Cortex Analyst. Le pipeline devient : dbt documente → dbt_semantic_view publie les vues sémantiques → Cortex Analyst répond aux questions en langage naturel.

Un projet dbt bien documenté est à 2-4 semaines d'un talk to my data fonctionnel, pas à 6 mois. Le travail restant (formaliser quelques métriques, distinguer dimensions et mesures, ajouter des synonymes métier) est marginal comparé à ce qui est déjà en place. C'est un quick win qui ne se débloque qu'avec une condition : avoir traité la documentation comme une contrainte dès le début, pas comme une tâche à faire plus tard.

## Ce que j'en retiens

Au final, ce qui a marché pour moi, c'est de traiter la documentation des données avec le même sérieux que le code de transformation. Pas comme un nice-to-have qu'on fera "quand on aura le temps." Comme une partie intégrante du pipeline, vérifiée en CI, propagée automatiquement.

Les YAML de dbt sont l'endroit idéal pour ça. Un seul fichier qui sert à la documentation humaine, aux tests automatisés, et à la gouvernance de données. Ça sonne fancy dit comme ça, mais en pratique c'est juste des fichiers YAML qui font leur job.
