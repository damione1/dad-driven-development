---
title: "Giving an LLM Root SSH Access to My Unraid Server"
date: 2026-05-16
draft: false
translationKey: "homelab-ssh"
description: "Why I let Claude Code operate my Unraid homelab over SSH: I know the devops concepts, I just don't have its speed. And how Ansible ended up reducing the need for remote operations."
tags: ["LLM", "Homelab", "Unraid", "DevOps", "SSH", "Ansible", "Claude Code"]
categories: ["DevOps", "AI"]
---

I have an Unraid server running about forty Docker containers, a few compose stacks, two AMD GPUs, VMs, Zigbee, media, home automation. In short, the typical homelab of someone who accumulated too much over the years. And I gave root SSH access to an LLM.

I already know what you're thinking. Yeah, root. On my domestic production.

Let me explain why.

## Six Years of Homelab, Then a Baby

I've had this server for roughly six years. In the beginning it was the classic playground: I tried things, I broke them, I fixed them, I changed my mind the following week. If Plex was down for two days, it didn't matter. If Home Assistant wouldn't boot, I'd eventually find the bug over a rainy evening.

Then I became a dad.

Overnight, my free time turned into thirty-minute blocks between naps and a load of diapers. When I really want to dig into a problem, I have to sacrifice a night. And nights, when you have a baby, are not the kind of resource you want to burn.

The problem is that in parallel, the server settled into the household's habits. Plex plays cartoons for the little one. Vaultwarden holds the whole family's passwords. Immich replaced Google Photos for the baby's pictures. Home Assistant runs the lights, the heating, the baby monitor.

It isn't my playground anymore, it's family infrastructure. And I find myself needing to hold a serious level of reliability on a stretched parent's schedule, with no desire to build a cluster that would make Hydro-Québec smile in January. No Kubernetes at home. Just one server, well configured, that doesn't fall over.

## What I Can Do, and What I Can't Do Fast Enough

Let's be clear about the nature of the problem, because this is where most discussions about "LLMs and ops" go off the rails.

I'm not delegating because I don't understand. I've been doing devops for years, it's part of my job. I know the concepts, I know what a network namespace is, a capability, a cgroup, a systemd unit, a Docker bridge, a device mapping. When Claude proposes a fix, I can read it, say why it's good or bad, and turn it down.

What I don't have is its speed.

Claude chains twenty `grep` calls with correct regexes in a few seconds. It composes a `find | xargs | awk` pipeline on the first try, without looking up syntax in a man page. It inspects thirty containers, cross-references the results and produces the conclusion while I'd still be recalling the argument order of `docker inspect --format`.

That's the whole difference. Between "I roughly know where to look" and "here's the answer" there are ten minutes for me and ten seconds for it. Over a full evening, that means nothing. On a thirty-minute block, it's the entire block.

Put another way: it isn't a skill I'm missing, it's throughput. And throughput can be outsourced.

## Encoding the Context Once

My server's context lived in my head, and it lived there badly. "Is Zigbee2MQTT on host or bridge?" "Which GPU does the transcoding?" "Where are the templates again?" Every time I debugged something at 11pm between two bottles, I spent a quarter of an hour recovering context before I could even start thinking about the problem.

So I wrote a Claude Code skill, `homelab-devops`, holding the hardware (CPU, RAM, GPUs, USB devices), the important paths, the network topology, the server's conventions, and an inventory of the containers with how each one is deployed.

When I say "look at why container X won't start," the skill loads on its own. Claude already knows how to reach the server, where the logs are, whether it's a compose container or an Unraid template, which paths are user shares and which are cache pools.

A side effect I hadn't planned for: writing that skill forced me to document, in plain words, what my server actually is. Documentation I had never taken the time to write in six years. It serves me as much as it serves the model.

## The Access Stays Mine

On the "giving root" part: it isn't permanent access, and it isn't set in stone.

Claude works through a dedicated SSH keypair that I can enable or revoke at will. I tell it where the private key is, it composes its own `ssh root@...` calls. At the end of a sensitive session, I pull the public key out of `authorized_keys` and the access simply stops existing.

Then, two modes of collaboration.

**Strict validation.** Every command that writes is approved before it runs. I read it, I understand it, I approve it. Slow, but it's the mode I use without exception the moment anything critical is involved.

**Direct mode.** Claude executes without asking, following a plan we defined together.

