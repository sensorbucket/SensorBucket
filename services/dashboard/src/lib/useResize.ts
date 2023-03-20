import { beforeUpdate } from 'svelte'

export function resize(el: HTMLElement, [store, updater]) {
    let dirty: boolean;
    beforeUpdate(() => {
        if (dirty) store.set(updater(el))
        dirty = false;
    })

    if (ResizeObserver) {
        const resizeObserver = new ResizeObserver(entries => {
            for (let entry of entries) {
                dirty = true;
                //force before update to run;
                store.set(null)
            }
        });

        resizeObserver.observe(el);

        return {
            destroy() { resizeObserver.unobserve(el); }
        }
    }
}

