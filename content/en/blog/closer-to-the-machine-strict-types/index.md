---
title: "Strict Types: Rust Enums and Contracts That Hold"
date: 2026-09-17
lastmod: 2026-10-28
draft: false
translationKey: "strict-types"
description: "Why I use Rust enums, exhaustive matches and shared contracts to catch missed changes, especially when an AI agent writes the patch."
tags: ["Rust", "Go", "Types", "Protobuf", "gRPC", "Software Design"]
categories: ["Development"]
images: ["closer-to-the-machine-strict-types-featured.jpg"]
---

I use Rust enums to put the allowed values in the type, then let exhaustive `match` expressions tell me what I forgot. The same obsession follows me into generated clients, API contracts and database schemas: declare the set once, and make the rest of the code answer to it.

That matters even more when an AI agent writes the patch. We say what we want, a model writes it, we reread, we rerun. Good practice still has to come from somewhere. The agent won't apply it out of politeness. You have to dictate it.

A vibe-coded project isn't bad by nature. It becomes fragile if you only say the *what*. The model invents the *how*, and the how is where it rots: a string instead of an enum, an unexplained `null`, the same set copied into three files. My *how* is strict types. More precisely: **encode the constraints the type system can enforce, and check the rest at the boundaries**.

## From PHP and JavaScript to Go and Rust

It didn't come from a paper. For years I wrote PHP and JavaScript. In those projects, the contract between two functions often lived in a parameter name and a comment. It ships. You add an `if`.

Then I spent time in Go, then in Rust. Something flipped. Structs, constructors, the habit of no longer passing `map[string]interface{}` everywhere. Go got my foot in the door. Rust turned the lock.

The important distinction is what the type system can express and what the build actually checks. [PHP has enums](https://www.php.net/manual/en/language.enumerations.overview.php). [TypeScript can check discriminated unions exhaustively](https://www.typescriptlang.org/docs/handbook/2/narrowing#exhaustiveness-checking). Compiling to native code doesn't give you those guarantees by itself.

Go's named types and `iota` constants helped me make intent visible, but they don't close a set or make a `switch` exhaustive. A Rust `enum` does describe a fixed set of variants, including the data each variant carries. A `match` covering each variant explicitly breaks the build when you add one. That's the friction I want.

## Rust enums instead of magic strings

Typing a function that accepts a `string` is already better than nothing. It says: not a number, not an object. It does not say *which* string.

```
setTheme("dark")
setTheme("Dark")
setTheme("DARK")
setTheme("sombre")
setTheme("")
```

Five calls, one of them is “the right one”, and even that depends on who wrote the `if`. You have to validate. You have to document. You have to remember to validate at the next call site. The day the set changes, nothing breaks at compile time. Two paths, two spellings, two behaviours. The compiler has no reason to complain.

Force the set into the type:

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

`theme_label("sombre")` doesn't compile. Neither does adding `Theme::Sepia` to the enum while leaving this `match` alone:

```text
error[E0004]: non-exhaustive patterns: `Theme::Sepia` not covered
```

That's the [exhaustiveness check](https://doc.rust-lang.org/book/ch06-02-match.html). The compiler points at the forgotten branch. No grep, no checklist somebody has to remember.

The config file still contains text, though. The type doesn't reach backwards and fix a typo on disk.

## Parse once at the boundary

Alexis King's [*Parse, don't validate*](https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/) captures the useful part: turn doubtful input into a type that preserves what you've checked. Then pass that type around.

For the theme, the boundary can be this small:

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

The parser decides what to accept. Once it returns `Theme`, the rest of the code doesn't need to keep asking whether `"Dark"` means `"dark"`. A wildcard makes sense here: the input is any string. It would hide a missed variant in `theme_label`, where the input is our closed enum.

That is the obsession. **Eliminate a class of bugs by making invalid states unrepresentable in the code that consumes them.** You still need the parser. You just stop throwing away the information it gave you.

## A boolean is yes or no

Go heritage, and it still sits right with me. A `bool` has two values. If I need a third state, I want to name it: an `Option<bool>` in Rust, a pointer or a dedicated type in Go, an enum when the states deserve names.

That choice still needs care at the boundary. Go's [`encoding/json`](https://go.dev/src/encoding/json/decode.go) doesn't reject `null` when decoding into a plain `bool` or `string`: it leaves the value unchanged without an error. On a fresh struct, a missing boolean and an explicit `false` can look identical. If that distinction matters, the decoder and the model have to preserve it.

