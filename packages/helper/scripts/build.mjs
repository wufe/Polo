#!/usr/bin/env zx

import 'zx/globals';
import {sassPlugin} from "esbuild-sass-plugin";
import esbuild from 'esbuild';

const publicPath = '/_polo_/public/';

const buildOutput = await esbuild.build({
    entryPoints: {
        'helper': 'src/components/app.tsx',
    },
    bundle: true,
    outdir: '../../public',
    entryNames: argv.dev ? '[dir]/[name]' : '[dir]/[name]-[hash]',
    plugins: [sassPlugin()],
    jsxFactory: 'h',
    jsxFragment: 'Fragment',
    inject: ['./scripts/preact-shim.js'],
    sourcemap: argv.dev,
    watch: argv.watch,
    metafile: true,
    minify: !argv.dev,
    publicPath,
});

const outputFiles = Object.keys(buildOutput.metafile.outputs)
    .map(x => path.basename(x))
    .filter(x => !x.endsWith('.map'));