---
title: "dbt : tes transformations de données comme de l'infrastructure"
date: 2025-12-26
draft: false
description: "dbt applique à tes pipelines de données la même rigueur que Terraform à l'infra - approche déclarative, DAG, Dynamic Tables, manifest et comparaison avec SQLMesh."
tags: ["dbt", "Snowflake", "Data Engineering", "SQL"]
categories: ["Data Engineering"]
---

Snowflake est fondamentalement SQL-first. C'est sa force : tout se pilote en SQL, des grants à la création d'objets en passant par les transformations. L'infrastructure, on a vu comment la dompter avec Terraform dans [l'article précédent]({{< ref "/blog/snowflake-terraform-infrastructure-as-code/" >}}). Mais les transformations de données, elles, tombent dans un angle mort. Des scripts SQL éparpillés, pas de tests, pas de versioning sérieux, un seul collègue qui sait dans quel ordre tout lancer.

C'est là que dbt entre en jeu. Et ce n'est pas un hasard si les deux outils fonctionnent si naturellement ensemble : Snowflake abstrait l'infrastructure physique, dbt abstrait l'orchestration des transformations. Les deux sont déclaratifs, SQL-first, conçus pour que le code soit la source de vérité. dbt n'est pas un framework générique de transformation de données, c'est l'outil qui complète Snowflake là où Snowflake ne complète pas lui-même : l'organisation, les tests, la documentation et le déploiement reproductible de tout ce SQL.

## Le problème avec les scripts SQL de transformation

Avant dbt, le pipeline de données classique ressemblait à ça : des scripts SQL numérotés, un scheduler quelconque, et un README qui expliquait dans quel ordre lancer quoi. Ou pire, des procédures stockées dans Snowflake avec une logique métier enfouie dedans.

Le problème, c'est le même qu'avec les scripts SQL d'infrastructure : les DDL s'accumulent. Tu as `create_orders_table.sql`, puis `add_status_column.sql`, puis `alter_orders_add_shipping_address.sql`. Au bout d'un moment, personne ne sait ce que la table est *censée* être. Tu dois exécuter tous les scripts dans l'ordre pour reconstituer l'état réel, et tu espères qu'il n'y en a pas un qui manque.

Et quand tu veux ajouter une colonne dans un pipeline en 5 phases (staging, enrichissement, agrégation, export, rapport), tu dois ouvrir 5 fichiers, ajouter la colonne dans chacun, puis coordonner le déploiement dans le bon ordre en prod. Sans framework, tu gères ça à la main. Et tu oublies presque toujours un modèle quelque part.

Quant aux tests : écrire un test de non-nullité sur une colonne sans framework, c'est écrire un `SELECT COUNT(*) WHERE col IS NULL` et vérifier que le résultat est 0. Faisable une fois. Pénible à maintenir sur 200 colonnes.

## Ce que dbt change fondamentalement

