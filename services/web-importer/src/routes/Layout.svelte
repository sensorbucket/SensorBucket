<script lang="ts">
    import Sidebar from "./Sidebar.svelte";

    const {children} = $props();
</script>

<Sidebar/>
<div class="mt-12 lg:(ml-$sidebar mt-0)">
    {@render children?.()}
</div>