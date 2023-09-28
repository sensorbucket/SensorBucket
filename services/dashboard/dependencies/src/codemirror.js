import { EditorView, keymap } from "@codemirror/view"
import { indentWithTab } from "@codemirror/commands"
import { basicSetup } from 'codemirror';
import { indentUnit } from '@codemirror/language';
import {python} from '@codemirror/lang-python';

export function createEditor(el, userCode) {
    let view = new EditorView({
        doc: userCode,
        extensions: [
            basicSetup,
            keymap.of([indentWithTab]),
            python(),
            indentUnit.of('    '),
        ],
        parent: el,
    })
    return view
}
