// Code generated by qtc from "deviceListPage.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line services/dashboard/views/deviceListPage.qtpl:1
package views

//line services/dashboard/views/deviceListPage.qtpl:1
import "sensorbucket.nl/sensorbucket/services/core/devices"

//line services/dashboard/views/deviceListPage.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line services/dashboard/views/deviceListPage.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line services/dashboard/views/deviceListPage.qtpl:3
func (p *DeviceListPage) StreamBody(qw422016 *qt422016.Writer) {
//line services/dashboard/views/deviceListPage.qtpl:3
	qw422016.N().S(`
    <div class="mx-auto w-full xl:w-2/3 flex flex-col gap-6">
        <div class="grid lg:grid-cols-4 gap-x-4 gap-y-1">
            `)
//line services/dashboard/views/deviceListPage.qtpl:6
	StreamRenderFilters(qw422016, p.SensorGroup, false)
//line services/dashboard/views/deviceListPage.qtpl:6
	qw422016.N().S(`
        </div>

        <div class="bg-white border rounded-md" >
            `)
//line services/dashboard/views/deviceListPage.qtpl:10
	StreamRenderMap(qw422016, p.SensorGroup)
//line services/dashboard/views/deviceListPage.qtpl:10
	qw422016.N().S(`
        </div>

        <div class="bg-white border rounded-md min-h-96 max-h-[90vh] lg:max-h-[50vh] overflow-y-auto" id="device-table-wrapper">
            `)
//line services/dashboard/views/deviceListPage.qtpl:14
	StreamRenderDeviceTable(qw422016, p.Devices)
//line services/dashboard/views/deviceListPage.qtpl:14
	qw422016.N().S(`
        </div>
    </div>
`)
//line services/dashboard/views/deviceListPage.qtpl:17
}

