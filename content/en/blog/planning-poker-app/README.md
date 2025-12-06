# Image Organization for Blog Posts

This directory uses Hugo's **Page Bundle** structure to keep all post-related files together.

## Structure

```
content/en/blog/planning-poker-app/
├── index.md                    # The blog post content
├── planning-poker-featured.jpg # Featured/cover image
└── other-image.png            # Any additional images
```

## Adding Images

1. **Place images in this directory** - Keep all images for this post in the same folder as `index.md`

2. **Reference in frontmatter** - Add to the `images` array in the frontmatter:

   ```yaml
   images: ["planning-poker-featured.jpg"]
   ```

3. **Reference in content** - Use relative paths in markdown:
   ```markdown
   ![Alt text](planning-poker-featured.jpg)
   ```

## Benefits

- ✅ All post files stay together
- ✅ Easy to move or delete posts
- ✅ Relative paths work automatically
- ✅ Hugo can process images efficiently
- ✅ Better organization as blog grows

## Image Naming Convention

- Featured images: `{post-slug}-featured.{ext}`
- Additional images: descriptive names like `architecture-diagram.png`
