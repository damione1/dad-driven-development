---
title: "Types stricts : des enums Rust aux contrats partagés"
date: 2026-09-17
lastmod: 2026-10-28
draft: false
translationKey: "strict-types"
description: "Pourquoi j’utilise les enums Rust, les match exhaustifs et les contrats partagés pour repérer les oublis, surtout quand une IA écrit le patch."
tags: ["Rust", "Go", "Types", "Protobuf", "gRPC", "Software Design"]
categories: ["Développement"]
images: ["closer-to-the-machine-strict-types-featured.jpg"]
---

J’utilise les enums Rust pour mettre les valeurs autorisées dans le type, puis laisser les `match` exhaustifs me dire ce que j’ai oublié. La même obsession me suit dans les clients générés, les contrats d’API et les schémas de base de données : déclarer le set une fois, et faire répondre le reste du code devant lui.

Ça compte encore plus quand un agent IA écrit le patch. On dit ce qu’on veut, un modèle l’écrit, on relit, on relance. Les bonnes pratiques doivent toujours venir de quelque part. L’agent ne les appliquera pas par politesse. Il faut les lui dicter.

Un projet vibe coded n’est pas mauvais par nature. Il devient fragile si tu ne dis que le *quoi*. L’IA invente le *comment*, et c’est le comment qui pourrit : une string à la place d’un enum, un `null` inexpliqué, le même set recopié dans trois fichiers. Mon *comment*, c’est les types stricts. Plus précisément : **encoder les contraintes que le système de types peut faire respecter, et vérifier le reste aux frontières**.

## De PHP et JavaScript à Go et Rust

Ça n’est pas venu d’un paper. Pendant des années j’ai écrit du PHP et du JavaScript. Dans ces projets, le contrat entre deux fonctions tenait souvent dans le nom d’un paramètre et un commentaire. Ça marche. On livre. On ajoute un `if`.

Puis j’ai passé du temps en Go, ensuite en Rust. Quelque chose s’est inversé. Structs, constructeurs, l’habitude de ne plus passer `map[string]interface{}` partout. Go m’a mis le pied dans la porte. Rust a verrouillé le verrou.

