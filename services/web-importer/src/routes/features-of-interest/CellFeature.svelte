<script lang="ts">
    interface Props {
        encoding_type: string;
        feature: any;
    }

    const {encoding_type, feature}: Props = $props();
</script>

{#if encoding_type === 'application/geo+json'}
    <span>{feature.coordinates[0]}, {feature.coordinates[1]}</span>
{:else}
    <pre>Can't display this feature type</pre>
{/if}
