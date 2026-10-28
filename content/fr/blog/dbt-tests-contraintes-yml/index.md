---
title: "Tests dbt en YAML : arrêter de prier pour ses données"
translationKey: "dbt-tests-constraints"
date: 2026-02-27
lastmod: 2026-10-28
draft: false
description: "Mes tests dbt en YAML pour repérer les nulls et les clés en double, puis vérifier les données en CI sans gaspiller les crédits Snowflake."
tags: ["dbt", "Data Quality", "Data Engineering", "Testing"]
categories: ["Data Engineering", "Data Quality"]
images: ["dbt-tests-constraints-yml-featured.jpg"]
---

Un rapport sort des chiffres bizarres. Un analyste te dit « les totaux matchent pas », et tu passes ta journée à remonter la chaîne. Le réflexe auquel je reviens est assez banal : **des tests de données dbt, déclarés en YAML à côté des colonnes, et vraiment exécutés en CI**. Dire au pipeline ce qu’on attend avant qu’un dashboard nous dise ce qui a dérapé.

## Ce que les tests de données dbt vérifient vraiment

Les [data tests de dbt](https://docs.getdbt.com/docs/build/data-tests) sont des vérifications SQL sur les données. Les tests génériques permettent de les déclarer dans les mêmes fichiers YAML que la documentation des modèles et des sources. dbt génère le SQL du test ; tu l’exécutes avec `dbt test`, ou tu construis et testes les ressources sélectionnées avec `dbt build`. `dbt run` seul ne lance pas les tests.

Les quatre tests génériques natifs sont un bon point de départ :

- **`not_null`** : aucun `NULL` SQL dans cette colonne. Une string vide est une autre valeur.
- **`unique`** : aucun doublon parmi les valeurs non nulles. À combiner avec `not_null` pour un identifiant obligatoire.
- **`accepted_values`** : les valeurs non nulles appartiennent à la liste déclarée.
- **`relationships`** : les valeurs non nulles existent dans la relation et la colonne référencées.

Le [traitement des nulls](https://docs.getdbt.com/reference/resource-properties/data-tests) compte. `accepted_values` et `relationships` ne rejettent pas les `NULL` ; il faut ajouter `not_null` si l’absence est invalide.

Ce sont des assertions sur les données que tu interroges. Pas des contraintes de base de données qui empêchent l’insertion d’une mauvaise ligne. Un test vert dit que cette assertion était vraie sur les données vérifiées, à ce moment-là. Il ne prouve pas que le modèle entier est correct.

## Un exemple YAML que je peux relire au merge

Sur un modèle `orders`, je veux un identifiant présent et unique, et un statut à la fois présent et tiré d’un ensemble connu :

```yaml
# models/orders.yml
version: 2

models:
  - name: orders
    description: "Une ligne par commande"
    columns:
      - name: order_id
        description: "Identifiant de la commande"
        data_tests:
          - not_null
          - unique
      - name: status
        description: "Statut de la commande"
        data_tests:
          - not_null
          - accepted_values:
              arguments:
                values: ['PENDING', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED']
```

La syntaxe `arguments:` est disponible à partir de dbt 1.10.5 ; les anciennes versions placent les arguments directement sous le nom du test. `tests:` reste un alias de `data_tests:`, mais j’utilise le nom explicite ici.

Pour vérifier le modèle existant, ou le reconstruire et le vérifier :

```bash
dbt test --select orders
dbt build --select orders
```

C’est aussi pour ça que j’aime garder descriptions et tests ensemble. Tu peux lire le contrat attendu sans fouiller dans une suite de tests à part. Je pousse cette idée dans [l’article sur la documentation dbt et la gouvernance]({{< ref "/blog/dbt-documentation-gouvernance-yml/" >}}).

Quand un nouveau statut de commande arrive, le test nous donne quelque chose de concret à discuter. La source a-t-elle changé légitimement ? Est-ce qu’on a oublié un cas ? Est-ce que la donnée est fausse ? Mettre le YAML à jour vient après cette discussion. Ce n’est pas le remède automatique à un test qui fail.

## Quand l’unicité porte sur plusieurs colonnes

Un snapshot quotidien des commandes peut répéter un `order_id` sur plusieurs dates. Ce qui doit rester unique, c’est le couple `order_id` et `snapshot_date`. Tester chaque colonne seule décrirait le mauvais grain.

[`dbt_utils.unique_combination_of_columns`](https://github.com/dbt-labs/dbt-utils#unique_combination_of_columns-source) fait cette vérification, une fois un package `dbt-utils` compatible installé. L’exemple suit son format d’arguments documenté pour dbt Core 1.10.6 et plus récent :

```yaml
models:
  - name: order_daily_snapshots
    data_tests:
      - dbt_utils.unique_combination_of_columns:
          arguments:
            combination_of_columns:
              - order_id
              - snapshot_date
    columns:
      - name: order_id
        data_tests:
          - not_null
      - name: snapshot_date
        data_tests:
          - not_null
```

Le test de combinaison cherche les doublons ; il ne rend pas chaque composante obligatoire. Je garde `not_null` sur les deux, parce qu’un snapshot sans commande ou sans date ne respecte pas le contrat de ce modèle. Ce sont des tests de données, pas une déclaration de clé primaire composite dans Snowflake.

Au-delà, il reste des règles que les quatre tests natifs ne savent pas exprimer. « Cette colonne facultative ne devrait pas être presque toujours nulle. » « Le sous-total plus les taxes plus la livraison doit donner le total de commande. » Les packages et les tests SQL custom peuvent couvrir ça. Le vrai travail, c’est de choisir une règle qui veut dire quelque chose, puis de lire le SQL censé la faire respecter.

## Les tests sur les sources : une première piste

Un pattern que j’apprécie particulièrement : tester les données source, pas juste les modèles transformés.

Quand les données arrivent dans Snowflake via Fivetran, Airbyte ou un autre outil de réplication, je ne pars pas du principe qu’elles sont propres. Un test `unique` sur un identifiant source ou `accepted_values` sur un statut me donne un endroit où regarder plus tôt quand ça casse en aval.

C’est souvent comme ça que je découvre un changement côté dev qui n’est pas passé par la case « prévenir l’équipe data ». Une colonne renommée, un nouveau statut, un type qui a changé en prod. Un test source peut fail ou partir en erreur avant que je passe l’après-midi à débugger la mauvaise transformation. Ça me donne une piste ; ça ne prouve pas quel service est responsable.

Les quatre tests génériques ne vérifient pas le schéma complet d’une source ni le type de chaque colonne. Pour ces garanties, il me faut des contrôles de schéma explicites. Lister les colonnes dans un YAML ne suffit pas.

La fraîcheur est encore une autre question. La [configuration de freshness](https://docs.getdbt.com/reference/resource-configs/freshness) compare un timestamp configuré ou les métadonnées d’un entrepôt supporté aux seuils déclarés. `dbt source freshness` vérifie les sources ; dbt v2 propose aussi `dbt freshness` pour les sources et modèles configurés. `dbt build` n’inclut pas ces contrôles. Un timestamp d’ingestion récent est utile, mais il ne prouve pas à lui seul que les événements métier sont récents.

## La CI transforme les tests en condition de merge

Les tests ne servent à rien si personne ne les roule. Ma règle, c’est que le job CI fasse les vérifications et que son résultat décide si la PR peut être mergée.

Le cycle que je veux est assez simple :

1. Vérifier la fraîcheur des sources quand le changement en dépend.
2. Construire et tester les ressources sélectionnées dans une cible CI isolée.
3. Valider la complétude de la documentation.
4. Bloquer le merge si un contrôle obligatoire échoue.

[`dbt build`](https://docs.getdbt.com/reference/commands/build) exécute les modèles et leurs tests dans l’ordre des dépendances. Pour les tests de données qui doivent bloquer le merge, je garde une sévérité d’erreur et je relis les [seuils ou réglages de warning](https://docs.getdbt.com/reference/resource-configs/severity). Transformer un échec gênant en warning, ce n’est pas corriger les données.

### Sélectionner les changements, puis utiliser `--defer`

Sur un gros DAG, tout reconstruire à chaque PR finit par coûter cher. Voici une commande CI illustrative, avec un manifest de production sauvegardé dans `artifacts/prod/` :

```bash
dbt build --target ci --select "state:modified+" \
  --state artifacts/prod --defer
```

Chaque morceau a son rôle. `state:modified` compare au manifest sauvegardé. Le [`+`](https://docs.getdbt.com/reference/node-selection/graph-operators) à la fin ajoute les descendants en aval, pas seulement le modèle suivant. [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) résout les dépendances non sélectionnées qui s’y prêtent vers l’environnement sauvegardé, plutôt que de t’obliger à les reconstruire.

`--defer` ne sélectionne pas les modèles modifiés. Par défaut, une relation non sélectionnée déjà présente dans la cible CI peut encore être utilisée ; `--favor-state` change cette préférence. Un test avec plusieurs parents peut aussi lire à la fois en CI et en prod, donc je vérifie ce qu’il va vraiment comparer. Un manifest sauvegardé ne garantit pas un run petit ou autonome.

Je garde cet état sauvegardé hors du `target/` du run en cours. Le [DAG et les cibles décrits dans mon article sur dbt comme infrastructure]({{< ref "/blog/dbt-data-infrastructure-as-code/" >}}) rendent cette séparation plus facile à tenir.

### Garder les ressources Snowflake de CI sous contrôle

Quelques points méritent une règle explicite pour la CI.

**Dimensionner le warehouse selon le travail.** Partir petit, puis mesurer le build et sa concurrence. Un XS n’est pas une réponse universelle, surtout quand plusieurs PR tournent ensemble. Les tests exécutent toujours du SQL ; une déclaration YAML courte ne fait pas disparaître son coût d’entrepôt.

**Utiliser du stockage transient quand les données sont jetables.** Les [databases et tables transient](https://docs.snowflake.com/en/user-guide/tables-temp-transient) évitent le stockage Fail-safe, mais offrent aussi moins de protection pour la récupération. Ça convient à des données CI que je peux reconstruire. Pas à une optimisation gratuite sur des données que je dois garder.

**Contrôler explicitement les refresh des Dynamic Tables.** Le [target lag](https://docs.snowflake.com/en/user-guide/dynamic-tables/target-lag) est un objectif de fraîcheur, pas un intervalle de refresh ni un interrupteur. Un lag très long ne garantit pas l’absence de refresh automatique. Si un contrôle CI a besoin d’un contenu fixe, il faut initialiser les tables CI, attendre la fin de ce chargement, puis [suspendre explicitement les refresh planifiés](https://docs.snowflake.com/en/user-guide/dynamic-tables/manage#suspend-a-dynamic-table) avant de les vérifier. La suspension touche aussi les Dynamic Tables en aval ; la séquence CI doit tenir compte du graphe.

**Nettoyer même quand le job plante.** Supprimer la database CI dédiée dans un step de cleanup, et vérifier les databases abandonnées. Un pipeline en erreur ne devrait pas laisser ses Dynamic Tables se rafraîchir jusqu’à ce que quelqu’un remarque la facture.

## Les erreurs qui m’ont convaincu

**Le null de trop.** Une colonne boolean contenait `NULL` en plus de `true` et `false`. C’est autorisé par un type SQL nullable, mais ce n’était pas le contrat qu’on voulait. Le code source traitait l’absence comme faux ; la transformation ne faisait pas forcément pareil. Sur un booléen obligatoire, `not_null` rend cette exigence explicite. Si j’utilise aussi `accepted_values` pour des booléens, je mets `quote: false` pour ne pas les traiter comme des strings.

**L’identifiant en double.** Un bug de migration a dupliqué des enregistrements dans un système source. Sans test d’unicité, les doublons pouvaient traverser les transformations sans que rien ne crie. Une petite vérification sur l’identifiant était plus utile qu’un autre après-midi à fixer le rapport final.

## Ce que j’en retiens

Je commence toujours par les vérifications banales. Un identifiant obligatoire reçoit `not_null` et `unique`. Un ensemble fermé de statuts reçoit `accepted_values`, avec `not_null` si l’absence est invalide. Une relation est vérifiée là où le modèle s’appuie dessus. Choisir les tests qui décrivent le vrai rôle de la colonne, puis les exécuter.

Les tests plus complexes viennent après, quand je peux lire ce qu’ils font. C’est là où j’ai appris à mes dépens : déléguer leur génération à un LLM sans vraie relecture peut produire de la fausse confiance. J’ai relu des tests générés que j’avais laissés passer au merge et trouvé des logiques inversées, la mauvaise table référencée, des seuils sans sens métier. Ils tournaient. Ils passaient. Ils ne testaient pas ce que je pensais qu’ils testaient.

La règle que j’applique maintenant est simple : des vérifications déclaratives que je comprends, imposées en CI. Des tests custom quand j’ai le temps de les lire ligne par ligne. Je préfère moins d’assertions auxquelles je fais confiance qu’une longue liste de checks verts que personne ne sait expliquer.
