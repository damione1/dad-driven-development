---
title: "Mon SSD est mort pendant que j'étais parti, et j'ai refait mon backup"
date: 2026-08-08
draft: false
translationKey: "homelab-backup-r2"
description: "Cinquante notifications Uptime Kuma, un SSD mourant, une session Claude à distance, et le backup hors site chiffré que j'ai monté ensuite sur Cloudflare R2 avec restic et rclone."
tags:
  [
    "Homelab",
    "Unraid",
    "Backup",
    "Cloudflare R2",
    "restic",
    "DevOps",
    "Claude Code",
  ]
categories: ["DevOps"]
---

Je suis parti de la maison pour la journée. En milieu de matinée, mon téléphone se met à vibrer sans s'arrêter. Une cinquantaine de notifications Uptime Kuma, d'un coup.

Tous mes conteneurs sont down, la domotique avec. Aucune maintenance prévue. Ce n'est pas normal.

## Trente minutes, à distance, depuis mon téléphone

J'ouvre l'app Claude. Une session à distance est encore active sur mon ordinateur, celle que j'avais laissée ouverte la veille. Je lui décris la situation et je lui demande de s'occuper du problème.

Ce qui suit, je ne l'ai pas piloté. Grâce au [skill qui encode le contexte de mon serveur]({{< ref "/blog/llm-ssh-homelab-unraid/" >}}), il a pris les initiatives dans le bon ordre :

Diagnostic d'abord. Le SSD qui porte toutes les VM et tous les conteneurs est en train de mourir. Pas une panne franche, la pire des morts : celle où le disque répond encore, par intermittence, en corrompant au passage.

Puis l'action. Redémarrage du serveur. Le SSD refuse de remonter dans son pool, alors il le monte manuellement en dehors du pool. Copie de tout ce qu'il contient vers l'array de disques durs. Relance des conteneurs et des VM sur les données rapatriées.

Trente minutes plus tard, tout tournait de nouveau.

Et il revient vers moi avec le verdict : c'est reparti, mais ton SSD est mort, il faut le remplacer. Voici quelques modèles que je te suggère. J'ai vérifié le SMART plus tard dans la soirée, il avait raison, tout était au rouge. Le disque n'était plus fiable.

J'ai ouvert les liens Amazon, j'en ai commandé un. Tant qu'à sortir le serveur du rack, j'ai pris un disque dur de plus pour l'array de stockage. J'étais dû de toute façon.

En rentrant le soir, la domotique fonctionnait, Plex fonctionnait, mon cloud maison fonctionnait, tout ça temporairement sur les disques durs. Le lendemain, les deux disques sont arrivés. Je les ai branchés à chaud et j'ai dit à Claude de finir le travail. Il les a initialisés, a reconstruit le pool SSD, a agrandi l'array de disques durs et a remis les conteneurs à leur place. Retour à la normale.

## Le contrecoup

Ça fait six ou sept ans que j'ai ce homelab. Au début, c'était un laptop et des disques durs classiques, puis j'ai migré vers un vrai serveur d'occasion, avec des SSD et des disques durs NAS. En six ans, jamais un seul disque n'avait lâché. Rétrospectivement, j'ai surtout été très chanceux.

Le stress est arrivé après coup, une fois tout réparé. J'ai fait l'inventaire de ce que j'aurais réellement perdu si la copie de sauvetage n'avait pas marché.

**Les conteneurs**, honnêtement, ce n'est presque que de la configuration Ansible. Un playbook et ils reviennent. C'est précisément pour ça que j'avais fait cette migration.

**La VM**, c'est Home Assistant, avec une sauvegarde quotidienne. Rien de grave.

**Les bases de données**, par contre. Celle de Vaultwarden, par exemple. Ça, ce serait très chiant. Pas catastrophique au sens où mes mots de passe sont aussi en cache dans les clients, mais très chiant.

Côté array de disques durs, il y a de la redondance, donc c'est moins stressant. Mais on n'est jamais à l'abri de deux disques qui lâchent en même temps, ou pire, en silence. Et sur cet array, il y a deux catégories de données très différentes. Un gros paquet de films et de séries, qui n'est pas irrécupérable, juste long à reconstituer. Et mes photos et vidéos dans Immich, plus les documents de mon OpenCloud.

