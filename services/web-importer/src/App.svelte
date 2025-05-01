<script lang="ts">
    import Layout from "./routes/Layout.svelte";
    import {Tabs} from "bits-ui";
    import DevicesPage from "./routes/devices/DevicesPage.svelte";
    import FeaturesPage from "./routes/features-of-interest/FeaturesPage.svelte";
</script>

<Layout>
    <main class="w-full lg:w-screen-lg mx-auto mt-4 border border-slate-100 rounded bg-white">
        <Tabs.Root value="devices">
            <div class="p-4 mb-4 pb-0">
                <Tabs.List
                        class="w-full grid cols-2 p-1 gap-4 bg-stone-300 shadow-inner rounded-9px font-bold text-stone-800">
                    <Tabs.Trigger value="devices"
                                  class="cursor-pointer px-3 py-2 rounded-md  data-[state=active]:(bg-white)">Devices
                        &amp; Sensors
                    </Tabs.Trigger>
                    <Tabs.Trigger value="features"
                                  class="cursor-pointer px-3 py-2 rounded-md  data-[state=active]:(bg-white)">Features
                        of Interest
                    </Tabs.Trigger>
                </Tabs.List>
            </div>
            <Tabs.Content value="devices">
                <DevicesPage/>
            </Tabs.Content>
            <Tabs.Content value="features">
                <FeaturesPage/>
            </Tabs.Content>
        </Tabs.Root>
    </main>
</Layout>