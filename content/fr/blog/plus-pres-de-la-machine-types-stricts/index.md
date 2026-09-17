---
title: "Plus près de la machine : une obsession pour les types stricts"
date: 2026-09-17
draft: false
translationKey: "strict-types"
description: "2026, on code presque plus : ce n'est pas une raison de lâcher la rigueur. On dicte le quoi et le comment à l'IA. Mon comment, c'est les types stricts : un set dans le type, un compilateur qui crie, un contrat qui traverse les machines."
tags: ["Rust", "Go", "Types", "Protobuf", "gRPC", "Software Design"]
categories: ["Développement"]
images: ["closer-to-the-machine-strict-types-featured.jpg"]
---

2026, c'est l'heure du développement agentique. On code presque plus. On dit ce qu'on veut, un modèle l'écrit, on relit, on relance. Ce n'est pas une raison de perdre les bonnes pratiques. C'est même le contraire : elles sont la seule chose qui tient encore quand ce n'est plus toi qui poses chaque ligne. Et il faut que l'IA les applique. Elle ne le fera pas par politesse. Il faut les lui dicter.

Un projet vibe coded n'est pas mauvais par nature. Il le devient si tu ne dis que le *quoi*. L'IA invente le *comment*, et c'est le comment qui pourrit : une string à la place d'un enum, un `null` à la place d'un bool, le même set recopié dans trois fichiers. À nous de dicter la rigueur. Ce qu'on veut implémenter, et surtout comment.

Mon *comment*, l'obsession que je recolle dans chaque prompt, c'est **les types stricts**. Le set de valeurs vit dans le type, une fois. Le compilateur refuse le reste. Le « contrat » (protobuf, client généré, `match` exhaustif) n'est pas un autre sujet. C'est le type, poussé jusqu'à l'autre fichier, jusqu'à l'autre machine.

Ça n'est pas venu d'un paper. Pendant des années j'ai écrit du PHP et du JavaScript. Le programme était un texte que la machine relisait à chaque exécution, et le contrat entre deux fonctions tenait souvent dans le nom d'un paramètre et un commentaire. Ça marche. On livre. On ajoute un `if`.

Puis j'ai passé du temps en Go, ensuite en Rust. Quelque chose s'est inversé. Ce n'est pas que ces langages soient magiques. C'est que le programme n'est plus un script qu'on interprète. Il est abaissé, une fois, vers quelque chose que la machine connaît déjà. Et si tu lui donnes un type au lieu d'une string, elle n'a plus à deviner. Rendre une app déterministe, c'est ça : **forcer le set de valeurs attendu dans le type**.

## Compilé, ça veut dire décidé trop tôt. C'est le point

Un langage lu (PHP, JS, Python au quotidien) te laisse remettre la décision à plus tard. La fonction prend une `string`. Au runtime tu vérifies, tu normalises, tu logs un warning, tu tombes dans un `else`. Le set de valeurs « réelles » vit dans la tête des gens, dans un wiki, dans un `if` copié trois fois.

Un langage compilé te force à décider avant. Pas « plus près du silicium » au sens romantique. Plus près de la représentation : ce qui existe, ce qui n'existe pas, combien ça pèse, qui a le droit de le construire.

Go m'a mis le pied dans la porte. Structs, constructeurs, `iota`, l'habitude de ne plus passer `map[string]interface{}` partout. Rust a verrouillé le verrou. Un `enum` n'est pas une liste de constantes. C'est un type somme. Le compilateur connaît le set. Un `match` sans `_` casse le build le jour où tu ajoutes une variante. En JavaScript, le même oubli se retrouve en production sous la forme d'un thème `"Dark"` qui n'allume jamais le dark mode, parce que quelqu'un avait écrit `"dark"`.

## Une string, ce n'est pas un contrat

Typer une fonction qui accepte une `string`, c'est déjà mieux que rien. Ça dit : pas un nombre, pas un objet. Ça ne dit pas _laquelle_.

```
setTheme("dark")
setTheme("Dark")
setTheme("DARK")
setTheme("sombre")
setTheme("")
```

Cinq appels, un seul est « le bon », et encore : ça dépend de qui a écrit le `if`. Tu dois valider. Tu dois documenter. Tu dois te souvenir de valider au prochain call site. Le jour où le set change, rien ne casse à la compilation. L'app devient non déterministe au sens qui m'intéresse : deux chemins, deux orthographes, deux comportements.

Si tu forces le set dans le type :

```rust
enum Theme { Dark, Light, System }

fn set_theme(theme: Theme) { /* ... */ }
```

