---
title: "Donner un accès SSH root à un LLM sur mon serveur Unraid"
date: 2026-05-16
draft: false
translationKey: "homelab-ssh"
description: "Pourquoi je laisse Claude Code opérer mon homelab Unraid en SSH : je connais les concepts devops, je n'ai simplement pas sa vitesse. Et comment Ansible a fini par réduire le besoin d'opérations à distance."
tags: ["LLM", "Homelab", "Unraid", "DevOps", "SSH", "Ansible", "Claude Code"]
categories: ["DevOps", "AI"]
---

J'ai un serveur Unraid qui fait tourner une quarantaine de conteneurs Docker, des stacks compose, deux GPU AMD, des VM, du Zigbee, du media, de la domotique. Bref, un homelab typique de quelqu'un qui en a trop accumulé avec le temps. Et j'ai donné l'accès SSH root à un LLM.

Je sais déjà ce que tu penses. Ouais, root. Sur ma prod domestique.

Laisse-moi expliquer pourquoi.

## Six ans de homelab, puis un bébé

J'ai ce serveur depuis à peu près six ans. Au début, c'était le terrain de jeu classique : j'essayais des affaires, je cassais, je réparais, je changeais d'idée la semaine d'après. Si Plex était down deux jours, ce n'était pas grave. Si Home Assistant ne bootait pas, je finissais par trouver le bug pendant une soirée pluvieuse.

Puis je suis devenu papa.

D'un coup, mon temps libre est devenu des blocs de trente minutes, entre deux siestes et une brassée de couches. Quand je veux vraiment fouiller un problème en profondeur, il faut que je sacrifie une nuit. Et les nuits, quand t'as un bébé, ce n'est pas le genre de ressource que tu veux brûler.

Le problème, c'est qu'en parallèle, le serveur s'est ancré dans les habitudes de la maison. Plex joue les dessins animés au petit. Vaultwarden contient les mots de passe de toute la famille. Immich remplace Google Photos pour les photos du bébé. Home Assistant contrôle l'éclairage, le chauffage, le baby monitor.

Ce n'est plus mon terrain de jeu, c'est de l'infrastructure familiale. Et je me retrouve à devoir tenir un niveau de fiabilité sérieux avec le temps d'un parent débordé, sans envie de monter un cluster qui va faire sourire Hydro-Québec en janvier. Pas de Kubernetes à la maison. Juste un serveur, bien configuré, qui ne tombe pas.

## Ce que je sais faire, et ce que je ne fais pas assez vite

Il faut être clair sur la nature du problème, parce que c'est là que la plupart des discussions sur « les LLM et l'ops » dérapent.

Je ne délègue pas parce que je ne comprends pas. Je fais du devops depuis des années, c'est une partie de mon métier. Je connais les concepts, je sais ce qu'est un namespace réseau, une capability, un cgroup, une unité systemd, un bridge Docker, un mapping de périphérique. Quand Claude me propose un correctif, je suis capable de le lire, de dire pourquoi il est bon ou mauvais, et de le refuser.

Ce que je n'ai pas, c'est sa vitesse.

Claude enchaîne vingt `grep` avec des regex correctes en quelques secondes. Il compose un pipeline `find | xargs | awk` du premier coup, sans chercher la syntaxe dans un man page. Il inspecte trente conteneurs, croise les résultats et sort la conclusion pendant que j'en serais encore à me rappeler l'ordre des arguments de `docker inspect --format`.

C'est là toute la différence. Entre « je sais à peu près où chercher » et « voici la réponse », il y a dix minutes pour moi et dix secondes pour lui. Sur une soirée entière, ça n'a aucune importance. Sur un bloc de trente minutes, c'est la totalité du bloc.

Autrement dit : ce n'est pas une compétence qui me manque, c'est du débit. Et le débit, ça se sous-traite.

## Encoder le contexte une bonne fois

Le contexte de mon serveur vivait dans ma tête, et il y vivait mal. « Zigbee2MQTT est en host ou en bridge ? » « C'est quel GPU qui fait le transcodage ? » « Les templates sont où déjà ? » Chaque fois que je debuggais quelque chose à 23h entre deux biberons, je passais un quart d'heure à retrouver le contexte avant même de commencer à penser au problème.

