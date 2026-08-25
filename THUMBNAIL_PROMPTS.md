# Prompts thumbnails — Nano Banana / Gemini Image

Un prompt par article du blog. Style commun : illustration éditoriale vectorielle, fond sombre
(cohérent avec le thème Blowfish `noir`, `defaultAppearance = "dark"`), une métaphore visuelle par
article, une couleur d'accent par sujet.

Les prompts sont en anglais : les modèles d'image sont nettement plus fiables dans cette langue.

---

## Méthode recommandée

1. **Génère l'article n°1 en premier** (Snowflake + Terraform). C'est celui qui définit le mieux la
   grammaire visuelle.
2. **Réutilise l'image obtenue comme référence de style** pour les 10 suivants. Nano Banana accepte
   une image en entrée : joins-la au prompt avec la phrase

   > `Use the attached image as the exact style reference: same background, same line weight, same flat vector treatment, same grain. Only the subject changes.`

   C'est le seul moyen d'obtenir une série cohérente. Sans ça, tu auras 11 styles différents.
3. **Ratio** : 16:9. Cible finale 1200×630 (Hugo/Blowfish redimensionne, mais pars plus grand :
   1920×1080, puis `make optimize` convertit en WebP).
4. **Zéro texte dans l'image.** Les modèles d'image écrivent mal. Le titre de l'article s'affiche
   déjà sous la card. Tous les prompts ci-dessous demandent explicitement `no text`.

## Bloc de style commun

