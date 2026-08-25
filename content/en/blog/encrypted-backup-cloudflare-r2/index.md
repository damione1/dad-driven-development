---
title: "My SSD Died While I Was Out, and I Rebuilt My Backup"
date: 2026-08-08
draft: false
translationKey: "homelab-backup-r2"
description: "Fifty Uptime Kuma notifications, a dying SSD, a remote Claude session, and the encrypted off-site backup I built afterwards on Cloudflare R2 with restic and rclone."
tags:
  ["Homelab", "Unraid", "Backup", "Cloudflare R2", "restic", "DevOps", "Claude Code"]
categories: ["DevOps"]
images: ["encrypted-backup-cloudflare-r2-featured.jpg"]
---

I left the house for the day. Mid-morning, my phone starts buzzing and doesn't stop. About fifty Uptime Kuma notifications, all at once.

Every container is down, home automation with it. No maintenance scheduled. This isn't normal.

## Thirty Minutes, Remotely, From My Phone

I open the Claude app. A remote session is still live on my computer, the one I'd left open the day before. I describe the situation and ask it to deal with the problem.

What follows, I didn't steer. Thanks to the [skill that encodes my server's context]({{< ref "/blog/llm-ssh-homelab-unraid/" >}}), it took the initiative in the right order:

Diagnosis first. The SSD carrying every VM and every container is dying. Not a clean failure, the worst kind of death: the one where the drive still answers, intermittently, corrupting things along the way.

Then the action. Reboot the server. The SSD refuses to come back into its pool, so it mounts it manually outside the pool. Copies everything it holds onto the hard drive array. Restarts the containers and VMs on the recovered data.

Thirty minutes later, everything was running again.

And it comes back with the verdict: you're up, but your SSD is dead and needs replacing. Here are a few models I'd suggest. I checked SMART later that evening; it was right, everything was red. The drive was no longer trustworthy.

I opened the Amazon links and ordered one. As long as I was pulling the server out of the rack, I grabbed an extra hard drive for the storage array. I was due anyway.

Coming home that evening, home automation worked, Plex worked, my home cloud worked, all of it temporarily on the hard drives. The next day, both drives arrived. I hot-plugged them and told Claude to finish the job. It initialized them, rebuilt the SSD pool, expanded the hard drive array and put the containers back where they belonged. Back to normal.

## The Aftershock

I've had this homelab for six or seven years. In the beginning it was a laptop and ordinary hard drives, then I moved to a real second-hand server, with SSDs and NAS drives. In six years, not one drive had ever failed. In hindsight, I'd mostly been very lucky.

The stress arrived afterwards, once everything was fixed. I took inventory of what I would actually have lost if the rescue copy hadn't worked.

**The containers**, honestly, are almost nothing but Ansible configuration. One playbook and they come back. That's precisely why I did that migration.

**The VM** is Home Assistant, with a daily backup. Nothing serious.

**The databases**, on the other hand. Vaultwarden's, for instance. That one would be a real pain. Not catastrophic, in the sense that my passwords are also cached in the clients, but a real pain.

On the hard drive array there's redundancy, so it's less stressful. But you're never safe from two drives failing at once, or worse, silently. And that array holds two very different categories of data. A big pile of movies and shows, which isn't unrecoverable, just long to rebuild. And my photos and videos in Immich, plus the documents in my OpenCloud.

That, I cannot lose. Full stop.

## My 3-2-1 Had a Weak Link

I already had an off-site backup, because the 3-2-1 rule isn't negotiable when you host your own family photos. Three copies, two media, one off-site.

Except my off-site destination was OneDrive. Two problems.

It isn't **encrypted on my end**. My files are readable by the provider, with the telemetry and terms of service that come with it, on personal documents and family photos.

And it's **expensive for what I get out of it**, around 12 dollars per terabyte. A Microsoft subscription makes sense when you use the suite that comes with it. I wasn't using anything else on that account: I was paying for disk space at the price of an office suite I never open.

I figured I could do better with something more serious.

## Why a Bucket

I'm getting to know object storage well, I work with it a lot professionally, between AWS and Snowflake. The properties I care about here are exactly the ones that make it valuable in a company: storage is cheap, you can set lifecycle rules to automatically move cold data to a cheaper class, and capacity is effectively unlimited. No drive to buy, no quota to watch.

Except that for an individual, it stays pricier than you'd think.

