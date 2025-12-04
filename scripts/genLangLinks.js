#!/usr/bin/env node
// Language link generation
// Adapted from n9o.xyz

const fs = require('fs');
const path = require('path');

function findContentFiles(dir, files = []) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });

    for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        if (entry.isDirectory()) {
            findContentFiles(fullPath, files);
        } else if (entry.name.endsWith('.md')) {
            files.push(fullPath);
        }
    }

    return files;
}

const enFiles = findContentFiles('content/en');

enFiles.forEach(enFile => {
    const relativePath = path.relative('content/en', enFile);
    const frFile = path.join('content/fr', relativePath);

    if (!fs.existsSync(frFile)) {
        console.log(`Missing French translation: ${relativePath}`);
    }
});
