(() => {
    // XXX: Not actually used, for now plots are initialized next to their elements
    htmx.defineExtension('uplot', {
        onEvent: function(name, evt) {
            if (name == "htmx:afterProcessNode") {
                const opts = {
                    width: evt.target.clientWidth,
                    height: evt.target.clientHeight
                }
                const data = []
                const plot = new uPlot(opts, data, evt.target)
                evt.target.uplot = plot;
                htmx.trigger(evt.target, "uplot:init")
            }
        }
    })
})()