`set_theme("sombre")` n'existe plus. Il n'y a plus rien à valider à l'entrée. Le set est le type. Les cinq call sites, l'UI, le fichier de config, le log : ils parlent de la même chose, ou le compilateur refuse.

C'est ça, l'obsession. Pas « j'aime les generics ». **Éliminer une classe entière d'erreurs en rendant l'état invalide inconstructible.**

Alexis King a un nom pour ça : *parse, don't validate*. Transformer une donnée douteuse en un type qui rend l'état invalide impossible, une fois, à la frontière, plutôt que la valider à chaque call site en la laissant repartir sous sa forme large. `set_theme` ne valide pas une string à l'entrée. Il refuse de compiler avec autre chose qu'un `Theme`.

## Un booléen, c'est oui ou non

Héritage Go, et ça me va. Chaque type a une *zero value*. Un `bool`, c'est `true` ou `false`. Pas `null`. Pas `undefined`. Pas un troisième état « on sait pas encore » qui se faufile dans un JSON. Si tu as vraiment besoin d'un troisième état, tu le nommes : un `Option<bool>`, un enum. Tu ne le laisses pas arriver tout seul.

Une `string` peut être vide. `""` veut dire : le champ est là, personne n'a rien mis. Ça existe. Ce qui n'existe pas, c'est la string absente. Si elle « n'existe pas », c'est qu'en amont quelqu'un n'a pas initialisé, pas sérialisé, pas collé le champ. Le trou n'est pas une valeur du type. C'est un bug.

Ça choque un peu quand on vient de JavaScript, où tout est `string | null | undefined` et où `if (name)` mélange le vide, le zéro et l'oubli. Go refuse ce mélange pour les scalaires. Rust te force à écrire `Option` quand l'absence est réelle. Les deux disent la même chose : **ne pas exister n'est pas une valeur**.

## Closed set, open set

Tout n'est pas un enum. C'est même le piège.

Un **ensemble fermé**, c'est un set dont les valeurs sont décidées par le code, pas par l'utilisateur ni par les données. Thème, méthode de collage, niveau de log, onglet de settings, état d'enregistrement. Ajouter une valeur, c'est éditer le type. Donc c'est un enum, déclaré une fois, converti exhaustivement à chaque frontière.

Un **ensemble ouvert**, c'est l'inverse. L'id d'un modèle, l'UID d'un micro, le nom d'un template que l'utilisateur a tapé. Les mettre dans un enum de 400 variantes, c'est mentir : le set n'est pas fermé. Le backend envoie un catalogue, le front lit une liste. Un newtype (`DeviceUid`) si tu veux que ça ne se mélange pas avec une autre `String`. Pas un enum « au cas où ».

J'ai lu les [Framework Design Guidelines](https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/enum) de Microsoft en travaillant là-dessus. Ce n'est pas le même langage. L'obsession est la même : enum pour un petit set fermé, pas pour les noms de tes amis, pas de variante réservée « pour plus tard », pas de sentinelle `All` fourrée dans le type de l'item. En Rust on va plus loin, parce que le `match` est exhaustif. Ajouter une variante _doit_ casser le build. En C#, un `switch` non exhaustif ne déclenche qu'un warning, pas un blocage à la compilation. Chez nous, ce warning-là, on le traite comme un crash. On ne met pas de `_ =>` « pour la compat ».

## Une seule source de vérité

Le corollaire, c'est organisationnel autant que technique.

Si le thème est un enum dans le backend, une string dans l'UI, et une autre string dans le fichier de settings, tu n'as pas un type. Tu as trois copies d'un set, et elles vont dériver. Je l'ai payé. Dans [Soufflé]({{< ref "/blog/souffle-transcription-locale-macos/" >}}), le contrat visuel et le contrat moteur doivent parler des mêmes onglets, des mêmes thèmes, des mêmes méthodes de collage. Une déclaration. Une conversion exhaustive de chaque côté de la frontière. Jamais la même littérale écrite deux fois.

Ça ressemble à ce que je raconte déjà sur la [state machine]({{< ref "/blog/souffle-transcription-locale-macos/" >}}) : quatre booléens décrivent seize combinaisons, dont la moitié sont du non-sens. Un seul type d'état, avec exactement les données qui ont du sens pour cet état, et les combinaisons invalides cessent d'exister. Le typing strict, c'est le même réflexe appliqué aux _valeurs_ plutôt qu'au _cycle de vie_.

## Le compilateur guide l'agent

