---
title: "Snowflake + Terraform : Arrêter de gérer son infra data en SQL"
date: 2025-12-05
draft: false
description: "Comment passer des scripts SQL et migrations à une infrastructure Snowflake gérée avec Terraform - rôles en couches, future grants, masquage et reproductibilité."
tags: ["Snowflake", "Terraform", "Data Engineering", "Infrastructure as Code"]
categories: ["Data Engineering", "Infrastructure"]
---

Il y a un moment, dans la vie d'un data engineer, où tu te retrouves avec un fichier SQL de 300 lignes qui crée des rôles, des grants, des warehouses, et tu te demandes comment t'en es arrivé là. C'est mon histoire.

## Snowflake : tout en SQL, pour le meilleur et pour le pire

Snowflake a réussi quelque chose que peu de plateformes accomplissent : abstraire complètement l'infrastructure physique tout en conservant un contrôle granulaire sur tout ce qui compte. Pas de serveurs à maintenir, pas de clusters à dimensionner. Juste des bases de données, des schémas, des warehouses, des rôles, et des commandes SQL pour tout piloter.

Snowsight, son interface graphique, est capable. Tu peux y créer des objets, gérer des accès, visualiser tes données. Mais comme tout outil qui existe à la fois en version UI et en version programmatique, la vraie puissance est côté code. L'UI te donne accès aux fonctionnalités. Le SQL te donne le contrôle. La manipulation programmatique te donne la reproductibilité.

C'est l'analogie classique avec les outils en ligne de commande : ils paraissent moins accessibles qu'une interface graphique, mais ils se scriptent, se versionnent, s'automatisent. Chaque action devient reproductible, auditable, intégrable dans un pipeline. Un `GRANT` exécuté dans Snowsight disparaît dans le néant dès que tu fermes l'onglet. Le même `GRANT` déclaré dans du code Terraform est tracé, reviewable, réversible.

Le problème, c'est que Snowflake est tellement accessible en SQL ad hoc qu'on finit par tout faire comme ça. Un GRANT par-ci, un nouveau rôle via la console, un warehouse créé en urgence un mardi soir. Chacun de ces gestes est anodin. Ensemble, ils deviennent une infrastructure impossible à auditer.

## L'ère des scripts SQL

Au départ, quand j'ai commencé à monter une infrastructure Snowflake, j'ai fait ce que tout le monde fait : des fichiers `.sql`. Un script pour créer les databases, un autre pour les schémas, un autre pour les rôles. Simple, direct, ça marche.

Sauf que ça marche jusqu'à un certain point.

Première semaine : 5 fichiers SQL, bien organisés. Premier mois : 15 fichiers, quelques `IF NOT EXISTS` pour l'idempotence. Troisième semaine : tu te rends compte que tu dois modifier un rôle et tu parcours 4 fichiers pour être sûr de pas oublier un grant quelque part. Et tu finis par te demander si le état actuel de Snowflake correspond vraiment à ce qui est dans tes scripts, ou si quelqu'un a fait un `GRANT` à la main dans la console un mardi soir.

## L'épisode des migrations

Ensuite j'ai essayé les migrations. Comme en développement applicatif : des fichiers numérotés, chacun avec un changement incrémental. `001_create_databases.sql`, `002_add_marketing_role.sql`, `003_fix_grant_on_silver.sql`...

Sur le papier, c'est mieux. T'as un historique. Tu peux retracer l'évolution. Mais en pratique :
- Si quelqu'un fait un changement manuel, les migrations et la réalité divergent silencieusement
- Faire un rollback d'un `REVOKE` ou d'un `DROP ROLE`, c'est pas comme un `ALTER TABLE`, les effets en cascade sont imprévisibles
- T'as aucune idée de l'état actuel sans exécuter un audit complet
- Et surtout, tu te retrouves avec 50 fichiers de migrations et plus personne sait ce que l'infrastructure est *censée* être, juste ce qu'elle est *devenue*

## L'évidence : Terraform

Et puis un jour, je me suis demandé s'il existait un provider Terraform pour Snowflake. La réponse : oui, et il est [maintenu par Snowflake directement](https://github.com/snowflakedb/terraform-provider-snowflake).

Ça m'a paru comme une évidence. Terraform, c'est exactement le bon outil pour ce problème :
- **Déclaratif** : tu décris l'état souhaité, pas les étapes pour y arriver
- **Plan avant apply** : `terraform plan` te montre exactement ce qui va changer avant que ça change
- **Source de vérité unique** : le code Terraform décrit l'état souhaité de l'infrastructure, pas une approximation, pas un historique de migrations
- **Historique via git** : chaque changement est un commit, reviewable, réversible
- **Idempotent** : tu peux l'exécuter 10 fois, ça donne le même résultat

