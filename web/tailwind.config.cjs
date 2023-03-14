/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#EB5424',
        red: '#EB5424',
        green: '#4BAB4E',
        blue: '#2454BB',
        yellow: '#FFF535',
        text: '#313638',
        'gray-50': '#FAF8F3',
        'gray-100': '#F0EFEA',
        'gray-200': '#E0DFD5',
        'gray-300': '#CBCAC2',
        'gray-400': '#B5B5AE',
        'gray-500': '#AAABA5',
        'gray-600': '#9FA09B',
        'gray-700': '#898B87',
        'gray-800': '#5D6160',
        'gray-900': '#313638',
      },
    },
  },
  plugins: [],
};
