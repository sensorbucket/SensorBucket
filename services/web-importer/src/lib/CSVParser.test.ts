import {test, expect} from 'vitest';
import {CSVParser} from "$lib/CSVParser";

test('CSVParser should call column parsers', async () => {
    const csvParser = new CSVParser({code: ""});
    csvParser.addColumn(/^device code$/g, () => (context, value) => {
        context.userData.code = value;
    })
    const file = 'other,data,notrelevant\ndevice code,device description,device properties\n1,2,3\n'
    const result = await csvParser.parse(file)
    expect(result.code == "1")
})

test('CSVParser should pass context to every onRowFinish', async () => {
    const csvParser = new CSVParser({code: "", other: "no"});
    csvParser.addColumn(/^device code$/g, () => (context, value) => {
        context.userData.code = value;
    })
    let pass = 0
    csvParser.afterRowParse = (context) => {
        if (pass === 0) {
            context.userData.other = "yes"
        } else {
            expect(context.userData.other === "yes");
        }
        pass++;
    }
    const file = 'device code,device description,device properties\n1,2,3\n1,2,3\n'
    await csvParser.parse(file)
})