<script lang="ts">
	import "../app.css";
	import { fade } from "svelte/transition";
	import { onMount } from "svelte";

	let visible = false;

	onMount(() => {
		visible = true;
	});
</script>

{#if visible}
	<main id="container" transition:fade>
		<slot />
	</main>
{/if}

<style>
	/* Mobile-first: full-width column with safe horizontal padding.
	   The old fixed percentages (60%/90%) + child vw widths caused
	   horizontal overflow on phones. */
	main {
		align-items: stretch;
		display: grid;
		grid-column-gap: 20px;
		grid-row-gap: 24px;
		grid-template-columns: minmax(0, 1fr);
		grid-template-rows: auto;
		justify-items: stretch;
		margin-left: auto;
		margin-right: auto;
		margin-top: 24px;
		margin-bottom: 24px;
		padding-left: 16px;
		padding-right: 16px;
		width: 100%;
		max-width: 100%;
	}

	@media screen and (min-width: 768px) {
		main {
			margin-top: 5vh;
			padding-left: 0;
			padding-right: 0;
			width: 90%;
		}
	}

	@media screen and (min-width: 1280px) {
		main {
			width: 60%;
		}
	}
</style>