On [AWS S3](https://aws.amazon.com/s3/pricing/), standard storage is 0.023 dollars per gigabyte per month. For a terabyte, that's roughly 23 dollars a month, double my OneDrive. And above all there are **egress fees**: about 0.09 dollars per gigabyte past the first hundred. Which means the day I actually need my backup, the day I have to pull a terabyte back because all my drives are dead, that's the day I get billed an extra 80 dollars. There's something deeply unhealthy about charging for the disaster.

I looked at the competition. Azure Blob, too expensive. Google Cloud Storage, too expensive as well. Archive classes like Glacier crater the storage price but add delays and retrieval fees that reproduce the same problem.

Then I remembered that Cloudflare had launched their offering.

## The Twist: R2

[Cloudflare R2](https://developers.cloudflare.com/r2/pricing/) does two things that change the equation.

**No egress fees.** Zero. Pulling all of my data back costs nothing. For a backup, that isn't a pricing detail, it's the heart of the matter: the restore scenario is the only one that truly counts, and it's the one everybody else charges a premium for.

**And cheap storage**, 0.015 dollars per gigabyte per month for standard, 0.010 for infrequent access, with an S3-compatible API, so all the usual tooling works unchanged.

Going by my bill, I land around 10 dollars a month.

Let's be honest: on price alone, the gap with OneDrive isn't spectacular. A bit cheaper, not revolutionary. Price isn't the argument, it's what lets the argument exist: at a comparable rate, I move from consumer disk space, unencrypted on my end and bundled with a suite I don't use, to object storage encrypted before it leaves, with no retrieval fees, and where I set the rules. At equal quality I would have hesitated. At equal price and better quality, there's no debate.

## Two Data Classes, Two Tools, Two Buckets

I didn't want to treat everything the same way, because not all my data carries the same risk profile.

**The rotating bucket**, handled by **[restic](https://restic.net/)**, holds what's re-generable or already replicated locally: Immich's PostgreSQL dumps, OpenCloud's configuration. Client-side AES-256 encryption, deduplication, and a standard retention of fourteen daily, eight weekly and six monthly snapshots. The server is allowed to prune it itself.

**The archive bucket**, handled by **[rclone](https://rclone.org/crypt/)** in encrypted mode, holds what cannot be lost: the Immich originals, the OpenCloud documents, the Vaultwarden dumps, and the Unraid flash drive configuration.

And here's a decision that matters: that bucket is fed with `copy`, never `sync`. The server can only **add**. A local deletion never propagates. On top of that, the bucket carries an indefinite immutability lock on Cloudflare's side.

The reasoning is simple: a backup that faithfully syncs with the source also replicates disasters. Ransomware encrypting my files, or one unfortunate command, propagates into the backup on the next pass. Here, even if my server is fully compromised, it cannot destroy the archive. Purging that archive is a manual admin operation, with a token that exists nowhere on the server.

It's the same principle I keep repeating at work about production data: a system's write permissions and the right to destroy its backups must never be carried by the same identity.

## The Detail That Matters: Where the Keys Live

This is the part tutorials rush through, and it's the one that decides whether your backup is worth anything.

The buckets are created in **Terraform**, versioned like the rest. Secrets are distributed according to what each one can destroy:

- The read-write token rendered onto the server can write, but **cannot lift the lock** on the archive.
- The admin tokens, the ones that can lift the lock and purge, live in the macOS Keychain behind Touch ID, and never set foot on the server.
- The Ansible vault password is in the Keychain too, with no plaintext file left anywhere.

And the non-negotiable rule: the passwords that decrypt the backups, restic's and rclone's, exist as an **offline copy**. Not in Vaultwarden, that would be circular, since Vaultwarden is precisely what I'm trying to restore. Not at Cloudflare either, putting the key next to the ciphertext defeats the point of encrypting. On paper, somewhere safe.

If the server disappears tomorrow, I still have the ciphertext at Cloudflare. Without those two passwords, that ciphertext is worth nothing.

The whole thing runs nightly: consistent dumps first, then the rotating bucket, then the archive.

## Twenty-Four Hours of Upload Later

The first push took a full day. Since then, only the deltas go out.

What I have now, and didn't before: my personal data encrypted before it leaves the house, at a host that cannot read it. No telemetry on my files, no model training on my family photos, no content scanning. And a restore I can trigger without being handed an invoice at the exact moment things are already going badly.

There's still the part everyone forgets and that I need to do properly: test the restore. A backup you've never restored is a hypothesis, not a backup.

One last thing, for perspective. The SSD and the hard drive I bought that day cost me almost as much as the entire server did six years ago. And that server is an EPYC 7302P, sixteen cores and thirty-two threads, on a Supermicro board, with 128 GB of RAM. Storage inflation stings. I hope I don't have to replace any more of it soon.