Ça, ce n'est pas perdable. Point.

## Mon 3-2-1 avait un maillon faible

J'avais déjà une sauvegarde hors site, parce que la règle 3-2-1 n'est pas négociable quand on héberge soi-même les photos de sa famille. Trois copies, deux supports, une hors site.

Sauf que ma destination hors site était un OneDrive. Deux problèmes.

Ce n'est **pas chiffré de mon côté**. Mes fichiers sont lisibles par le fournisseur, avec la télémétrie et les conditions d'utilisation qui vont avec, sur des documents personnels et des photos de famille.

Et c'est **cher pour ce que j'en fais**, autour de 12 dollars par téraoctet. Un abonnement Microsoft se justifie quand on utilise la suite qui va avec. Moi, je ne me servais de rien d'autre sur ce compte : je payais un espace disque au prix d'une suite bureautique que je n'ouvre jamais.

Je me suis dit que je pouvais faire mieux avec une solution plus sérieuse.

## Pourquoi un bucket

Je commence à bien connaître le stockage objet, j'en fais beaucoup au travail, entre AWS et Snowflake. Les propriétés qui m'intéressent ici sont exactement celles qui font sa valeur en entreprise : le stockage est bon marché, on peut définir des règles de cycle de vie pour basculer automatiquement les données froides vers une classe moins chère, et la capacité est effectivement illimitée. Pas de disque à racheter, pas de quota à surveiller.

Sauf que pour un particulier, ça reste plus cher qu'on ne le croit.

