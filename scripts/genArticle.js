#!/usr/bin/env node
// Article generation helper
// Adapted from n9o.xyz

const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
if (args.length < 2) {
    console.log('Usage: genArticle.js <section> <title>');
    console.log('Example: genArticle.js blog "My New Post"');
    process.exit(1);
}

const [section, title] = args;
const slug = title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
const date = new Date().toISOString().split('T')[0];

const frontMatter = `---
title: "${title}"
date: ${date}
draft: true
tags: []
stack: []
description: ""
showAuthor: true
showDate: true
showTableOfContents: true
---

Write your content here...
`;

const filePath = path.join('content', 'en', section, `${slug}.md`);
fs.mkdirSync(path.dirname(filePath), { recursive: true });
fs.writeFileSync(filePath, frontMatter);

console.log(`Created: ${filePath}`);
