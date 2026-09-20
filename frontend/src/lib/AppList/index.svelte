<script lang="ts">
    import { fade } from "svelte/transition";
    import AppGroup from "./AppGroup.svelte";

    export let apps: any;
    export let showGroups: boolean;
    export let defaultIcon: string = "mdi:application";
    export let showUrl: boolean = true;
    export let showInfo: boolean = true;
    export let showStatus: boolean = true;
    export let targetBlank: boolean = false;
</script>

<div class="apps">
    <h3>Applications</h3>
    {#if apps.length === 0}
        <p>No apps here...yet</p>
    {:else}
        <div class="apps_loop" class:grouped={showGroups}>
            {#each apps as group}
                {#if showGroups}
                    <div class="links_item" in:fade={{ duration: 300 }}>
                        <h4>{group.group}</h4>
                        <div class="apps_group">
                            <AppGroup
                                {group}
                                {showUrl}
                                {showInfo}
                                {showStatus}
                                {defaultIcon}
                                {targetBlank}
                            />
                        </div>
                    </div>
                {:else}
                    <AppGroup
                        {group}
                        {showUrl}
                        {showInfo}
                        {showStatus}
                        {defaultIcon}
                        {targetBlank}
                    />
                {/if}
            {/each}
        </div>
    {/if}
</div>

<style>
    /* Fluid tile grid: column count is decided by available space, not by
       media queries. `minmax(min(240px, 100%), 1fr)` lets tracks shrink
       below the ideal minimum on narrow screens so long app names can
       never force horizontal overflow. Phones get one column, tablets
       2-3, desktops 4+. */
    .apps_loop {
        display: grid;
        grid-template-columns: repeat(
            auto-fill,
            minmax(min(240px, 100%), 1fr)
        );
        grid-template-rows: auto;
        column-gap: 20px;
        row-gap: 10px;
        padding-bottom: var(--module-spacing);
    }

    .apps_group {
        display: grid;
        grid-template-columns: repeat(
            auto-fill,
            minmax(min(220px, 100%), 1fr)
        );
        column-gap: 16px;
        row-gap: 10px;
    }

    .links_item h4 {
        color: var(--color-text-acc);
    }
</style>
