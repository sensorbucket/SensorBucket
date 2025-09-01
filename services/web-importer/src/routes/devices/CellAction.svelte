<script lang="ts">
    import {Action, type Status} from "$lib/reconciliation";

    interface Props {
        status: Status
        action: Action
    }

    let {status, action}: Props = $props()
</script>

{status}