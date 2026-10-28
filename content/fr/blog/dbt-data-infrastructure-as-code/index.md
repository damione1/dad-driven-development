---
title: "dbt et Terraform : ma stack Snowflake en code"
date: 2025-12-26
lastmod: 2026-10-28
draft: false
description: "Comment je répartis ma stack Snowflake entre Terraform et dbt : modèles, Dynamic Tables, targets isolées, sélection CI et limites du manifest."
tags: ["dbt", "Snowflake", "Data Engineering", "SQL"]
categories: ["Data Engineering"]
images: ["dbt-data-infrastructure-as-code-featured.jpg"]
---

J’utilise Terraform pour provisionner mon environnement Snowflake et dbt pour décrire les transformations à l’intérieur. Les databases, les rôles et les warehouses d’un côté ; les modèles, les tests et la documentation de l’autre. Les deux sont du code, relu dans git et déployé par un pipeline.

Snowflake est fondamentalement SQL-first. C’est sa force : grants, objets, transformations, tout se pilote en SQL. Dans [l’article sur Terraform]({{< ref "/blog/snowflake-terraform-infrastructure-as-code/" >}}), j’expliquais comment j’avais arrêté de gérer l’infrastructure avec un tas de scripts SQL. Les transformations méritaient le même traitement.

Sinon, tu te retrouves avec des fichiers SQL éparpillés, un README qui explique l’ordre et un seul collègue qui sait quel script doit vraiment passer en premier. Je ne veux pas que le plan de déploiement vive dans la tête de ce collègue.

## dbt vs Terraform : la frontière que je choisis

Les deux outils se complètent dans cette stack. Voilà où je mets la responsabilité :

| Besoin | Ce que j’utilise |
|---|---|
| Databases, warehouses, rôles et leurs privilèges | Terraform |
| SQL de transformation, dépendances et matérialisations | dbt |
| Tests de données, descriptions et lignage du projet | dbt |
| Rafraîchissement continu des Dynamic Tables déployées | Le scheduler Snowflake, quand il est configuré |
| Revue et déploiement des changements | La CI autour de ces outils |

C’est un choix de responsabilité, pas une affirmation que leurs capacités ne se recouvrent jamais. dbt peut gérer certains grants ; Terraform peut gérer des objets de base de données. La règle qui compte, c’est qu’un même objet n’ait pas deux outils qui essaient chacun d’être son autorité.

## Des scripts SQL aux modèles dbt et au DAG

Avant cette structure, les DDL s’accumulent. `create_orders_table.sql`, puis `add_status_column.sql`, puis `alter_orders_add_shipping_address.sql`. Au bout d’un moment, reconstituer ce que la table est censée être commence à ressembler à de l’archéologie.

Les [modèles dbt](https://docs.getdbt.com/docs/build/models) donnent un endroit à la transformation. Un modèle SQL décrit un résultat ; sa matérialisation détermine si dbt construit une vue, une table ou un autre objet pris en charge. Les références aux autres modèles établissent les dépendances.

```sql
-- models/silver/orders_enriched.sql
select
    o.order_id,
    o.created_at,
    o.status,
    c.email as customer_email,
    p.name as product_name
from {{ ref('stg_orders') }} o
left join {{ ref('stg_customers') }} c on o.customer_id = c.customer_id
left join {{ ref('stg_products') }} p on o.product_id = p.product_id
```

[`ref`](https://docs.getdbt.com/reference/dbt-jinja-functions/ref) résout le nom de la relation et donne la dépendance à dbt. Il sait que ce modèle dépend de ces trois modèles de staging, et peut ordonner les builds sélectionnés en conséquence.

Il ne propage pas toutes les nouvelles colonnes à travers le SQL à notre place. Si j’ajoute `shipping_country` en amont, ce `SELECT` explicite doit encore la sélectionner en aval. Ce qui disparaît, c’est l’ordre des scripts entretenu à la main, pas la nécessité de décider ce que contient chaque modèle.

Les tests et les descriptions vivent à côté de ce code. Je les développe séparément dans [l’article sur les tests YAML]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}) et [celui sur la documentation et la gouvernance]({{< ref "/blog/dbt-documentation-gouvernance-yml/" >}}).

## Jusqu’où tient l’analogie avec Terraform

