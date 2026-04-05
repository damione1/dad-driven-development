---
title: "dbt-guard : Mon premier package Python (et pourquoi j'en avais besoin)"
date: 2026-04-05
draft: false
description: "Créer dbt-guard, un outil d'analyse statique offline qui détecte les breaking changes dans un DAG dbt en comparant deux manifests - sans connexion à Snowflake."
tags: ["dbt", "Python", "Data Engineering", "Open Source"]
categories: ["Data Engineering", "Open Source"]
---

Publier un package sur PyPI. C'est un de ces trucs qui a l'air intimidant de l'extérieur, mais qui, finalement, s'avère être une question de bon timing et d'un problème assez précis à résoudre.

## Le problème

Quand tu travailles avec dbt, tu finis par avoir un DAG (directed acyclic graph) avec des dizaines, voire des centaines de modèles qui dépendent les uns des autres. Le truc, c'est que dbt ne te dit pas quand tu casses quelque chose en aval. Tu renommes une colonne dans un modèle silver, tu pushes ta PR, les tests passent... et trois jours plus tard, quelqu'un se rend compte qu'un dashboard gold est cassé parce qu'il référençait cette colonne.

C'est le genre de problème silencieux. Pas d'erreur à la compilation. Pas de test qui fail. Juste un consommateur en aval qui se retrouve avec des données manquantes.

## Mais ça existe sûrement déjà ?

Avant de me lancer, j'ai fait ce que tout le monde fait : chercher si quelqu'un avait déjà résolu le problème.

Il y a quelques outils dans l'écosystème qui s'en approchent :

