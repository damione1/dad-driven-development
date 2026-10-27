#!/usr/bin/env python3
"""Check the built site's SEO contract with only the Python standard library.

Usage: python3 scripts/check-seo.py public [--inventory path.json]
"""
import argparse
import collections
from html.parser import HTMLParser
import json
from pathlib import Path
import re
import sys
from urllib.parse import unquote, urlsplit
import xml.etree.ElementTree as ET


class Page(HTMLParser):
    def __init__(self):
        super().__init__()
        self.meta = {}
        self.canonical = []
        self.alternates = {}
        self.schemas = []
        self.links = []
        self.listing_links = []
        self._tags = []
        self.h1 = 0
        self._json = None

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if tag not in {"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr"}:
            self._tags.append((tag, set(attrs.get("class", "").split())))
        if tag == "meta":
            key = attrs.get("name", attrs.get("property", attrs.get("http-equiv", ""))).lower()
            self.meta[key] = attrs.get("content", "")
        if tag == "link":
            if attrs.get("rel") == "canonical":
                self.canonical.append(attrs.get("href", ""))
            if attrs.get("hreflang"):
                self.alternates[attrs["hreflang"]] = attrs.get("href", "")
        if tag == "a" and attrs.get("href"):
            self.links.append(attrs["href"])
            # Editorial introductions can legitimately link to articles that
            # are outside this taxonomy. Only compare the actual listing.
            if not any("prose" in classes for _, classes in self._tags):
                self.listing_links.append(attrs["href"])
        if tag == "h1":
            self.h1 += 1
        if tag == "script" and attrs.get("type") == "application/ld+json":
            self._json = ""

    def handle_data(self, data):
        if self._json is not None:
            self._json += data

    def handle_endtag(self, tag):
        if tag == "script" and self._json is not None:
            self.schemas.append(json.loads(self._json))
            self._json = None
        for index in range(len(self._tags) - 1, -1, -1):
            if self._tags[index][0] == tag:
                self._tags = self._tags[:index]
                break


