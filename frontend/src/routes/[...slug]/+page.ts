import { error } from '@sveltejs/kit';
import { api } from '$lib/api';
import type { PageLoad } from './$types';

type Startpage = {
    id: string;
    name: string;
    title: string;
    theme: string;
    lightTheme: string;
    darkTheme: string;
    customThemes: Array<Record<string, string>>;
    showSearch: string;
    showGreeting: boolean;
    showApps: boolean;
    showAppGroups: boolean;
    showAppUrls: boolean;
    showAppInfo: boolean;
    showAppStatus: boolean;
    defaultAppIcon: string;
    showBookmarks: boolean;
    showBookmarkGroups: boolean;
    showGlobalBookmarks: boolean;
    alwaysTargetBlank: boolean;
    defaultSearchProvider: string;
    searchProviders: Array<Record<string, string>>;
    bookmarks: any;
};

export const load: PageLoad = async ({ fetch, params, url }) => {
    const { slug } = params;

    const groupParam = url.searchParams.get('group') ?? url.searchParams.get('g');
    const groupSuffix = groupParam !== null ? `?group=${encodeURIComponent(groupParam)}` : '';

    const [startpage, apps, bookmarks] = await Promise.all([
        api(fetch, 'GET', `startpage/${slug}`),
        api(fetch, 'GET', `apps${groupSuffix}`),
        api(fetch, 'GET', `bookmarks${groupSuffix}`)
    ]);

    if (await startpage.status !== 200) {
        let data = await startpage.json();
        throw error(startpage.status, data.status);
    }

    // Group impersonation requires admin membership server-side. For
    // anyone else a ?group= link would 403 both fetches and the JSON
    // parsing below would crash the page into a 500 — degrade to the
    // requester's own view instead.
    const asJson = async (resp: Response, fallbackFetch: () => Promise<Response>) => {
        if (resp.ok) return resp.json();
        if (groupSuffix) return (await fallbackFetch()).json();
        return [];
    };

    return {
        startpage: (await startpage.json() as Startpage),
        apps: (await asJson(apps, () => api(fetch, 'GET', 'apps'))),
        globalBookmarks: (await asJson(bookmarks, () => api(fetch, 'GET', 'bookmarks'))),
        slug
    }

};