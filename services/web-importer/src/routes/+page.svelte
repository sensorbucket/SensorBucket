<script lang="ts">
    import {createQuery} from '@tanstack/svelte-query'
    import {listDevicesOptions} from '$lib/sensorbucket/@tanstack/svelte-query.gen'
    import {createClient} from "@hey-api/client-fetch";

    const client = createClient({baseUrl: "/api"})r 
    const q = createQuery(listDevicesOptions({client}))

    // const apiKey = "OTE4OTg5Njc3NjEzNzkzMDU4MTplYzM5MmYyZjAyNjE1YzFmZDc1MTljODVkMWE5MzZjZQ";
</script>
{#if $q.isPending}
    loading...
{/if}
{#if $q.error}
    error!
{/if}
{#if $q.isSuccess}
    success
    <ul>
        {#each $q.data.data as device}
            <li>{device.code}</li>
        {/each}
    </ul>
{/if}