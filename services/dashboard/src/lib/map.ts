import { getContext } from 'svelte';

export const mapKey = Symbol();
export const layerGroupKey = Symbol();

type getLayer = () => L.Map | L.LayerGroup;
export const getLayer: getLayer = () => (getContext(layerGroupKey) as getLayer)();
