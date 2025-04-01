<script lang="ts">
    import {QueryClient, QueryClientProvider} from "@tanstack/svelte-query";

    const client = new QueryClient()
    const {children} = $props();
</script>

<QueryClientProvider {client}>
    {@render children?.()}
</QueryClientProvider>