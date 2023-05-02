// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
// and what to do when importing types
declare namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface Platform {}
}


// default export
declare module 'leaflet?client' {
    import all from 'leaflet'
    export = all
}

// fallback
declare module '*?client'
declare module '*?server'
