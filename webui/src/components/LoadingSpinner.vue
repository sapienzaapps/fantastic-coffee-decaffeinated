<script>
export default {
	props: ["loading"]
}
</script>

<template>
	<div v-if="loading" class="text-center py-5">
		<div class="spinner-border text-primary" role="status">
			<span class="visually-hidden">Loading...</span>
		</div>
		<p class="text-body-secondary mt-2 mb-0">Loading...</p>
	</div>
	<div v-else><slot /></div>
</template>

<style></style>
