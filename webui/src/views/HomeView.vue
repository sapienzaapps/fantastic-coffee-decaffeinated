<script>
export default {
	data() {
		return {
			errormsg: null,
			loading: false,
			saving: false,
			name: null,
			draft: '',
		}
	},
	methods: {
		// Read the name from the API. A 404 means no name has been stored yet.
		async loadName() {
			this.loading = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.get("/name");
				this.name = response.data.name;
				this.draft = response.data.name;
			} catch (e) {
				if (e.response && e.response.status === 404) {
					this.name = null;
					this.draft = '';
				} else {
					this.errormsg = this.errorMessage(e);
				}
			}
			this.loading = false;
		},
		// Store the name through the API and update the page with the saved value.
		async saveName() {
			this.saving = true;
			this.errormsg = null;
			try {
				const response = await this.$axios.put("/name", { name: this.draft });
				this.name = response.data.name;
				this.draft = response.data.name;
			} catch (e) {
				this.errormsg = this.errorMessage(e);
			}
			this.saving = false;
		},
		errorMessage(e) {
			if (e.response) {
				return `Request failed (HTTP ${e.response.status})`;
			}
			return e.toString();
		},
	},
	mounted() {
		this.loadName();
	},
}
</script>

<template>
	<div class="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-4">
		<div>
			<h1 class="h3 mb-1">Home page</h1>
			<p class="text-body-secondary mb-0">Example of reading and writing data through the API.</p>
		</div>
		<button class="btn btn-outline-secondary btn-sm" type="button" :disabled="loading" @click="loadName">
			<i class="bi bi-arrow-clockwise me-1"></i> Refresh
		</button>
	</div>

	<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

	<div class="card shadow-sm">
		<div class="card-body">
			<LoadingSpinner :loading="loading">
				<p v-if="name === null" class="text-body-secondary">No name has been stored yet. Set one below.</p>
				<p v-else class="mb-3">
					The stored name is
					<span class="badge text-bg-primary fs-6 ms-1">{{ name }}</span>
				</p>

				<form class="row g-2 align-items-center" @submit.prevent="saveName">
					<label for="nameInput" class="col-auto col-form-label">Name</label>
					<div class="col-sm-6 col-md-4">
						<input id="nameInput" v-model="draft" type="text" class="form-control"
							placeholder="Type a name..." maxlength="255">
					</div>
					<div class="col-auto">
						<button type="submit" class="btn btn-primary" :disabled="saving || draft.trim() === ''">
							<i class="bi bi-check-lg me-1"></i> Save
						</button>
					</div>
				</form>
			</LoadingSpinner>
		</div>
	</div>
</template>

<style>
</style>
