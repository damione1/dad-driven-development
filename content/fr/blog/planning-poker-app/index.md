---
title: "Construire Planning Poker : Une app de collaboration temps réel (que j'ai peut-être trop complexifiée)"
date: 2025-11-15
draft: false
description: "Construire une application Planning Poker temps réel avec Go, PocketBase, htmx et WebSockets - et peut-être aller un peu trop loin avec l'infrastructure de déploiement."
tags: ["Go", "PocketBase", "htmx", "WebSockets", "DevOps", "Terraform", "AWS"]
categories: ["Développement", "Backend"]
images: ["planning-poker-featured.jpg"]
---

Fait que j'ai construit cette app de Planning Poker. Tu sais, le truc agile d'estimation où l'équipe se rassemble pour voter sur les story points ? Ouais, j'ai décidé de la faire web et temps réel. Et puis... disons juste que je suis devenu un peu ambitieux avec le déploiement.

## L'idée

Les sessions de Planning Poker c'est généralement le chaos dans une salle de conférence — quelqu'un écrit sur un tableau, quelqu'un d'autre crie son estimation, pis la moitié de l'équipe est dans la lune. Le problème est simple : t'as besoin de quelque chose qui marche instantanément, qui demande pas d'inscription, pis qui juste... marche.

Fait que j'ai pris Go pour le backend (rapide, compilé, un bon modèle de concurrence), PocketBase comme couche tout-en-un pour la base de données et l'authentification, pis j'ai ajouté htmx + Alpine.js sur le frontend pour cette sensation réactive sans bâtir une app React complète. WebSockets pour les mises à jour temps réel. Simple.

## Pourquoi PocketBase ? (Le point de vue du gars WordPress)

Voici l'affaire — j'ai passé des années à faire du développement WordPress. Je l'ai aimé, détesté, la routine habituelle. Quand j'ai transitionné vers du vrai développement logiciel, j'ai commencé à faire toute la danse : choisir un ORM, configurer les migrations, brancher l'authentification, bâtir ton routeur, gérer les changements de schéma de base de données... c'est épuisant.

PocketBase m'a fait penser à l'équivalent Go de ce que WordPress était pour moi — une base de référence sensée tout-en-un. T'as une couche de base de données, l'authentification, une interface d'admin, des migrations qui marchent pour vrai, pis un routeur HTTP. C'est assez structuré pour avancer vite, mais assez flexible pour étendre.

Je voulais pas passer la prochaine semaine à gosser avec Gorm, écrire des fichiers de migration, bâtir un middleware d'authentification pis configurer un routeur. Je voulais livrer quelque chose en quelques jours. PocketBase m'a permis de me concentrer sur le vrai problème : faire marcher Planning Poker en temps réel.

Pis si ça se transforme un jour en quelque chose que je veux monétiser comme SaaS, je peux bâtir par-dessus cette fondation. La nature headless de PocketBase veut dire que je peux éventuellement changer le frontend, ajouter une couche de tarification, peu importe. C'est une base solide.

## SQLite : Le héros méconnu des bases de données

Pour vrai : j'adore Postgres. Sérieusement. Mais pour ce projet-là ? SQLite c'était le bon choix.

La plupart du monde réalise pas ça, mais SQLite c'est le moteur de base de données [le plus déployé](https://sqlite.org/mostdeployed.html) au monde. Ton téléphone a probablement une douzaine de bases SQLite qui roulent en ce moment. Android, iOS, Firefox, Chrome — tout SQLite. C'est plate, fiable, pis vraiment sous-estimé.

Pour une app Planning Poker qui a pas besoin de mise à l'échelle horizontale ou de transactions multi-utilisateurs complexes à grande échelle, SQLite c'est _exactement_ ce qu'il te faut. Pas de serveur de base de données séparé à gérer, pas de maux de tête de pooling de connexions, pas d'appels de crise "est-ce que la DB est down ?" à 3h du matin. Ça vit dans un fichier. Tu peux le sauvegarder, le versionner, le déplacer.

J'ai fait un choix délibéré cette fois : pas trop complexifier la base de données. L'app a pas besoin de la complexité de Postgres. SQLite gère 20 000 connexions WebSocket concurrentes sur un t3.micro sans broncher. C'est en masse.

## Ce que ça fait

La boucle principale est pas mal directe :

- Créer une salle, pas besoin de connexion
- Le monde rejoint avec un nom
- Tu configures un tour de vote avec Fibonacci ou des valeurs personnalisées
- Tout le monde vote en même temps
- Révélation et discussion
- On recommence

Y'a aussi des affaires basées sur les rôles — tu peux être un voteur ou juste spectateur. Les créateurs de salle peuvent verrouiller les choses comme ils veulent. Tout se synchronise en temps réel via WebSocket, fait que quand quelqu'un vote, tout le monde le voit instantanément (ou voit qu'il a voté, du moins — les votes sont cachés jusqu'à la révélation).

La gestion d'état est propre. Les tours ont des états : vote, révélé, terminé. Les participants suivent qui est où. Les votes vivent dans la base de données. Les salles expirent automatiquement après 24 heures fait que tu stockes pas de données mortes pour toujours.

## Performance ? Ouais, j'ai vérifié

