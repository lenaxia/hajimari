<script lang="ts">
    import BookmarkGroup from "./BookmarkGroup.svelte";

    export let header: string = "Bookmarks";
    export let bookmarks: [];
    export let showGroups: boolean;
    export let targetBlank: boolean;
</script>

{#if bookmarks.length === 0}
    <div class="links">
        <h3>{header}</h3>
        <p>No bookmarks here...yet</p>
    </div>
{:else}
    <div class="links">
        <h3>{header}</h3>
        <div class="links_loop">
            {#each bookmarks as bookmarkGroup}
                {#if showGroups}
                    <div class="links_item">
                        <h4>{bookmarkGroup.group}</h4>
                        <BookmarkGroup {bookmarkGroup} {targetBlank} />
                    </div>
                {:else}
                    <BookmarkGroup {bookmarkGroup} {targetBlank} />
                {/if}
            {/each}
        </div>
    </div>
{/if}

<style>
    /* Fluid bookmark grid — column count follows available space. */
    .links_loop {
        display: grid;
        grid-template-columns: repeat(
            auto-fill,
            minmax(min(240px, 100%), 1fr)
        );
        grid-template-rows: auto;
        column-gap: 20px;
        row-gap: 12px;
    }

    .links_item {
        line-height: 1.5rem;
        margin-bottom: 2em;
        min-width: 0;
        webkit-font-smoothing: antialiased;
    }

    .links_item h4 {
        color: var(--color-text-acc);
    }
</style>
