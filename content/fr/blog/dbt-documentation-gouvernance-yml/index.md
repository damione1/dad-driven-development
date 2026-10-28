---
title: "Documentation dbt : du YAML à la gouvernance"
translationKey: "dbt-documentation-governance"
date: 2026-01-16
lastmod: 2026-10-28
draft: false
description: "Descriptions YAML, persist_docs et contrôles CI : comment je documente mes modèles dbt dans Snowflake et relie leurs métadonnées à la gouvernance."
tags: ["dbt", "Snowflake", "Data Governance", "Data Engineering"]
categories: ["Data Engineering", "Data Governance"]
images: ["dbt-documentation-governance-yml-featured.jpg"]
---

Dans mes projets dbt, le YAML rassemble les descriptions, les tests et les responsables de la donnée. J’utilise `persist_docs` pour faire remonter les descriptions des modèles dans Snowflake, puis des contrôles CI pour éviter que la doc décroche tranquillement du code. La gouvernance vient de ce qu’on construit autour de ces métadonnées. Elle n’arrive pas parce qu’un fichier est en YAML.

La documentation reste le truc que personne n’a envie de faire. Surtout en data. Tu as des centaines de colonnes dans des dizaines de tables, et quelqu’un demande ce que veut dire le champ `status` dans `orders`. La réponse honnête, c’est souvent « euh... un enum qui doit probablement vouloir dire X ».

C’est ce problème-là que je veux faire résoudre au YAML. Un endroit pour écrire ce qu’on sait, à côté du code qui s’en sert, et un build qui remarque quand on arrête de le faire.

## La documentation dbt commence par les descriptions

Un projet dbt peut avoir du SQL qui tourne très bien et presque aucune documentation utile. Une colonne appelée `type` ne dit toujours pas à un analyste si c’est une catégorie de client, un moyen de paiement ou un reste d’un système source dont personne ne se souvient.

