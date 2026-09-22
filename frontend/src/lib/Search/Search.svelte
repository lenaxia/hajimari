<script lang="ts">
    import { appQuery } from "$lib/stores.js";
    import Icon from "@iconify/svelte";

    export let providers = [
        {
            name: "Google",
            token: "g",
            icon: "simple-icons:google",
            searchUrl: "https://www.google.com/search?q={query}",
            url: "https://www.google.com",
        },
        {
            name: "DuckDuckGo",
            token: "d",
            icon: "simple-icons:duckduckgo",
            searchUrl: "https://duckduckgo.com/?q={query}",
            url: "https://duckduckgo.com",
        },
        {
            name: "IMDB",
            token: "i",
            icon: "simple-icons:imdb",
            searchUrl: "https://www.imdb.com/find?q={query}",
            url: "https://www.imdb.com",
        },
        {
            name: "Reddit",
            token: "r",
            icon: "simple-icons:reddit",
            searchUrl: "https://www.reddit.com/search?q={query}",
            url: "https://www.reddit.com",
        },
        {
            name: "YouTube",
            token: "y",
            icon: "simple-icons:youtube",
            searchUrl: "https://www.youtube.com/results?search_query={query}",
            url: "https://www.youtube.com",
        },
        {
            name: "Spotify",
            token: "s",
            icon: "simple-icons:spotify",
            searchUrl: "hhttps://open.spotify.com/search/{query}",
            url: "https://open.spotify.com",
        },
        {
            name: "ABC",
            token: "a",
            icon: "mdi:test-tube",
            url: "https://example.com",
        }
    ];

    export let defaultProvider = "Google";

    let query = "";

    // All provider lookups are case-insensitive: config files routinely
    // disagree on capitalization (e.g. defaultSearchProvider: "Kagi" vs
    // provider name "kagi") and an exact-match miss renders an empty icon
    // or a dead token with no other symptom.
    const findByKey = (key: string, value: string) =>
        providers.find(
            (provider) =>
                String(provider[key] ?? "").toLowerCase() ===
                String(value ?? "").toLowerCase()
        );

    // Non-fatal validation: duplicate names or tokens are a config mistake
    // that silently breaks lookups (the first entry wins). Surface it in
    // the console AND under the search bar instead of crashing.
    let duplicateWarning: string | null = null;
    {
        const seenNames = new Map<string, string>();
        const seenTokens = new Map<string, string>();
        const problems: string[] = [];
        for (const provider of providers) {
            const name = String(provider.name ?? "").toLowerCase();
            const token = String(provider.token ?? "").toLowerCase();
            if (name) {
                if (seenNames.has(name)) {
                    problems.push(
                        `duplicate provider name "${provider.name}" (already used by "${seenNames.get(name)}")`
                    );
                } else {
                    seenNames.set(name, provider.name);
                }
            }
            if (token) {
                if (seenTokens.has(token)) {
                    problems.push(
                        `duplicate token "@${provider.token}" (already used by "${seenTokens.get(token)}")`
                    );
                } else {
                    seenTokens.set(token, provider.name);
                }
            }
        }
        if (problems.length) {
            duplicateWarning = problems.join("; ");
            console.warn(
                `[hajimari] search provider config issues: ${duplicateWarning}; the first matching entry wins`
            );
        }
    }

    let defaultProviderRecord = findByKey("name", defaultProvider);
    let icon = defaultProviderRecord?.icon;

    $: {
        let matches = query.match(/\/(.*)/);
        if (matches) {
            $appQuery = matches[1];
            icon = "mdi:apps";
        } else {
            $appQuery = "";
            icon = defaultProviderRecord?.icon;
        }
    }

    $: {
        let matches = query.match(/@(\w+)\s?(.*)/);
        if (matches) {
            let provider = findByKey("token", matches[1]);
            if (provider) {
                icon = provider.icon;
            }
        }
    }

    const handleSubmit = () => {
        query = query.replaceAll("+", "%2B");

        let matches = query.match(/@(\w+)\s?(.*)/);
        if (matches) {
            let token = matches[1];
            let queryText = matches[2];

            let provider = findByKey("token", token);

            if (provider?.searchUrl && queryText) {
                window.location.assign(
                    provider.searchUrl.replaceAll("{query}", queryText)
                );
            } else if (provider) {
                window.location.assign(provider.url);
            }
        } else if (validURL(query)) {
            if (containsProtocol(query)) {
                window.location.assign(query);
            } else {
                window.location.assign("https://" + query);
            }
        } else {
            let provider = defaultProviderRecord;
            if (provider?.searchUrl) {
                window.location.assign(
                    provider.searchUrl.replaceAll("{query}", query)
                );
            }
        }
    };

    // Source: https://stackoverflow.com/questions/5717093/check-if-a-javascript-string-is-a-url
    function validURL(str: string) {
        var pattern = new RegExp(
            "^(https?:\\/\\/)?" + // protocol
                "((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|" + // domain name
                "((\\d{1,3}\\.){3}\\d{1,3}))" + // OR ip (v4) address
                "(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*" + // port and path
                "(\\?[;&a-z\\d%_.~+=-]*)?" + // query string
                "(\\#[-a-z\\d_]*)?$",
            "i"
        ); // fragment locator
        return !!pattern.test(str);
    }

    function containsProtocol(str: string) {
        var pattern = new RegExp("^(https?:\\/\\/){1}.*", "i");
        return !!pattern.test(str);
    }
</script>

<section id="search">
    <form on:submit|preventDefault={handleSubmit}>
        <!-- svelte-ignore a11y-autofocus -->
        <Icon {icon}/>
        <input
            bind:value={query}
            type="text"
            id="keywords"
            spellcheck="false"
            autofocus={true}
        />
    </form>
    {#if duplicateWarning}
        <p class="config_warning" title={duplicateWarning}>
            <Icon icon="mdi:alert-outline" />
            search providers: {duplicateWarning} — the first matching entry
            wins
        </p>
    {/if}
</section>

<style>
    #search :global(svg) {
        font-size: 1.5em;
        position: absolute;
        /* deterministic placement: anchored inside the form box instead of
           riding its static position, which could push past the input edge */
        left: 0.6em;
        top: 50%;
        transform: translateY(-50%);
    }

    #search {
        position: relative;
        margin-bottom: 3vh;
    }

    #search form {
        position: relative;
    }

    input {
        font-size: max(0.8em, 16px);
        text-indent: 3em;
        min-width: 0;
        max-width: 100%;
    }

    .config_warning {
        color: var(--color-text-acc);
        font-size: 0.8em;
        margin: 0.4em 0 0 0.2em;
        overflow-wrap: anywhere;
    }

    .config_warning :global(svg) {
       	font-size: 1em;
        position: static;
        margin: 0 0.2em 0.15em 0;
        transform: none;
        vertical-align: middle;
        display: inline-block;
    }
</style>