Chez [AWS S3](https://aws.amazon.com/s3/pricing/), le stockage standard est à 0,023 dollar par gigaoctet et par mois. Pour un téraoctet, on est à environ 23 dollars mensuels, soit le double de mon OneDrive. Et surtout, il y a les **frais de sortie** : environ 0,09 dollar par gigaoctet au-delà des cent premiers. Autrement dit, le jour où j'ai vraiment besoin de mon backup, où je dois rapatrier mon téraoctet parce que tous mes disques sont morts, ce jour-là on me facture 80 dollars supplémentaires. Il y a quelque chose de profondément malsain à faire payer le sinistre.

J'ai regardé la concurrence. Azure Blob, trop cher. Google Cloud Storage, trop cher aussi. Les classes d'archivage type Glacier font chuter le prix de stockage, mais ajoutent des délais et des frais de restitution qui reproduisent le même problème.

Puis je me suis rappelé que Cloudflare avait lancé son offre.

## Le twist : R2

[Cloudflare R2](https://developers.cloudflare.com/r2/pricing/) fait deux choses qui changent l'équation.

**Aucun frais de sortie.** Zéro. Récupérer l'intégralité de mes données ne coûte rien. Pour une sauvegarde, ce n'est pas un détail tarifaire, c'est le cœur du sujet : le scénario de restauration est le seul qui compte vraiment, et c'est celui que tous les autres facturent au prix fort.

**Et un stockage à prix cassé**, 0,015 dollar par gigaoctet et par mois en standard, 0,010 en accès peu fréquent, avec une API compatible S3, donc tout l'outillage habituel fonctionne sans adaptation.

D'après ma facture, je m'en sors autour de 10 dollars par mois.

Soyons honnête : sur le prix seul, l'écart avec OneDrive n'est pas spectaculaire. C'est un peu moins cher, pas révolutionnaire. Le prix n'est pas l'argument, c'est ce qui permet à l'argument d'exister : à tarif comparable, je passe d'un espace disque grand public, non chiffré de mon côté et vendu avec une suite que je n'utilise pas, à un stockage objet chiffré avant l'envoi, sans frais de restitution, et sur lequel c'est moi qui définis les règles. À qualité égale j'aurais hésité. À prix égal et qualité supérieure, il n'y a pas de débat.

## Deux classes de données, deux outils, deux buckets

Je n'ai pas voulu tout traiter de la même façon, parce que toutes mes données n'ont pas le même profil de risque.

**Le bucket rotatif**, géré par **[restic](https://restic.net/)**, contient ce qui est re-générable ou déjà répliqué localement : les dumps PostgreSQL d'Immich, la configuration d'OpenCloud. Chiffrement AES-256 côté client, déduplication, et une rétention classique de quatorze sauvegardes quotidiennes, huit hebdomadaires et six mensuelles. Le serveur a le droit d'y faire le ménage lui-même.

**Le bucket d'archive**, géré par **[rclone](https://rclone.org/crypt/)** en mode chiffré, contient ce qui n'est pas perdable : les originaux Immich, les documents OpenCloud, les dumps Vaultwarden, et la configuration de la clé USB Unraid.

Et là, une décision qui compte : ce bucket est alimenté en `copy`, jamais en `sync`. Le serveur peut seulement **ajouter**. Une suppression locale ne se propage jamais. Le bucket porte en plus un verrou d'immuabilité côté Cloudflare, indéfini.

La raison est simple : un backup qui se synchronise fidèlement avec la source réplique aussi les catastrophes. Un rançongiciel qui chiffre mes fichiers, ou une commande malheureuse, se propage dans la sauvegarde à la passe suivante. Ici, même si mon serveur est entièrement compromis, il ne peut pas détruire l'archive. Purger cette archive est une opération manuelle d'administration, avec un jeton qui ne se trouve nulle part sur le serveur.

C'est le même principe que celui que je répète au travail sur les données de production : les droits d'écriture d'un système et les droits de destruction de ses sauvegardes ne doivent jamais être portés par la même identité.

## Le détail qui compte : où vivent les clés

C'est la partie que les tutoriels bâclent, et c'est pourtant celle qui décide si ta sauvegarde vaut quelque chose.

Les buckets sont créés en **Terraform**, versionnés comme le reste. Les secrets sont répartis selon ce que chacun peut détruire :

- Le jeton de lecture-écriture rendu sur le serveur peut écrire, mais **ne peut pas lever le verrou** de l'archive.
- Les jetons d'administration, ceux qui peuvent lever le verrou et purger, vivent dans le trousseau macOS, protégés par Touch ID, et ne mettent jamais un pied sur le serveur.
- Le mot de passe du vault Ansible est également dans le trousseau, plus aucun fichier en clair.

Et la règle non négociable : les mots de passe qui déchiffrent les sauvegardes, celui de restic et celui de rclone, existent en **copie hors ligne**. Pas dans Vaultwarden, ce serait circulaire, puisque c'est précisément ce que je cherche à restaurer. Pas chez Cloudflare non plus, mettre la clé à côté du texte chiffré annule l'intérêt du chiffrement. Sur papier, dans un endroit sûr.

Si le serveur disparaît demain, j'ai toujours le chiffré chez Cloudflare. Sans ces deux mots de passe, ce chiffré ne vaut rien.

Le tout tourne toutes les nuits : dumps cohérents d'abord, bucket rotatif ensuite, archive en dernier.

## Vingt-quatre heures d'upload plus tard

Le premier envoi a pris une bonne journée. Depuis, seuls les deltas partent.

Ce que j'ai maintenant, et que je n'avais pas : mes données personnelles chiffrées avant de quitter la maison, chez un hébergeur qui ne peut pas les lire. Pas de télémétrie sur mes fichiers, pas d'entraînement de modèle sur mes photos de famille, pas de scan de contenu. Et une restauration que je peux déclencher sans qu'on me présente une facture au moment précis où ça va déjà mal.

Reste la partie que tout le monde oublie et que je dois encore faire sérieusement : tester la restauration. Une sauvegarde qu'on n'a jamais restaurée est une hypothèse, pas une sauvegarde.

Une dernière chose, pour la perspective. Le SSD et le disque dur que j'ai achetés ce jour-là m'ont coûté presque aussi cher que le serveur entier il y a six ans. Et ce serveur, c'est un EPYC 7302P, seize cœurs et trente-deux threads, sur une carte Supermicro, avec 128 Go de RAM. L'inflation sur le stockage, ça fait mal. J'espère ne pas avoir à en remplacer d'autre de sitôt.
