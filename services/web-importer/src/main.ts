import {mount} from 'svelte'
import App from './App.svelte'
import 'virtual:uno.css';
import './theme.css';

const app = mount(App, {
    target: document.getElementById('app')!,
})

export default app
