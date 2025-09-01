<script lang="ts">
    import {Progress, useId} from "bits-ui";

    interface Props {
        value: number;
        max: number;
        label: string;
    }

    let {value, max, label}: Props = $props();

    let progress = $derived(Math.ceil(value / (max || 1) * 100));

    const labelId = useId();
</script>

<div class="flex w-full flex-col gap-2">
    <div class="flex items-center justify-between text-sm font-medium">
        <span id={labelId}> {label} </span>
        <span>{progress}%</span>
    </div>
    <Progress.Root
            aria-labelledby={labelId}
            {value}
            {max}
            class="bg-dark-100 shadow-inner relative h-2 w-full overflow-hidden rounded-full"
    >
        <div
                class="bg-emerald-400 shadow-inner h-full w-full flex-1 rounded-full transition-all duration-200 ease-in-out"
                style={`transform: translateX(-${100 - (100 * ((progress ?? 0))) / 100}%)`}
        ></div>
    </Progress.Root>
</div>