[dbt](https://www.getdbt.com/) (data build tool) prend une idée simple : les transformations de données, c'est du code. Et le code, ça se gère avec les mêmes pratiques que le reste : version control, tests, documentation, CI/CD.

Concrètement, tu écris des modèles : des fichiers `.sql` qui définissent chacun une table ou une vue. Un modèle référence d'autres modèles avec une syntaxe `{{ ref('nom_du_modele') }}`. dbt résout automatiquement les dépendances et construit un [DAG](https://docs.getdbt.com/terms/dag) (graphe acyclique dirigé) de tes transformations.

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

dbt sait que `orders_enriched` dépend de `stg_orders`, `stg_customers` et `stg_products`. Il les construit dans le bon ordre, automatiquement. Quand tu ajoutes une colonne à `stg_orders`, elle est disponible dans tous les modèles en aval sans que tu aies à coordonner quoi que ce soit.

## L'analogie avec Terraform

Le parallèle avec Terraform n'est pas superficiel. Les deux outils partagent la même philosophie fondamentale :

**État souhaité, pas étapes.** Tu ne dis pas "exécute cette transformation après celle-là". Tu décris ce que chaque table doit contenir, et dbt figure comment y arriver.

**Déclaratif.** Ton code décrit le résultat, pas le processus. `SELECT ... FROM ref('source')` dit "cette table contient ces colonnes calculées depuis cette source", pas "prends cette table, jointure-la avec ça, filtre ça".

**Reproductible.** `dbt build` reconstruit tout depuis zéro, dans le bon ordre, à chaque fois. Comme `terraform apply` reconstruit ton infra depuis les fichiers HCL.

**Versionné.** Chaque changement dans une transformation est un commit. Tu peux voir qui a modifié quoi, pourquoi, et revenir en arrière.

L'analogie a quand même une limite importante : dbt n'a pas de state distant. Si quelqu'un supprime une table dans Snowflake à la main, dbt n'en sait rien. Le prochain `dbt run` va simplement la recréer sans te signaler le drift. Terraform, lui, detecterait l'écart entre l'état réel et l'état souhaité et te proposerait de le corriger. Ce n'est pas un défaut critique, mais c'est à garder en tête : dbt est déclaratif sur ce qu'il crée, pas sur ce qui existe.

## L'extensibilité : ce qui compense largement

Ce manque de state distant, c'est un des rares points où Terraform fait mieux. Mais dbt compense avec quelque chose de différent : une extensibilité remarquable.

Le [dbt Hub](https://hub.getdbt.com/) regroupe des centaines de packages communautaires. Des tests additionnels, des macros utilitaires, des intégrations avec des sources spécifiques. [dbt-utils](https://github.com/dbt-labs/dbt-utils) est probablement dans tous les projets dbt sérieux : il ajoute des dizaines de macros et de tests que tu n'aurais pas envie d'écrire toi-même.

Mais la vraie puissance, c'est le système de [macros Jinja](https://docs.getdbt.com/docs/build/jinja-macros). Tout ce que dbt fait en interne, tu peux le faire aussi. Ce qui veut dire que les fonctionnalités réservées à dbt Cloud sont souvent reproductibles en dbt Core avec quelques macros.

La détection de breaking changes dans les schémas ? On en parle dans [un prochain article]({{< ref "/blog/dbt-guard-package-python/" >}}), et c'est exactement ce genre de feature qu'on peut implémenter soi-même. L'alerting sur la fraîcheur des sources ? Des macros. La génération automatique de documentation ? Des macros. dbt Core n'est pas une version appauvrie de dbt Cloud, c'est une fondation sur laquelle tu construis ce dont tu as besoin.

## dbt comme infrastructure, pas juste comme ETL runner

Beaucoup d'équipes utilisent dbt comme un runner ETL : elles matérialisent des tables, et font tourner `dbt run` toutes les heures dans Airflow ou Prefect. C'est une utilisation valide. Mais c'est loin d'épuiser ce que dbt peut faire.

Snowflake a un concept qui change l'équation : les [Dynamic Tables](https://docs.snowflake.com/en/user-guide/dynamic-tables-intro). Au lieu de matérialiser une table et la rafraîchir manuellement, tu déclares une Dynamic Table avec un lag cible ("cette table doit être à jour à moins de 1 heure"). Snowflake gère le rafraîchissement automatiquement, en propageant les changements dans le DAG.

Combiné avec dbt, ça veut dire que tu n'as plus besoin d'un orchestrateur externe pour gérer les rafraîchissements. Tu déclares tes modèles comme des Dynamic Tables dans dbt, tu déploies, et Snowflake s'occupe du reste. dbt devient ce qu'il est vraiment : un outil de déclaration d'infrastructure de données, pas un scheduler.

```yaml
# dbt_project.yml
models:
  mon_projet:
    silver:
      +materialized: dynamic_table
      +target_lag: '1 hour'
      +snowflake_warehouse: TRANSFORMING_L
```

## Les targets : un environnement par contexte

Un des concepts les plus pratiques de dbt, c'est le système de [targets](https://docs.getdbt.com/docs/core/connect-data-platform/connection-profiles). Un target, c'est une configuration de déploiement : quelle database, quel schéma, quel warehouse utiliser.

En pratique, chaque développeur a un target `dev` par défaut qui pointe vers sa propre database isolée, avec un petit warehouse pour ne pas bruler des crédits inutilement :

```yaml
# profiles.yml
mon_projet:
  target: dev
  outputs:
    dev:
      database: DEV_JEAN
      schema: silver
      warehouse: DEV_XS
    prod:
      database: PROD
      schema: silver
      warehouse: TRANSFORMING_L
```

Quand Jean lance `dbt run`, il déploie dans `DEV_JEAN.silver`. Il ne peut pas accidentellement écraser la production. Pour déployer en prod, il faut être explicite : `dbt run --target prod`. C'est un opt-in volontaire, pas le comportement par défaut.

Ce qui est élégant, c'est que dbt permet d'[overrider la database, le schéma, et même le warehouse](https://docs.getdbt.com/docs/build/custom-schemas) à plusieurs niveaux : au niveau du projet, du dossier, du modèle individuel. Tu peux avoir un modèle spécifique qui tourne toujours sur un warehouse dédié, indépendamment du target. La configuration se compose.

## Sélectionner avec précision dans le DAG

Une fois que tu as un DAG de 200 modèles, tu ne veux pas systématiquement tout rebuilder. dbt a un [système de sélection](https://docs.getdbt.com/reference/node-selection/syntax) expressif pour cibler exactement ce dont tu as besoin.

La syntaxe `+` contrôle la direction dans le DAG :

```bash
# Juste le modèle
dbt run --select orders_enriched

# Le modèle et tout ce qui est en amont (ses dépendances)
dbt run --select +orders_enriched

# Le modèle et tout ce qui est en aval (ce qui en dépend)
dbt run --select orders_enriched+

# Le modèle, ses dépendances ET ses dépendants
dbt run --select +orders_enriched+
```

Tu peux aussi sélectionner par tag. Si tu tagges tes modèles retail :

```yaml
-- models/silver/orders_enriched.sql
{{ config(tags=['retail', 'orders']) }}
```

Tu peux alors cibler tout le domaine retail en une commande :

```bash
dbt run --select tag:retail
```

Combiné avec `state:modified`, ça donne un contrôle très précis sur ce qui tourne en CI : seulement les modèles modifiés et leurs dépendants directs.

## Sources et exposures : les deux bouts du DAG

dbt gère non seulement les transformations intermédiaires, mais aussi les deux extrémités du pipeline.

**Les sources** sont les tables que tu ne contrôles pas : tes données brutes qui arrivent via Fivetran, Airbyte, ou un autre outil de réplication. Tu les [déclares dans les YAML](https://docs.getdbt.com/docs/build/sources) :

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

Ça te permet de les référencer avec `{{ source('shopify', 'orders') }}` dans tes modèles, de leur associer des tests, et de monitorer leur fraîcheur. Si les données Shopify n'ont pas été mises à jour depuis 6 heures, dbt peut te l'alerter avant que tes rapports soient périmés.

Ces descriptions de sources s'intègrent aussi avec Snowflake Horizon : les métadonnées que tu déclares dans dbt remontent dans le [catalogue de données Snowflake](https://docs.snowflake.com/en/user-guide/snowflake-horizon), visible par tous ceux qui ont accès à l'instance.

**Les exposures** sont l'autre bout : les consommateurs de tes données que dbt ne gère pas. Un dashboard Tableau, une API, un fichier exporté vers un partenaire. Tu les [déclares aussi](https://docs.getdbt.com/docs/build/exposures) :

```yaml
exposures:
  - name: tableau_revenue_dashboard
    type: dashboard
    owner:
      name: Équipe BI
    depends_on:
      - ref('daily_revenue')
      - ref('top_products')
```

Ça complète le lignage : tu peux voir dans le DAG non seulement comment les données se transforment, mais aussi où elles finissent. Quand tu modifies `daily_revenue`, dbt sait que le dashboard Tableau en dépend et peut t'alerter.

## Les assets complémentaires : seeds et UDFs

Deux éléments méritent d'être mentionnés pour compléter le tableau.

**Les [seeds](https://docs.getdbt.com/docs/build/seeds)** sont des fichiers CSV que dbt gère comme des tables. Des tables de référence, des mappings, des configurations statiques. Au lieu d'avoir un fichier CSV quelque part sur un serveur ou dans un Google Sheet, tu le versiones dans le repo dbt et dbt crée la table correspondante dans Snowflake.

```
seeds/
  product_categories.csv
  country_codes.csv
  shipping_carriers.csv
```

**Les UDFs** (User Defined Functions) Snowflake ne sont pas gérées nativement par dbt, mais tu peux les déployer via des [pre-hooks ou des macros](https://docs.getdbt.com/docs/build/jinja-macros). C'est la limite de la métaphore infrastructure-as-code : certains objets Snowflake restent dans le territoire Terraform.

## Le manifest : le plan compilé de ta stack

Quand tu lances [`dbt parse`](https://docs.getdbt.com/reference/commands/parse), dbt compile tout ton projet et produit un [`manifest.json`](https://docs.getdbt.com/reference/artifacts/manifest-json). C'est l'artifact central : une représentation complète et lisible par machine de tout ce que ton projet sait faire.

Le manifest contient les modèles, leurs colonnes déclarées, leurs dépendances, les tests, les sources, les exposures. Tout.

C'est l'équivalent du state Terraform pour tes données. Et comme le state Terraform, il peut être utilisé pour faire des comparaisons : qu'est-ce qui a changé entre la version précédente et la version actuelle ? C'est exactement ce qu'on exploitera dans [un prochain article]({{< ref "/blog/dbt-guard-package-python/" >}}) pour détecter les breaking changes avant qu'ils arrivent en production.

## Pourquoi ça change la dynamique d'équipe

Avant dbt, les transformations étaient souvent la propriété d'une personne. Quelqu'un savait dans quel ordre lancer les scripts, quelles tables dépendaient de quoi, où était la logique de déduplication. La connaissance était tribale.

Avec dbt, n'importe quel data engineer peut ouvrir le repo, voir le DAG, comprendre les dépendances, modifier un modèle, et déployer dans son environnement dev sans risque. La connaissance est dans le code.

C'est le même gain qu'on a eu avec Terraform pour l'infra. On passe de "demande à la personne qui sait" à "lis le code".

## Pourquoi dbt et pas SQLMesh ?

Quand j'ai commencé à évaluer des outils pour gérer mes transformations de façon déclarative, [SQLMesh](https://sqlmesh.com/) était sur ma liste. C'est un outil sérieux : open source, Apache 2.0, rétrocompatible avec les projets dbt existants, avec des concepts intéressants comme la détection automatique des breaking changes et l'évaluation incrémentale native.

Sur le papier, SQLMesh est techniquement plus rigoureux que dbt sur plusieurs points. Mais j'ai choisi dbt pour une raison simple : l'effet de masse.

dbt est de loin l'outil le plus utilisé dans l'écosystème data. Ce qui veut dire :
- Des centaines de packages dans le [dbt Hub](https://hub.getdbt.com/) - des tests, des macros, des utilitaires déjà écrits
- Une communauté active avec des réponses à presque tous les problèmes que tu vas rencontrer
- Des intégrations natives dans tous les outils de la stack : Fivetran, Hightouch, Metabase, Tableau, et l'essentiel des data catalogues
- Des data engineers qui connaissent déjà dbt quand ils arrivent dans une équipe

Et depuis 2024, Snowflake a formalisé ce partenariat avec une intégration encore plus profonde : la [dbt Snowflake Native App](https://docs.getdbt.com/docs/cloud-integrations/snowflake-native-app). Tu peux désormais orchestrer et monitorer tes jobs dbt directement depuis Snowflake, sans infrastructure externe. dbt tourne dans ton instance Snowflake, pas à côté.

SQLMesh aurait probablement aussi bien fait le job. Mais quand tu choisis une infrastructure, tu choisis aussi un écosystème. Et l'écosystème dbt est imbattable pour l'instant.

## dbt + Snowflake : le duo pour un data engineer qui vient du dev

Ce qui rend cette combinaison particulièrement solide, c'est que les deux outils partagent la même philosophie et qu'ils se complètent sans se chevaucher.

Terraform déclare l'infrastructure : les databases, les schémas, les rôles, les permissions, les politiques de masquage. dbt déclare les transformations : les modèles, les tests, la documentation, le lignage. Les deux sont du code, versionnés dans git, déployés via des pipelines CI, avec des environnements isolés par développeur.

Mais au-delà de la philosophie partagée, l'intégration est concrète. Les Dynamic Tables Snowflake se matérialisent nativement dans dbt. Les descriptions de colonnes dans tes YAML remontent dans Snowflake Horizon via `persist_docs`. Et depuis 2024, dbt peut tourner directement dans ton instance Snowflake via la [dbt Snowflake Native App](https://docs.getdbt.com/docs/cloud-integrations/snowflake-native-app), sans infrastructure externe à gérer.

La stack complète ressemble à ça : Terraform gère ce qui est au-dessus des données, dbt gère ce qui est dedans, et Snowflake est le socle sur lequel les deux s'appuient. Chaque couche a sa responsabilité, chaque couche est du code. Pour quelqu'un qui vient du dev, c'est exactement ce à quoi une stack data devrait ressembler. Pas des scripts SQL dans un dossier partagé, pas de connaissances tribales, pas de déploiement manuel un vendredi soir.

La rigueur qu'on applique au code applicatif n'a aucune raison de ne pas s'appliquer aux données. dbt est l'outil qui rend ça possible, et Snowflake est la plateforme sur laquelle cette rigueur prend tout son sens.
