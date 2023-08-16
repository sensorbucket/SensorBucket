(() => {
    htmx.defineExtension('leaflet', {
        onEvent: function(name, evt) {
            const target = evt.target;

            switch (name) {
                case "htmx:afterProcessNode":
                    if (target.getAttribute("hx-ext") != "leaflet") {
                        return true;
                    }
                    if (target.leafet == undefined) {
                        target.innerHTML = "";
                        const m = L.map(target).setView([51.55, 3.9], 9)
                        L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
                            maxZoom: 19,
                            attribution: '© OpenStreetMap'
                        }).addTo(m);
                        target.leaflet = m
                    }
                    break;
                case "htmx:beforeCleanupElement":
                    if (target.getAttribute("hx-ext") != "leaflet") {
                        return true;
                    }
                    target.leaflet.remove()
                    break;
            }
        }
    })
})()