À coller **au début** de chaque prompt (ou à remplacer par l'image de référence après le n°1) :

```
Editorial spot illustration for a software engineering blog thumbnail, 16:9 landscape.
Flat vector style with a subtle film grain. Very dark charcoal background (#0b0b0d) with one
soft radial glow behind the subject. Restrained palette: near-black surfaces, off-white line
work, thin consistent 2-3px strokes, and a single accent colour. Geometric, schematic, calm.
Generous negative space, centred composition, safe margins on all sides. Faint dot-grid texture.
No text, no letters, no numbers, no logos, no brand marks, no photorealism, no 3D render,
no lens flare, no human faces.
```

---

## 1. Snowflake + Terraform — infrastructure as code

`content/{en,fr}/blog/snowflake-terraform-infrastructure-as-code/`
Accent : **bleu glacier `#7fd8ff` + violet `#8b5cf6`**

```
A large six-armed crystalline snowflake floating on the right side, but instead of ice its arms
are assembled from small modular rectangular blocks locked together like building bricks.
On the left, three stacked schematic module cards float in a vertical column, each connected by
thin violet dashed lines that feed into the snowflake and become its structure. The snowflake is
half-built: the lower-left arm is still forming, with a few blocks in transit along the lines.
Ice-blue line work, violet connector lines, near-black background with a soft cold glow behind
the crystal.
```

## 2. dbt — transformations traitées comme de l'infrastructure

`content/{en,fr}/blog/dbt-data-infrastructure-as-code/`
Accent : **corail `#ff694a`**

```
A directed acyclic graph flowing from left to right, drawn like an architect's blueprint.
On the far left, two dark database cylinders. From them, thin lines fan out into a lattice of
rounded rectangular nodes arranged in three columns, each column slightly higher-tech than the
last, converging into one single larger node on the right that glows warm coral.
Scaffolding poles and construction-blueprint guide marks are faintly visible behind the graph,
suggesting the pipeline is a building under construction. Off-white lines, coral accents on the
final node and its incoming edges only.
```

## 3. dbt — quand tes YAML deviennent ta gouvernance

`content/{en,fr}/blog/dbt-documentation-{governance,gouvernance}-yml/`
Accent : **ambre `#f5b342`**

```
An open structured document floating at a slight angle, its content rendered as abstract indented
line-blocks of varying length (never readable text, just bars). Three of those lines are marked
with small glowing amber tags. Rising out of the document, growing from its own lines, a large
geometric shield with a padlock at its centre — the shield's outline is literally drawn by the
document's lines continuing upward and curving into shape. Amber accent on the shield outline,
the tags and the lock. Everything else in off-white line work on near-black.
```

## 4. dbt — les tests en YAML

`content/{en,fr}/blog/dbt-tests-{constraints,contraintes}-yml/`
Accent : **vert `#4ade80` + rouge `#f87171`**

```
A vertical stream of small identical spheres falling from the top of the frame toward a single
horizontal sieve bar that spans the middle of the composition. Most spheres pass cleanly through
the slots and continue downward, glowing soft green with faint motion trails. Two spheres are
caught on top of the bar, sitting in their slot, glowing red with a small crack across them.
The sieve bar is a precise machined component, off-white line work, near-black background,
soft green glow below the bar and a small red glow above it.
```

## 5. Documenter une base source avec des LLM multi-agents

`content/{en,fr}/blog/dbt-document{-sources,er-source}-llm-multi-agent/`
Accent : **violet `#a78bfa`**

```
On the left, a tall dark database cylinder, closed and opaque, faintly cracked open with a dim
interior. Three hexagonal agent modules float in a vertical column in the centre, each with a
glowing violet core, each sending a thin dashed scanning beam into a different part of the
cylinder. On the right, their combined output stacks into three neat clean document cards with
abstract indented line-blocks, the top card fully complete and lit. Violet accent on the cores,
the beams and the finished card. Off-white line work, near-black background.
```

## 6. dbt-guard — mon premier package Python

`content/{en,fr}/blog/dbt-guard-package-python/`
Accent : **corail `#ff694a` + vert `#4ade80`**

```
A horizontal chain of three rounded schematic nodes connected left to right. Between the second
and the third node, the connecting link is violently fractured: a jagged red break with the
severed ends recoiling apart, and the third node is going dark. Hovering exactly in the gap, a
clean geometric shield intercepts the fracture, its edge catching the break, glowing calm green.
The contrast is between the chaotic red split and the precise green shield. Off-white line work,
near-black background, soft green glow around the shield.
```

## 7. Mon SSD est mort — backup chiffré sur Cloudflare R2

`content/{en,fr}/blog/{encrypted-backup,backup-chiffre}-cloudflare-r2/`
Accent : **orange `#f6821f` + rouge `#ef4444`**

```
Left third: a solid-state drive lying at an angle, cracked across its body, sparking, with red
light bleeding from the fracture and small loose data fragments spilling out of it in a chaotic
scatter. Centre: those scattered fragments are pulled into a large geometric padlock, and on the
far side of the lock they come out reordered into a tidy, evenly spaced column of identical
orange blocks. Right third: the ordered blocks rise into a simple geometric cloud outline that
receives them. The story reads left to right: chaos, lock, order, safety. Orange accent, red only
on the broken drive, near-black background.
```

## 8. Donner un accès SSH root à un LLM sur mon serveur Unraid

`content/{en,fr}/blog/llm-ssh-homelab-unraid/`
Accent : **turquoise `#2dd4bf` + orange `#f15a2c`**

```
A home server rack standing in a dark room, seen slightly from the side, its drive bays glowing
faint orange. From outside the frame on the left, a single thin turquoise tunnel of light reaches
across and connects to the rack, and travelling inside that tunnel is a simple geometric key,
mid-flight, about to arrive. Floating just above the rack, a small calm geometric core of
turquoise light watches over it — a machine presence, abstract, no face, no robot body.
Warm orange from the rack, cold turquoise from the tunnel, near-black room, soft glow on the
floor beneath the rack.
```

## 9. Maslow Desktop — un contrôleur CNC bâti sur une machine à états

`content/{en,fr}/blog/maslow-desktop-{cnc-controller,controleur-cnc}/`
Accent : **ambre `#f5a524`**

```
Four circular state nodes arranged in a perfect ring, connected by curved arrows that flow
clockwise around the circle — a state machine diagram. Inside the ring, at the centre, a small
precise CNC gantry: a horizontal rail with a carriage and a router bit pointing down at a piece
of stock, cutting. A faint amber toolpath curve traces across the whole background behind the
ring, passing behind the nodes. One of the four nodes is lit amber and filled, the other three
are dark outlines. Off-white line work, near-black background.
```

## 10. Planning Poker — app collaborative temps réel

`content/{en,fr}/blog/planning-poker-app/`
Accent : **bleu `#60a5fa`**

```
Three playing cards fanned out and floating in the centre of the frame, slightly overlapping,
the middle one raised and lit blue. The card faces are blank except for a single large geometric
pip in the centre — a circle on one, a triangle on the next, a small cluster of dots on the third.
Four abstract avatar tokens (simple circle-and-shoulders silhouettes, no faces) sit at the four
corners of the frame, each connected to the cards by a thin dashed blue line with a small pulse
travelling along it toward the centre. Off-white line work, near-black background, soft blue glow
behind the raised card.
```

> Variante si tu veux les valeurs Fibonacci : remplace la phrase sur les pips par
> `the three cards show the large numerals 5, 8 and 13 in a clean geometric sans-serif`.
> À vérifier à l'œil, c'est là que le modèle se plante le plus souvent.

## 11. Soufflé — transcription locale sur macOS

`content/{en,fr}/blog/souffle-{local-transcription-macos,transcription-locale-macos}/`
Accent : **ambre chaud `#e0a13c` + crème `#f3ead8`**

```
Exception to the palette: this one is warm. Left half: a horizontal audio waveform made of
rounded vertical bars of varying height, in warm amber and cream. Moving to the right, the bars
progressively stop being sound and become text — they stretch, flatten and settle into three neat
stacked paragraph blocks of abstract line-bars, alternating between two tints (amber and a muted
terracotta) to show two different speakers. The transformation is continuous and readable across
the frame: sound on the left, structured transcript on the right. Warm cream and amber on a dark
charcoal background, soft warm glow behind the waveform.
```

---

## Intégration dans Hugo

Une fois les PNG récupérés, une seule image par article, partagée entre EN et FR.

Nom de fichier, en suivant la convention de `IMAGE_ORGANIZATION.md` : `<slug-en>-featured.png`.

```bash
cd ~/Projects/dad-driven-development
# exemple pour un article
cp ~/Downloads/snowflake-terraform-featured.png \
   content/en/blog/snowflake-terraform-infrastructure-as-code/
cp ~/Downloads/snowflake-terraform-featured.png \
   content/fr/blog/snowflake-terraform-infrastructure-as-code/
```

Puis dans le frontmatter des **deux** `index.md` :

```yaml
images: ["snowflake-terraform-featured.png"]
```

C'est ce champ qui pilote la vignette des cards sur `/blog/` (vérifié dans le HTML généré),
et il sert aussi d'`og:image` au partage.

### Correspondance EN ↔ FR

| # | Slug EN | Slug FR |
|---|---|---|
| 1 | `snowflake-terraform-infrastructure-as-code` | `snowflake-terraform-infrastructure-as-code` |
| 2 | `dbt-data-infrastructure-as-code` | `dbt-data-infrastructure-as-code` |
| 3 | `dbt-documentation-governance-yml` | `dbt-documentation-gouvernance-yml` |
| 4 | `dbt-tests-constraints-yml` | `dbt-tests-contraintes-yml` |
| 5 | `dbt-document-sources-llm-multi-agent` | `dbt-documenter-source-llm-multi-agent` |
| 6 | `dbt-guard-package-python` | `dbt-guard-package-python` |
| 7 | `encrypted-backup-cloudflare-r2` | `backup-chiffre-cloudflare-r2` |
| 8 | `llm-ssh-homelab-unraid` | `llm-ssh-homelab-unraid` |
| 9 | `maslow-desktop-cnc-controller` | `maslow-desktop-controleur-cnc` |
| 10 | `planning-poker-app` | `planning-poker-app` |
| 11 | `souffle-local-transcription-macos` | `souffle-transcription-locale-macos` |

### Deux points relevés en passant

- **Planning Poker : image cassée.** Le frontmatter déclare
  `images: ["planning-poker-featured.jpg"]` alors que le fichier présent dans le bundle s'appelle
  `planning-pocker-featured.png` (faute de frappe + mauvaise extension). Résultat : la card de cet
  article n'a aucune vignette sur `/blog/`. À corriger dans les deux langues.
- **Soufflé et Maslow ont déjà une vignette** (`souffle-featured.png`,
  `maslow-run-toolpath.png`), et ce sont de vrais screenshots d'app. Une illustration générée sera
  probablement moins parlante que le screenshot pour ces deux-là. Si tu génères quand même les
  illustrations, garde le screenshot en premier dans `images:` et l'illustration en second, ou
  l'inverse selon ce que tu préfères voir dans la grille.
- Les 8 autres articles n'ont actuellement **aucune vignette** : leurs cards s'affichent sans
  image sur `/blog/`.
