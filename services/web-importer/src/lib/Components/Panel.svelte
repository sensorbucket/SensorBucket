<script lang="ts">
    import type {Snippet} from "svelte";

    interface Props {
        title: string
        children?: Snippet
    }

    let {title, children}: Props = $props()
</script>

<div class="flex flex-col gap-6 relative p-4 rounded-xl border border-dashed border-2 border-blue-100">
    <h2 class="absolute left-4 top-0 -translate-y-1/2 text-sm bg-white font-bold px-2">{title}</h2>
    {@render children?.()}
</div>