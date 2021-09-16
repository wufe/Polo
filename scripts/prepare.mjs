#!/usr/bin/env zx

import 'zx/globals';

// Installing Husky's hooks
await $`husky install`;

const cwd = process.cwd();
const poloProPath = path.resolve(cwd, 'packages/server/third_party/polo-pro');
const poloProGoModPath = path.join(poloProPath, 'go.mod');

if (!fs.existsSync(poloProGoModPath)) {
    await $`mkdir -p ${poloProPath}`;
    await $`echo "module github.com/wufe/polo-pro\n\ngo 1.15\n" > ${path.join(poloProPath, 'go.mod')}`;
}