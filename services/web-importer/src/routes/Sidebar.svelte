<script lang="ts">
    let is_open = $state(false);
</script>

<aside class="w-screen lg:w-$sidebar fixed lg:h-screen top-0 left-0 flex flex-col bg-$secondary z-[5000]">
    <header class="h-12 lg:h-header flex justify-between lg:flex-col lg:justify-end">
        <a class="h-full block p-3 lg:p-0 lg:pl-8 lg:pr-12 lg:h-auto" href="/dashboard"><img
                src="/dashboard/static/logo-white.png" alt="SensorBucket" class="h-full"></a>
        <button class="h-8 w-8 self-center lg:hidden">
            <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"
                 class="text-white transition duration-150" class:rotate-180={is_open}
                 onclick={() => is_open = !is_open}>
                <!-- Icon from Google Material Icons by Material Design Authors - https://github.com/material-icons/material-icons/blob/master/LICENSE -->
                <path fill="currentColor"
                      d="M8.12 9.29L12 13.17l3.88-3.88a.996.996 0 1 1 1.41 1.41l-4.59 4.59a.996.996 0 0 1-1.41 0L6.7 10.7a.996.996 0 0 1 0-1.41c.39-.38 1.03-.39 1.42 0"/>
            </svg>
        </button>
    </header>
    <nav class="mt-4 px-8 pb-4 lg:flex lg:flex-1 flex-col justify-between" class:h-auto={is_open}
         class:hidden={!is_open}>
        <!-- TOP Nav items -->
        <ul>
            <li class="pb-2 text-xs font-bold text-$secondary-300 uppercase">Navigation</li>

            <li>
                <a href="/dashboard/"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" class="mr-2">
                        <!-- Icon from Material Symbols by Google - https://github.com/google/material-design-icons/blob/master/LICENSE -->
                        <path fill="currentColor"
                              d="M6 19h3v-5q0-.425.288-.712T10 13h4q.425 0 .713.288T15 14v5h3v-9l-6-4.5L6 10zm-2 0v-9q0-.475.213-.9t.587-.7l6-4.5q.525-.4 1.2-.4t1.2.4l6 4.5q.375.275.588.7T20 10v9q0 .825-.588 1.413T18 21h-4q-.425 0-.712-.288T13 20v-5h-2v5q0 .425-.288.713T10 21H6q-.825 0-1.412-.587T4 19m8-6.75"/>
                    </svg>
                    <span>
                        Overview
                    </span>
                </a>
            </li>

            <li>
                <a href="/dashboard/workers"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" class="mr-2">
                        <!-- Icon from Unicons by Iconscout - https://github.com/Iconscout/unicons/blob/master/LICENSE -->
                        <path fill="currentColor"
                              d="M20.49 7.52a.2.2 0 0 1 0-.08a.2.2 0 0 1 0-.07v-.09l-.06-.15a.5.5 0 0 0-.09-.11l-.09-.08h-.05l-3.94-2.49l-3.72-2.3a.9.9 0 0 0-.29-.15h-.08a.8.8 0 0 0-.27 0h-.1a1.1 1.1 0 0 0-.33.13L4 6.78l-.09.07l-.09.08l-.1.07l-.05.06l-.06.15v.15a.7.7 0 0 0 0 .2v8.73a1 1 0 0 0 .47.85l7.5 4.64l.15.06h.08a.86.86 0 0 0 .52 0h.08l.15-.06L20 17.21a1 1 0 0 0 .47-.85V7.63s.02-.07.02-.11M12 4.17l1.78 1.1l-5.59 3.46l-1.79-1.1Zm-1 15l-5.5-3.36V9.42l5.5 3.4Zm1-8.11l-1.91-1.15l5.59-3.47l1.92 1.19Zm6.5 4.72L13 19.2v-6.38l5.5-3.4Z"/>
                    </svg>
                    <span>
                        Workers
                    </span>
                </a>
            </li>

            <li>
                <a href="/dashboard/pipelines"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" class="mr-2">
                        <!-- Icon from Material Design Icons by Pictogrammers - https://github.com/Templarian/MaterialDesign/blob/master/LICENSE -->
                        <path fill="currentColor"
                              d="M22 14h-2v2h-6v-3h2v-2h-2V6a2 2 0 0 0-2-2H4V2H2v8h2V8h6v3H8v2h2v5a2 2 0 0 0 2 2h8v2h2"/>
                    </svg>
                    <span>
                        Pipelines
                    </span>
                </a>
            </li>

            <li>
                <a href="/importer"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 16 16" class="mr-2">
                        <!-- Icon from Lsicon by Wis Design - https://github.com/wisdesignsystem/lsicon/blob/main/LICENSE -->
                        <path fill="currentColor" fill-rule="evenodd"
                              d="M3 1.5a.5.5 0 0 1 .5-.5H9a.5.5 0 0 1 .354.146l3.5 3.5A.5.5 0 0 1 13 5v2.5h-1v-2H9a.5.5 0 0 1-.5-.5V2H4v12h5.5v1h-6a.5.5 0 0 1-.5-.5zm6.5 1.207L11.293 4.5H9.5zm-2.854 7.94l2-2l.708.707L8.207 10.5H13v1H8.207l1.147 1.146l-.708.708l-2-2a.5.5 0 0 1 0-.708"
                              clip-rule="evenodd"/>
                    </svg>
                    <span>
                        Import
                    </span>
                </a>
            </li>

        </ul>

        <!-- BOTTOM Nav items -->
        <ul>
            <li class="pb-2 text-xs font-bold text-$secondary-300 uppercase">Other</li>

            <li>
                <a href="/tenants/switch?success_url=/dashboard"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" class="mr-2">
                        <!-- Icon from Material Design Icons by Pictogrammers - https://github.com/Templarian/MaterialDesign/blob/master/LICENSE -->
                        <path fill="currentColor"
                              d="M5 3v18h6v-3.5h2V21h6V3zm2 2h2v2H7zm4 0h2v2h-2zm4 0h2v2h-2zM7 9h2v2H7zm4 0h2v2h-2zm4 0h2v2h-2zm-8 4h2v2H7zm4 0h2v2h-2zm4 0h2v2h-2zm-8 4h2v2H7zm8 0h2v2h-2z"/>
                    </svg>
                    <span>
                        Switch Organisation
                    </span>
                </a>
            </li>

            <li>
                <a href="/tenants/auth/settings"
                   class="flex items-center py-1.5 text-sm text-$secondary-300 hover:text-white transition-colors duration-150">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" class="mr-2">
                        <!-- Icon from Bootstrap Icons by The Bootstrap Authors - https://github.com/twbs/icons/blob/main/LICENSE.md -->
                        <path fill="currentColor"
                              d="M11 5a3 3 0 1 1-6 0a3 3 0 0 1 6 0m-9 8c0 1 1 1 1 1h5.256A4.5 4.5 0 0 1 8 12.5a4.5 4.5 0 0 1 1.544-3.393Q8.844 9.002 8 9c-5 0-6 3-6 4m9.886-3.54c.18-.613 1.048-.613 1.229 0l.043.148a.64.64 0 0 0 .921.382l.136-.074c.561-.306 1.175.308.87.869l-.075.136a.64.64 0 0 0 .382.92l.149.045c.612.18.612 1.048 0 1.229l-.15.043a.64.64 0 0 0-.38.921l.074.136c.305.561-.309 1.175-.87.87l-.136-.075a.64.64 0 0 0-.92.382l-.045.149c-.18.612-1.048.612-1.229 0l-.043-.15a.64.64 0 0 0-.921-.38l-.136.074c-.561.305-1.175-.309-.87-.87l.075-.136a.64.64 0 0 0-.382-.92l-.148-.045c-.613-.18-.613-1.048 0-1.229l.148-.043a.64.64 0 0 0 .382-.921l-.074-.136c-.306-.561.308-1.175.869-.87l.136.075a.64.64 0 0 0 .92-.382zM14 12.5a1.5 1.5 0 1 0-3 0a1.5 1.5 0 0 0 3 0"/>
                    </svg>
                    <span>
                        Settings
                    </span>
                </a>
            </li>

        </ul>
    </nav>
</aside>