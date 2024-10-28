/** @type {import('tailwindcss').Config} */
module.exports = {
    // content: [],
    // content: ["/css/**/*.{html,js}"],
    // content: ["/templates/*"],
    content: ["../pkg/routes/templates/**/*.{html,tmpl}"], //relative to where tailwind is executed
    safelist: [
        // 'bg-red-500',
        // 'text-3xl',
        // 'lg:text-4xl',
        // {
        //   pattern: /([a-zA-Z]+)-./, // all of tailwind
        // },
    ],
    darkMode: 'class',
    theme: {
        extend: {},
    },
    plugins: [],
}
