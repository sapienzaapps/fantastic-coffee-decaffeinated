<script>
export default {
	props: ['msg']
}
</script>

<template>
	<div class="alert alert-danger d-flex align-items-center gap-2" role="alert">
		<i class="bi bi-exclamation-triangle-fill"></i>
		<div>{{ msg }}</div>
	</div>
</template>

<style>
</style>
