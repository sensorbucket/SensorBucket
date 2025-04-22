<script lang="ts">
    let fileHover = $state(false)

    interface Props {
        onFileSubmit: (file: File | null) => void
        class?: string
    }

    let {class: extraClasses = '', onFileSubmit}: Props = $props();

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
        const file = ev.target?.files.item(0)
        if (file == null) return;
        onFileSubmit(file)
    }
</script>

<input type="file" class={`rounded-lg border-2 border-sky border-dashed p-6 ${extraClasses}`}
       class:bg-sky-200={fileHover}
       class:bg-sky-50={!fileHover}
       ondragenter={() => fileHover = true} ondragleave={()=> fileHover = false}
       ondrop={onDrop}
       onchange={onSelect}
       bind:this={inputEl}
/>