#!/usr/bin/env zx

import 'zx/globals';

const cwd = process.cwd();

const poloProDestPath = path.resolve(cwd, 'third_party/polo-pro');
const poloSettingsPath = path.resolve(cwd, '.vscode/settings.json');

// Removing link to source folder
await $`rm -rf ${poloProDestPath}`;

// Restoring git submodule to pro repository
await $`git submodule update --remote`;

// Chaning IDE settings to disable pro features development
const settings = JSON.parse(fs.readFileSync(poloSettingsPath, 'utf8'));
settings['gopls']['build.buildFlags'] = ['-tags=standard'];

fs.writeFileSync(poloSettingsPath, JSON.stringify(settings, null, 4), 'utf8');