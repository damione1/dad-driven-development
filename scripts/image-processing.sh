#!/bin/bash
# Image optimization using ImageMagick
set -e

CONTENT_DIR="content"
STATIC_DIR="static/images"

echo "Processing images..."

find "$CONTENT_DIR" -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) | while read img; do
    filename=$(basename "$img")
    dirname=$(dirname "$img")

    # Convert to WebP
    convert "$img" -quality 85 -define webp:lossless=false "${dirname}/${filename%.*}.webp"

    # Create thumbnails
    convert "$img" -resize 400x400^ -gravity center -extent 400x400 "${dirname}/thumb_${filename}"

    echo "Processed: $filename"
done

echo "Image processing complete!"
