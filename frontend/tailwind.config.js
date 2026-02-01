/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{svelte,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                'app-bg': '#1b2636',
                'card-bg': '#2a3b55',
                'primary': '#3e84f4',
                'primary-hover': '#5a96f5',
            }
        },
    },
    plugins: [],
}
