# Dad Driven Development mark

`ddd-mark.svg` is the editable vector source. Three nested D/data strokes form one compact mark, using mint `#8be0c1` and charcoal `#222b2e`. It has no font, embedded bitmap, external resource or filter. The rounded tile stays visible on a light browser surface and blends into the blog's dark palette.

Design brief: a simple geometric monogram for Dad Driven Development / Damien Goehrig, legible at 16 and 32 px, using one or two flat colors and suitable for direct vector editing.

Created directly as native SVG; no image-generation API, raster tracing, stock art or third-party glyph was used. Chrome's SVG renderer produced the PNG exports; Pillow packaged the multi-resolution ICO. The Apple and Android exports have an opaque background and inset artwork inside the maskable safe circle.

The browser set lives under `static/`: SVG, 16/32/96 px PNG, ICO (16–256 px), 180 px Apple touch icon and 192/512 px web app icons. `layouts/partials/favicons.html` references these with a versioned query so a new identity is not hidden behind the one-year asset cache. Bump the query in that partial and `site.webmanifest` when changing the mark.