C'est ici que ça devient un peu absurde. Sur un t3.micro (1 vCPU, 1GB RAM), ce truc-là peut gérer :

- 2 000-3 000 salles concurrentes
- 20 000-30 000 connexions WebSocket

C'est... pas mal plus que ce dont t'aurais jamais besoin pour du Planning Poker. Mais j'allais pas livrer quelque chose qui pourrait pas gérer son propre succès, non ?

J'ai intégré de la diffusion asynchrone avec livraison de messages non-bloquante, des canaux d'envoi par client avec mise en mémoire tampon, du verrouillage fin. Y'a des endpoints de surveillance pour que tu puisses jeter un œil aux métriques en temps réel. Les clients lents se font détecter pis nettoyer automatiquement. C'est beaucoup trop complexe pour ce que ça fait, mais ça _marche_.

## Le rabbit hole du déploiement

Pis là j'suis arrivé au déploiement.

Au lieu de juste le lancer sur un serveur avec SSH, j'ai décidé d'aller en mode entreprise. Voici la configuration :

1. **GitHub Actions** surveille les tags git
2. Construit une image Docker multi-architecture
3. Pousse vers GitHub Container Registry
4. Déclenche une commande AWS Systems Manager Run Command
5. L'instance EC2 tire l'image
6. Docker Compose démarre les conteneurs
7. Les vérifications de santé valident que tout est en ligne

Zéro clé SSH exposée. Aucun port ouvert sauf HTTP/HTTPS. Tout est audité dans CloudTrail. Pis ouais, j'suis allé full Terraform sur l'infrastructure — EC2, groupes de sécurité, rôles IAM, toute la patente.

Est-ce excessif pour une app Planning Poker ? 100%. Est-ce que j'aurais pu juste SSH dans une machine pis la rouler ? Ouais. Mais de cette façon-là, déployer c'est littéralement juste `git tag v1.0.0 && git push origin v1.0.0`. Deux minutes plus tard c'est en ligne. Pis t'as une piste d'audit, des capacités de rollback automatique, pis de l'infrastructure-as-code. Fait que vraiment, je suis juste minutieux.

## Stack technique (la vraie)

- **Backend** : Go 1.25, PocketBase 0.30 (qui est essentiellement Echo + SQLite + interface d'admin dans un seul binaire)
- **Frontend** : htmx 2.0 pour AJAX/WebSocket, Alpine.js 3.14 pour l'interactivité, Templ pour le templating
- **Base de données** : SQLite (inclus dans PocketBase)
- **Déploiement** : Docker, Docker Compose, Terraform, GitHub Actions, AWS SSM
- **Surveillance** : Endpoint de métriques intégré, vérifications de santé

## Pourquoi cette stack ?

PocketBase c'était la vraie découverte. Tout le monde veut bâtir un backend, mais PocketBase t'en donne juste un. Base de données, migrations, interface d'admin, authentification, tout. J'ai juste eu à brancher le hub WebSocket pis la logique métier. C'est du temps pas passé à gosser sur du boilerplate.

htmx + Alpine c'est sous-estimé pour ce genre de projets. Pas de maux de tête de pipeline de build, pas de fatigue de framework JavaScript, juste des attributs HTML déclaratifs qui font ce que t'attends. Amélioration progressive, hypermédia, toutes ces bonnes affaires-là. T'écris moins de code, c'est plus facile à suivre, pis ton frontend devient pas un cauchemar de maintenance dans six mois.

Pis les goroutines de Go ont rendu le hub WebSocket trivial. Diffuser des messages à des milliers de connexions ? Goroutines avec channels. Fait. C'est le vrai avantage de Go pour ce cas d'utilisation — de la concurrence qui te rend pas fou.

## Le vrai défi

Honnêtement ? Réussir la machine d'état. Les tours doivent s'enchaîner proprement : vote → révélé → terminé → nouveau tour. Les participants doivent rester synchronisés. Les connexions tombent pis se reconnectent — tu peux pas perdre le vote de quelqu'un parce que son WiFi a flanché.

Ça m'a pris plus de réflexion que le déploiement. L'infrastructure faisait juste... ce que l'infrastructure fait. La partie difficile c'était de s'assurer que la logique de vote était solide.

## Fait que c'est fini ?

Ouais, ça marche. Tu peux rouler `make dev` pis ça démarre localement avec rechargement en direct. Y'a des tests d'intégration. Ça a des métriques, des vérifications de santé, de la gestion d'erreurs correcte. Tu pourrais vraiment utiliser ça pour rouler des sessions Planning Poker maintenant.

Est-ce que le déploiement est beaucoup trop complexe pour une app Planning Poker qui va probablement jamais suer à grande échelle ? Absolument. Mais bon, c'est là, ça marche, pis maintenant j'ai un template pour déployer des apps Go sur AWS sans toucher à SSH. Ça vaut quelque chose, non ?

La vraie leçon par contre ? Complexifie pas trop la base de données. SQLite a fait la job. PocketBase m'a sauvé de passer trois jours sur du boilerplate. Les primitives de concurrence de Go ont rendu les parties difficiles faciles. Pis des fois, la meilleure infrastructure c'est celle qui se tasse de ton chemin pis qui marche juste.
