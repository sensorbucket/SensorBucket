<script lang="ts">
	import { ToStream, QueryMeasurements } from './api';
	import uPlot from 'uplot';
	import 'uplot/dist/uPlot.min.css';
	import { parseISO } from 'date-fns';
	import type { Datastream, Measurement } from './models';

	export let datastream: Datastream;
	export let start = new Date(Date.now() - 24 * 60 * 60 * 1000000);
	export let end = new Date();

	let chart: uPlot;
	function useChart(container: HTMLElement) {
		let opts = {
			title: 'Datastream measurements',
			width: container.clientWidth - 10,
			height: container.clientHeight - 70,
			scales: {
				x: {
					time: true
				}
			},
			series: [{}, { stroke: 'blue' }],
			axes: [{ label: 'Timestamp' }]
		};

		chart = new uPlot(opts, [], container);
	}

	let data: [number[], number[]] = [[], []];
	let stream: ReadableStream<Measurement[]>;
	let transform = () =>
		new TransformStream({
			transform(chunk, ctrl) {
				let x: number[] = [];
				let y: number[] = [];
				for (let m of chunk.reverse()) {
					x.push(parseISO(m.measurement_timestamp).getTime() / 1000);
					y.push(m.measurement_value);
				}
				ctrl.enqueue([x, y]);
			}
		});
	let writer = () =>
		new WritableStream({
			start() {
				data = [[], []];
			},
			write(chunk) {
				data[0] = [...chunk[0], ...data[0]];
				data[1] = [...chunk[1], ...data[1]];
				data = data;
			}
		});

	$: {
		if (stream) {
			stream.cancel();
		}
		stream = ToStream(QueryMeasurements(start, end, { datastream: datastream.id }));
		stream.pipeThrough(transform()).pipeTo(writer());
	}

	$: {
		(() => {
			if (!chart) return;
			chart.delSeries(1);
			chart.addSeries(
				{
					label: `${datastream.observed_property} (${datastream.unit_of_measurement})`,
					stroke: 'blue'
				},
				1
			);
			let labels = chart.root.querySelectorAll('.u-legend .u-label');
			labels[1].textContent = chart.axes[1].label = datastream.unit_of_measurement;
			chart.setData(data);
		})();
	}
</script>

<div use:useChart class="w-full h-full" />