Same with strings. Empty and absent can mean different things. Sometimes absence is a bug; sometimes it's a PATCH request saying “leave this field alone”. I want that decision in the contract, rather than buried in an `if (name)` that mixes empty, zero and forgotten. **Absence deserves a type when it's part of the domain.**

## Closed set, open set

Not everything is an enum. That is the trap.

A **closed set** is a set whose values are decided by the code, not by the user or by data. Theme, paste method, log level, settings tab, recording state. Adding a value means editing the type. So it is an enum, declared once, converted exhaustively at every boundary.

An **open set** is the opposite. A model id, a microphone UID, a template name the user typed. Putting those in a 400-variant enum is a lie: the set is not closed. The backend ships a catalogue, the front reads a list. A newtype (`DeviceUid`) if you do not want it mixed up with some other `String`. Not an enum “just in case”.

I read Microsoft's [Framework Design Guidelines](https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/enum) while working through this. Different language. Same obsession: enums for small closed sets, not for the names of your friends, no reserved variant “for later”, no `All` sentinel stuffed into the item-kind type. In my Rust domain code, I want each variant named in the `match`. Adding a variant should break the build wherever it hasn't been considered. I don't add `_ =>` just to get the old code compiling again. External enums designed to evolve are a different contract; they need a deliberate unknown-case policy.

## One source of truth

The corollary is organisational as much as technical.

If theme is an enum in the backend, a string in the UI, and another string in the settings file, you do not have a type. You have three copies of a set, and they will drift. I paid for that. In [Soufflé]({{< ref "/blog/souffle-local-transcription-macos/" >}}), the visual contract and the engine contract have to talk about the same tabs, the same themes, the same paste methods. One declaration. Exhaustive conversion on both sides of the boundary. Never the same literal written twice.

It's the same instinct I wrote about with Soufflé's state machine: four booleans describe sixteen combinations, half of them nonsense. One state type, carrying exactly the data that makes sense for that state, and the illegal combinations stop existing. Strict typing is that reflex applied to *values*, not only to *lifecycle*.

## The compiler guides the agent

In 2026 the patch often no longer comes from a human who holds the whole graph in their head. You ask an AI for a fix. It reads three files, it touches a fourth, it never opened the fifth.

Without types, the fifth file still compiles. An `if (theme === "dark")` somewhere in an “unrelated” header silently takes the other branch. You find it in QA, or in prod, or never.

Yes, the model might grep for the value it just replaced. Or it might not. Depending on the word, it gets three hits, or three thousand. Then it will read them one by one. Or miss some. Or stop at the twentieth because the context window is full. The compiler just yells. File, line, missing variant. That is what we want.

With an enum and an exhaustive `match`, especially in Rust, **the compiler is the first reviewer, and it is verbose on purpose**. You add `Theme::Sepia`. `rustc` does not say “error”. It says: non-exhaustive patterns, here are the sites, here is the missing variant, file and line. Often a file the agent did not have in context. The agent reads the diagnostic, patches the `match`, reruns. The loop fits in the terminal.

That is why, in Soufflé, we deny the wildcard `match` on domain enums expected to grow (`clippy::wildcard_enum_match_arm`). A `_ =>` is a hole the agent (or I) stuffs the new variant into without handling it. On an enum that will not move again, the rule would not do anything. On ours, it earns its keep every time: we want the crash. The crash is the guide.

The wall only holds if the checks run. In my workflow, a compile error or a denied Clippy lint blocks the merge. The agent still has to fix the right behaviour; adding a fallback just to silence the diagnostic defeats the point. But it no longer has to remember to open every consumer. The type gives it the list.

## The same set, on the other side of the network

So far the contract lives in one process. The reflex does not stop at the binary's edge.

