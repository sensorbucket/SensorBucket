(() => {
    htmx.defineExtension('leaflet', {
        /**
         * Handle events for the leaflet extension.
         * @param {string} name - The name of the event.
         * @param {Event & {target: HTMLElement}} evt - The event object.
         */
        onEvent: function(name, evt) {
            const t = evt.target;
            console.log("leaflet ev: ", name, " with details: ", evt)
            if (name === "htmx:afterProcessNode") {
                let view = [
                    parseFloat(t.getAttribute("data-latitude")) || 51.55,
                    parseFloat(t.getAttribute("data-longitude")) || 3.89,
                ]
                let viewZoom = parseInt(t.getAttribute("data-zoom")) || 9

                var map = L.map(evt.target).setView(view, viewZoom);
                L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
                    maxZoom: 19,
                }).addTo(map);

                /** @type{{lat: number, lng: number}[]} */
                let markers = JSON.parse(evt.target.getAttribute("data-markers"))
                for (let ix = 0; ix < markers.length; ix++) {
                    const { lat, lng } = markers[ix]
                    L.marker([lat, lng]).addTo(map)
                }
            }
        }
    });
})()