J'ai donc écrit un skill Claude Code, `homelab-devops`, qui contient le matériel (CPU, RAM, GPU, périphériques USB), les chemins importants, la topologie réseau, les conventions du serveur, et l'inventaire des conteneurs avec leur mode de déploiement.

Quand je dis « regarde pourquoi le conteneur X ne démarre pas », le skill se charge tout seul. Claude sait déjà comment joindre le serveur, où sont les logs, si c'est un conteneur compose ou un template Unraid, quels chemins sont des user shares et lesquels sont des pools cache.

Effet secondaire que je n'avais pas prévu : écrire ce skill m'a forcé à documenter noir sur blanc ce qu'est mon serveur. Une documentation que je n'avais jamais pris le temps de faire en six ans. Elle profite autant à moi qu'au modèle.

## L'accès reste le mien

Sur le fait de « donner root » : ce n'est pas un accès permanent ni gravé dans la pierre.

Claude travaille via une paire de clés SSH dédiée, que je peux activer ou révoquer à la volée. Je lui indique où est la clé privée, il compose ses `ssh root@...` tout seul. À la fin d'une session sensible, je retire la clé publique des `authorized_keys` et l'accès n'existe simplement plus.

Ensuite, deux modes de collaboration.

**Validation stricte.** Chaque commande qui écrit est validée avant exécution. Je lis, je comprends, j'approuve. Lent, mais c'est le mode que j'utilise sans exception dès qu'on touche à du critique.

**Mode direct.** Claude exécute sans demander, en suivant un plan qu'on a défini ensemble.

Ma règle est simple : le mode direct est réservé aux actions dont la pire conséquence est « oups, faut refaire », jamais « oups, faut tout restaurer depuis un backup ». Un diagnostic en lecture seule, un audit d'inventaire, un grep dans les logs, aucun problème. Un redémarrage de conteneur dont dépend Home Assistant un samedi matin, je valide chaque étape même si ça prend trois fois plus de temps.

## Puis j'ai arrêté de tout faire en SSH

C'est l'évolution la plus utile de toute cette histoire, et elle réduit précisément le rôle décrit plus haut.

Opérer un serveur en SSH, même très vite, même très bien, reste de l'impératif. Chaque intervention est un geste ponctuel, non reproductible, et l'état réel du serveur finit par diverger de ce que tu crois savoir. Un LLM rapide ne règle pas ce problème, il le rend juste moins douloureux. Ce qui est déjà bien, mais ce n'est pas une solution.

J'ai donc migré tous mes conteneurs vers **[Ansible](https://www.ansible.com/)**. Un dépôt qui est maintenant la source de vérité de ce qui tourne sur le serveur : il crée les réseaux Docker, pousse les fichiers compose, génère les fichiers `.env` à partir de secrets chiffrés avec Ansible Vault, et déploie les stacks.

Une contrainte amusante au passage : Unraid n'a pas de Python 3, donc les modules Docker d'Ansible ne peuvent pas tourner sur la cible. Ils tournent sur mon Mac et parlent au démon Docker à distance, via SSH. C'est inhabituel, mais ça marche très bien et ça évite d'installer un runtime sur un OS qui n'est pas fait pour ça.

Ce que ça change concrètement :

- **Les fichiers compose vivent dans le dépôt, pas sur le serveur.** On édite ici, on pousse. Éditer directement sur la machine est devenu interdit, y compris pour moi.
- **Les secrets sont chiffrés et versionnés**, plus éparpillés dans des variables d'environnement que personne ne se rappelle avoir posées.
- **Reconstruire un service est une commande**, pas une séance d'archéologie.
- **Le nombre d'opérations à distance s'effondre.** Claude n'a plus besoin de bricoler le serveur : il modifie une configuration versionnée, que je relis comme une pull request, et l'application est un `ansible-playbook`.

