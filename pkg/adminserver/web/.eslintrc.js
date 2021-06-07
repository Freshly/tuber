module.exports = {
	'env': {
		'browser':  true,
		'es2020':   true,
		'commonjs': true,
	},
	'extends': [
		'eslint:recommended',
		'plugin:react/recommended',
	],
	globals:         { process: 'readonly' },
	parser:          '@typescript-eslint/parser',
	'parserOptions': {
		'ecmaFeatures': { 'jsx': true },
		'ecmaVersion':  11,
		'sourceType':   'module',
	},
	'plugins': [
		'react',
		'@typescript-eslint',
	],
	'rules': {
		'comma-dangle':              ['error', 'always-multiline'],
		'indent':                    'off',
		'@typescript-eslint/indent': ['error', 'tab'],
		'linebreak-style':           ['error', 'unix'],
		'quotes':                    ['error', 'single' ],
		'semi':                      ['error', 'never' ],
		'object-curly-newline':      ['error', { 'multiline': true }],
		'block-spacing':             ['error', 'always'],
		'object-curly-spacing':      ['error', 'always'],
		'key-spacing':               ['error', { align: 'value' }],
		'comma-spacing':             ['error'],
		'react/prop-types':          'off',
	},
}
