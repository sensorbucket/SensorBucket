<script lang="ts">
    interface Props {
        title: string;
    }
    let {title}: Props = $props();
</script>

<div class="-mx-2 px-2 relative flex flex-col border-t-2 border-blue">
    <h2 class="absolute top-0 text-sm -translate-y-1/2 bg-white font-bold px-2">{title}</h2>
</div>