C'est le bon ordre des choses. Le LLM est excellent pour opérer vite quand il faut opérer vite. Mais la vraie amélioration, ce n'est pas d'opérer vite, c'est de ne plus avoir besoin d'opérer. L'accès SSH sert maintenant à diagnostiquer et à gérer l'imprévu. Le reste est déclaratif.

## Les gardes-fous

Donner root à un LLM, ça ne se fait pas n'importe comment.

**Confirmation avant écriture.** Les commandes en lecture seule passent librement, tout ce qui modifie est validé.

**Rien de destructif sans plan écrit.** `rm -rf`, `docker system prune`, un reset git. Jamais sans que le plan soit posé noir sur blanc, en dry-run quand ça existe.

**Pas d'accès à l'interface web Unraid.** Il travaille en SSH sur des fichiers et via le CLI Docker, là où tout est traçable.

**Les secrets restent hors contexte.** Ils vivent dans des fichiers `.env` sur le serveur ou dans la voûte Ansible chiffrée, jamais en clair dans un dépôt ni dans une conversation. Un skill dédié documente comment s'en servir : on les redirige par des pipes vers la commande qui en a besoin, sans jamais afficher leur valeur. Claude sait qu'un secret existe, il sait où il est, il sait le passer à un processus, mais il ne le voit pas passer et il ne finit pas dans l'historique de session.

Ce n'est pas encore optimal, et je ne vais pas prétendre le contraire : ça repose sur une discipline documentée plutôt que sur une barrière technique, et rien n'empêcherait matériellement un `cat` de trop. Ça fonctionne en pratique, et je travaille sur une meilleure solution.

## Ce qui marche moins bien

**Les commandes interactives.** Un prompt `y/N`, un `docker exec -it`, un wizard d'installation. Il faut reformuler en non-interactif ou le faire à la main.

**Les problèmes visuels.** Un graphe Grafana qui déconne, un dashboard bizarre : l'outil perd beaucoup de son intérêt. J'ai fini par faire des captures d'écran et décrire ce que je voyais.

**Les erreurs silencieuses.** Une commande qui retourne 0 sans avoir fait ce qu'on croyait. J'ai pris l'habitude de demander systématiquement « vérifie que c'est bien appliqué » après tout correctif non trivial.

**Les décisions d'architecture.** « Est-ce que je passe à TrueNAS », « quel orchestrateur » : c'est une discussion, pas une exécution. Il aide à réfléchir, je tranche.

## La vraie leçon

Un LLM avec un accès SSH sur mon homelab, ce n'est pas un remplaçant d'administrateur système. C'est un multiplicateur de débit pour quelqu'un qui sait déjà ce qu'il fait, mais qui n'a plus le temps de le faire à la vitesse humaine.

Ce que j'y gagne : un debug de deux heures qui rentre dans un bloc de trente minutes. Des micro-tâches qui traînaient depuis six mois et qui se font en passant. Des audits « pourquoi ça va mal » qui deviennent systématiques parce qu'ils ne coûtent plus rien à lancer, donc je trouve les problèmes avant qu'ils me réveillent à 3h du matin.

Ce que je n'y gagne pas : la responsabilité. C'est toujours mon serveur, mes données, mes choix. Si ça casse, c'est moi qui répare. Le LLM est un outil, pas un coupable. Et si je délègue sans lire, je n'apprends rien : je lis ce qu'il fait, je me fais expliquer les paramètres que je ne connais pas.

Au final, mon objectif n'est pas que mon serveur soit impressionnant. C'est qu'il soit ennuyant. Qu'il ne tombe pas. Qu'il ne me réveille pas. Qu'il laisse Plex jouer les dessins animés un samedi matin sans que papa ait à se lever.

Claude Code m'en rapproche, mais pas de la façon dont on l'imagine. Pas en rendant le serveur plus fiable en soi, mais en me redonnant assez de temps pour faire les petites améliorations qui, cumulées, font la différence entre une infra qu'on subit et une infra qui se tient debout toute seule. Ansible en est la preuve : c'est exactement le genre de chantier que je repoussais depuis des années.
