---
title: "Building Planning Poker: A Real-time Collaboration App (That I Might Have Over-Engineered)"
date: 2025-11-15
draft: false
description: "Building a real-time Planning Poker app with Go, PocketBase, htmx, and WebSockets - and maybe going a bit overboard with the deployment infrastructure."
tags: ["Go", "PocketBase", "htmx", "WebSockets", "DevOps", "Terraform", "AWS"]
categories: ["Development", "Backend"]
images: ["planning-poker-featured.jpg"]
---

So I built this Planning Poker app. You know, that agile estimation thing where teams gather around and vote on story points? Yeah, I decided to make it web-based and real-time [pokerplanning.net](https://pokerplanning.net/). And then... well, let's just say I got a little ambitious with the deployment setup.

## The Idea

Planning Poker sessions are usually chaos in a conference room — someone's writing on a whiteboard, someone else is yelling their estimate, and half the team's zoning out. The problem's simple: you need something that works instantly, doesn't require sign-ups, and just... works.

So I grabbed Go for the backend (fast, compiled, solid concurrency model), PocketBase as my all-in-one database and auth layer, and added htmx + Alpine.js on the frontend for that reactive feel without building a full React app. WebSockets for real-time updates. Simple.

## Why PocketBase? (The WordPress Guy's Take)

Here's the thing — I spent years doing WordPress development. Loved it, hated it, the usual. When I transitioned into actual software development, I started doing the whole dance: pick an ORM, set up migrations, wire up authentication, build your router, handle database schema changes... it's exhausting.

PocketBase felt like the Go equivalent of what WordPress was to me — a sensible all-in-one baseline. You get a database layer, authentication, an admin UI, migrations that actually work, and an HTTP router. It's structured enough to move fast, but flexible enough to extend.

I didn't want to spend the next week messing around with Gorm, writing migration files, building auth middleware, and configuring a router. I wanted to ship something in a few days. PocketBase let me focus on the actual problem: making Planning Poker work in real-time.

And if this ever turns into something I want to monetize as a SaaS, I can build on top of this foundation. The headless nature of PocketBase means I can eventually swap out the frontend, add a pricing layer, whatever. It's a solid base.

## SQLite: The Unsung Database Hero

Real talk: I love Postgres. Seriously. But for this project? SQLite was the right call.

Most people don't realize this, but SQLite is the [most deployed](https://sqlite.org/mostdeployed.html) database engine in the world. Your phone probably has a dozen SQLite databases running right now. Android, iOS, Firefox, Chrome — all SQLite. It's boring, reliable, and seriously underestimated.

For a Planning Poker app that doesn't need horizontal scaling or complex multi-user transactions at massive scale, SQLite is _exactly_ what you need. No separate database server to manage, no connection pooling headaches, no "is the DB down?" crisis calls at 3 AM. It lives in a file. You can back it up, version it, move it around.

I made a deliberate choice this time: don't over-engineer the database. The app doesn't need the complexity of Postgres. SQLite handles 20,000 concurrent WebSocket connections on a t3.micro without breaking a sweat. That's plenty.

## What It Does

The core loop is pretty straightforward:

- Create a room, no login needed
- People join with a name
- You set up a voting round with Fibonacci or custom values
- Everyone votes at the same time
- Reveal and discuss
- Repeat

There's role-based stuff too — you can be a voter or just a spectator. Room creators can lock things down however they want. Everything syncs in real-time via WebSocket, so when someone votes, everyone sees it instantly (or sees that they voted, at least — votes are hidden until reveal).

The state management is clean. Rounds have states: voting, revealed, completed. Participants track who's where. Votes live in the database. Rooms auto-expire after 24 hours so you're not storing dead data forever.

## Performance? Yeah, I Checked

Here's where it gets a bit absurd. On a t3.micro (1 vCPU, 1GB RAM), this thing can handle:

- 2,000-3,000 concurrent rooms
- 20,000-30,000 WebSocket connections

That's... a lot more than you'd ever need for Planning Poker. But I wasn't about to ship something that couldn't handle its own success, right?

I built in async broadcasting with non-blocking message delivery, per-client send channels with buffering, fine-grained locking. There's monitoring endpoints so you can peek at real-time metrics. Slow clients get detected and cleaned up automatically. It's way too complex for what it does, but it _works_.

## The Deployment Rabbit Hole

And then I got to deployment.

Instead of just throwing it on a server with SSH, I decided to go full enterprise mode. Here's the setup:

1. **GitHub Actions** watches for git tags
2. Builds a multi-architecture Docker image
3. Pushes to GitHub Container Registry
4. Triggers an AWS Systems Manager Run Command
5. EC2 instance pulls the image
6. Docker Compose spins up the containers
7. Health checks validate everything's live

Zero SSH keys exposed. No ports open except HTTP/HTTPS. Everything's audited in CloudTrail. And yeah, I went full Terraform on the infrastructure — EC2, security groups, IAM roles, the whole thing.

Is it overkill for a Planning Poker app? 100%. Could I have just SSHed into a box and ran it? Yeah. But this way, deploying is literally just `git tag v1.0.0 && git push origin v1.0.0`. Two minutes later it's live. And you've got an audit trail, automatic rollback capabilities, and infrastructure-as-code. So really, I'm just being thorough.

## Tech Stack (The Real One)

- **Backend**: Go 1.25, PocketBase 0.30 (which is essentially Echo + SQLite + admin UI in one binary)
- **Frontend**: htmx 2.0 for AJAX/WebSocket, Alpine.js 3.14 for interactivity, Templ for templating
- **Database**: SQLite (bundled in PocketBase)
- **Deployment**: Docker, Docker Compose, Terraform, GitHub Actions, AWS SSM
- **Monitoring**: Built-in metrics endpoint, health checks

## Why This Stack?

PocketBase was the real discovery here. Everyone wants to build a backend, but PocketBase just gives you one. Database, migrations, admin UI, auth, all of it. I just had to wire up the WebSocket hub and business logic. That's time not spent messing around with boilerplate.

htmx + Alpine is underrated for this kind of project. No build pipeline headaches, no JavaScript framework fatigue, just declarative HTML attributes that do what you'd expect. Progressive enhancement, hypermedia, all that good stuff. You write less code, it's easier to follow, and your frontend doesn't become a maintenance nightmare in six months.

And Go's goroutines made the WebSocket hub trivial. Broadcasting messages to thousands of connections? Goroutines with channels. Done. That's the real advantage of Go for this use case — concurrency that doesn't drive you crazy.

## The Real Challenge

Honestly? Getting the state machine right. Rounds need to flow properly: voting → revealed → completed → new round. Participants need to stay synced. Connections drop and reconnect — you can't lose someone's vote because their WiFi dropped out.

That took more thought than the deployment did. The infrastructure was just... doing what infrastructure does. The hard part was making sure the voting logic was solid.

## So Is It Done?

Yeah, it works. You can run `make dev` and it spins up locally with live reload. There's integration tests. It's got metrics, health checks, proper error handling. You could actually use this to run Planning Poker sessions right now.

Is the deployment way too complex for a Planning Poker app that probably won't ever break a sweat at scale? Absolutely. But hey, it's there, it works, and now I've got a template for deploying Go apps to AWS without touching SSH. That's worth something, right?

The real lesson though? Don't over-engineer the database. SQLite did the job. PocketBase saved me from spending three days on boilerplate. Go's concurrency primitives made the hard parts easy. And sometimes the best infrastructure is the one that gets out of your way and just works.

You could actually use this to run Planning Poker sessions right now at [pokerplanning.net](https://pokerplanning.net/).
