#!/usr/bin/env zx

import 'zx/globals';

const cwd = process.cwd();

const poloProSrcPath = path.resolve(cwd, '../polo-pro');
const poloProDestPath = path.resolve(cwd, 'third_party/polo-pro');
const poloSettingsPath = path.resolve(cwd, '.vscode/settings.json');

if (!fs.existsSync(poloProSrcPath))
    process.exit(0);

// Removing submodule folder
await $`rm -rf ${poloProDestPath}`;

// Linking source folder
await $`ln -sf ${poloProSrcPath} ${poloProDestPath}`;

// Changing IDE settings to enable pro features development
const settings = JSON.parse(fs.readFileSync(poloSettingsPath, 'utf8'));
settings['gopls']['build.buildFlags'] = ['-tags=pro'];

fs.writeFileSync(poloSettingsPath, JSON.stringify(settings, null, 4), 'utf8');