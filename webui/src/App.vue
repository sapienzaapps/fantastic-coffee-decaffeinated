<script>
export default {
	data() {
		return {
			theme: localStorage.getItem('theme') ||
				(window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'),
		}
	},
	watch: {
		theme: {
			immediate: true,
			handler(value) {
				document.documentElement.setAttribute('data-bs-theme', value)
				localStorage.setItem('theme', value)
			},
		},
	},
	methods: {
		toggleTheme() {
			this.theme = this.theme === 'dark' ? 'light' : 'dark'
		},
	},
}
</script>

<template>
	<div class="app-body">
		<aside id="sidebarMenu" class="sidebar offcanvas-md offcanvas-start" tabindex="-1"
			aria-labelledby="sidebarMenuLabel">
			<div class="sidebar-inner">
				<div class="offcanvas-header border-bottom">
					<span id="sidebarMenuLabel" class="offcanvas-title fw-bold">
						<i class="bi bi-cup-hot me-1"></i> Example App
					</span>
					<button type="button" class="btn-close" data-bs-dismiss="offcanvas"
						data-bs-target="#sidebarMenu" aria-label="Close"></button>
				</div>

				<div class="sidebar-brand d-none d-md-flex">
					<i class="bi bi-cup-hot text-primary"></i> Fantastic Coffee
				</div>

				<nav class="sidebar-nav">
					<p class="sidebar-heading">General</p>
					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink to="/" class="nav-link" exact-active-class="active">
								<i class="bi bi-house"></i> Home
							</RouterLink>
						</li>
						<li class="nav-item">
							<RouterLink to="/link1" class="nav-link" active-class="active">
								<i class="bi bi-grid-1x2"></i> Menu item 1
							</RouterLink>
						</li>
						<li class="nav-item">
							<RouterLink to="/link2" class="nav-link" active-class="active">
								<i class="bi bi-key"></i> Menu item 2
							</RouterLink>
						</li>
					</ul>

					<p class="sidebar-heading">Secondary menu</p>
					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink :to="'/some/' + 'variable_here' + '/path'" class="nav-link"
								active-class="active">
								<i class="bi bi-file-text"></i> Item 1
							</RouterLink>
						</li>
					</ul>
				</nav>
			</div>
		</aside>

		<div class="app-content d-flex flex-grow-1 flex-column">
			<header class="topbar sticky-top d-flex align-items-center gap-2 border-bottom bg-body px-3 px-md-4 py-2">
				<button class="btn btn-sm btn-outline-secondary d-md-none" type="button"
					data-bs-toggle="offcanvas" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu"
					aria-label="Toggle navigation">
					<i class="bi bi-list"></i>
				</button>
				<span class="fw-semibold d-md-none">Example App</span>

				<div class="ms-auto">
					<button class="btn btn-sm btn-outline-secondary" type="button"
						:title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
						@click="toggleTheme">
						<i :class="theme === 'dark' ? 'bi bi-sun' : 'bi bi-moon-stars'"></i>
					</button>
				</div>
			</header>

			<main class="flex-grow-1 p-3 p-md-4">
				<RouterView />
			</main>
		</div>
	</div>
</template>

<style>
</style>
