#!/bin/bash
# Image optimization using ImageMagick
set -e

# Check if ImageMagick is available
if ! command -v convert &> /dev/null; then
    echo "Warning: ImageMagick (convert) not found. Skipping image preprocessing."
    echo "Images will be processed by Hugo's built-in image processing instead."
    exit 0
fi

CONTENT_DIR="content"
STATIC_DIR="static/images"

echo "Processing images with ImageMagick..."

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
