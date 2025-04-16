import {sveltekit} from '@sveltejs/kit/vite';
import {defineConfig} from 'vite';
import UnoCSS from 'unocss/vite'
import {presetWind3, presetWind4, transformerVariantGroup} from "unocss";
import extractorSvelte from "@unocss/extractor-svelte";

export default defineConfig({
    plugins: [sveltekit(), UnoCSS({
        presets: [presetWind3()],
        transformers: [transformerVariantGroup()],
        extractors: [extractorSvelte()]
    })],
    server: {
        port: 3000,
        host: '0.0.0.0',
    },
});