I am very fond of [Protocol Buffers](https://protobuf.dev/). For two reasons that are really one.

The first: it unlocks [gRPC](https://grpc.io/). Between my own services, between a worker and an API, that is my default: a binary protocol, compact, streaming, where the contract sits on the mandatory path. You do not renegotiate the set on every hop.

When I expose a JSON API, typically outward-facing or to a browser that doesn't speak protobuf, I go through [OpenAPI](https://www.openapis.org/), with generated clients and validation at the boundary. JSON can carry the same disciplined contract. It just needs that schema and those checks to be on the actual request path.

The second reason: **it's the same typing habit, extended between machines**. Frontend and backend. Worker and API. The proto is the source. You declare the enum once, generate the clients, and stop retyping the list in each language.

There is a catch worth keeping visible. [Proto3 enums are open](https://protobuf.dev/programming-guides/enum/), and adding a value is [compatible on the wire](https://protobuf.dev/programming-guides/proto3/#updating). It doesn't automatically break generation or the builds. An old client can receive a number it doesn't know. Generated types give us a shared vocabulary; converting them into a closed domain enum and deciding what to do with unknown values is still our job. So is deploying compatible versions.

### Down to the database

The same obsession can reach the database. Postgres has [native enums](https://www.postgresql.org/docs/current/datatype-enum.html): the named set exists in the schema. A [`CHECK` constraint](https://www.postgresql.org/docs/current/ddl-constraints.html#DDL-CONSTRAINTS-CHECK-CONSTRAINTS) also enforces allowed values. Both are real constraints. My concern is where the set is declared, and how it reaches the application without being copied by hand.

Tools can carry that contract in either direction. [sqlc](https://sqlc.dev/) generates code from SQL schema and queries; [SQLBoiler](https://github.com/volatiletech/sqlboiler) inspects the database. A code-first ORM can work too. What I don't want is three definitions pretending to be the authority.

Migrations are where that gets uncomfortable. A tool can mistake a rename for a drop and an addition. The migration you review and run is what changes production. Generated code helps with the shape; it doesn't relieve you of reading the migration.

### In my projects

On the [thread art generator](https://github.com/damione1/thread-art-generator), the contract is protobuf. [`ArtStatus`](https://github.com/damione1/thread-art-generator/blob/master/proto/art.proto) lives in the schema, with its unspecified, pending-image, processing, complete, failed and archived values. `buf generate` emits the Go and TypeScript types. The browser speaks binary protobuf on `/rpc`; services speak gRPC. Changing the set starts in one file, then the generated code and its consumers have to be checked. The generator doesn't promise that every consumer handled the new status.

On [Maslow Desktop]({{< ref "/blog/maslow-desktop-cnc-controller/" >}}), the same `proto/maslow/v1/*.proto` files feed the app, an HTTP API, gRPC, and an MCP server so an LLM can drive the machine. The docs site is not hand-written: it is generated from those artifacts. The shared [`WsState` schema](https://github.com/damione1/maslow-desktop/blob/master/proto/maslow/v1/common.proto) names unspecified, disconnected and connected. The transports and the MCP client get their wire vocabulary from that definition.

Soufflé, in the Tauri era, was the in-process version of the same trick: a Rust enum, `specta`, a generated TypeScript client, CI that rejects a hand-edited `generated.ts`. Engine and UI were not allowed to drift on `Theme` or `PasteMethod`. I am pulling it out of the webview, onto [Slint](https://slint.dev/), for a native Metal UI. The contract changes shape with it: Slint enums and explicit Rust conversions. Those mappings need checking too. The rule I keep coming back to is the same: **one authority, generated types where possible, explicit conversions where needed**.

The proto is not architecture paperwork. It is the type, put on the wire.

## What the types change at runtime

There is a representation benefit too, but it's easy to oversell.

A small fieldless enum lets code branch on a variant instead of repeatedly parsing a word. That doesn't mean every string allocates: a Rust `&str` can point at a literal. Nor is every enum a word-sized integer. Its [layout depends on its variants and representation](https://doc.rust-lang.org/reference/type-layout.html). I'd measure a hot path before claiming a speedup.

On a settings panel, the benefit I care about is already concrete: fewer `to_string` / parse round-trips, no `"Dark"` versus `"dark"`, fewer places maintaining the same list.

Sum types go further. `Mode::Meeting { id }` puts the meeting id in the variant where it belongs, instead of leaving `mode: String` and `meeting_id: Option<String>` to drift apart. The win is the impossible combination you can no longer construct. [`Option` doesn't imply a heap allocation](https://doc.rust-lang.org/std/option/index.html#representation); the reason to change the model is the invariant.

## More predictable changes

I used to call this making the app deterministic. That's too broad. Types don't prove the business logic correct, stop every race or upgrade a service running an old binary. What I want is more specific:

- A theme is one of the variants we declared. A meeting mode carries its meeting id. Absence is explicit where the domain needs it.
- Adding a variant makes the incomplete Rust matches fail. The agent gets file, line and missing case, including files it hadn't read.
- UI, engine and services share a source for their vocabulary, with conversions and compatibility rules checked at the boundaries.
- Strings stay where they belong: external input, identifiers, free text. Inside a closed domain, code handles the type.

The type asks me to decide earlier. The compiler and CI make me carry that decision through. When an agent writes the patch, that's one less piece of the design left to whatever happened to fit in its context.

More friction on Tuesday. Less folklore on Thursday. I picked Tuesday.