My rule is simple: direct mode is for actions whose worst outcome is "oops, have to redo it," never "oops, have to restore from a backup." A read-only diagnosis, an inventory audit, a grep through logs: no problem. Restarting a container Home Assistant depends on, on a Saturday morning, I approve every step even if it takes three times longer.

## Then I Stopped Doing Everything Over SSH

This is the most useful development in the whole story, and it specifically shrinks the role described above.

Operating a server over SSH, however fast, however well, is still imperative. Every intervention is a one-off, non-reproducible gesture, and the server's real state eventually drifts from what you think you know. A fast LLM doesn't solve that problem, it just makes it less painful. Which is already worth something, but it isn't a solution.

So I migrated all my containers to **[Ansible](https://www.ansible.com/)**. A repository that is now the source of truth for what runs on the server: it creates the Docker networks, pushes the compose files, generates the `.env` files from secrets encrypted with Ansible Vault, and deploys the stacks.

An amusing constraint along the way: Unraid has no Python 3, so Ansible's Docker modules can't run on the target. They run on my Mac and talk to the Docker daemon remotely, over SSH. It's unusual, but it works well and it avoids installing a runtime on an OS that isn't built for one.

What that changes in practice:

- **Compose files live in the repo, not on the server.** You edit here, you push. Editing directly on the machine is now forbidden, including for me.
- **Secrets are encrypted and versioned**, rather than scattered across environment variables nobody remembers setting.
- **Rebuilding a service is one command**, not an archaeology session.
- **The number of remote operations collapses.** Claude no longer needs to tinker with the server: it edits a versioned configuration, which I review like a pull request, and applying it is an `ansible-playbook` run.

That's the right order of things. The LLM is excellent at operating fast when you have to operate fast. But the real improvement isn't operating fast, it's no longer needing to operate. SSH access now exists for diagnosis and for the unexpected. The rest is declarative.

## The Guardrails

Giving root to an LLM isn't something you do carelessly.

**Confirmation before writing.** Read-only commands pass freely, anything that modifies gets approved.

**Nothing destructive without a written plan.** `rm -rf`, `docker system prune`, a git reset. Never without the plan spelled out, in dry-run where that exists.

**No access to the Unraid web UI.** It works over SSH on files and through the Docker CLI, where everything is traceable.

**Secrets stay out of context.** They live in `.env` files on the server or in the encrypted Ansible vault, never in plaintext in a repo or in a conversation. A dedicated skill documents how to use them: they're piped into the command that needs them, without ever printing their value. Claude knows a secret exists, knows where it is, knows how to hand it to a process, but it doesn't watch it go by and it doesn't end up in the session history.

That last one isn't optimal yet, and I won't pretend otherwise: it rests on documented discipline rather than a technical barrier, and nothing physically prevents one `cat` too many. It works in practice, and I'm working on something better.

## What Works Less Well

**Interactive commands.** A `y/N` prompt, a `docker exec -it`, an install wizard. You have to rephrase it non-interactively or do it by hand.

**Visual problems.** A Grafana graph misbehaving, a dashboard acting strange: the tool loses much of its value. I've ended up taking screenshots and describing what I was seeing.

**Silent failures.** A command that returns 0 without having done what you assumed. I've made a habit of always asking "verify that it actually applied" after any non-trivial fix.

**Architecture decisions.** "Should I move to TrueNAS," "which orchestrator": that's a discussion, not an execution. It helps me think, I decide.

## The Real Lesson

An LLM with SSH access to my homelab isn't a replacement sysadmin. It's a throughput multiplier for someone who already knows what he's doing but no longer has time to do it at human speed.

What I gain: a two-hour debug that fits into a thirty-minute block. Micro-tasks that had been sitting around for six months getting done in passing. "Why is this misbehaving" audits becoming routine because they cost nothing to launch, so I find problems before they wake me at 3am.

What I don't gain: responsibility. It's still my server, my data, my choices. If it breaks, I'm the one fixing it. The LLM is a tool, not a scapegoat. And if I delegate without reading, I learn nothing: I read what it does, I ask about the parameters I don't know.

In the end, my goal isn't for my server to be impressive. It's for it to be boring. To not fall over. To not wake me up. To let Plex play cartoons on a Saturday morning without dad having to get up.

Claude Code gets me closer, but not the way people imagine. Not by making the server more reliable in itself, but by giving me back enough time to make the small improvements that, stacked up, are the difference between infrastructure you endure and infrastructure that stands on its own. Ansible is the proof: it's exactly the kind of project I had been putting off for years.
