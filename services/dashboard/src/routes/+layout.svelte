<script lang="ts">
	import '../app.css';
	import 'iconify-icon';
	import logo from '../logo-white.png';
	import Link from './Link.svelte';
</script>

<svelte:head>
	<title>Sensorbucket</title>
</svelte:head>

<sidebar class="fixed left-0 top-0 w-sidebar h-full bg-secondary-600 flex flex-col">
	<header class="h-header flex justify-center items-center">
		<img src={logo} alt="SensorBucket" class="w-1/2" />
	</header>
	<nav class="mt-4 px-8">
		<ul>
			<li class="pb-2 text-xs font-bold text-secondary-300 uppercase">Navigation</li>
			<li><Link to="/" icon="material-symbols:home-outline-rounded">Homepage</Link></li>
		</ul>
	</nav>
</sidebar>

<section class="ml-sidebar h-full">
	<header class="h-header bg-white border-b" />
	<main class="mt-4 px-4 bg-gray-100">
		<slot />
	</main>
</section>
