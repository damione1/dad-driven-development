---
title: "Closer to the Machine: Getting Obsessed With Strict Types"
date: 2026-09-17
draft: true
translationKey: "strict-types"
description: "2026, we barely write the code anymore: not a reason to drop the rigor. You tell the AI the what and the how. My how is strict types: the set lives in the type, the compiler yells, the contract crosses machines."
tags: ["Rust", "Go", "Types", "Protobuf", "gRPC", "Software Design"]
categories: ["Development"]
images: ["closer-to-the-machine-strict-types-featured.jpg"]
---

2026 is the agentic hour. We barely write the code anymore. We say what we want, a model writes it, we reread, we rerun. That is not a reason to drop good practice. It is the opposite: practice is the only thing that still holds when you are no longer the one putting down every line. And the AI has to apply it. It will not do so out of politeness. You have to dictate it.

A vibe-coded project is not bad by nature. It becomes bad if you only say the *what*. The model invents the *how*, and the how is where it rots: a string instead of an enum, a `null` instead of a bool, the same set copied into three files. We dictate the rigor. What to implement, and especially how.

My *how*, the obsession I paste into every prompt, is **strict types**. The set of values lives in the type, once. The compiler refuses the rest. The “contract” (protobuf, generated client, exhaustive `match`) is not a second topic. It is the type, pushed to the other file, to the other machine.

It did not come from a paper. For years I wrote PHP and JavaScript. The program was a text the machine reread on every run, and the contract between two functions often lived in a parameter name and a comment. It ships. You add an `if`.

Then I spent time in Go, then in Rust. Something flipped. These languages are not magic. The program is no longer a script you interpret. It is lowered, once, into something the machine already knows. And if you give it a type instead of a string, it no longer has to guess. Making an app deterministic is that: **force the expected set of values into the type**.

## Compiled means decided too early. That is the point

A language you read at runtime (PHP, JS, day-to-day Python) lets you postpone the decision. The function takes a `string`. At runtime you validate, you normalize, you log a warning, you fall into an `else`. The “real” set of values lives in people's heads, in a wiki, in an `if` copied three times.

A compiled language makes you decide before. Not “closer to the silicon” in the romantic sense. Closer to the representation: what exists, what does not, how much it weighs, who is allowed to construct it.

Go got my foot in the door. Structs, constructors, `iota`, the habit of not passing `map[string]interface{}` everywhere. Rust turned the lock. An `enum` is not a list of constants. It is a sum type. The compiler knows the set. A `match` with no `_` breaks the build the day you add a variant. In JavaScript, the same miss shows up in production as a `"Dark"` theme that never turns the UI dark, because someone had written `"dark"`.

## A string is not a contract

Typing a function that accepts a `string` is already better than nothing. It says: not a number, not an object. It does not say *which* string.

```
setTheme("dark")
setTheme("Dark")
setTheme("DARK")
setTheme("sombre")
setTheme("")
```

Five calls, one of them is “the right one”, and even that depends on who wrote the `if`. You have to validate. You have to document. You have to remember to validate at the next call site. The day the set changes, nothing breaks at compile time. The app becomes non-deterministic in the sense I care about: two paths, two spellings, two behaviours.

Force the set into the type:

```rust
enum Theme { Dark, Light, System }

fn set_theme(theme: Theme) { /* ... */ }
```

`set_theme("sombre")` no longer exists. There is nothing left to validate at the door. The set *is* the type. The five call sites, the UI, the settings file, the log: they talk about the same thing, or the compiler refuses.

That is the obsession. Not “I like generics”. **Eliminate a whole class of bugs by making invalid state unrepresentable.**

Alexis King has a name for that: *parse, don't validate*. Turn a suspect piece of data into a type that makes the invalid state impossible, once, at the boundary, instead of validating it at every call site while it keeps travelling under its wide type. `set_theme` does not validate a string at the door. It refuses to compile with anything but a `Theme`.

## A boolean is yes or no

Go heritage, and it still sits right with me. Every type has a zero value. A `bool` is `true` or `false`. Not `null`. Not `undefined`. Not a third “we don't know yet” state sneaking in through JSON. If you genuinely need a third state, you name it: an `Option<bool>`, an enum. You do not let it arrive on its own.

A `string` can be empty. `""` means: the field is there, nobody put anything in it. That exists. What does not exist is the absent string. If it “does not exist”, someone upstream failed to initialize, serialize, or attach the field. The hole is not a value of the type. It is a bug.