Le mécanisme de base est simple : des [descriptions de modèles et de colonnes](https://docs.getdbt.com/reference/resource-properties/description) dans un fichier de propriétés comme `schema.yml`. Le nom du fichier n’est pas le contrat. Son contenu, oui.

Un petit exemple, avec les conventions de responsabilité et de classification que je veux à côté du texte :

```yaml
version: 2
models:
  - name: dim_customers
    description: "Une ligne par client, y compris les comptes inactifs."
    config:
      meta:
        owner: data-platform
    columns:
      - name: customer_id
        description: "Identifiant client stable provenant du système source."
      - name: email
        description: "Email de contact du client ; peut être absent."
        config:
          meta:
            classification: pii
            semantic_type: email
            steward: customer-operations
```

Ces clés `meta`, ce sont nos conventions. dbt accepte des [métadonnées personnalisées](https://docs.getdbt.com/reference/resource-configs/meta) ; il ne décide pas ce que `pii` veut dire et ne fait pas respecter une politique de confidentialité à notre place. La syntaxe actuelle place les métadonnées de colonne sous `config.meta` (changement en dbt 1.10, rétroporté en 1.9).

Je peux aussi y mettre la sensibilité, une règle de rétention ou un responsable métier. Ce qui sert, c’est d’avoir un vocabulaire sur lequel l’équipe s’est mise d’accord, plutôt que cinq orthographes de « confidentiel » et une page de wiki que personne ne met à jour.

## Générer la doc dbt, puis la rendre utile

Pour un site de documentation local, les [commandes dbt docs](https://docs.getdbt.com/reference/commands/cmd-docs) donnent un endroit où lire ces descriptions :

```bash
dbt docs generate
dbt docs serve
```

Ça rend le projet navigable. Ça ne rend pas les descriptions justes. « Identifiant client » est techniquement une description ; savoir si elle explique la différence entre un compte et une personne, c’est une autre question. Générer un site et relire son contenu sont deux tâches différentes.

## persist_docs : du YAML aux commentaires Snowflake

[`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs), c’est l’option qui évite aux descriptions de rester enfermées dans dbt. Dans `dbt_project.yml` :

```yaml
models:
  mon_projet:
    +persist_docs:
      relation: true
      columns: true
```

Pour les matérialisations de modèles prises en charge, dbt persiste les descriptions des relations et colonnes comme commentaires dans l’entrepôt lorsqu’il construit les objets. Quelqu’un qui les consulte dans Snowsight peut lire les descriptions là aussi. On arrête d’entretenir une deuxième copie à la main.

Il y a des frontières. Le réglage ci-dessus concerne les relations et colonnes des modèles. dbt **n’implémente pas** `persist_docs` pour les sources. Pour écrire des commentaires sur les tables sources brutes, il faut un script ou une macro séparée utilisant [Snowflake `COMMENT`](https://docs.snowflake.com/en/sql-reference/sql/comment), avec les bons privilèges sur les objets. Et les commentaires ne transforment pas les champs `meta` personnalisés en tags Snowflake ou en politiques de masquage.

C’est cette séparation que je garde explicite : les descriptions deviennent des commentaires grâce à une fonctionnalité dbt prise en charge ; le reste des métadonnées a besoin de sa propre intégration.

## Forcer la documentation en CI

Mon avis n’a pas changé : la documentation fait partie du modèle. Si une PR ajoute une colonne inexpliquée, « on la documentera plus tard » n’est pas un plan de déploiement très convaincant.

Les contrôles que je veux sont concrets :

- Chaque modèle Silver et Gold a une description.
- Chaque colonne de la relation construite est couverte dans le YAML.
- Chaque colonne documentée a une description non vide.

Un contrôle statique peut inspecter les descriptions du manifest avant le build. Comparer les colonnes déclarées à celles de la relation réelle, c’est un autre contrôle : la relation doit exister.

Le package [`dbt-meta-testing`](https://github.com/tnightengale/dbt-meta-testing) est une manière de faire ce deuxième travail. Sa macro `required_docs` vérifie les modèles configurés après leur construction ; les modèles éphémères sont exclus. `required_tests` vérifie la couverture de tests configurée. Ce sont des opérations à mettre dans le pipeline, pas des contrôles que dbt lance discrètement pour tous les projets. Je figerais la version du package et je la vérifierais avec la version de dbt qu’on utilise.

Une séquence CI avec ce package peut construire les modèles concernés, lancer `dbt run-operation required_docs`, puis faire échouer le job si la documentation est incomplète. Les [tests de données]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}), séparément, vérifient les assertions sur les lignes. Les deux comptent. Une belle description ne repère pas un identifiant dupliqué.

## Des métadonnées aux tags et au masquage

Un champ YAML qui dit `classification: pii` ne masque rien tout seul.

L’implémentation a plusieurs morceaux. Les tags et les politiques de masquage Snowflake sont créés et maintenus. Un hook personnalisé ou une étape de déploiement traduit nos valeurs de métadonnées convenues en tags existants. Snowflake applique ensuite la politique associée au moment de la requête. Le [masquage basé sur les tags](https://docs.snowflake.com/en/user-guide/tag-based-masking-policies) demande Enterprise Edition ou supérieur et les privilèges appropriés.

Il reste du SQL de politique à écrire, des comportements de rôles à vérifier et une intégration à maintenir. Je veux que le YAML fournisse la classification, plutôt que de la recopier dans chaque vue. Mais je testerais le résultat avec les rôles censés voir des données masquées. La présence d’un tag ne prouve pas que tout le montage fait ce qu’on voulait.

La responsabilité fonctionne pareil. Un owner ou un steward dans `meta` aide quelqu’un qui lit la doc dbt. Pour rendre ce contact visible ailleurs dans Snowflake, il faut une traduction explicite en commentaires, en tags ou une autre intégration de catalogue. Le YAML est l’entrée de ce travail, pas un moyen de transport à lui seul.

## Les tests vérifient des assertions, pas chaque phrase

La boucle reste intéressante : tu documentes qu’un `order_id` est une clé, tu ajoutes `unique` et `not_null`. Tu documentes les statuts permis, tu ajoutes `accepted_values`. Tu documentes une référence vers une autre table, tu ajoutes `relationships`.

Le même YAML porte l’explication et le contrôle. C’est ça que j’aime.

Mais des tests qui passent ne soutiennent que les assertions qu’ils vérifient vraiment, sur les données qu’ils ont interrogées. Ils ne prouvent pas qu’une définition métier est juste, qu’un responsable est toujours le bon ou qu’une classification remplit une obligation légale. Quelqu’un doit encore relire ces décisions. Je développe la partie tests dans [l’article sur les tests dbt et la CI]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}).

## La documentation aide la couche sémantique ; elle ne la remplace pas

Les descriptions ont un deuxième public : les outils qui aident les gens à poser des questions à leurs données.

[Snowflake Cortex Analyst](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst) génère du SQL à partir de questions en langage naturel et d’un contexte sémantique. La documentation actuelle de Snowflake recommande les Semantic Views pour ce contexte, et oriente les nouvelles implémentations vers Cortex Agents. Le prérequis qui m’intéresse reste le même : quelqu’un doit définir ce que les données veulent dire.

De bonnes descriptions de colonnes sont une entrée utile. Ce n’est pas un modèle sémantique complet. Il faut encore les relations, les dimensions, les faits, les métriques métier et des exemples de questions auxquelles on veut que le système réponde.

Le package Snowflake Labs [`dbt_semantic_view`](https://github.com/Snowflake-Labs/dbt_semantic_view) permet de définir un modèle `semantic_view` avec le SQL des vues sémantiques Snowflake. Il ne convertit pas automatiquement des descriptions dbt ordinaires, ni un modèle sémantique MetricFlow, en couche conversationnelle qui fonctionne. Les définitions restent à écrire.

Et il faut évaluer les réponses générées. Les [évaluations Cortex Analyst](https://docs.snowflake.com/en/user-guide/snowflake-cortex/cortex-analyst-evaluations) comparent les résultats du SQL généré à ceux de requêtes vérifiées. Je ne promettrais pas une date de livraison parce que le YAML est propre. Je commencerais par un petit domaine métier, des questions qu’on peut contrôler et des définitions dont quelqu’un accepte d’être responsable.

Le travail de documentation n’est pas perdu. C’est la partie qu’il faudrait sinon reconstruire plus tard, en essayant en plus de déboguer une réponse qui paraît plausible.

## Ce que j’en retiens

Documenter chaque colonne prend du temps. Demander à la même personne ce qu’un champ veut dire chaque fois qu’on prépare un nouveau rapport, aussi. Les LLM peuvent aider à produire un premier jet, comme je l’ai raconté dans [le projet de documentation d’une base source]({{< ref "/blog/dbt-documenter-source-llm-multi-agent/" >}}). Relire le sens reste notre boulot.

Ce qui a marché pour moi, c’est de traiter la documentation avec le même sérieux que le [code de transformation]({{< ref "/blog/dbt-data-infrastructure-as-code/" >}}) : versionnée, relue et contrôlée en CI. Les descriptions restent à côté des modèles ; les fonctionnalités prises en charge les transportent dans l’entrepôt ; des intégrations explicites emmènent les métadonnées de gouvernance plus loin.

Un seul fichier ne nous donne pas une gouvernance des données. Il nous donne un endroit pour rendre les règles visibles, et un pipeline dans lequel faire respecter celles qu’on a vraiment implémentées. C’est déjà beaucoup plus utile que du texte optionnel que personne ne lit.
