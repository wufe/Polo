#!/usr/bin/env zx

import 'zx/globals';

const cwd = process.cwd();

process.env.POLO_CWD = cwd;
process.env.GO_ENV = 'development';

const startServer = async () => {
    cd(`packages/server`);
    return await nothrow($`go run cmd/server/main.go`);
}

const startClient = async () => {
    cd(cwd);
    return await nothrow($`yarn serve`);
}

if (argv['server-only']) {
    await startServer();
} else if (argv['client-only']) {
    await startClient();
} else {
    await Promise.all([
        startServer(),
        startClient(),
    ]);
}