En comparaison avec les scripts SQL : si quelqu'un fait un changement manuel dans la console Snowflake, le prochain `terraform plan` te le montre comme un drift. Tu vois la différence entre l'état réel et l'état souhaité. Avec des scripts SQL, tu ne vois rien.

(En vrai, Terraform a ses propres irritants : le state qui se corrompt, les `import` pénibles, les changements breaking du provider entre deux versions. Mais ces problèmes sont gérables, et le gain en visibilité compense largement.)

## Les rôles : penser en couches

Le concept qui m'a le plus aidé, c'est de structurer les rôles en trois niveaux. C'est une pratique recommandée par Snowflake dans leur [documentation sur le contrôle d'accès](https://docs.snowflake.com/en/user-guide/security-access-control-considerations), et documentée par plusieurs architectes Snowflake ([exemple sur le blog officiel](https://medium.com/snowflake/rbac-and-cloning-for-devops-950c800c594e)).

**Niveau 1 : Les rôles système Snowflake.** ACCOUNTADMIN, SYSADMIN, SECURITYADMIN. Tu ne les crées pas, ils existent déjà. Mais tu configures Terraform pour utiliser le bon rôle au bon endroit, SYSADMIN pour créer des databases et des warehouses, SECURITYADMIN pour les rôles et les grants. Principe du moindre privilège.

**Niveau 2 : Les rôles d'accès (access roles).** Ce sont des rôles techniques, granulaires, qui donnent accès à un schéma spécifique à un niveau spécifique. Genre `SILVER_RO` (lecture seule sur le schéma Silver), `GOLD_RW` (lecture-écriture sur Gold), `STAGING_FULL` (accès complet au schéma de staging). La convention de nommage est importante : en lisant le nom du rôle, tu sais exactement ce qu'il fait.

**Niveau 3 : Les rôles fonctionnels (functional roles).** Ce sont les rôles business, ceux que tu assignes aux humains. Un rôle `ANALYST`, un rôle `ENGINEER`, un rôle `REPORTING`. Chaque rôle fonctionnel agrège plusieurs rôles d'accès. Le rôle Analyst obtient `SILVER_RO` + `GOLD_RO`. L'engineer obtient un accès plus large.

Le flux est simple : **Utilisateur → Rôle fonctionnel → Rôles d'accès → Permissions sur les schémas.**

L'avantage de cette approche, c'est que quand un nouvel analyste arrive, tu l'assignes au rôle fonctionnel ANALYST et il a automatiquement accès à tout ce dont il a besoin. Pas de liste de grants à maintenir manuellement.

## Les grants en cascade

Un des aspects les plus satisfaisants de cette approche, c'est la cascade des grants. Dans Snowflake, tu peux créer une hiérarchie de rôles : un rôle peut "contenir" un autre rôle.

Concrètement, si tu as trois niveaux d'accès pour un schéma (lecture, écriture, création), tu structures ça en cascade :
- Le rôle Create hérite du rôle Read-Write
- Le rôle Read-Write hérite du rôle Read-Only
- Tu n'as besoin d'accorder les privilèges qu'une seule fois à chaque niveau

Quand tu assignes le rôle Create à quelqu'un, il obtient automatiquement les permissions d'écriture et de lecture par héritage. Pas de duplication de grants, pas de maintenance supplémentaire.

## Les future grants : anticiper l'avenir

Un pattern critique dans Snowflake : les [future grants](https://docs.snowflake.com/en/sql-reference/sql/grant-privilege.html). Quand tu crées un rôle avec `SELECT` sur un schéma, ça s'applique aux tables qui existent *maintenant*. Mais quand dbt crée un nouveau modèle demain ? Sans future grants, personne n'y a accès.

Terraform permet de déclarer des future grants : "tous les objets futurs dans ce schéma hériteront automatiquement de ces permissions." C'est le genre de détail qui fait la différence entre une infra qui fonctionne le jour 1 et une infra qui fonctionne encore 6 mois plus tard quand l'équipe data a ajouté 50 modèles.

## Les utilisateurs : configuration, pas code

L'ajout d'un utilisateur, dans cette approche, ne nécessite pas d'écrire du code Terraform. Tu remplis une configuration :

- Son identité (nom, email)
- Son équipe (marketing, finance, data engineering...)
- Est-ce un admin ?
- A-t-il besoin d'un sandbox de développement ?

À partir de cette configuration, Terraform calcule et génère automatiquement :
- Le compte utilisateur
- L'assignation au rôle fonctionnel de son équipe
- Le warehouse par défaut
- L'accès sandbox si applicable

Ajouter un nouveau membre, c'est modifier quelques lignes de configuration, faire un `terraform plan` pour vérifier, et `terraform apply`. Pas de SQL à écrire, pas de grants à chercher.

## La frontière Terraform / dbt

Un point important : où s'arrête Terraform et où commence dbt ?

Ma vision : **Terraform gère tout ce qui est au-dessus du schéma. dbt gère tout ce qui est en dessous.**

Terraform s'occupe de :
- Créer les databases et les schémas
- Créer les warehouses
- Gérer les rôles et les permissions
- Définir les politiques de masquage
- Configurer le monitoring

dbt s'occupe de :
- Créer les tables et les vues dans les schémas
- Transformer les données
- Appliquer les tags de gouvernance aux colonnes
- Documenter les modèles
- Tester la qualité des données

L'infrastructure (Terraform) évolue lentement, un nouveau schéma par mois, un nouveau rôle par trimestre. Les transformations (dbt) évoluent tous les jours : nouveaux modèles, logique métier, corrections.

Cette séparation fait que deux équipes peuvent travailler en parallèle sans se marcher dessus. L'ingénieur plateforme gère le Terraform, le data engineer gère le dbt. Chacun dans son repo, avec son propre cycle de déploiement.

## Le masquage de données : un bon exemple d'intégration

Un exemple concret de comment Terraform et dbt collaborent : le masquage de données.

Terraform crée les [politiques de masquage par tag](https://docs.snowflake.com/en/user-guide/tag-based-masking-policies.html) : "si une colonne est taguée PII, masquer la valeur sauf pour les rôles qui ont le droit de la voir." Il crée aussi les tags et les rôles de démasquage.

dbt, de son côté, applique les tags aux colonnes quand il crée les modèles : "cette colonne email est PII, cette colonne montant est FINANCIAL."

Le résultat : quand un utilisateur marketing fait un `SELECT * FROM clients`, les emails sont masqués automatiquement. L'utilisateur avec le bon rôle voit les valeurs réelles. Personne n'a eu à écrire de logique de masquage dans le SQL, c'est géré par la combinaison infrastructure + metadata.

## Auditer et itérer

Un des bénéfices les plus sous-estimés de cette approche : l'auditabilité.

Quand un audit de sécurité demande "qui a accès à quoi ?", la réponse est dans le code. Pas dans une console Snowflake où il faut naviguer 15 pages. Pas dans un document Excel maintenu à la main. Dans le code, versionné, avec l'historique complet de qui a changé quoi et quand.

Et quand tu veux ajouter une nouvelle équipe, un nouveau schéma, ou modifier des permissions, c'est un processus standard : modifier la config, ouvrir une PR, reviewer, merger, appliquer. Le même workflow que pour du code applicatif.

## Assisté par un LLM

Un dernier point qui vaut la peine d'être mentionné : une fois que ton infrastructure est déclarée sous forme de code structuré, un LLM devient un assistant redoutablement efficace. Tu peux lui demander de créer un nouveau rôle fonctionnel, d'ajouter un utilisateur, de modifier des permissions, et il produit du Terraform valide qui suit tes conventions existantes.

Avec des scripts SQL ad hoc, c'est beaucoup plus risqué. L'IA ne sait pas quel est l'état actuel de l'infra. Avec Terraform, l'état souhaité est dans le code, et le `plan` valide que le résultat est correct avant toute application. L'IA propose, Terraform vérifie.

## Le vrai gain

Au final, ce qui a changé, c'est pas la technologie, c'est la confiance. Tu sais que l'état de ton infrastructure correspond au code. Tu sais que les permissions sont correctes. Tu sais que personne n'a fait un changement non documenté. Et quand quelqu'un te demande d'ajouter un accès ou de créer un nouveau schéma, c'est 5 minutes de configuration au lieu d'une demi-heure de scripts SQL avec la peur de casser quelque chose.

Les scripts SQL, c'est de l'artisanat. Les migrations, c'est de l'artisanat mieux organisé. Terraform, c'est pas parfait non plus : le state peut être fragile, les erreurs de plan sont parfois cryptiques, et le provider Snowflake a ses propres bugs. Mais au moins, tu sais où t'en es. Et quand quelque chose casse, tu sais pourquoi.
