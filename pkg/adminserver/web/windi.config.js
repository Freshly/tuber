import { defineConfig } from 'windicss/helpers'

export default defineConfig({
	darkMode: 'media',
	safelist: ['text-lg'],

	extract: {
		include: ['./src/**/*.{js,ts,tsx,html}', './pages/**/*.{js,ts,tsx,html}'],
		exclude: ['node_modules', '.git', '.next'],
	},
})