Le parallèle est utile : décrire le résultat voulu, le versionner, relire le changement et laisser un outil le faire traverser le graphe de dépendances.

Mais aucune des deux commandes ne veut dire « tout détruire et reconstruire ». [`dbt build`](https://docs.getdbt.com/reference/commands/build) exécute les ressources sélectionnées et leurs tests dans l’ordre du DAG, selon leurs matérialisations. Un modèle incrémental peut être mis à jour de façon incrémentale. Une Dynamic Table a son propre comportement de déploiement et de rafraîchissement. La sélection et `--full-refresh` changent ce qui se passe.

De la même manière, Terraform planifie les changements sur l’infrastructure qu’il gère, plutôt que de recréer toute la stack à chaque `apply`.

La plus grande différence, c’est le [state Terraform](https://developer.hashicorp.com/terraform/language/state/purpose) : il garde l’association entre les ressources configurées et les objets réels. Un [`terraform plan`](https://developer.hashicorp.com/terraform/cli/commands/plan) normal actualise les informations sur ces objets et propose des changements.

Les artifacts du projet dbt ne fournissent pas cette même réconciliation générale de l’infrastructure. Ils servent à décrire le projet et à comparer ses versions. Ils ne prouvent pas que personne n’a modifié une table à la main hier.

Je continue à voir dbt comme de l’infrastructure as code. Je ne demande simplement pas à l’analogie de faire un travail que l’outil ne fait pas.

## Dynamic Tables : déclarer le modèle, choisir qui le rafraîchit

Les [Dynamic Tables Snowflake](https://docs.snowflake.com/en/user-guide/dynamic-tables-intro), c’est là où cette manière de penser devient particulièrement utile. On déclare une requête, et Snowflake peut en entretenir le résultat par des rafraîchissements au lieu de planifier chaque transformation comme un batch séparé.

L’[adaptateur dbt Snowflake](https://docs.getdbt.com/reference/resource-configs/snowflake-configs) prend en charge une matérialisation `dynamic_table`. Cet exemple utilise la configuration de planification de l’adaptateur pour dbt 1.12+ et choisit volontairement le rafraîchissement géré par Snowflake :

```yaml
# dbt_project.yml
models:
  mon_projet:
    silver:
      +materialized: dynamic_table
      +snowflake_warehouse: TRANSFORMING_L
      +scheduler: ENABLE
      +target_lag: '1 hour'
```

Le SQL du modèle définit toujours la donnée. La configuration définit comment ce résultat est maintenu. Le [target lag](https://docs.snowflake.com/en/user-guide/dynamic-tables/target-lag) est un objectif de fraîcheur par rapport aux données sous-jacentes, pas « lance une exécution pile toutes les heures » ni une garantie que cet objectif sera toujours tenu.

La planification et le mode de rafraîchissement sont deux décisions différentes. Snowflake peut faire un rafraîchissement complet ou incrémental quand la requête et la configuration le permettent. La version de l’adaptateur compte aussi : les configurations actuelles distinguent la planification gérée par Snowflake du rafraîchissement géré par dbt. Je vérifierais l’adaptateur installé avant de copier une configuration d’une autre version.

Sur cette partie du DAG, Snowflake peut reprendre le travail de rafraîchissement récurrent. Il faut encore déployer les changements, lancer les tests adaptés, vérifier la fraîcheur des sources et surveiller les échecs. Le reste du pipeline ne disparaît pas. Je développe le côté CI dans [l’article sur les tests dbt]({{< ref "/blog/dbt-tests-contraintes-yml/" >}}).

## Les targets : un environnement par contexte

Un [target dbt](https://docs.getdbt.com/docs/local/profiles.yml) choisit les paramètres de connexion d’une exécution. Je veux que la cible par défaut d’un développeur pointe vers une database de développement isolée, avec un warehouse dimensionné pour ce travail.

Voici la partie environnements d’un `profiles.yml`. C’est un extrait : les réglages de compte, d’utilisateur et d’authentification restent à fournir avec la configuration prise en charge par l’adaptateur Snowflake.

```yaml
mon_projet:
  target: dev
  outputs:
    dev:
      type: snowflake
      database: DEV_JEAN
      schema: silver
      warehouse: DEV_XS
      role: DBT_DEV
    prod:
      type: snowflake
      database: PROD
      schema: silver
      warehouse: TRANSFORMING_L
      role: DBT_PROD
```

`dbt run` utilise le target par défaut ; `dbt run --target prod` choisit explicitement la production. C’est un garde-fou utile. La protection contre une écriture en production vient des privilèges Snowflake et des identifiants utilisés par le job, pas du fait d’avoir nommé une cible `dev`.

Le projet peut aussi redéfinir les paramètres de database et de schéma. Je veux que ces redéfinitions soient relues avec le même soin que le SQL. Une cible de développement isolée ne sert que si les modèles restent réellement dans la frontière qu’on lui a donnée.

## Sélectionner la partie du DAG qui a besoin de travail

Quand un projet a des centaines de modèles, tout reconstruire à chaque changement est rarement ce que je veux. Les [opérateurs de graphe](https://docs.getdbt.com/reference/node-selection/graph-operators) rendent le périmètre explicite :

```bash
# Juste le modèle
dbt run --select orders_enriched

# Le modèle et ses dépendances en amont
dbt run --select +orders_enriched

# Le modèle et tous ses descendants en aval
dbt run --select orders_enriched+

# Les deux directions
dbt run --select +orders_enriched+
```

Les tags donnent un autre point d’entrée. Un modèle avec `{{ config(tags=['retail', 'orders']) }}` peut être sélectionné par `dbt run --select tag:retail`. Il faut encore choisir si on inclut ses dépendances.

En CI, un manifest de production précédent permet de sélectionner les ressources modifiées. Avec un target `ci` configuré séparément et propre, une commande peut ressembler à ça :

```bash
dbt build --select "state:modified+" --state artifacts/prod --defer --target ci
```

`state:modified+` sélectionne les ressources modifiées et tous leurs descendants, pas seulement les enfants directs. [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) est le mécanisme séparé qui peut résoudre les références amont non sélectionnées vers les relations de production existantes. Les relations déjà présentes dans la cible courante peuvent influencer cette résolution, d’où l’intérêt d’une CI isolée. Les artifacts de production sauvegardés restent hors du `target/` de l’exécution courante, où ils ne seront pas écrasés.

C’est la partie que j’aime : le périmètre est écrit. On peut l’inspecter, le relire et le changer. « Ne lancer que ce qui a changé » arrête d’être une hypothèse enfouie dans le scheduler.

## Sources et exposures : les deux bouts du lignage

Une [déclaration de source](https://docs.getdbt.com/docs/build/sources) donne un nom à une donnée qui arrive déjà par un autre système. Elle ne charge pas cette donnée à notre place.

```yaml
sources:
  - name: shopify
    database: RAW
    schema: shopify
    tables:
      - name: orders
      - name: customers
      - name: products
```

Les modèles peuvent référencer `{{ source('shopify', 'orders') }}`. On peut déclarer des tests et configurer des contrôles de fraîcheur sur ces sources, puis lancer réellement `dbt source freshness` dans le pipeline. Une déclaration de source seule n’est pas un service d’alerte.

La même frontière s’applique à la documentation. Les descriptions dans dbt servent sa propre doc ; pousser des commentaires sur les tables sources demande une implémentation séparée. [`persist_docs`]({{< ref "/blog/dbt-documentation-gouvernance-yml/" >}}) traite les descriptions des modèles pris en charge, pas les commentaires de sources ni les tags de gouvernance personnalisés.

Les [exposures](https://docs.getdbt.com/docs/build/exposures) décrivent les consommateurs à l’autre bout : un dashboard, une API, un export.

```yaml
exposures:
  - name: tableau_revenue_dashboard
    type: dashboard
    owner:
      name: Équipe BI
      email: analytics@example.com
    depends_on:
      - ref('daily_revenue')
      - ref('top_products')
```

Le projet garde maintenant le fait que le dashboard dépend de ces modèles. Ça permet une revue d’impact. Envoyer une notification quand ils changent reste à la charge du workflow ou d’une intégration. Le YAML n’inscrit pas l’équipe BI à des emails tout seul.

## Seeds, fonctions et macros

Les [seeds](https://docs.getdbt.com/docs/build/seeds) sont des fichiers CSV que dbt peut charger comme tables : petites listes de référence, mappings, configuration statique. Je préfère voir leurs changements dans le repo plutôt que de me demander quelle copie d’un tableur le job a ramassée.

Les [fonctions définies par l’utilisateur](https://docs.getdbt.com/docs/build/udfs) ont aussi un support natif à partir de dbt 1.11, selon l’adaptateur et le langage de la fonction. Elles ne rentrent plus dans une catégorie générale « dbt ne sait pas gérer ça ». D’anciens projets peuvent toujours les déployer par hooks ou macros ; la version qu’on utilise détermine les options.

Le système de [macros Jinja](https://docs.getdbt.com/docs/build/jinja-macros) et des packages comme [dbt-utils](https://github.com/dbt-labs/dbt-utils) sont une autre raison pour laquelle l’outil me convient. Je peux réutiliser une transformation ou un test, et ajouter des contrôles propres au projet quand le comportement intégré ne couvre pas le besoin. C’est un point d’extension utile, pas une promesse de reproduire chaque fonctionnalité d’une plateforme hébergée avec quelques lignes de Jinja.

## Le manifest : état du projet, pas de l’entrepôt

[`dbt parse`](https://docs.getdbt.com/reference/commands/parse) parse et valide le projet, puis écrit `manifest.json`. Il ne se connecte pas à l’entrepôt et ne compile pas le SQL des modèles.

Le [manifest](https://docs.getdbt.com/reference/artifacts/manifest-json) décrit les ressources, les configurations, les colonnes déclarées et les dépendances. Les commandes qui exécutent des nœuds peuvent ajouter le SQL compilé de ces nœuds. C’est un artifact de ce que dbt comprend du projet, pas un inventaire en direct de tout Snowflake.

Ça reste extrêmement utile. On peut comparer les versions, sélectionner des ressources modifiées et inspecter les contrats. C’est ce que j’utilise dans [dbt-guard]({{< ref "/blog/dbt-guard-package-python/" >}}) pour chercher les breaking changes avant le déploiement. Mais comparer des manifests et contrôler la dérive d’une infrastructure répond à deux questions différentes.

## Pourquoi j’ai choisi dbt plutôt que SQLMesh

[SQLMesh](https://sqlmesh.readthedocs.io/en/stable/) était sur ma liste quand j’ai évalué cette approche. Ses [plans et catégories de changements](https://sqlmesh.readthedocs.io/en/stable/concepts/plans/) rendent les changements et les backfills explicites. Il a aussi un [mode de compatibilité dbt](https://sqlmesh.readthedocs.io/en/stable/integrations/dbt/), avec des limites à vérifier sur le projet réel.

J’ai choisi dbt pour l’écosystème dans lequel je voulais travailler : les packages du [dbt Hub](https://hub.getdbt.com/), les intégrations déjà présentes dans ma stack et la familiarité du workflow. C’était un choix pratique, pas un benchmark prouvant qu’un outil gagne dans chaque catégorie.

Il y a aussi plusieurs endroits où faire tourner dbt. [dbt Projects on Snowflake](https://docs.snowflake.com/en/user-guide/data-engineering/dbt-projects-on-snowflake) fournit des runtimes gérés dans Snowflake, avec ses propres options de développement et de planification. C’est un produit d’exécution précis. Utiliser l’adaptateur Snowflake de dbt depuis un portable, un runner CI ou un service dbt hébergé ne veut pas dire à lui seul que le runtime dbt est dans Snowflake.

SQLMesh aurait aussi pu être un choix raisonnable. Choisir une infrastructure, c’est aussi choisir qui la comprendra, comment elle sera déployée et quelles parties on devra entretenir nous-mêmes.

## Ce que ça change pour l’équipe

La partie que je ramène du développement logiciel est simple : quelqu’un d’autre doit pouvoir ouvrir le repo et comprendre comment le système est censé fonctionner.

Terraform déclare l’environnement. dbt déclare les transformations et leurs contrôles. Snowflake exécute le SQL et, là où on a choisi des Dynamic Tables, gère leurs rafraîchissements. Chaque couche a une responsabilité qu’on peut montrer, et les changements passent par une revue plutôt que par un déploiement manuel le vendredi soir.

La connaissance ne devient pas complète parce qu’elle est dans le code. Mais le graphe de dépendances, la configuration et les tests donnent à la prochaine personne quelque chose de bien meilleur que « demande au collègue qui sait ».

La rigueur qu’on applique au code applicatif n’a aucune raison de ne pas s’appliquer aux données. C’est ce que je veux de cette stack.
