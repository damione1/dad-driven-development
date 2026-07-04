---
title: "Maslow Desktop : un contrôleur CNC construit autour d'une machine à états"
date: 2026-07-04
draft: false
translationKey: "maslow-desktop"
description: "Comment une CNC restée un an et demi dans son carton m'a poussé à écrire un client desktop pour la Maslow 4 : une machine à états extraite du firmware, un assistant de calibration, et un serveur MCP tant qu'à faire."
tags: ["Rust", "Tauri", "CNC", "Open Source", "MCP", "LLM"]
categories: ["Développement", "Open Source"]
images: ["maslow-run-toolpath.png"]
---

J'ai participé au Kickstarter de la [Maslow 4](https://www.maslowcnc.com/). Je l'ai reçue, je l'ai montée, et elle est restée dans son carton pendant un an et demi.

Ça arrive. Le genre de projet qu'on finance avec enthousiasme et qu'on ressort quand la vie laisse enfin une fenêtre.

Quand je l'ai enfin allumée, j'ai reconnu l'interface tout de suite : c'est [FluidNC](https://github.com/bdring/FluidNC), un excellent projet de contrôleur dérivé de GRBL qui tient dans un ESP32. Je connais bien cette UI : j'avais déjà installé ce firmware sur une carte d'imprimante 3D, une MKS TinyBee, pour piloter ma [thread art machine](https://github.com/damione1/thread-art-generator). Ça fonctionne, c'est fiable, et ce n'est pas très sexy.

Alors je me suis dit : pourquoi pas en faire ma version.

![L'onglet Run avec la prévisualisation du parcours d'outil](maslow-run-toolpath.png)

Une précision avant d'aller plus loin, parce que ce genre d'article se lit vite de travers. J'aime beaucoup le projet Maslow. Une CNC grand format, open source, à un prix qui la met à portée d'un atelier de garage, portée par Bar Smith et une petite communauté qui documente tout et répond aux questions : c'est exactement le genre de projet que je veux voir exister. Le firmware et l'interface embarquée sont le travail de gens qui font ça bien, sur un problème difficile, avec des moyens sans commune mesure avec ceux d'un fabricant industriel.

Ce qui suit n'est pas une critique de leur travail. C'est l'histoire d'un gars qui s'est écrit un client desktop pour sa propre machine, parce qu'il avait envie de se la rendre agréable, et qui l'a publié tant qu'à faire.

Concrètement, l'interface embarquée tourne dans l'ESP32 de la machine, et y faire tenir une UI web en plus du contrôle temps réel de quatre moteurs relève déjà de la prouesse. Elle fait beaucoup avec très peu, mais elle reste minimale par nécessité : tous les boutons sont là en permanence, rien ne t'indique dans quel ordre les presser, et un clic hors séquence met la machine en alarme. En sortir demande de rétracter les quatre courroies puis de les redéployer entièrement, c'est long, et à chaque cycle le mécanisme peut mâchouiller une courroie. Ce n'est pas théorique, ça m'est arrivé. Je n'avais pas envie de corriger leur interface, j'avais envie de la mienne : mon contrôleur, ma surcouche, à ma sauce.

## Le point de départ : une machine à états

Avant de dessiner le moindre écran, j'ai voulu répondre à une seule question : **quelles transitions le firmware autorise-t-il réellement ?**

C'est une information qui existe, mais elle est dispersée dans le code C++ du firmware, dans les gardes qui acceptent ou refusent les changements d'état. J'ai donc demandé à Claude d'analyser le firmware Maslow, de retrouver la fonction qui arbitre les changements d'état, et d'en extraire la liste exhaustive des transitions permises, avec leurs conditions.

À partir de ça, j'ai construit dans le backend Rust un modèle typé qui reflète le firmware : dix états explicites, du « inconnu » au « prêt à couper » en passant par le déploiement des courroies et le calcul de calibration. Chaque état sait ce qu'il autorise, et le backend en déduit une politique d'actions que l'interface consomme telle quelle.

La conséquence est simple : **l'UI ne propose que ce que la machine peut réellement faire à cet instant**. Un bouton qui mènerait à une transition refusée est désactivé, avec la raison affichée à côté plutôt qu'un cul-de-sac silencieux. On ne peut plus se mettre en alarme par une mauvaise séquence de clics, parce que la mauvaise séquence n'est pas cliquable.

C'est exactement le même principe que celui que j'avais mis au centre de [Soufflé]({{< ref "/blog/souffle-transcription-locale-macos/" >}}), mon autre projet en Rust : rendre l'état explicite et interdire les combinaisons invalides par construction. Sauf qu'ici, l'enjeu n'est pas un bouton grisé au mauvais moment. C'est une défonceuse qui tourne.

Un détail qui a son importance : le firmware reste la source de vérité. Mon modèle est un miroir conservateur du sien, pas une autorité concurrente. S'il refuse quelque chose que j'autorisais, c'est mon modèle qui a tort.

## Un assistant de calibration en langage courant

![L'assistant de calibration guidé](maslow-calibrate-wizard.png)

Une fois les transitions connues, l'assistant s'écrit presque tout seul. Chaque étape est expliquée en langage courant, elle avance automatiquement quand le firmware signale sa progression, et l'ordre des opérations n'est plus quelque chose que l'utilisateur doit mémoriser. C'est le rôle du programme de connaître l'ordre.

J'y ai ajouté les deux gestes que je faisais le plus souvent : une reprise quotidienne en un geste, quand la machine est déjà calibrée et qu'il suffit de remettre les courroies en tension, et un relâchement de tension pour laisser reposer les courroies et le cadre pendant la nuit.

## Une app desktop, en réutilisant ce que je savais déjà

![L'onglet principal avec le jog et l'affichage des positions](maslow-main-jog.png)

Neuf jours entre le premier commit et une version installable. C'est court, et ce n'est possible que parce que je ne partais pas de zéro : j'ai repris la stack de Soufflé, [Tauri](https://tauri.app/) avec un cœur Rust et Svelte 5 pour l'interface. Le côté Rust possède la connexion à la machine, l'envoi du G-code et le modèle de calibration ; le frontend affiche.

La contrainte de design que je me suis imposée : **une seule interface, pensée pour le tactile**. Un contrôleur de CNC s'utilise debout, à côté de la machine, souvent avec les mains sales. Donc de gros boutons, un ABORT rouge toujours accessible, un pied de page qui montre l'état de la machine en permanence, et une grammaire de couleurs stricte (bleu pour agir, orange pour les origines, vert pour ce qui tourne, rouge pour arrêter).

J'ai regardé ce qui se fait ailleurs, du côté des contrôleurs CNC sur écran tactile, pour voir quelles conventions valaient la peine d'être reprises. Le résultat est une mise en page unique qui s'adapte d'une tablette en portrait montée près de la machine jusqu'à une fenêtre de bureau. Pas de mode « mobile » et de mode « desktop » à maintenir séparément, ce qui est surtout une décision de paresse assumée : deux interfaces, c'est deux fois les bugs.

## Et un MCP, tant qu'à faire

Celui-là est né pendant les séances de débogage.

Piloter une CNC à la main pour reproduire un bug, c'est lent. À chaque essai, il faut refaire la même séquence de gestes. À un moment, en plein débogage, je me suis dit que ce serait quand même pratique que l'agent qui m'aide puisse manipuler le contrôleur lui-même plutôt que de me dicter les étapes.

L'app expose donc la surface de contrôle de la machine en HTTP, gRPC et [MCP](https://modelcontextprotocol.io/), derrière une clé d'API qu'on génère soi-même. C'est **désactivé par défaut**, ce qui est le minimum syndical quand on parle d'une API qui fait bouger un outil coupant.

Ce n'est pas une fonctionnalité que j'aurais imaginée sur le papier. Elle vient d'un agacement pratique, et c'est probablement pour ça qu'elle sert.

## Ce que c'est, et ce que ce n'est pas

Soyons clairs sur le périmètre : Maslow Desktop fait à peu près ce que fait l'interface web embarquée. Piloter les axes, charger et lancer un job, parcourir la carte SD, régler la configuration, taper des commandes brutes dans une console. C'est un client desktop, et n'importe qui ayant envie de gratter le protocole pourrait en écrire un.

Ce qu'il apporte tient en deux choses : il connaît la machine à états et refuse de t'y faire déroger, et il te guide pas à pas dans la calibration au lieu de te laisser reconstituer l'ordre tout seul.

C'est un petit projet, né d'une irritation personnelle, et il résout exactement mon problème : ma CNC est sortie du carton, et je ne repasse plus une soirée à rétracter et redéployer des courroies parce que j'ai cliqué dans le mauvais ordre.

## Le distribuer proprement

Je l'ai écrit pour moi, mais publier coûtait peu et pouvait servir à quelqu'un d'autre avec la même machine et le même agacement.

Comme pour Soufflé, j'ai fait l'effort d'aller jusqu'au bout de la chaîne de distribution plutôt que de m'arrêter à « clonez et compilez ». La version macOS est signée et notarisée chez Apple, donc elle s'ouvre sans avertissement et sans manipulation obscure. Il y a aussi un exécutable Windows. Rien de plus à faire que télécharger et installer.

Il manque encore les stores. J'aimerais publier une version tablette, sur l'App Store et sur Android, parce que c'est là que cette app a le plus de sens : un écran tactile monté à côté de la machine, pas un laptop posé en équilibre sur un bout de contreplaqué. C'est surtout du travail administratif et de signature, donc ça attendra un moment où je peux y consacrer du temps sans le prendre sur autre chose. Projet futur, si j'y arrive.

---

Maslow Desktop est sur [GitHub](https://github.com/damione1/maslow-desktop), sous licence GPL-3.0, avec des binaires macOS signés et notarisés et un installeur Windows sur la page des [releases](https://github.com/damione1/maslow-desktop/releases/latest). Il faut une Maslow sous FluidNC accessible sur le réseau.
