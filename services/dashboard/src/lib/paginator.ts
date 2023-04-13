import type { Subscriber, Unsubscriber } from "svelte/store";

type Invalidator<T> = (data: T) => void;
const noop = () => { }

export class Paginator<T> {
    private pages: T[][] = [];
    private _page: number = 0;
    private it: AsyncIterator<T[]>;

    constructor(it: AsyncIterator<T[]>) {
        this.it = it
        this.page = 0
    }

    private subscribers: Set<Subscriber<any>> = new Set()
    subscribe(run: Subscriber<any>, _invalidate: Invalidator<T> = noop): Unsubscriber {
        this.subscribers.add(run);
        run(this)
        return () => {
            this.subscribers.delete(run);
        };
    }

    private notify() {
        for (let subscriber of this.subscribers) {
            subscriber(this)
        }
    }

    get data(): T[] {
        return this.pages[this._page] ?? [];
    }

    get page() {
        return this._page;
    }

    get length() {
        return this.pages.length
    }

    set page(value: number) {
        if (value < 0) return;
        // Requested page is within already fetched pages
        if (value >= 0 && value < this.pages.length) {
            console.log(`Within valid lengths`);
            this._page = value
            this.notify()
        } else {
            console.log(`Out of range, request new`);
            // Fetch next
            this.it.next().then(({ value, done }) => {
                if (done) {
                    return
                }
                this.pages.push(value)
                this._page = this.pages.length - 1
                this.notify()
            })
        }
    }



}
