---
title: "Documenter une base de données source avec des LLM multi-agents"
date: 2026-02-06
draft: false
description: "Une approche multi-agent pour générer automatiquement la documentation YAML des sources dbt - ce que les LLM découvrent, leurs limites, et le workflow complet."
tags: ["dbt", "LLM", "Data Engineering", "AI"]
categories: ["Data Engineering", "AI"]
---

Documenter les colonnes d'une base source, c'est le genre de tâche que personne ne veut faire. T'as un système opérationnel avec des centaines de tables, des milliers de colonnes, et une documentation qui va de "inexistante" à "un commentaire de 2017 qui dit `TODO: document this`."

## Le contexte

Quand tu travailles avec dbt et que tu [définis tes sources](https://docs.getdbt.com/docs/build/sources), tu veux idéalement documenter chaque colonne. Pas juste son nom et son type, mais ce qu'elle représente réellement, ses particularités, ses valeurs possibles, ses relations avec d'autres tables.

C'est la couche bronze (les données brutes telles qu'elles arrivent des systèmes sources) qui est la plus difficile à documenter. Contrairement aux couches silver et gold, où la transformation elle-même est une forme de documentation (le SQL dit ce que la donnée est censée être), la couche bronze hérite des conventions, des bugs et des décisions de design du système qui l'alimente. La connaissance ne vit pas dans dbt, elle vit dans la codebase applicative.

C'est là aussi que tout repose. Si tu ne sais pas ce que signifie une colonne en bronze, tu ne peux pas documenter correctement sa transformation en silver, ni la métrique business qu'elle alimente en gold. La documentation se construit de bas en haut, et le bas, c'est le plus dur.

Le problème, c'est que cette connaissance est souvent dispersée. Elle est dans le code applicatif qui écrit dans ces tables. Elle est dans la tête des développeurs backend. Elle est parfois dans un wiki que personne n'a mis à jour depuis 2019.

Et personne n'a envie de passer 3 semaines à éplucher du code legacy pour comprendre ce que `legacy_field_42` veut dire.

## L'idée : des agents LLM spécialisés

L'approche que j'ai expérimentée, c'est d'utiliser des agents LLM pour faire le gros du travail d'investigation. Pas un seul prompt géant qui essaie de tout comprendre d'un coup, mais une approche multi-agent où chaque agent a un rôle spécifique.

Le principe :
1. **Agent explorateur** : parcourt le schéma de la base source, identifie les tables et les colonnes, note les types, les FK apparentes, les patterns de nommage
2. **Agent analyste de code** : prend le code applicatif qui interagit avec chaque table et analyse comment chaque colonne est utilisée : en lecture, en écriture, les validations appliquées, les transformations
3. **Agent documentaliste** : synthétise les informations des deux premiers agents et produit une documentation structurée au format YAML de dbt

Chaque agent travaille table par table, colonne par colonne. C'est méthodique et systématique.

## Ce que les agents découvrent

Le plus intéressant, c'est ce que les agents trouvent que personne ne savait (ou avait oublié) :

**Les colonnes détournées.** Une colonne `notes` qui en théorie contient du texte libre, mais qui en pratique stocke du JSON sérialisé avec une structure spécifique que le frontend parse.

**Les valeurs magiques.** Un `status` qui vaut 0, 1, 2, 3, 4. Mais personne ne sait que 3 veut dire "en attente de validation manuelle" et 4 c'est "annulé automatiquement par le système." L'agent qui analyse le code trouve les constantes et les conditions.

**Les contraintes implicites.** Une colonne qui n'a pas de contrainte NOT NULL en base, mais que le code applicatif ne laisse jamais vide. Ou une colonne qui devrait être unique mais qui a des doublons à cause d'un bug corrigé il y a 3 ans.

**Les données sérialisées.** Du JSON, du XML, des formats propriétaires dans un champ texte. L'agent identifie le format et documente la structure interne.

**Les relations non documentées.** Des FK qui n'existent pas en base mais que le code utilise systématiquement. Des colonnes qui référencent d'autres tables via une convention de nommage que personne n'a formalisée.

## L'intégration dans dbt

Une fois les YAML générés et validés, ils s'intègrent directement dans le projet dbt comme définitions de sources. Avec `persist_docs` activé, les descriptions remontent dans Snowflake et les métadonnées de classification alimentent les politiques de gouvernance. Ce mécanisme est couvert en détail dans [l'article sur les YAML comme gouvernance]({{< ref "/blog/dbt-documentation-gouvernance-yml/" >}}).

