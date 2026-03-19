const config = {
    plugins: {
        '@tailwindcss/postcss': {
            config: './tailwind.config.js',
            content: ['./components/**/*.{html,js,templ}', './pages/**/*.{html,js,templ}'],
        },
    },
};
export default config;