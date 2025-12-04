# Image Organization Guide

## Overview

This Hugo site uses **Page Bundles** to organize images with blog posts and projects. Each post/project gets its own directory containing the content file and all related images.

## Structure

### Blog Posts

```
content/
└── en/
    └── blog/
        └── planning-poker-app/          # Post directory (page bundle)
            ├── index.md                 # Post content
            ├── planning-poker-featured.jpg  # Featured image
            └── other-diagram.png        # Additional images
```

### Projects

Same structure applies to projects:

```
content/
└── en/
    └── projects/
        └── thread-art-generator/
            ├── index.md
            └── screenshots/
```

## How to Add Images

### 1. Create the Page Bundle Structure

When creating a new post, Hugo will create a directory automatically if you use:
```bash
hugo new blog/my-new-post/index.md
```

Or manually:
```bash
mkdir -p content/en/blog/my-new-post
# Then create index.md inside
```

### 2. Add Images to the Directory

Place all images for that post in the same directory:
```bash
cp my-image.jpg content/en/blog/my-new-post/
```

### 3. Reference in Frontmatter

Add to the `images` array in the frontmatter:
```yaml
---
title: "My Post"
date: 2024-12-01
images: ["my-image.jpg"]
---
```

### 4. Reference in Content

Use relative paths in markdown:
```markdown
![Description](my-image.jpg)
```

## Benefits

- **Organization**: All related files stay together
- **Portability**: Easy to move or archive posts
- **No Broken Links**: Relative paths work automatically
- **Hugo Processing**: Images can be processed/resized by Hugo
- **Scalability**: Clean structure as content grows

## Image Processing

The site includes an image processing script (`scripts/image-processing.sh`) that:
- Converts images to WebP format
- Creates thumbnails
- Processes images in page bundles

Run with:
```bash
make optimize
```

## Naming Conventions

- **Featured images**: `{post-slug}-featured.{ext}`
- **Screenshots**: `screenshot-{number}.{ext}`
- **Diagrams**: `{descriptive-name}.{ext}`

## Static vs Page Bundle Images

- **Page Bundle** (recommended): Images specific to a post/project
  - Location: `content/en/blog/post-name/image.jpg`
  - Use for: Post images, project screenshots, diagrams

- **Static Directory**: Site-wide images
  - Location: `static/img/` or `assets/img/`
  - Use for: Profile photos, site logos, shared assets