Ce qui compte ici : les agents transforment un exercice de documentation fastidieux en base concrète pour une gouvernance active, sans que ça soit un projet séparé.

## Du bronze aux couches supérieures

Une fois la couche bronze documentée, quelque chose change dans la façon dont on documente le reste.

En silver, chaque modèle dbt est une transformation explicite depuis des sources connues. Le SQL lui-même dit beaucoup : une colonne `total_amount` calculée par `unit_price * quantity` n'a pas besoin d'une longue description. Ce qui compte, c'est de documenter les décisions de nettoyage, les règles de déduplication, les cas limites. Et ça, un LLM peut l'inférer en lisant le SQL et la documentation bronze en parallèle.

En gold, les modèles sont souvent des agrégations business. Les colonnes correspondent à des métriques dont le sens est dans la logique métier, pas dans le code. C'est là que la documentation devient plus manuelle, mais au moins tu pars d'une base solide. Tu sais exactement ce que chaque champ upstream représente, ce qui rend la documentation des métriques dérivées beaucoup plus précise.

L'effet de levier est réel : la couche bronze est la plus longue à documenter et la plus difficile à automatiser partiellement. Les couches supérieures bénéficient directement de ce travail de fondation. Chaque colonne bronze correctement décrite se propage dans le lignage et réduit le travail de documentation des couches qui en dépendent.

C'est aussi ce qui rend la documentation bronze si rentable à faire en premier, malgré l'effort : c'est le seul endroit où la connaissance est enfouie dans une codebase externe, et donc le seul endroit où les agents LLM ont un vrai avantage sur un data engineer qui ne connaît pas ce code.

## Les limites

Soyons honnêtes sur ce qui marche moins bien :

**Le contexte métier.** Un LLM peut comprendre que `creation_date` est une date de création. Il ne peut pas savoir que dans votre contexte, cette date a une signification contractuelle précise qui affecte d'autres calculs en aval. Le contexte métier fin, ça reste humain.

**Le code legacy illisible.** Quand le code qui interagit avec une table est un fichier de 3000 lignes sans structure claire, même un LLM a du mal à en extraire une documentation cohérente.

**La validation.** Tout ce que produit un LLM doit être validé par quelqu'un qui connaît le domaine. Les agents font le gros du boulot, mais la validation, la correction et l'ajout de contexte métier restent essentiels et irremplaçables. Pis comme je le dis souvent, on est responsable de notre utilisation de l'IA, et ça inclut la validation de ce qu'elle produit.

## Le workflow complet

En pratique :

1. Tu donnes à tes agents le dump du schéma de la base source et le code applicatif
2. Les agents produisent des fichiers YAML documentés, table par table
3. Un humain review, corrige les erreurs, ajoute le contexte métier manquant
4. Les YAML corrigés deviennent les définitions de sources dans dbt
5. Les métadonnées sont poussées vers Snowflake via [`persist_docs`](https://docs.getdbt.com/reference/resource-configs/persist_docs)
6. Les classifications alimentent les politiques de gouvernance

Le temps total ? Pour une base de quelques centaines de tables : quelques jours d'agents + quelques jours de review humaine. Sans les agents, c'est des semaines, voire des mois, de travail manuel que personne ne veut faire.

Un dernier conseil pratique : quitte à faire ça, autant y aller franchement. Prompt de reverse engineering complet, mode deep thinking activé, review systématique table par table. Ça veut dire brûler quelques millions de tokens chez nos amis d'OpenAI, Anthropic ou Google, mais c'est un investissement ponctuel pour un actif qui dure. Lance ça de nuit. Le matin, t'as une première version de doc sur toute ta couche bronze, et tu n'as pas eu à te taper une seule ligne de `legacy_field_42` à la main.

## La leçon

La documentation de source, c'est un des meilleurs use cases pour les LLM dans le data engineering. C'est pas glamour, c'est pas du machine learning, c'est pas de la data science. C'est du travail de fond, fastidieux mais essentiel, que les LLM font bien parce que c'est systématique, que le contexte est dans le code, et que la sortie est structurée.

Et contrairement à d'autres applications de LLM, ici la validation est simple : un data engineer ou un développeur backend peut vérifier la documentation produite en quelques minutes par table. Les erreurs sont faciles à repérer et à corriger.

C'est pas magique. C'est juste un bon outil appliqué au bon problème.
