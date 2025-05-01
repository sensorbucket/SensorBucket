import Papa, {type ParseResult} from "papaparse";

type ColumnParser<T> = (context: CSVParserContext<T>, value: string) => void
type ColumnParserBuilder<T> = (field: string) => ColumnParser<T>

function papaParse(file: File | string): Promise<ParseResult<string[]>> {
    // @ts-ignore
    return new Promise((resolve, reject) => Papa.parse(file, {
        worker: false,
        encoding: "utf-8",
        skipEmptyLines: "greedy",
        complete(results: ParseResult<any>, _: File) {
            resolve(results)
        },
        error(error: Error, _: string) {
            reject(error)
        }
    }))
}

class CSVParserContext<T> {
    csv: ParseResult<string[]>
    columnParsers: ColumnParser<T>[];
    currentRow: number;
    userData: T;

    constructor(csv: ParseResult<string[]>, emptyData: T) {
        this.csv = csv;
        this.columnParsers = [];
        this.currentRow = 0;
        this.userData = emptyData;
    }

    getRow(row: number) {
        this.currentRow = row;
        return this.csv.data[row];
    }

    [Symbol.iterator]() {
        return this;
    }

    next() {
        let row = this.csv.data[++this.currentRow];
        return {done: this.currentRow == this.csv.data.length, value: row}
    }
}

export class CSVParser<T = unknown> {
    protected columnParserBuilders: [RegExp, ColumnParserBuilder<T>][] = [];
    public beforeRowParse?: (context: CSVParserContext<T>) => void;
    public afterRowParse?: (context: CSVParserContext<T>) => void;
    protected initialUserData: () => T;

    constructor(initialUserData: () => T) {
        this.initialUserData = initialUserData
    }

    public addColumn(match: RegExp, parser: ColumnParserBuilder<T>) {
        this.columnParserBuilders.push([match, parser]);
        return this;
    }

    protected getRowColumnMatchCount(header: string[]) {
        let count = 0;
        for (const column of header) {
            for (const [regex] of this.columnParserBuilders) {
                if (column.match(regex)) {
                    count++;
                    break;
                }
            }
        }
        return count;
    }

    protected findHeaderRow(rows: string[][]) {
        let maxMatches = 0;
        let maxMatchIndex = -1;
        const MAX_ROWS = 10;
        const rowsToCheck = Math.min(rows.length, MAX_ROWS);

        for (let ix = 0; ix < Math.min(rowsToCheck, rows.length); ix++) {
            const matches = this.getRowColumnMatchCount(rows[ix]);
            if (matches > maxMatches) {
                maxMatches = matches;
                maxMatchIndex = ix;
            }
        }
        return maxMatchIndex;
    }

    protected buildColumnParsers(context: CSVParserContext<T>, header: string[]) {
        for (const column of header) {
            let found = false;
            for (const [regex, builder] of this.columnParserBuilders) {
                if (column.match(regex)) {
                    found = true
                    context.columnParsers.push(builder(column));
                    break;
                }
            }
            if (!found) context.columnParsers.push(() => {
            })
        }
    }

    protected parseRow(context: CSVParserContext<T>, row: string[]) {
        for (let ix = 0; ix < row.length && ix < context.columnParsers.length; ix++) {
            if (row[ix] === undefined || row[ix].trim() === "") continue;
            context.columnParsers[ix](context, row[ix]);
        }
        return context;
    }

    public async parse(file: File | string) {
        // Parse the file as CSV
        const csv = await papaParse(file)
        // Find the header row, this is the first row with at least 3 columns that match the columnBuilders
        const headerRow = this.findHeaderRow(csv.data);
        if (headerRow === -1) {
            return new Error("No header row found");
        }
        //
        const context = new CSVParserContext<T>(csv, this.initialUserData());
        this.buildColumnParsers(context, context.getRow(headerRow));
        //
        for (const row of context) {
            try {
                this.beforeRowParse?.(context);
                this.parseRow(context, row);
                this.afterRowParse?.(context);
            } catch (e) {
                if (e instanceof Error) {
                    return e
                }
                return new Error(e as any) // ?
            }
        }
        return context.userData;
    }
}