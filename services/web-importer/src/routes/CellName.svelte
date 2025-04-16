<script lang="ts">
    interface Props {
        depth: number;
        value: any;
    }

    const {depth, value}: Props = $props();
</script>

<div style={`padding-left: ${depth * 2}rem`}>
    <!--{#if row.getCanExpand()}-->
    <!--    <button onclick={row.getToggleExpandedHandler()}>-->
    <!--        {row.getIsExpanded() ? '👇' : '👉'}-->
    <!--    </button>-->
    <!--{:else}-->
    <!--    &nbsp;-->
    <!--{/if}-->
    {value}
</div>