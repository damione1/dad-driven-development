---
title: "dbt : Les tests dans les YAML, ou comment arrêter de prier pour que les données soient correctes"
date: 2026-02-27
draft: false
description: "Les tests déclaratifs dans dbt - des quatre tests natifs aux tests avancés, avec des exemples concrets de problèmes attrapés avant la prod."
tags: ["dbt", "Data Quality", "Data Engineering", "Testing"]
categories: ["Data Engineering", "Data Quality"]
---

Tu connais cette sensation : un rapport qui sort des chiffres bizarres, un analyste qui te dit "les totaux matchent pas", et tu passes ta journée à remonter la chaîne pour trouver où les données ont dérapé. Souvent, le problème aurait pu être détecté automatiquement si quelqu'un avait mis un test quelque part.

## Les tests déclaratifs dans dbt

dbt a un système de tests intégré directement dans les YAML de documentation. C'est le même fichier qui documente tes colonnes et qui déclare tes tests. L'idée est simple : tu décris tes attentes sur les données, et dbt les vérifie à chaque exécution.

Les [quatre tests natifs](https://docs.getdbt.com/docs/build/data-tests) :
- **not_null** : cette colonne ne devrait jamais être vide
- **unique** : pas de doublons sur cette colonne
- **accepted_values** : les seules valeurs possibles sont cette liste
- **relationships** : cette colonne référence une autre table (intégrité référentielle)

C'est déclaratif. Tu n'écris pas de SQL de test, tu déclares des contraintes.

## Au-delà des tests de base

Les quatre tests de base couvrent une bonne partie des besoins, mais pas tout. Pour le reste, il y a les packages de tests et les tests custom.

**Les tests de combinaison.** "Cette combinaison de colonnes doit être unique." Par exemple, une commande ne devrait apparaître qu'une seule fois par date et par client. C'est pas un simple `unique` sur une colonne, c'est une contrainte composite. [`dbt-utils`](https://github.com/dbt-labs/dbt-utils) fournit `unique_combination_of_columns` pour ça.

**Les tests de distribution.** "Cette colonne ne devrait pas avoir plus de X% de valeurs nulles." Utile pour les colonnes qui *peuvent* être nulles mais ne devraient pas l'être *trop souvent*.

**Les tests de fraîcheur.** "La donnée la plus récente dans cette source ne devrait pas avoir plus de 24 heures." Techniquement c'est un mécanisme séparé dans dbt ([`dbt source freshness`](https://docs.getdbt.com/docs/build/sources#source-data-freshness)), mais ça se déclare au même endroit dans les YAML. Si ta source arrête d'envoyer des données et que personne ne le remarque pendant une semaine, t'as un problème.

**Les tests de cohérence.** "Le sous-total + taxes + livraison devrait être égal au total de commande." C'est le genre de test qui attrape les bugs d'arrondi et les incohérences de calcul avant qu'un client ou fournisseur te le signale.

## Tests sur les sources : la première ligne de défense

Un pattern que j'apprécie particulièrement : tester les données source, pas juste les modèles transformés.

Quand tes données arrivent dans Snowflake via un outil de réplication (genre Fivetran, Airbyte), tu n'as aucune garantie sur leur qualité. Le système source peut avoir des bugs. La réplication peut avoir des problèmes. Les types peuvent changer sans prévenir.

En mettant des tests directement sur les définitions de sources dans dbt, tu crées une première ligne de défense :
- Est-ce que les colonnes que tu attends sont toujours là ?
- Est-ce que les types sont corrects ?
- Est-ce que les IDs sont bien uniques ?
- Est-ce qu'il y a des données récentes ?

Quand un test source fail, ça te dit "le problème vient d'en amont, pas de ta transformation." C'est de l'information précieuse pour le debug.

En pratique, c'est souvent le meilleur moyen de découvrir qu'un collègue du côté dev a fait un changement sur l'un de ses services sans passer par la case "prévenir l'équipe data". Une colonne renommée, un nouveau statut ajouté, un type qui change silencieusement en prod. Sans tests sur les sources, tu le découvres quand un dashboard est cassé. Avec, tu catches rapidement pourquoi la dynamic table et tout le lineage fail. Tu roules tes tests, ça te donne une première piste, et tu peux aller poser la bonne question à la bonne équipe avant de partir débugger dans le mauvais sens.

## La CI comme filet de sécurité

Les tests ne servent à rien si personne ne les roule. La pipeline CI est là pour ça.

Chaque PR qui touche aux modèles dbt déclenche un cycle complet :
1. Build des modèles en environnement CI
2. Exécution de tous les tests
3. Validation de la documentation (complétude)
4. Si tout passe, la PR peut être mergée

Le point clé : la CI fail si un test fail. Pas de warning ignoré, pas de "on corrigera plus tard." Si tes données ne passent pas les contraintes que tu as déclarées, le code ne va pas en production.

### CI sur Snowflake : quelques réglages pour ne pas saigner des crédits

Si ta CI build une stack éphémère complète à chaque PR, il y a quelques réglages qui font une vraie différence sur la facture.

**Dimensionner les warehouses selon le volume de CI.** Un XS suffit pour la plupart des builds CI, inutile de sur-dimensionner. Le vrai paramètre à ajuster, c'est combien de runs parallèles tu peux avoir en simultané sur une journée chargée.

**Utiliser des databases transient pour la CI.** Une [database transient dans Snowflake](https://docs.snowflake.com/en/user-guide/tables-temp-transient) ne conserve pas de Fail-Safe (la rétention de données en cas de corruption ou suppression accidentelle qui est activée par défaut sur les tables standard). Pour de la donnée CI qui est de toute façon recréée à chaque run, payer pour le Fail-Safe n'a aucun sens. Déclarer la database cible de la CI comme transient coupe ce coût sans aucun impact fonctionnel.

**Nettoyer proprement à la fin.** Le step de cleanup de ta CI doit dropper la database entière, pas juste les tables créées pendant le run. Un pipeline qui plante à mi-chemin sans cleanup laisse des objets orphelins qui tournent, notamment les Dynamic Tables, qui continuent à se rafraîchir et à consommer des crédits jusqu'à ce que quelqu'un les supprime manuellement. Vérifier de temps en temps que des vieilles CI databases ne traînent pas est une bonne hygiène.

**Overrider le lag des Dynamic Tables en CI.** Par défaut, une Dynamic Table déployée en CI va essayer de se rafraîchir selon son lag cible : toutes les heures, toutes les 5 minutes, selon ce qui est déclaré en prod. En CI, tu veux exactement le contraire : qu'elles ne se rafraîchissent jamais toutes seules. La solution est d'overrider le `target_lag` à une valeur longue (genre `8760 hours`, soit un an) dans ton profil CI. La table est créée, les tests tournent sur le contenu initial, et aucun refresh automatique ne vient perturber ou prolonger l'exécution.

**Utiliser `--defer` avec un manifest de prod.** C'est probablement l'optimisation la plus impactante. dbt a une option [`--defer`](https://docs.getdbt.com/reference/node-selection/defer) qui, combinée avec le manifest de la branche principale, permet de ne builder que les modèles modifiés dans la PR. Pour les modèles non touchés, dbt les "proxie" vers la version prod existante plutôt que de les recréer from scratch. Une PR qui modifie 3 modèles dans un DAG de 200 ne build que ces 3 modèles et leurs dépendants directs, pas la stack entière. Le gain en temps et en crédits est considérable sur les projets de taille respectable.

## Les tests comme documentation vivante

Ce qui est élégant avec les tests déclaratifs dans les YAML, c'est qu'ils servent aussi de documentation. Quand tu vois :

```yaml
- name: status
  description: "Statut de la commande"
  tests:
    - not_null
    - accepted_values:
        values: ['PENDING', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED']
```

Tu sais immédiatement trois choses :
1. Ce que la colonne contient (description)
2. Qu'elle ne peut pas être vide (not_null)
3. Quelles sont ses valeurs possibles (accepted_values)

C'est de la documentation qui se vérifie automatiquement. Quand un nouveau statut de commande apparaît dans les données, le test fail, la documentation est mise à jour, et tout le monde est au courant.

## Les erreurs qui m'ont convaincu

Quelques exemples concrets de problèmes attrapés par des tests :

**Le type boolean fantôme.** Une colonne qui devrait être boolean mais qui contient des `NULL` en plus de `true/false`. Le code source traite `NULL` comme `false`, mais ta transformation dbt ne fait pas forcément pareil. Un test `accepted_values: [true, false]` combiné à `not_null` clarifie l'intention.

**L'ID en double.** Un système source qui, suite à un bug de migration, a dupliqué quelques milliers d'enregistrements. Sans test `unique`, ces doublons se propagent silencieusement dans toute la chaîne de transformation.

## Ce que j'en retiens

L'approche qui a marché pour moi part d'une hiérarchie simple.

**En premier : les contraintes YAML.** `not_null`, `unique`, `accepted_values`, `relationships`. Ces tests ne sont pas séparables de la documentation : ils *sont* la documentation. Déclarer qu'une colonne `status` accepte `['PENDING', 'SHIPPED', 'DELIVERED']`, c'est à la fois documenter le contrat et le vérifier à chaque run. Le coût est quasi nul (une ligne de YAML), et ça donne une couverture de base sur toutes les colonnes sans effort particulier. C'est le minimum non négociable.

**Ensuite seulement : les tests complexes.** Tests de cohérence, de distribution, de mapping : ceux qui nécessitent du SQL custom ou des packages comme `dbt-utils`. Ceux-là sont précieux, mais ils ont un coût : il faut les relire.

C'est là où j'ai appris à ma dépens : déléguer la génération de tests à un LLM sans passer par une relecture sérieuse, c'est se retrouver avec de la fausse couverture. Des tests qui s'exécutent, qui passent, et qui ne testent pas vraiment ce qu'ils prétendent tester. C'est pire que pas de tests du tout, parce que ça donne une confiance non méritée. J'ai relu des tests générés par LLM que je n'avais pas vérifiés au moment du merge, et certains étaient tout simplement à côté de la plaque : logique inversée, mauvaise table référencée, seuil arbitraire sans sens métier. Ça passe mais ça ne sert à rien.

La règle que j'applique maintenant : les contraintes YAML, toujours, systématiquement, vérifiées en CI. Les tests complexes, seulement quand j'ai le temps de les relire ligne par ligne avant de les merger. Une couverture de 30% de tests bien compris vaut mieux qu'une couverture de 80% de tests dont personne ne sait vraiment ce qu'ils vérifient.
