import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import fs from 'fs';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		https: {
			key: fs.readFileSync('../certs/localhost+2-key.pem'),
			cert: fs.readFileSync('../certs/localhost+2.pem'),
		},
	},
});