That jars a bit when you come from JavaScript, where everything is `string | null | undefined` and `if (name)` mixes empty, zero, and forgotten. Go refuses that mix for scalars. Rust makes you write `Option` when absence is real. Both say the same thing: **not existing is not a value**.

## Closed set, open set

Not everything is an enum. That is the trap.

A **closed set** is a set whose values are decided by the code, not by the user or by data. Theme, paste method, log level, settings tab, recording state. Adding a value means editing the type. So it is an enum, declared once, converted exhaustively at every boundary.

An **open set** is the opposite. A model id, a microphone UID, a template name the user typed. Putting those in a 400-variant enum is a lie: the set is not closed. The backend ships a catalogue, the front reads a list. A newtype (`DeviceUid`) if you do not want it mixed up with some other `String`. Not an enum “just in case”.

I read Microsoft's [Framework Design Guidelines](https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/enum) while working through this. Different language. Same obsession: enums for small closed sets, not for the names of your friends, no reserved variant “for later”, no `All` sentinel stuffed into the item-kind type. Rust goes further, because `match` is exhaustive. Adding a variant *must* break the build. In C#, a non-exhaustive `switch` only triggers a warning, not a compile-time block. Here, we treat that warning like a crash. We do not add `_ =>` “for forward compat”.

## One source of truth

The corollary is organisational as much as technical.

If theme is an enum in the backend, a string in the UI, and another string in the settings file, you do not have a type. You have three copies of a set, and they will drift. I paid for that. In [Soufflé]({{< ref "/blog/souffle-local-transcription-macos/" >}}), the visual contract and the engine contract have to talk about the same tabs, the same themes, the same paste methods. One declaration. Exhaustive conversion on both sides of the boundary. Never the same literal written twice.

It is the same instinct I already wrote about with the [state machine]({{< ref "/blog/souffle-local-transcription-macos/" >}}): four booleans describe sixteen combinations, half of them nonsense. One state type, carrying exactly the data that makes sense for that state, and the illegal combinations stop existing. Strict typing is that reflex applied to *values*, not only to *lifecycle*.

## The compiler guides the agent

In 2026 the patch often no longer comes from a human who holds the whole graph in their head. You ask an AI for a fix. It reads three files, it touches a fourth, it never opened the fifth.

Without types, the fifth file still compiles. An `if (theme === "dark")` somewhere in an “unrelated” header silently takes the other branch. You find it in QA, or in prod, or never.

Yes, the model might grep for the value it just replaced. Or it might not. Depending on the word, it gets three hits, or three thousand. Then it will read them one by one. Or miss some. Or stop at the twentieth because the context window is full. The compiler just yells. File, line, missing variant. That is what we want.

With an enum and an exhaustive `match`, especially in Rust, **the compiler is the first reviewer, and it is verbose on purpose**. You add `Theme::Sepia`. `rustc` does not say “error”. It says: non-exhaustive patterns, here are the sites, here is the missing variant, file and line. Often a file the agent did not have in context. The agent reads the diagnostic, patches the `match`, reruns. The loop fits in the terminal.

That is why, in Soufflé, we deny the wildcard `match` on domain enums expected to grow (`clippy::wildcard_enum_match_arm`). A `_ =>` is a hole the agent (or I) stuffs the new variant into without handling it. On an enum that will not move again, the rule would not do anything. On ours, it earns its keep every time: we want the crash. The crash is the guide.

A language you read gives you an optional linter, which an agent can ignore or work around. A typed compiler gives you a wall. You do not merge while the wall is up. The AI does not have to “remember to check the other file”. The type *is* the checklist.

## The same set, on the other side of the network

So far the contract lives in one process. The reflex does not stop at the binary's edge.