En 2026, le patch ne vient souvent plus d'un humain qui a le graphe entier dans la tête. Tu demandes un fix à une IA. Elle lit trois fichiers, elle touche le quatrième, elle n'a pas ouvert le cinquième.

Sans types, le cinquième fichier continue de compiler. Un `if (theme === "dark")` quelque part dans un header « non lié » prend silencieusement l'autre branche. Tu le découvres en QA, ou en prod, ou jamais.

Oui, l'IA va peut-être grepper la valeur qu'elle vient de remplacer. Ou pas, justement. Selon le mot, elle récupère trois hits, ou trois mille. Ensuite elle les lira un par un. Ou elle en manquera. Ou elle s'arrêtera au vingtième parce que le contexte est plein. Le compilateur, lui, crie. Fichier, ligne, variante manquante. C'est ce qu'on veut.

Avec un enum et un `match` exhaustif, surtout en Rust, **le compilateur est le premier reviewer, et il est verbeux exprès**. Tu ajoutes `Theme::Sepia`. `rustc` ne dit pas « error ». Il dit : non-exhaustive patterns, voici les sites, voici la variante manquante, fichier et ligne. Souvent un fichier que l'agent n'avait pas dans son contexte. L'agent relit le diagnostic, patch le `match`, relance. La boucle tient dans le terminal.

C'est pour ça que, dans Soufflé, on interdit le `match` joker sur les enums métier amenés à grossir (`clippy::wildcard_enum_match_arm`). Un `_ =>` est un trou dans lequel l'agent (ou moi) fourre la nouvelle variante sans la traiter. Sur un enum qui ne bougera plus, la règle ne servirait à rien. Sur les nôtres, elle sert à chaque fois : on veut le crash. Le crash est le guide.

Un langage lu te donne un linter optionnel, qu'un agent peut ignorer ou contourner. Un compilateur typé te donne un mur. Tu ne merges pas tant que le mur est là. L'IA n'a pas à « penser à checker l'autre fichier ». Le type _est_ la checklist.

## Le même set, de l'autre côté du réseau

Jusqu'ici le contrat vit dans un process. Le réflexe ne s'arrête pas à la frontière du binaire.

J'aime beaucoup [Protocol Buffers](https://protobuf.dev/). Pour deux raisons, qui n'en sont qu'une.