def check(root):
    errors, pages, inventory = [], {}, []
    redirects = {}
    redirect_file = root / "_redirects"
    if redirect_file.is_file():
        for line in redirect_file.read_text().splitlines():
            fields = line.split()
            if len(fields) == 3 and fields[2] in {"301", "308"} and all(value.startswith("/") and "*" not in value and ":" not in value for value in fields[:2]):
                redirects[fields[0]] = fields[1]

    def require(condition, path, message):
        if not condition:
            errors.append(f"{path}: {message}")

    def local(url):
        path = unquote(urlsplit(url).path).lstrip("/")
        return root / (path + "index.html" if not Path(path).suffix else path)

    for file in sorted(root.rglob("*.html")):
        relative = file.relative_to(root).as_posix()
        page = Page()
        try:
            page.feed(file.read_text())
        except (ValueError, json.JSONDecodeError) as exc:
            errors.append(f"{relative}: invalid HTML/JSON-LD: {exc}")
            continue
        if "refresh" in page.meta or relative.endswith("404.html"):
            continue
        route = "/" + relative.removesuffix("index.html")
        if route in redirects:
            continue
        pages[route] = page
        require(len(page.canonical) == 1, route, "expected one canonical")
        canonical = page.canonical[0] if page.canonical else ""
        require(unquote(urlsplit(canonical).path) == route, route, "canonical must identify this page, including pagination")
        require(page.h1 == 1, route, "expected one main heading")
        require(bool(page.meta.get("description")), route, "missing description")
        require("Backend Software Engineer (Go)" not in page.meta.get("description", ""), route, "obsolete fallback description")
        require(page.meta.get("og:description") == page.meta.get("description"), route, "social/HTML description mismatch")
        image = page.meta.get("og:image", "")
        require(bool(image) and local(image).is_file(), route, "social image must exist in build")
        nodes = [node for schema in page.schemas for node in schema.get("@graph", [schema])]
        require(len(nodes) == 1, route, "expected one primary structured entity")
        if nodes:
            node = nodes[0]
            require(node.get("url") == canonical, route, "structured URL differs from canonical")
            is_article = "/blog/" in route and "/page/" not in route and route.rstrip("/").split("/")[-1] != "blog"
            require(node.get("@type") == "BlogPosting" if is_article else node.get("@type") != "BlogPosting", route, "structured type does not match page type")
            if is_article:
                require(node.get("image") == image, route, "article image mismatch")
                require(node.get("mainEntityOfPage") == canonical, route, "article mainEntityOfPage is not its URL")
                require(local(node.get("author", {}).get("url", "")).is_file(), route, "missing author profile")
                require({"en", "fr", "x-default"} <= page.alternates.keys(), route, "article translation association missing")
        require(page.meta.get("og:title") == page.meta.get("title"), route, "social/HTML title mismatch")
        if re.fullmatch(r"/(?:fr/)?(?:tags|categories)/[^/]+/", route):
            articles = sorted(set(urlsplit(link).path for link in page.listing_links if "/blog/" in link and "/page/" not in link and urlsplit(link).path.rstrip("/").split("/")[-1] != "blog" and not urlsplit(link).fragment))
            inventory.append({"url": canonical, "description": page.meta.get("description"), "articles": articles, "article_count": len(articles), "robots": page.meta.get("robots", "index, follow"), "decision": "retain; selective GSC consolidation applied"})

    for source, target in redirects.items():
        require(source != target and target not in redirects, source, "redirect loop or chain")
        require(local(target).is_file(), source, f"redirect target missing: {target}")
        if re.fullmatch(r"/(?:fr/)?(?:tags|categories)/[^/]+/", source):
            source_feed, target_feed = local(source + "index.xml"), local(target + "index.xml")
            require(source_feed.is_file() and target_feed.is_file(), source, "redirected archive feeds missing")
            if source_feed.is_file() and target_feed.is_file():
                def members(feed):
                    return {item.findtext("link") for item in ET.parse(feed).findall(".//item")}
                require(bool(members(source_feed)) and members(source_feed) == members(target_feed), source, "redirected archives have different article sets")

    for route, page in pages.items():
        for language, url in page.alternates.items():
            target = unquote(urlsplit(url).path)
            require(target in pages, route, f"hreflang {language} points to a missing page: {target}")
            if target in pages:
                require(page.canonical[0] in pages[target].alternates.values(), route, f"hreflang {language} has no return link")

    # RSS is rendered separately from the HTML head. Comparing membership
    # catches a paginator initialized with the wrong collection, which can
    # produce correct metadata but empty HTML archives.
    for route, page in pages.items():
        is_blog = route in ("/blog/", "/fr/blog/")
        is_term = re.fullmatch(r"/(?:fr/)?(?:tags|categories)/[^/]+/", route)
        if not (is_blog or is_term):
            continue
        feed = root / route.lstrip("/") / "index.xml"
        require(feed.is_file(), route, "archive RSS is missing")
        if not feed.is_file():
            continue
        expected = {unquote(urlsplit(item.findtext("link", "")).path) for item in ET.parse(feed).findall(".//item")}
        archive_pages = [candidate for path, candidate in pages.items() if path == route or re.fullmatch(re.escape(route) + r"page/\d+/", path)]
        listed = {unquote(urlsplit(link).path) for candidate in archive_pages for link in candidate.listing_links if "/blog/" in link and "/page/" not in link and urlsplit(link).path.rstrip("/").split("/")[-1] != "blog" and not urlsplit(link).fragment}
        require(bool(expected), route, "archive contains no articles in RSS")
        require(expected == listed, route, f"HTML archive membership differs from RSS: missing={sorted(expected - listed)}, extra={sorted(listed - expected)}")

    # Protect the high-value routes independently of the generic template checks.
    for route in ["/", "/fr/", "/resume/", "/fr/resume/"]:
        require(route in pages, route, "profile route missing")
        if route in pages:
            title = pages[route].meta.get("title", "")
            require("Damien GOEHRIG · Damien GOEHRIG" not in title and len(title) > 20, route, "profile title is not descriptive")
    for route in ["/blog/page/2/", "/fr/blog/page/2/"]:
        require(route in pages, route, "blog pagination missing")
    # All sitemap locations must resolve to canonical, indexable pages (or sitemap files).
    for file in root.rglob("sitemap.xml"):
        tree = ET.parse(file)
        for item in tree.getroot():
            url = item.findtext("{*}loc")
            if not url:
                continue
            target = unquote(urlsplit(url).path)
            if target.endswith(".xml"):
                require(local(url).is_file(), file, f"missing sitemap: {target}")
            else:
                require(target in pages, file, f"sitemap page missing: {target}")
                if target in pages:
                    require(url in pages[target].canonical, file, f"noncanonical sitemap entry: {url}")
                    require("noindex" not in pages[target].meta.get("robots", ""), file, f"noindex sitemap entry: {url}")
    return errors, pages, inventory


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("build", type=Path)
    parser.add_argument("--inventory", type=Path)
    args = parser.parse_args()
    errors, pages, inventory = check(args.build)
    if args.inventory:
        args.inventory.parent.mkdir(parents=True, exist_ok=True)
        groups = collections.defaultdict(list)
        for page in inventory:
            groups[tuple(page["articles"])].append(page["url"])
        args.inventory.write_text(json.dumps({"archives": inventory, "same_article_sets": [urls for articles, urls in groups.items() if articles and len(urls) > 1]}, ensure_ascii=False, indent=2) + "\n")
    if errors:
        print("\n".join(errors), file=sys.stderr)
        sys.exit(1)
    print(f"SEO checks passed: {len(pages)} pages; canonical, hreflang, schema, images, sitemaps, redirects and archive content.")
