import {svelte} from '@sveltejs/vite-plugin-svelte';
import {defineConfig} from 'vite';
import UnoCSS from 'unocss/vite';
import {presetWind4, transformerVariantGroup} from 'unocss';
import extractorSvelte from '@unocss/extractor-svelte';
import * as path from 'node:path';

export default defineConfig({
    base: "/importer",
    server: {
        host: true,
    },
    resolve: {
        alias: {
            '$lib': path.resolve(__dirname, './src/lib'),
            '$assets': path.resolve(__dirname, './src/assets'),
        }
    },
    plugins: [UnoCSS({
        presets: [presetWind4()],
        extractors: [extractorSvelte()],
        transformers: [transformerVariantGroup()],
    }),
        svelte(),
    ]
});
