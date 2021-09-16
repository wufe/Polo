#!/usr/bin/env zx

import 'zx/globals';

try {
    await $`git diff --color --exit-code ./.vscode/settings.json`;
    await $`git diff --cached --color --exit-code ./.vscode/settings.json`;
} catch (e) {
    console.log(`\n${chalk.yellow('Settings file changed.\nUnstage your changes before committing.')}`);
    process.exit(1);
}