- **[dbt Core 1.5+ avec model contracts](https://docs.getdbt.com/docs/mesh/govern/model-contracts)** : si tu déclares `contract: enforced: true` sur un modèle, dbt détecte les breaking changes au moment du run. Mais ça exige une configuration explicite modèle par modèle. Sur un projet existant, c'est pas réaliste de tout rétroférer. Ok un LLM peut refactorer ton projet pour ajouter les contrats, mais j'aime pas vraiment cette utilisation des contracts. Je les vois plutot comme une spec OpenAPI (vision de developpeur) que comme un outil de validation de breaking changes (vision d'ops).
- **[Recce](https://github.com/DataRecce/recce)** : l'outil open source le plus complet pour la validation de changements dbt. Il utilise [SQLGlot](https://github.com/tobymao/sqlglot) pour analyser les breaking changes, il est sérieux et actif. Mais son workflow est pensé pour comparer deux environnements connectés (dev et prod) avec accès à la base de données.
- **dbt-manifest-differ** : compare deux manifests, mais pour déboguer pourquoi dbt a marqué un nœud comme `state:modified`. Pas pour détecter des breaking changes de colonnes.

Aucun ne répondait exactement à mon besoin : **comparer deux [manifests](https://docs.getdbt.com/reference/artifacts/manifest-json) en offline, sans connexion à Snowflake, sans configuration préalable des modèles**.

La raison pour laquelle ce gap existe, je pense, c'est une combinaison de facteurs. Les équipes qui ont des DAGs assez grands pour souffrir de ce problème sont souvent sur dbt Cloud, qui a la détection de breaking changes derrière son offre payante. Et dans la culture data engineering, l'outillage se connecte au warehouse par réflexe. L'idée d'analyse statique sur des fichiers JSON locaux est contre-intuitive dans cet écosystème.

Il y a aussi un prérequis implicite : pour que l'analyse statique sur le manifest fonctionne, il faut que les colonnes soient documentées dans les fichiers YAML. Ce qui n'est peut-être pas le cas dans beaucoup de projets dbt.

L'autre motivation, c'est le coût. Une pipeline CI dbt classique, ça fait un `dbt build` complet. Ça veut dire du compute Snowflake, des warehouses qui tournent, des credits qui brûlent. Chaque PR, chaque push. Quand t'as une stack data de taille respectable, ça chiffre vite. Et pour quoi ? Pour valider un renommage de colonne qui aurait pu être détecté sans jamais toucher à Snowflake.

## L'idée

Je voulais quelque chose de simple : un outil qui compare deux versions d'un manifest dbt (la branche base vs la PR) et qui te dit "attention, tu as retiré/renommé la colonne X dans le modèle Y, et les modèles Z1, Z2, Z3 en dépendent."

Pas de connexion à la base de données requise. Zéro compute Snowflake. 100% offline. Juste de l'analyse statique sur les fichiers [manifest.json](https://docs.getdbt.com/reference/artifacts/manifest-json) que dbt génère déjà. Tu roules [`dbt parse`](https://docs.getdbt.com/reference/commands/parse) (qui est quasi instantané et ne touche pas à ta base), tu compares les manifests, et c'est réglé.

## Comment ça marche

Le principe est assez direct :

1. Tu lui donnes deux dossiers : un avec le manifest de la branche principale, l'autre avec celui de ta PR
2. Il parse les deux manifests et extrait les colonnes de chaque modèle
3. Il compare et identifie les changements : colonnes supprimées, renommées, changements de type
4. Pour les changements cassants, il traverse le DAG en BFS (breadth-first search) pour trouver tous les modèles en aval impactés
5. En option, il trace le lignage au niveau des colonnes pour éliminer les faux positifs : si un modèle downstream ne référence pas la colonne modifiée, pas d'alerte

La distinction entre cassant et non-cassant est simple :
- **Cassant** : colonne supprimée, renommée, changement de type
- **Non-cassant** : nouvelle colonne, nouveau modèle

```bash
dbt-guard diff --base target/base --current target/current --dialect snowflake
```

Et ça sort un rapport en texte, JSON, ou en annotations GitHub Actions. Ce dernier format est pratique, les alertes apparaissent directement sur les lignes de code dans la PR.

## Le lignage au niveau des colonnes

C'est la feature qui m'a le plus donné de fil à retordre, et c'est aussi celle qui rend l'outil vraiment utile.

Sans lignage de colonnes, si tu modifies une colonne dans un modèle upstream, tous les modèles downstream sont flaggés comme potentiellement impactés. Avec des centaines de modèles, ça génère beaucoup de bruit.

Avec le lignage, dbt-guard trace quelles colonnes downstream référencent réellement la colonne modifiée. Si un modèle downstream fait un `SELECT col_a, col_b` et que tu as modifié `col_c`, pas d'alerte. C'est [SQLGlot](https://github.com/tobymao/sqlglot) qui fait cette magie en parsant le SQL et en construisant l'arbre de dépendances.

Évidemment, ça a ses limites. `SELECT *` est le cas classique qui complique les choses. Quand un modèle fait un `SELECT *`, on ne peut pas savoir statiquement quelles colonnes il consomme réellement. Et certains patterns SQL complexes peuvent tromper le parser. Mais pour la majorité des cas, ça fonctionne et ça réduit significativement le bruit.

## Publier sur PyPI

Premier package Python publié. Le processus en soi n'est pas sorcier : `pyproject.toml` bien configuré, quelques métadonnées, et `pip install build && python -m build && twine upload dist/*`. Mais c'est quand même satisfaisant de voir son package apparaître sur pypi.org et de pouvoir faire `pip install dbt-guard`.

J'ai mis en place les standards de base : pytest pour les tests, mypy pour le typage, ruff pour le linting, un seuil de couverture à 80%. Les tests utilisent des fixtures synthétiques plutôt que des bases de données réelles, cohérent avec la philosophie "pas de connexion requise" de l'outil.

## Ce que j'ai appris

Quelques leçons en vrac :

**La dégradation gracieuse, c'est important.** SQLGlot ne peut pas parser tous les patterns SQL imaginables. Plutôt que de crasher, dbt-guard retombe sur les colonnes documentées dans le manifest quand le parsing SQL échoue. C'est pas parfait, mais c'est mieux qu'une erreur bloquante.

**Garder les dépendances minimales.** Chaque dépendance que tu ajoutes est une source potentielle de conflits de version dans l'environnement de quelqu'un d'autre. Avec juste SQLGlot et Click, les chances de conflits sont faibles.

**Les fixtures synthétiques pour les tests.** Pas besoin d'une vraie base Snowflake pour tester un outil d'analyse statique. Des manifests JSON fabriqués à la main font l'affaire et les tests roulent en quelques secondes.

## Pourquoi la documentation comme prérequis n'est pas un problème

Un prérequis de dbt-guard, c'est que tes colonnes soient documentées dans les fichiers YAML. Pour quelqu'un qui vient du développement logiciel, ça semble évident : la documentation fait partie du contrat d'une API, pas d'un nice-to-have. Mais en data engineering, c'est loin d'être la norme.

Mon point de vue : avec les LLMs, il n'y a plus d'excuse. Documenter des centaines de colonnes manuellement était pénible. Aujourd'hui, un agent peut lire ton SQL, inférer le contexte métier et générer un premier jet de documentation YAML en quelques secondes. J'en parle plus en détail dans [l'article sur la documentation dbt comme gouvernance]({{< ref "/blog/dbt-documentation-gouvernance-yml/" >}}).

Et puis dans un CI, l'enforcement est automatique via [`dbt_meta_testing`](https://github.com/tnightengale/dbt-meta-testing) : si une PR ajoute un modèle ou une colonne sans description, elle ne passe pas. La documentation n'est pas optionnelle, elle est vérifiée au même titre que les tests.

## Un projet dbt, c'est de l'infrastructure testable offline

Ce qui m'a convaincu de l'approche statique, c'est une analogie simple : un projet dbt est une représentation de l'infrastructure de ta stack data. Le DAG, les colonnes, les dépendances, tout ça vit dans des fichiers JSON et YAML. Comme du Terraform pour ta donnée.

Et l'infrastructure, on peut la valider offline. `terraform plan` ne touche pas à ton cloud. [`dbt parse`](https://docs.getdbt.com/reference/commands/parse) ne touche pas à Snowflake. Le manifest qui en résulte est une description complète de ce que le projet est censé faire.

Tester ce manifest, c'est tester l'infrastructure avant de la déployer. Ça ne remplace pas les tests sur les données réelles (les contraintes d'unicité, les valeurs nulles, la fraîcheur), mais ça permet d'être strict sur la structure des pipelines sans aucun coût de compute. C'est un filtre en amont, rapide et gratuit, avant même de toucher à la base.

## L'argument du coût

Je reviens là-dessus parce que c'est un point sous-estimé. Une CI dbt "standard" qui fait `dbt build --target ci`, c'est :
- Un warehouse Snowflake qui se réveille
- Des modèles qui se matérialisent (même en mode CI)
- Des tests qui roulent sur des données réelles
- Des credits qui partent à chaque PR

Avec dbt-guard, la détection de breaking changes coûte exactement zéro credit Snowflake. Ça tourne sur le runner CI lui-même, en quelques secondes, avec des fichiers JSON locaux. Ça ne remplace pas ta CI dbt (tu en as toujours besoin pour valider la logique), mais ça attrape une catégorie entière de problèmes avant même de toucher à ta base. C'est un filtre rapide et gratuit en amont.

## Et la suite ?

L'outil fait ce qu'il doit faire. C'est un garde-fou qui s'intègre dans ta CI, et qui bloque (ou avertit) quand une PR risque de casser des consommateurs en aval. Pas plus, pas moins.

C'est mon premier package Python publié. C'est pas révolutionnaire, c'est pas un framework qui va changer le monde. C'est un outil qui résout un problème spécifique que j'avais, que personne d'autre n'avait encore résolu en open source, et qui pourrait être utile à d'autres.

Le code est sur [GitHub](https://github.com/damione1/dbt-guard), sous licence Apache 2.0. Si vous travaillez avec dbt et que vous avez déjà eu le plaisir de découvrir un breaking change en production un vendredi soir, ça vaut peut-être le coup d'y jeter un oeil.