La distinction qui compte, c’est ce que le système de types peut exprimer et ce que le build vérifie vraiment. [PHP a des enums](https://www.php.net/manual/fr/language.enumerations.overview.php). [TypeScript peut vérifier l’exhaustivité des unions discriminées](https://www.typescriptlang.org/docs/handbook/2/narrowing#exhaustiveness-checking). Compiler vers du code natif ne donne pas ces garanties à lui seul.

Les types nommés et les constantes `iota` de Go m’ont aidé à rendre l’intention visible, mais ils ne ferment pas un ensemble et ne rendent pas un `switch` exhaustif. Un `enum` Rust décrit bien un ensemble fixé de variantes, avec les données que chacune porte. Un `match` qui nomme chaque variante casse le build quand tu en ajoutes une. C’est cette friction-là que je veux.

## Des enums Rust plutôt que des strings magiques

Typer une fonction qui accepte une `string`, c’est déjà mieux que rien. Ça dit : pas un nombre, pas un objet. Ça ne dit pas _laquelle_.

```
setTheme("dark")
setTheme("Dark")
setTheme("DARK")
setTheme("sombre")
setTheme("")
```

Cinq appels, un seul est « le bon », et encore : ça dépend de qui a écrit le `if`. Tu dois valider. Tu dois documenter. Tu dois te souvenir de valider au prochain call site. Le jour où le set change, rien ne casse à la compilation. Deux chemins, deux orthographes, deux comportements. Le compilateur n’a aucune raison de se plaindre.

Si tu forces le set dans le type :

```rust
enum Theme { Dark, Light, System }

fn theme_label(theme: Theme) -> &'static str {
    match theme {
        Theme::Dark => "dark",
        Theme::Light => "light",
        Theme::System => "system",
    }
}
```

`theme_label("sombre")` ne compile pas. Ajouter `Theme::Sepia` à l’enum sans toucher au `match` non plus :

```text
error[E0004]: non-exhaustive patterns: `Theme::Sepia` not covered
```

C’est la [vérification d’exhaustivité](https://doc.rust-lang.org/book/ch06-02-match.html). Le compilateur pointe la branche oubliée. Pas de grep, pas de checklist dont quelqu’un doit se souvenir.

Le fichier de config contient toujours du texte, par contre. Le type ne remonte pas dans le temps pour corriger une faute sur le disque.

## Parser une fois à la frontière

Le [*Parse, don’t validate*](https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/) d’Alexis King décrit bien le réflexe : transformer une entrée douteuse en un type qui conserve ce qu’on a vérifié. Puis faire circuler ce type.

Pour le thème, la frontière peut tenir dans ça :

```rust
impl TryFrom<&str> for Theme {
    type Error = &'static str;

    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value {
            "dark" => Ok(Self::Dark),
            "light" => Ok(Self::Light),
            "system" => Ok(Self::System),
            _ => Err("unknown theme"),
        }
    }
}
```

Le parser décide ce qu’il accepte. Une fois qu’il renvoie un `Theme`, le reste du code n’a plus à se demander si `"Dark"` veut dire `"dark"`. Le joker a sa place ici : l’entrée peut être n’importe quelle string. Dans `theme_label`, où l’entrée est notre enum fermé, il masquerait une variante oubliée.

C’est ça, l’obsession. **Éliminer une classe d’erreurs en rendant les états invalides impossibles à représenter dans le code qui les consomme.** Tu as toujours besoin du parser. Tu arrêtes juste de jeter l’information qu’il t’a donnée.

## Un booléen, c’est oui ou non

Héritage Go, et ça me va. Un `bool` a deux valeurs. Si j’ai besoin d’un troisième état, je veux le nommer : un `Option<bool>` en Rust, un pointeur ou un type dédié en Go, un enum quand les états méritent un nom.

Ce choix demande quand même de l’attention à la frontière. Le [`encoding/json`](https://go.dev/src/encoding/json/decode.go) de Go ne rejette pas `null` quand il décode vers un simple `bool` ou une `string` : il laisse la valeur inchangée, sans erreur. Sur une struct fraîche, un booléen absent et un `false` explicite peuvent avoir la même tête. Si la distinction compte, le décodeur et le modèle doivent la préserver.

Pareil pour les strings. Vide et absent peuvent vouloir dire deux choses différentes. Parfois l’absence est un bug ; parfois c’est une requête PATCH qui dit « ne touche pas à ce champ ». Je veux cette décision dans le contrat, plutôt qu’enterrée dans un `if (name)` qui mélange le vide, le zéro et l’oubli. **L’absence mérite un type quand elle fait partie du métier.**

## Ensemble fermé, ensemble ouvert

Tout n’est pas un enum. C’est même le piège.

Un **ensemble fermé**, c’est un set dont les valeurs sont décidées par le code, pas par l’utilisateur ni par les données. Thème, méthode de collage, niveau de log, onglet de settings, état d’enregistrement. Ajouter une valeur, c’est éditer le type. Donc c’est un enum, déclaré une fois, converti exhaustivement à chaque frontière.

Un **ensemble ouvert**, c’est l’inverse. L’id d’un modèle, l’UID d’un micro, le nom d’un template que l’utilisateur a tapé. Les mettre dans un enum de 400 variantes, c’est mentir : le set n’est pas fermé. Le backend envoie un catalogue, le front lit une liste. Un newtype (`DeviceUid`) si tu veux que ça ne se mélange pas avec une autre `String`. Pas un enum « au cas où ».

J’ai lu les [Framework Design Guidelines](https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/enum) de Microsoft en travaillant là-dessus. Ce n’est pas le même langage. L’obsession est la même : enum pour un petit set fermé, pas pour les noms de tes amis, pas de variante réservée « pour plus tard », pas de sentinelle `All` fourrée dans le type de l’item. Dans mon code métier Rust, je veux chaque variante nommée dans le `match`. Ajouter une variante doit casser le build là où elle n’a pas été prise en compte. Je ne rajoute pas un `_ =>` juste pour refaire compiler le vieux code. Les enums externes conçus pour évoluer ont un autre contrat ; ils demandent une politique explicite pour les cas inconnus.

## Une seule source de vérité

Le corollaire, c’est organisationnel autant que technique.

Si le thème est un enum dans le backend, une string dans l’UI, et une autre string dans le fichier de settings, tu n’as pas un type. Tu as trois copies d’un set, et elles vont dériver. Je l’ai payé. Dans [Soufflé]({{< ref "/blog/souffle-transcription-locale-macos/" >}}), le contrat visuel et le contrat moteur doivent parler des mêmes onglets, des mêmes thèmes, des mêmes méthodes de collage. Une déclaration. Une conversion exhaustive de chaque côté de la frontière. Jamais la même littérale écrite deux fois.

C’est le même réflexe que celui que je racontais avec la machine à états de Soufflé : quatre booléens décrivent seize combinaisons, dont la moitié sont du non-sens. Un seul type d’état, avec exactement les données qui ont du sens pour cet état, et les combinaisons invalides cessent d’exister. Le typing strict, c’est le même réflexe appliqué aux _valeurs_ plutôt qu’au _cycle de vie_.

## Le compilateur guide l’agent

En 2026, le patch ne vient souvent plus d’un humain qui a le graphe entier dans la tête. Tu demandes un fix à une IA. Elle lit trois fichiers, elle touche le quatrième, elle n’a pas ouvert le cinquième.

Sans types, le cinquième fichier continue de compiler. Un `if (theme === "dark")` quelque part dans un header « non lié » prend silencieusement l’autre branche. Tu le découvres en QA, ou en prod, ou jamais.

Oui, l’IA va peut-être grepper la valeur qu’elle vient de remplacer. Ou pas, justement. Selon le mot, elle récupère trois hits, ou trois mille. Ensuite elle les lira un par un. Ou elle en manquera. Ou elle s’arrêtera au vingtième parce que le contexte est plein. Le compilateur, lui, crie. Fichier, ligne, variante manquante. C’est ce qu’on veut.

Avec un enum et un `match` exhaustif, surtout en Rust, **le compilateur est le premier reviewer, et il est verbeux exprès**. Tu ajoutes `Theme::Sepia`. `rustc` te renvoie les sites incomplets, la variante, le fichier et la ligne. Souvent un fichier que l’agent n’avait pas dans son contexte. L’agent relit le diagnostic, patch le `match`, relance. La boucle tient dans le terminal.

C’est pour ça que, dans Soufflé, on interdit le `match` joker sur les enums métier amenés à grossir (`clippy::wildcard_enum_match_arm`). Un `_ =>` est un trou dans lequel l’agent (ou moi) fourre la nouvelle variante sans la traiter. Sur un enum qui ne bougera plus, la règle ne servirait à rien. Sur les nôtres, elle sert à chaque fois : on veut le crash. Le crash est le guide.

Le mur ne tient que si les checks tournent. Dans mon workflow, une erreur de compilation ou un lint Clippy interdit bloque le merge. L’agent doit encore corriger le bon comportement ; ajouter un fallback pour faire taire le diagnostic rate tout l’intérêt. Mais il n’a plus à se souvenir d’ouvrir chaque consommateur. Le type lui donne la liste.

## Les types partagés de l’autre côté du réseau

Jusqu’ici le contrat vit dans un process. Le réflexe ne s’arrête pas à la frontière du binaire.

J’aime beaucoup [Protocol Buffers](https://protobuf.dev/). Pour deux raisons, qui n’en sont qu’une.

La première : ça ouvre [gRPC](https://grpc.io/). Entre mes services, entre un worker et une API, c’est mon défaut : un protocole binaire, compact, streaming, où le contrat est sur le chemin obligatoire. Tu ne rediscutes pas le set à chaque hop.

Quand j’expose une API JSON, typiquement vers l’extérieur ou un navigateur qui ne parle pas protobuf, je passe par [OpenAPI](https://www.openapis.org/), avec génération de clients et validation à la frontière. JSON peut porter un contrat tout aussi rigoureux. Il faut juste que le schéma et les checks soient sur le vrai chemin des requêtes.

La seconde : **c’est le même réflexe de typage, prolongé entre les machines**. Frontend et backend. Worker et API. Le proto est la source. Tu déclares l’enum une fois, tu génères les clients, tu arrêtes de recopier la liste dans chaque langage.

Il y a un piège qui mérite de rester visible. [Les enums proto3 sont ouverts](https://protobuf.dev/programming-guides/enum/), et ajouter une valeur est [compatible sur le fil](https://protobuf.dev/programming-guides/proto3/#updating). Ça ne casse automatiquement ni la génération ni les builds. Un vieux client peut recevoir un nombre qu’il ne connaît pas. Les types générés nous donnent un vocabulaire commun ; les convertir vers un enum métier fermé et décider quoi faire des valeurs inconnues reste notre travail. Déployer des versions compatibles aussi.

### Jusqu’à la base de données

La même obsession peut atteindre la base. Postgres a des [enums natifs](https://www.postgresql.org/docs/current/datatype-enum.html) : le set nommé existe dans le schéma. Une [contrainte `CHECK`](https://www.postgresql.org/docs/current/ddl-constraints.html#DDL-CONSTRAINTS-CHECK-CONSTRAINTS) fait aussi respecter les valeurs autorisées. Les deux sont de vraies contraintes. Ce qui m’intéresse, c’est où le set est déclaré et comment il arrive dans l’application sans être recopié à la main.

Les outils peuvent porter ce contrat dans les deux sens. [sqlc](https://sqlc.dev/) génère du code depuis le schéma et les requêtes SQL ; [SQLBoiler](https://github.com/volatiletech/sqlboiler) inspecte la base. Un ORM code-first peut fonctionner aussi. Ce que je ne veux pas, c’est trois définitions qui prétendent toutes faire autorité.

Les migrations sont l’endroit où ça devient inconfortable. Un outil peut prendre un renommage pour une suppression et un ajout. La migration qu’on relit et qu’on exécute, c’est elle qui change la prod. Le code généré aide sur la forme ; il ne nous dispense pas de lire la migration.

### Dans mes projets

Sur le [thread art generator](https://github.com/damione1/thread-art-generator), le contrat c’est protobuf. [`ArtStatus`](https://github.com/damione1/thread-art-generator/blob/master/proto/art.proto) vit dans le schéma, avec ses valeurs unspecified, pending-image, processing, complete, failed et archived. `buf generate` sort les types Go et TypeScript. Le browser parle protobuf en binaire sur `/rpc` ; les services se parlent en gRPC. Changer le set commence dans un fichier, puis il faut vérifier le code généré et ses consommateurs. Le générateur ne promet pas que chaque consommateur a traité le nouveau statut.

Sur [Maslow Desktop]({{< ref "/blog/maslow-desktop-controleur-cnc/" >}}), les mêmes fichiers `proto/maslow/v1/*.proto` alimentent l’app, une API HTTP, du gRPC, et un serveur MCP pour qu’un LLM pilote la machine. La doc du site n’est pas écrite à la main : elle est générée depuis ces artefacts. Le [schéma partagé `WsState`](https://github.com/damione1/maslow-desktop/blob/master/proto/maslow/v1/common.proto) nomme unspecified, disconnected et connected. Les transports et le client MCP tirent leur vocabulaire réseau de cette définition.

Soufflé, à l’époque Tauri, faisait la version in-process du même truc : un enum Rust, `specta`, un client TypeScript généré, CI qui refuse un `generated.ts` modifié à la main. Le moteur et l’UI n’avaient pas le droit de diverger sur `Theme` ou `PasteMethod`. Je suis en train de le sortir du webview, vers [Slint](https://slint.dev/), pour une UI native Metal. Le contrat change de forme avec lui : enums Slint et conversions Rust explicites. Ces mappings ont aussi besoin d’être vérifiés. La règle à laquelle je reviens reste la même : **une autorité, des types générés quand c’est possible, des conversions explicites quand c’est nécessaire**.

Le proto n’est pas une paperasse d’archi. C’est le type, mis sur le fil.

## Ce que les types changent au runtime

Il y a un bénéfice de représentation aussi, mais on a vite fait de le survendre.

Un petit enum sans données permet de brancher sur une variante plutôt que de parser un mot à répétition. Ça ne veut pas dire que chaque string alloue : un `&str` Rust peut pointer vers un littéral. Ni que chaque enum fait la taille d’un mot machine. Sa [représentation dépend de ses variantes et de son layout](https://doc.rust-lang.org/reference/type-layout.html). Je mesurerais un hot path avant de promettre un gain de vitesse.

Sur un settings panel, le bénéfice qui m’intéresse est déjà concret : moins de round-trips `to_string` / parse, plus de `"Dark"` contre `"dark"`, moins d’endroits qui entretiennent la même liste.

Les types somme vont plus loin. `Mode::Meeting { id }` met l’id de réunion dans la variante où il a un sens, au lieu de laisser `mode: String` et `meeting_id: Option<String>` dériver chacun de son côté. Le gain, c’est la combinaison impossible qu’on ne peut plus construire. [`Option` n’implique pas une allocation sur le tas](https://doc.rust-lang.org/std/option/index.html#representation) ; c’est l’invariant qui justifie de changer le modèle.

## Des changements plus prévisibles

J’appelais ça rendre l’app déterministe. C’est trop large. Les types ne prouvent pas que la logique métier est juste, n’empêchent pas toutes les courses et ne mettent pas à jour un service qui tourne sur un vieux binaire. Ce que je veux est plus précis :

- Un thème fait partie des variantes déclarées. Un mode réunion porte son id de réunion. L’absence est explicite là où le métier en a besoin.
- Ajouter une variante fait échouer les `match` Rust incomplets. L’agent récupère le fichier, la ligne et le cas manquant, y compris dans les fichiers qu’il n’avait pas lus.
- L’UI, le moteur et les services partagent une source pour leur vocabulaire, avec des conversions et des règles de compatibilité vérifiées aux frontières.
- Les strings restent à leur place : entrées externes, identifiants, texte libre. À l’intérieur d’un domaine fermé, le code manipule le type.

Le type me demande de décider plus tôt. Le compilateur et la CI me demandent de tenir cette décision jusqu’au bout. Quand un agent écrit le patch, c’est un bout de design de moins laissé à ce qui tenait par hasard dans son contexte.

C’est plus de friction le mardi. C’est moins de folklore le jeudi. J’ai choisi le mardi.
