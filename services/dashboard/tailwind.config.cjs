/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ['**/*.qtpl', 'style.css'],
	theme: {
		extend: {
			spacing: {
				sidebar: '16rem',
				header: '4rem'
			},
			colors: {
				primary: {
					50: '#E9FBF8',
					100: '#CFF7F1',
					200: '#A3F0E3',
					300: '#72E9D5',
					400: '#46E2C8',
					500: '#21CFB3',
					600: '#1BA790',
					700: '#147B6A',
					800: '#0D5448',
					900: '#062822'
				},
				secondary: {
					50: '#eaecef',
					100: '#d5dade',
					200: '#acb5be',
					300: '#828f9d',
					400: '#596a7d',
					500: '#2f455c',
					600: '#26374a',
					700: '#1c2937',
					800: '#131c25',
					900: '#090e12'
				}
			}
		}
	},
	plugins: []
};