La première : ça ouvre [gRPC](https://grpc.io/). Entre mes services, entre un worker et une API, c'est mon défaut : un protocole binaire, compact, streaming, où le contrat est sur le chemin obligatoire. Tu ne rediscutes pas le set à chaque hop.

Quand j'expose une API JSON, typiquement vers l'extérieur ou un navigateur qui ne parle pas protobuf, je passe par [OpenAPI](https://www.openapis.org/), avec génération de clients et un schéma forcé sur le payload. Je suis tout aussi strict là-dessus, un JSON Schema bien tenu, c'est déjà un contrat. Ce que je constate, c'est que par défaut JSON reste une string : le schéma est un add-on qu'on peut oublier, un `additionalProperties: true` que quelqu'un laissera passer, un `null` qui se faufile à la place d'un bool. gRPC met le contrat sur le chemin obligatoire par construction, sans qu'on ait à y penser.

La seconde : **c'est la même logique que le typing, prolongée entre deux machines**. Frontend et backend. Worker et API. Microservices. Le proto est la source. Tu déclares l'enum une fois, tu génères les clients dans chaque langage, tu t'assures d'une continuité. Le closed set traverse le réseau sans redevenir du folklore.

La même obsession peut descendre encore d'un cran, jusqu'à la base. Postgres a des enums natifs (`CREATE TYPE mood AS ENUM (...)`), pas un `varchar` avec un `CHECK` qu'on croise les doigts pour respecter. Le closed set existe dans le schéma, pas seulement dans le code applicatif. Reste à ne pas le recopier à la main côté Go ou Rust. Un ORM qui génère depuis la DB, [sqlc](https://sqlc.dev/) ou [SQLBoiler](https://github.com/volatiletech/sqlboiler) par exemple, part des migrations appliquées, lit le schéma réel, et sort les structs et les enums. La source, c'est la migration. Le code applicatif est un client généré, comme le proto. À l'inverse, un ORM code-first ne « recopie » pas forcément le type : si le modèle est bien typé, l'enum n'existe toujours qu'une fois, côté code. Le vrai risque est ailleurs. Lire un schéma existant pour en déduire du code est une réflexion mécanique et fidèle, comme `buf generate` pour un `.proto`. Déduire une migration d'un diff entre deux modèles est une inférence : un renommage de colonne se scaffold souvent en `DROP` puis `ADD`, ce qui perd les données, et il faut réécrire la migration à la main pour dire `RENAME COLUMN`. Cette migration éditée à la main, c'est elle qui tourne en prod, pas le modèle. Rien ne garantit qu'elle reste synchro au prochain diff. La source glisse, silencieusement, du modèle vers un fichier de migration devenu la vraie autorité, sans que le compilateur ne voie rien.

Concrètement, chez moi, ça ressemble à ça.

Sur le [thread art generator](https://github.com/damione1/thread-art-generator), le contrat c'est protobuf. `ArtStatus` (pending, processing, complete, failed, archived) n'est pas une string que le navigateur, l'API Go et le worker se refilent au pif. `buf generate` sort les types. Le browser parle protobuf en binaire sur `/rpc`. Les services se parlent en gRPC. Un statut ajouté dans le `.proto` casse la génération, puis le build, des deux côtés.

Sur [Maslow Desktop]({{< ref "/blog/maslow-desktop-controleur-cnc/" >}}), le même fichier `proto/maslow/v1/*.proto` alimente l'app, une API HTTP, du gRPC, et un serveur MCP pour qu'un LLM pilote la machine. La doc du site n'est pas écrite à la main : elle est générée depuis ces artefacts. Un `WsState` { disconnected, connected } n'existe qu'à un endroit. Les trois transports et l'agent MCP jouent avec le même set.

Soufflé, à l'époque Tauri, faisait la version in-process du même truc : un enum Rust, `specta`, un client TypeScript généré, CI qui refuse un `generated.ts` modifié à la main. Le moteur et l'UI n'avaient pas le droit de divergir sur `Theme` ou `PasteMethod`. Je suis en train de le sortir du webview, vers [Slint](https://slint.dev/), pour une UI native Metal. Le contrat n'a pas changé de métier, juste de forme : enum Slint, `match` Rust. L'idée n'a pas bougé : **une déclaration, des clients, pas de copie**.

Le proto n'est pas une paperasse d'archi. C'est le type, mis sur le fil.

## Ce que ça change en mémoire, vraiment

`"clipboard"` dans une string, c'est un pointeur, une longueur, et des octets sur le tas (ou l'équivalent). Tu le copies, tu le compares, tu le parses. À chaque couche, tu re-valides que c'est bien l'un des trois mots magiques.

Un enum `PasteMethod { Clipboard, Typing, Ax }`, au runtime, c'est un discriminant. Un petit entier. La machine ne transporte pas le mot « clipboard ». Le mot existe une fois, dans le compilateur, pour les humains. Les 10 000 call sites portent une étiquette de la taille d'un mot, pas la valeur lisible.

Ce n'est pas « les pointeurs c'est plus zen ». C'est : **tu paies le tag, tu ne paies pas la string, tu ne paies pas le `strcmp`**. Sur un hot path ça se mesure. Sur un settings panel, ça se sent surtout comme du calme : plus de round-trip `to_string` / parse, plus de `"Dark"` vs `"dark"`.

Rust pousse le même réflexe plus loin avec les types somme. `Mode::Meeting { id }` n'est pas `mode: String` plus `meeting_id: Option<String>`. Le payload n'existe que dans la variante où il a un sens. Tu n'alloues pas un `Option` vide pour les cas qui n'en ont pas besoin, et tu ne peux pas lire un id de meeting pendant une dictée.

## Déterministe, concrètement

Quand je dis déterministe, je ne parle pas de floating point ni de threads. Je parle de ça :

- L'app n'a pas d'état que le type refuse. « En réunion sans id de réunion » ne compile pas. Un bool n'est pas `null`. Une string absente n'est pas une valeur, c'est un oubli en amont.
- Un nouvel onglet, un nouveau thème, un nouveau niveau de log : le compilateur liste les sites. Pas un QA, pas un grep. Un agent qui n'avait pas le fichier dans son contexte se le fait recracher par `rustc`.
- L'UI et le moteur ne peuvent pas diverger sur le set, parce qu'il n'y a qu'un set. Deux machines non plus, si le set vit dans un `.proto`.
- Au runtime, tu n'interprètes pas un mot. Tu branches sur un tag que le compilateur a déjà vérifié. Sur le réseau, tu ne re-parses pas une string JSON pour retrouver le même enum.

Le programme lu te laisse être flou jusqu'à 3 h du matin. Le programme compilé, typé serré, te demande d'être précis à 10 h, une fois, et ensuite il tient le contrat tout seul.

C'est plus de friction le mardi. C'est moins de folklore le jeudi. J'ai choisi le mardi.
