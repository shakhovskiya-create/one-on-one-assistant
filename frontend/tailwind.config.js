/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // EKF Brand Colors
        ekf: {
          orange: '#E63312',
          'orange-dark': '#C42A0E',
          'orange-light': '#FF4D2A',
          dark: '#1A1A1A',
          gray: '#666666',
          'gray-light': '#999999',
          light: '#F5F5F5',
          white: '#FFFFFF',
        },
        // Semantic colors using EKF palette
        primary: {
          50: '#FEF2F0',
          100: '#FDE6E2',
          200: '#FBCCC5',
          300: '#F9A699',
          400: '#F5715C',
          500: '#E63312',
          600: '#C42A0E',
          700: '#A3230B',
          800: '#821C09',
          900: '#611507',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
