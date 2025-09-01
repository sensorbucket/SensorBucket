<script lang="ts">
    let fileHover = $state(false)

    interface Props {
        onFileSubmit: (file: File | null) => void;
        disabled?: boolean;
        class?: string
    }

    let {class: extraClasses = '', disabled = false, onFileSubmit}: Props = $props();
    let inputEl: HTMLInputElement;

    function updateInputValue(file: File | null) {
        if (file === null) {
            inputEl.files = null;
            return
        }
        const dt = new DataTransfer()
        dt.items.add(file)
        inputEl.files = dt.files;
    }

    function onDrop(ev: DragEvent) {
        ev.preventDefault();
        if (disabled) return;
        let file: File | null = null;

        if (ev.dataTransfer?.items) {
            const item = ev.dataTransfer.items[0]
            if (item.kind !== 'file') return
            file = item.getAsFile()
        } else if (ev.dataTransfer?.files) {
            file = ev.dataTransfer.files.item(0)
        }

        updateInputValue(file);
        onFileSubmit(file);
    }

    function onSelect(ev: Event) {
        ev.preventDefault();
        if (disabled) return;
        const file = ev.target?.files.item(0)
        if (file == null) return;
        onFileSubmit(file)
    }
</script>

<input type="file" class={`rounded-lg border-2 border-sky border-dashed p-6 ${extraClasses}`}
       class:bg-sky-200={!disabled && fileHover}
       class:bg-sky-50={!disabled && !fileHover}
       class:bg-gray-100={disabled}
       class:text-gray-300={disabled}
       ondragenter={() => !disabled && (fileHover = true)} ondragleave={()=> !disabled && (fileHover = false)}
       ondrop={onDrop}
       onchange={onSelect}
       {disabled}
       bind:this={inputEl}
/>