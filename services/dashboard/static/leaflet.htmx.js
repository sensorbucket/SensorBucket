(() => {

    /** 
     * @param el{HTMLElement}
     *
     * @returns {boolean}
     */
    function isRootElement(el) {
        const ext = el.getAttribute("hx-ext") ?? ""
        const exts = ext.split(",")
        return exts.includes("leaflet")
    }

    /** 
     * @param target{HTMLElement}
     */
    function initLeaflet(target) {
        const mapEl = target.querySelector(".leaflet-container") || document.createElement("div")
        mapEl.innerHTML = ""
        mapEl.classList.add("w-full", "h-full")
        target.appendChild(mapEl)

        const latitude = parseFloat(target.getAttribute("data-latitude") ?? "0") || 51.55
        const longitude = parseFloat(target.getAttribute("data-longitude") ?? "0") || 3.9

        const m = L.map(mapEl).setView([latitude, longitude], 9)
        L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
            attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>, &copy; <a href="https://carto.com/attributions">CARTO</a>',
            maxZoom: 19,
        }).addTo(m);
        target.leaflet = m
        // Fire initialize event
        htmx.trigger(target, "leaflet:init")
    }

    function isLeafletInitialized(target) {
        return target.leaflet != undefined;
    }

    htmx.defineExtension('leaflet', {
        onEvent: function(name, evt) {
            /** @type{HTMLElement} */
            const target = evt.target;
            switch (name) {
                case "htmx:afterProcessNode":
                    if (isRootElement(target) && !isLeafletInitialized(target)) {
                        initLeaflet(target)
                    }
                    break;
                case "htmx:beforeCleanupElement":
                    if (isRootElement(target) && isLeafletInitialized(target)) {
                        target.leaflet.remove()
                    }
                    break;
            }
        },
    })

    class MapMarker extends HTMLElement {
        constructor() {
            super()
        }

        connectedCallback() {
            let parent = this.parentElement;
            while (parent != document.body) {
                parent.addEventListener('leaflet:init', this.leafletInit.bind(this))
                parent = parent.parentElement;
            }
        }

        leafletInit(evt) {
            if (evt.target.contains(this)) {
                this.initMarker(evt.target)
            }
        }

        initMarker(target) {
            this.map = target.leaflet;

            this.marker = L.marker([this.latitude, this.longitude]).addTo(this.map)

            if (this.label) {
                this.marker.bindTooltip(this.label)
            }

            this.marker.on('click', (evt) => {
                this.dispatchEvent(new CustomEvent(evt.type, { detail: evt.detail }))
            })
        }

        disconnectedCallback() {
            this.marker.remove()
        }

        get latitude() {
            return parseFloat(this.getAttribute('latitude'));
        }

        get longitude() {
            return parseFloat(this.getAttribute('longitude'));
        }

        get label() {
            return this.getAttribute('label');
        }
    }
    customElements.define('map-marker', MapMarker)
})()