//line services/dashboard/views/deviceListPage.qtpl:17
func (p *DeviceListPage) WriteBody(qq422016 qtio422016.Writer) {
//line services/dashboard/views/deviceListPage.qtpl:17
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services/dashboard/views/deviceListPage.qtpl:17
	p.StreamBody(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:17
	qt422016.ReleaseWriter(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:17
}

//line services/dashboard/views/deviceListPage.qtpl:17
func (p *DeviceListPage) Body() string {
//line services/dashboard/views/deviceListPage.qtpl:17
	qb422016 := qt422016.AcquireByteBuffer()
//line services/dashboard/views/deviceListPage.qtpl:17
	p.WriteBody(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:17
	qs422016 := string(qb422016.B)
//line services/dashboard/views/deviceListPage.qtpl:17
	qt422016.ReleaseByteBuffer(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:17
	return qs422016
//line services/dashboard/views/deviceListPage.qtpl:17
}

//line services/dashboard/views/deviceListPage.qtpl:19
func StreamRenderFilters(qw422016 *qt422016.Writer, sensorGroup *devices.SensorGroup, oob bool) {
//line services/dashboard/views/deviceListPage.qtpl:19
	qw422016.N().S(`
    <fieldset
        id="filters"
        class="relative col-span-2"
        `)
//line services/dashboard/views/deviceListPage.qtpl:23
	if oob {
//line services/dashboard/views/deviceListPage.qtpl:23
		qw422016.N().S(`
            hx-swap-oob="true"
        `)
//line services/dashboard/views/deviceListPage.qtpl:25
	}
//line services/dashboard/views/deviceListPage.qtpl:25
	qw422016.N().S(`
    >
        <label for="sensor-group-search" class="ml-1 -mb-1 block"><small class="text-xs text-slate-500">Sensor group</small></label>
        `)
//line services/dashboard/views/deviceListPage.qtpl:28
	if sensorGroup == nil {
//line services/dashboard/views/deviceListPage.qtpl:28
		qw422016.N().S(`
            <input
                type="text" name="search" id="sensor-group-search"

                class="block w-full px-2 py-1 border rounded-md bg-white placeholder:text-slate-600"
                hx-trigger="keyup changed delay:500ms, search"
                hx-get="/overview/sensor-groups"
                hx-target="next ul"
                _="on blur wait 50ms add .hidden to next <ul/> on focus remove .hidden from next <ul/>"
            />
            <ul class="absolute top-full left-0 right-0 block rounded-md rounded-t-none bg-white z-[4000]
                        text-sm">
                
            </ul>
        `)
//line services/dashboard/views/deviceListPage.qtpl:42
	} else {
//line services/dashboard/views/deviceListPage.qtpl:42
		qw422016.N().S(`
            <a
                href="/overview"
                hx-get="/overview/devices/table"
                hx-target="#device-table-wrapper"
                class="w-full px-2 py-1 border rounded-md text-primary-600 flex justify-between items-center"
            >
                `)
//line services/dashboard/views/deviceListPage.qtpl:49
		qw422016.E().S(sensorGroup.Name)
//line services/dashboard/views/deviceListPage.qtpl:49
		qw422016.N().S(`
                <iconify-icon icon="charm:cross" class="text-rose-500"></iconify-icon>
            </a>
        `)
//line services/dashboard/views/deviceListPage.qtpl:52
	}
//line services/dashboard/views/deviceListPage.qtpl:52
	qw422016.N().S(`
    </fieldset>
`)
//line services/dashboard/views/deviceListPage.qtpl:54
}

//line services/dashboard/views/deviceListPage.qtpl:54
func WriteRenderFilters(qq422016 qtio422016.Writer, sensorGroup *devices.SensorGroup, oob bool) {
//line services/dashboard/views/deviceListPage.qtpl:54
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services/dashboard/views/deviceListPage.qtpl:54
	StreamRenderFilters(qw422016, sensorGroup, oob)
//line services/dashboard/views/deviceListPage.qtpl:54
	qt422016.ReleaseWriter(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:54
}

//line services/dashboard/views/deviceListPage.qtpl:54
func RenderFilters(sensorGroup *devices.SensorGroup, oob bool) string {
//line services/dashboard/views/deviceListPage.qtpl:54
	qb422016 := qt422016.AcquireByteBuffer()
//line services/dashboard/views/deviceListPage.qtpl:54
	WriteRenderFilters(qb422016, sensorGroup, oob)
//line services/dashboard/views/deviceListPage.qtpl:54
	qs422016 := string(qb422016.B)
//line services/dashboard/views/deviceListPage.qtpl:54
	qt422016.ReleaseByteBuffer(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:54
	return qs422016
//line services/dashboard/views/deviceListPage.qtpl:54
}

//line services/dashboard/views/deviceListPage.qtpl:56
func StreamSensorGroupSearch(qw422016 *qt422016.Writer, sgs []devices.SensorGroup) {
//line services/dashboard/views/deviceListPage.qtpl:56
	qw422016.N().S(`
    `)
//line services/dashboard/views/deviceListPage.qtpl:57
	for _, sg := range sgs {
//line services/dashboard/views/deviceListPage.qtpl:57
		qw422016.N().S(`
        <li>
            <a
                href="?sensor_group=`)
//line services/dashboard/views/deviceListPage.qtpl:60
		qw422016.N().DL(sg.ID)
//line services/dashboard/views/deviceListPage.qtpl:60
		qw422016.N().S(`"
                hx-get="/overview/devices/table?sensor_group=`)
//line services/dashboard/views/deviceListPage.qtpl:61
		qw422016.N().DL(sg.ID)
//line services/dashboard/views/deviceListPage.qtpl:61
		qw422016.N().S(`"
                hx-target="#device-table-wrapper"
                class="block hover:bg-primary-100 p-2"
            >`)
//line services/dashboard/views/deviceListPage.qtpl:64
		qw422016.E().S(sg.Name)
//line services/dashboard/views/deviceListPage.qtpl:64
		qw422016.N().S(`</a>
        </li>
    `)
//line services/dashboard/views/deviceListPage.qtpl:66
	}
//line services/dashboard/views/deviceListPage.qtpl:66
	qw422016.N().S(`
`)
//line services/dashboard/views/deviceListPage.qtpl:67
}

//line services/dashboard/views/deviceListPage.qtpl:67
func WriteSensorGroupSearch(qq422016 qtio422016.Writer, sgs []devices.SensorGroup) {
//line services/dashboard/views/deviceListPage.qtpl:67
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services/dashboard/views/deviceListPage.qtpl:67
	StreamSensorGroupSearch(qw422016, sgs)
//line services/dashboard/views/deviceListPage.qtpl:67
	qt422016.ReleaseWriter(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:67
}

//line services/dashboard/views/deviceListPage.qtpl:67
func SensorGroupSearch(sgs []devices.SensorGroup) string {
//line services/dashboard/views/deviceListPage.qtpl:67
	qb422016 := qt422016.AcquireByteBuffer()
//line services/dashboard/views/deviceListPage.qtpl:67
	WriteSensorGroupSearch(qb422016, sgs)
//line services/dashboard/views/deviceListPage.qtpl:67
	qs422016 := string(qb422016.B)
//line services/dashboard/views/deviceListPage.qtpl:67
	qt422016.ReleaseByteBuffer(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:67
	return qs422016
//line services/dashboard/views/deviceListPage.qtpl:67
}

//line services/dashboard/views/deviceListPage.qtpl:69
func StreamRenderDeviceTable(qw422016 *qt422016.Writer, devices []devices.Device) {
//line services/dashboard/views/deviceListPage.qtpl:69
	qw422016.N().S(`
    <table class="w-full text-sm border-separate border-spacing-0">
        <thead class="text-left text-slate-500 sticky top-0 bg-white">
            <tr class="h-10">
                <th class="font-normal border-b align-middle px-4">
                    Device ID
                </th>
                <th class="font-normal border-b align-middle px-4">
                    Device Code
                </th>
                <th class="font-normal border-b align-middle px-4">
                    Device Description
                </th>
                <th class="font-normal border-b align-middle px-4">
                    Device Location Description
                </th>
            </tr>
        </thead>
        <tbody>
            `)
//line services/dashboard/views/deviceListPage.qtpl:88
	for _, dev := range devices {
//line services/dashboard/views/deviceListPage.qtpl:88
		qw422016.N().S(`
            <tr class="hover:bg-slate-50 group">
                <td class="px-4 h-10 border-b">`)
//line services/dashboard/views/deviceListPage.qtpl:90
		qw422016.N().DL(dev.ID)
//line services/dashboard/views/deviceListPage.qtpl:90
		qw422016.N().S(`</td>
                <td class="border-b"><a
                    class="flex items-center px-4 h-10 text-primary-700 group-hover:underline"
                    href="/overview/devices/`)
//line services/dashboard/views/deviceListPage.qtpl:93
		qw422016.N().DL(dev.ID)
//line services/dashboard/views/deviceListPage.qtpl:93
		qw422016.N().S(`"
                >`)
//line services/dashboard/views/deviceListPage.qtpl:94
		qw422016.E().S(dev.Code)
//line services/dashboard/views/deviceListPage.qtpl:94
		qw422016.N().S(`</a></td>
                <td class="px-4 h-10 border-b">`)
//line services/dashboard/views/deviceListPage.qtpl:95
		qw422016.E().S(dev.Description)
//line services/dashboard/views/deviceListPage.qtpl:95
		qw422016.N().S(`</td>
                <td class="px-4 h-10 border-b">`)
//line services/dashboard/views/deviceListPage.qtpl:96
		qw422016.E().S(dev.LocationDescription)
//line services/dashboard/views/deviceListPage.qtpl:96
		qw422016.N().S(`</td>
            </tr>
            `)
//line services/dashboard/views/deviceListPage.qtpl:98
	}
//line services/dashboard/views/deviceListPage.qtpl:98
	qw422016.N().S(`
        </tbody>
    </table>
`)
//line services/dashboard/views/deviceListPage.qtpl:101
}

//line services/dashboard/views/deviceListPage.qtpl:101
func WriteRenderDeviceTable(qq422016 qtio422016.Writer, devices []devices.Device) {
//line services/dashboard/views/deviceListPage.qtpl:101
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services/dashboard/views/deviceListPage.qtpl:101
	StreamRenderDeviceTable(qw422016, devices)
//line services/dashboard/views/deviceListPage.qtpl:101
	qt422016.ReleaseWriter(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:101
}

//line services/dashboard/views/deviceListPage.qtpl:101
func RenderDeviceTable(devices []devices.Device) string {
//line services/dashboard/views/deviceListPage.qtpl:101
	qb422016 := qt422016.AcquireByteBuffer()
//line services/dashboard/views/deviceListPage.qtpl:101
	WriteRenderDeviceTable(qb422016, devices)
//line services/dashboard/views/deviceListPage.qtpl:101
	qs422016 := string(qb422016.B)
//line services/dashboard/views/deviceListPage.qtpl:101
	qt422016.ReleaseByteBuffer(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:101
	return qs422016
//line services/dashboard/views/deviceListPage.qtpl:101
}

//line services/dashboard/views/deviceListPage.qtpl:103
func StreamRenderMap(qw422016 *qt422016.Writer, sensorGroup *devices.SensorGroup) {
//line services/dashboard/views/deviceListPage.qtpl:103
	qw422016.N().S(`
    <div 
        hx-ext="leaflet"
        class="w-full h-96"
        id="device-map"
    >
        <script type="text/javascript">
        function getWebSocketURL(path) {
            const loc = window.location;
            let newUri;

            if (loc.protocol === "https:") {
                newUri = "wss:";
            } else {
                newUri = "ws:";
            }
            
            newUri += "//" + loc.host + path;
            
            return newUri;
        }
        (() => {
            const devmap = htmx.find("#device-map")
            let map;
            let markerLayer;
            let ws;

            function replaceDevices() {
                // Remove old
                if (ws != undefined) ws.close();
                if (markerLayer != undefined) markerLayer.remove()

                // New layer
                markerLayer = L.markerClusterGroup()
                map.addLayer(markerLayer);

                // Fetch news
                ws = new WebSocket(getWebSocketURL("/overview/devices/stream-map" + location.search))
                ws.onmessage = (event) => {
                    const data = JSON.parse(event.data)
                    const marker = L.marker(data.coordinates).addTo(markerLayer)
                    marker.on('click', evt => {
                        const url = "/overview/devices/" + data.device_id;
                        htmx.ajax("GET", url, "main").then(() => {
                            window.history.pushState({}, null, url)
                        })
                    })
                    marker.bindTooltip(data.device_code)
                }
            }
            htmx.on(document.body, "newDeviceList", () => {
                replaceDevices()
            })
            async function init() {
                map = devmap.leaflet;
                replaceDevices()
            }

            if (devmap.leaflet !== undefined) {
                init()
            } else {
                devmap.addEventListener('leaflet:init', () => init())
            }
        })()
    </script>
    </div>
`)
//line services/dashboard/views/deviceListPage.qtpl:169
}

//line services/dashboard/views/deviceListPage.qtpl:169
func WriteRenderMap(qq422016 qtio422016.Writer, sensorGroup *devices.SensorGroup) {
//line services/dashboard/views/deviceListPage.qtpl:169
	qw422016 := qt422016.AcquireWriter(qq422016)
//line services/dashboard/views/deviceListPage.qtpl:169
	StreamRenderMap(qw422016, sensorGroup)
//line services/dashboard/views/deviceListPage.qtpl:169
	qt422016.ReleaseWriter(qw422016)
//line services/dashboard/views/deviceListPage.qtpl:169
}

//line services/dashboard/views/deviceListPage.qtpl:169
func RenderMap(sensorGroup *devices.SensorGroup) string {
//line services/dashboard/views/deviceListPage.qtpl:169
	qb422016 := qt422016.AcquireByteBuffer()
//line services/dashboard/views/deviceListPage.qtpl:169
	WriteRenderMap(qb422016, sensorGroup)
//line services/dashboard/views/deviceListPage.qtpl:169
	qs422016 := string(qb422016.B)
//line services/dashboard/views/deviceListPage.qtpl:169
	qt422016.ReleaseByteBuffer(qb422016)
//line services/dashboard/views/deviceListPage.qtpl:169
	return qs422016
//line services/dashboard/views/deviceListPage.qtpl:169
}

//line services/dashboard/views/deviceListPage.qtpl:172
type DeviceListPage struct {
	BasePage
	Devices     []devices.Device
	SensorGroup *devices.SensorGroup
}