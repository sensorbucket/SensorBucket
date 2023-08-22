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
    function assertLeaflet(target) {
        const mapEl = target.querySelector(".leaflet-container") || document.createElement("div")
        mapEl.innerHTML = ""
        mapEl.classList.add("w-full", "h-full")
        target.appendChild(mapEl)

        const latitude = parseFloat(target.getAttribute("data-latitude") ?? "0") || 51.55
        const longitude = parseFloat(target.getAttribute("data-longitude") ?? "0") || 3.9

        const m = L.map(mapEl).setView([latitude, longitude], 9)
        L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
            maxZoom: 19,
        }).addTo(m);
        target.leaflet = m
    }

    htmx.defineExtension('leaflet', {
        onEvent: function(name, evt) {
            /** @type{HTMLElement} */
            const target = evt.target;
            switch (name) {
                case "htmx:afterProcessNode":
                    if (isRootElement(target)) {
                        assertLeaflet(target)
                        break;
                    }
                    break;
                case "htmx:beforeCleanupElement":
                    if (isRootElement(target)) {
                        target.leaflet.remove()
                    }
                    break;
            }
        }
    })

    class MapMarker extends HTMLElement {
        constructor() {
            super()
        }

        connectedCallback() {
            this.innerHTML = ""
            setTimeout(() => {
                let parentEl = this.parentElement
                while (parentEl) {
                    if (parentEl.leaflet) break;
                    parentEl = parentEl.parentElement;
                }
                if (parentEl == undefined) {
                    throw 'No parent found with leaflet instance';
                }

                this.parent = parentEl
                this.map = this.parent.leaflet;

                this.marker = L.marker([this.latitude, this.longitude]).addTo(this.map)

                if (this.label) {
                    this.marker.bindTooltip(this.label)
                }

                this.marker.on('click', (evt) => {
                    this.dispatchEvent(new CustomEvent(evt.type, { detail: evt.detail }))
                })
            }, 10)
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
