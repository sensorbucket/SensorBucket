<script lang="ts">
	import {
		Chart,
		LineController,
		LineElement,
		Legend,
		Tooltip,
		CategoryScale,
		TimeScale,
		LinearScale,
		PointElement,
		Colors
	} from 'chart.js';
	import 'chartjs-adapter-date-fns';
	Chart.register(
		LineController,
		LineElement,
		Legend,
		Tooltip,
		CategoryScale,
		TimeScale,
		LinearScale,
		PointElement,
		Colors
	);
	import type { PageData } from './$types';
	import { page } from '$app/stores';
	import type { Measurement } from '$lib/measurement';
	import type { Device } from '$lib/device';

	export let data: PageData;
	const timeseries = $page.data.timeseries as Measurement[];
	const device = $page.data.device as Device;

	interface Dataset {
		type: 'line';
		label: string;
		data: { x: any; y: any }[];
	}

	let chartData: Record<string, Dataset> = {};
	$: {
		chartData = timeseries.reduce<Record<string, Dataset>>((data, next) => {
			let group = data[next.measurement_type] ?? {
				type: 'line',
				label: next.measurement_type,
				data: []
			};
			group.data = [{ x: new Date(next.timestamp).getTime(), y: next.value }, ...group.data];
			//group.data.push({ x: new Date(next.timestamp).getTime(), y: next.value });
			data[next.measurement_type] = group;
			return data;
		}, {});
	}

	let chart: Chart;
	const usePlot = (el: HTMLCanvasElement) => {
		chart = new Chart(el, {
			data: {
				datasets: Object.values(chartData)
			},
			options: {
				parsing: false,
				animation: false,
				interaction: {
					intersect: false,
					mode: 'index',
					axis: 'x'
				},
				scales: {
					x: {
						type: 'time'
					}
				}
			}
		});
		return {
			destroy: () => chart.destroy()
		};
	};
</script>

<div class="grid grid-cols-12 gap-6">
	<div class="rounded bg-white p-4 col-span-12">
		<h2 class="text-lg">Device: {device.code}</h2>
		<hr class="my-1" />
		<canvas use:usePlot />
	</div>
</div>