I am very fond of [Protocol Buffers](https://protobuf.dev/). For two reasons that are really one.

The first: it unlocks [gRPC](https://grpc.io/). Between my own services, between a worker and an API, that is my default: a binary protocol, compact, streaming, where the contract sits on the mandatory path. You do not renegotiate the set on every hop.

When I expose a JSON API, typically outward-facing or to a browser that does not speak protobuf, I go through [OpenAPI](https://www.openapis.org/), with generated clients and a schema forced onto the payload. I am just as strict there. A well-kept JSON Schema is already a contract. What I notice is that, by default, JSON is still a string: the schema is an add-on you can forget, an `additionalProperties: true` someone will leave in, a `null` that sneaks in where a bool should be. gRPC puts the contract on the mandatory path by construction, without anyone having to think about it.

The second: **it is the same typing logic, extended between two machines**. Frontend and backend. Worker and API. Microservices. The proto is the source. You declare the enum once, you generate the clients in each language, you get continuity. The closed set crosses the network without turning back into folklore.

The same obsession can go one level lower still, into the database itself. Postgres has native enums (`CREATE TYPE mood AS ENUM (...)`), not a `varchar` with a `CHECK` you cross your fingers over. The closed set exists in the schema, not only in the application code. What is left is not re-typing it by hand on the Go or Rust side. An ORM that generates from the database, [sqlc](https://sqlc.dev/) or [SQLBoiler](https://github.com/volatiletech/sqlboiler) for instance, starts from the applied migrations, reads the real schema, and emits the structs and enums. The source is the migration. The application code is a generated client, same as the proto. The other way around, a code-first ORM does not necessarily duplicate the type: if the model is well typed, the enum still exists only once, on the code side. The real risk lies elsewhere. Reading an existing schema to derive code is a mechanical, faithful reflection, same as `buf generate` for a `.proto`. Deriving a migration from a diff between two models is an inference: a column rename often gets scaffolded as `DROP` then `ADD`, which loses data, so you rewrite the migration by hand to say `RENAME COLUMN` instead. That hand-edited migration is what runs in production, not the model. Nothing guarantees it stays in sync on the next diff. The source quietly slides from the model to a migration file that has become the real authority, and the compiler never sees it happen.

In my repos it looks like this.

On the [thread art generator](https://github.com/damione1/thread-art-generator), the contract is protobuf. `ArtStatus` (pending, processing, complete, failed, archived) is not a string the browser, the Go API, and the worker pass around by vibe. `buf generate` emits the types. The browser speaks binary protobuf on `/rpc`. Services speak gRPC to each other. A status added in the `.proto` breaks generation, then the build, on both sides.

On [Maslow Desktop]({{< ref "/blog/maslow-desktop-cnc-controller/" >}}), the same `proto/maslow/v1/*.proto` files feed the app, an HTTP API, gRPC, and an MCP server so an LLM can drive the machine. The docs site is not hand-written: it is generated from those artifacts. A `WsState` { disconnected, connected } exists in one place. The three transports and the MCP agent play with the same set.

Soufflé, in the Tauri era, was the in-process version of the same trick: a Rust enum, `specta`, a generated TypeScript client, CI that rejects a hand-edited `generated.ts`. Engine and UI were not allowed to drift on `Theme` or `PasteMethod`. I am pulling it out of the webview, onto [Slint](https://slint.dev/), for a native Metal UI. The contract did not change jobs, only shape: Slint enum, Rust `match`. The idea did not move: **one declaration, generated clients, no copies**.

The proto is not architecture paperwork. It is the type, put on the wire.

## What it actually changes in memory

People sometimes recite “performance” as a piety. Here is the concrete bit.

`"clipboard"` as a string is a pointer, a length, and bytes on the heap (or the moral equivalent). You copy it, you compare it, you parse it. At every layer you re-validate that it is one of the three magic words.

An enum `PasteMethod { Clipboard, Typing, Ax }`, at runtime, is a discriminant. A small integer. The machine does not carry the word “clipboard”. The word exists once, in the compiler, for humans. The 10,000 call sites carry a word-sized tag, not the readable value.

This is not “pointers are more zen”. It is: **you pay for the tag, you do not pay for the string, you do not pay for `strcmp`**. On a hot path you can measure it. On a settings panel you feel it as calm: no `to_string` / parse round-trip, no `"Dark"` versus `"dark"`.

Rust pushes the same reflex further with sum types. `Mode::Meeting { id }` is not `mode: String` plus `meeting_id: Option<String>`. The payload exists only in the variant where it belongs. You do not allocate an empty `Option` for the cases that do not need it, and you cannot read a meeting id while dictating.

## Deterministic, concretely

When I say deterministic, I do not mean floating point or threads. I mean this:

- The app has no state the type refuses. “In a meeting with no meeting id” does not compile. A bool is not `null`. An absent string is not a value, it is an upstream miss.
- A new tab, a new theme, a new log level: the compiler lists the sites. Not QA, not a grep. An agent that did not have the file in context gets it spat back by `rustc`.
- UI and engine cannot drift on the set, because there is only one set. Neither can two machines, if the set lives in a `.proto`.
- At runtime you do not interpret a word. You branch on a tag the compiler already checked. On the network, you do not re-parse a JSON string to recover the same enum.

The language you read lets you stay vague until 3am. The compiled, tightly typed program asks you to be precise at 10am, once, and then it holds the contract on its own.

More friction on Tuesday. Less folklore on Thursday. I picked Tuesday.
