---
title: "Planning Poker sans inscription : mon app en Go et htmx"
date: 2025-11-15
lastmod: 2026-10-27
draft: false
translationKey: "planning-poker-app"
description: "Mon app de planning poker en ligne sans inscription : salles partagées, cartes personnalisées et votes en direct, avec Go, PocketBase et une interface htmx."
tags: ["Go", "PocketBase", "htmx", "WebSockets", "DevOps", "Terraform", "AWS"]
categories: ["Développement", "Backend"]
images: ["planning-poker-app-featured.jpg"]
---

J'ai créé [pokerplanning.net](https://pokerplanning.net/) pour un besoin assez précis : faire une session de planning poker dans un navigateur sans demander à toute l'équipe de créer un compte de plus. Créer une salle, entrer son nom, partager le lien, voter, révéler, discuter. C'était ça, la partie utile.

Ensuite, je lui ai donné un backend Go, un hub WebSocket et une chaîne de déploiement avec Terraform et AWS Systems Manager. J'ai peut-être pris un peu trop de plaisir sur la partie infrastructure.

L'app est utilisable aujourd'hui. Voici ce qu'il y a derrière, et comment lancer une session si c'est surtout l'outil qui vous intéresse.

## Faire une session de planning poker sans s'inscrire

Le planning poker associe une estimation individuelle à une discussion d'équipe. Chacun choisit une valeur avant la révélation des cartes, puis l'équipe échange sur les écarts. Le [guide de Mountain Goat Software](https://www.mountaingoatsoftware.com/agile/story-points/planning-poker) explique la méthode ; l'app s'occupe de la salle partagée et des votes.

Le déroulement est court :

1. **Créer une salle.** Lui donner un nom et choisir Fibonacci, Fibonacci modifié, les tailles de t-shirt ou ses propres valeurs.
2. **La rejoindre soi-même**, puis partager le lien ou le QR code. Les autres participants entrent un nom et choisissent de voter ou d'observer.
3. **Discuter d'une story et voter.** La salle indique qui a voté, tout en gardant les valeurs masquées jusqu'à la révélation.
4. **Révéler et comparer.** Regarder la répartition, l'accord et, pour des votes numériques, la moyenne. Discuter des écarts plutôt que prendre la moyenne pour une décision de l'équipe.
5. **Revoter ou passer au tour suivant.** Les réglages contrôlent qui peut révéler, réinitialiser et avancer, et si la révélation se fait automatiquement une fois tous les votes reçus.

![Le formulaire de création d'une salle Planning Poker, avec le nom, l'échelle de vote et les réglages de session](planning-poker-room-setup.png)

Les salles expirent 24 heures après leur création. Ce sont des espaces temporaires pour une session, pas un backlog permanent ni un historique de toutes les estimations de l'équipe. On garde son outil de suivi et son appel habituel à côté.

## Pourquoi je suis parti sur PocketBase

J'ai fait du développement WordPress pendant des années. Une chose que j'y appréciais, c'était d'avoir un point de départ utilisable : les données, l'administration et des conventions déjà en place. En allant davantage vers Go, je n'avais pas spécialement envie que chaque projet personnel commence par le choix d'un ORM, le branchement de l'authentification et la fabrication d'une interface d'admin.

[PocketBase](https://pocketbase.io/docs/) m'a donné une base comparable en Go : SQLite, des collections, des migrations, des fonctions d'authentification et une interface d'administration. Le parcours public du planning poker ne demande pas de compte, mais les outils de gestion des données restent utiles côté backend.

J'ai donc pu passer du temps sur les salles, les participants, les tours et les votes. C'est ce que l'app doit réellement comprendre. Partir d'un backend qui gère déjà ce qu'il y a autour permet d'y arriver plus vite.

## SQLite pour les votes, WebSocket pour la salle

L'état persistant est stocké dans SQLite via PocketBase. Une salle a des participants et des tours ; les votes appartiennent à un tour. Les connexions actives sont gérées séparément par le hub WebSocket en Go.

Le stockage et la transmission restent donc séparés. Le hub suit les membres connectés et diffuse les mises à jour. Les files bornées et les limites de connexions placent des garde-fous autour de la charge, pendant que la base conserve les données de la session.

Pour ce projet, l'avantage pratique de SQLite est plus simple : pas de service de base de données séparé à exploiter pour un outil dont le modèle de données est petit et bien défini.

## htmx plutôt qu'une app frontend séparée

L'interface est rendue en Go avec [templ](https://templ.guide/). [htmx](https://htmx.org/docs/) gère les interactions HTML et l'intégration WebSocket ; Alpine.js s'occupe des contrôles côté navigateur, comme les réglages d'une salle.

J'aime garder l'interface proche du serveur pour ce genre d'app. Une salle n'est pas une plateforme frontend entière. Elle doit afficher les participants, accepter les votes, révéler les résultats et rester cohérente avec le tour courant du serveur.

La partie intéressante, c'est le cycle du tour : `vote → révélation → terminé`. Quand quelqu'un se reconnecte ou qu'un autre participant lance le tour suivant, les navigateurs doivent converger vers le même état. Il y a des tests d'intégration autour des votes, de la révélation automatique et des limites de participants, ainsi que des tests de concurrence autour du hub.

C'est là que le soin doit se porter. Un joli formulaire ne sert pas à grand-chose si deux personnes regardent des tours différents.

## Le détour par le déploiement

Le premier workflow de déploiement était plus élaboré que le formulaire ne le laissait penser. Un tag de version déclenchait les tests, la construction d'une image Docker et son envoi dans GitHub Container Registry. GitHub Actions utilisait ensuite AWS Systems Manager pour lancer la commande de déploiement sur EC2.

La production a depuis migré vers AWS Lightsail. Le pipeline actuel teste et construit toujours l'app à partir d'un tag de version, puis déploie le conteneur avec un certificat SSH Lightsail de courte durée et vérifie le point de contrôle public. Terraform décrit l'infrastructure, et Docker Compose fait tourner le service.

Le mécanisme a changé, mais le but reste le même : une publication reproductible qui vérifie ce qui s'est réellement passé, plutôt que supposer qu'une image envoyée signifie que l'app est en ligne.

J'aurais pu faire plus petit, certainement. Mais je voulais aller au bout d'un déploiement Go reproductible qui ne repose pas sur l'ouverture manuelle d'une session SSH. La chaîne faisait partie de ce que je voulais apprendre, même si une salle de planning poker n'exigeait pas tout cet exercice.

## Utiliser l'outil ou lire le code

Si votre équipe cherche une salle de planning poker en ligne sans inscription, [créez une salle sur pokerplanning.net](https://pokerplanning.net/). Choisissez l'échelle, rejoignez la salle en premier et envoyez le lien à l'équipe.

Le [code est sur GitHub](https://github.com/damione1/planning-poker), sous [Elastic License 2.0](https://github.com/damione1/planning-poker/blob/master/LICENSE). Pour l'implémentation, le gestionnaire de salles, le hub WebSocket et les tests d'intégration sont les points d'entrée utiles.

C'est un petit outil de collaboration construit autour d'un parcours volontairement court. Le déploiement a pris le chemin des écoliers. Rejoindre une salle ne devrait pas en faire autant.
