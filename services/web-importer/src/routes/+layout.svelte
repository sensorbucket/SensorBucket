<script lang="ts">
    import 'virtual:uno.css'
    import '@unocss/reset/tailwind.css'
    import {QueryClient, QueryClientProvider} from "@tanstack/svelte-query";
    import Sidebar from "./Sidebar.svelte";

    const client = new QueryClient()
    const {children} = $props();
</script>

<QueryClientProvider {client}>
    <Sidebar/>
    <div class="ml-9/10 lg:ml-[300px]">
        {@render children?.()}
    </div>
</QueryClientProvider>