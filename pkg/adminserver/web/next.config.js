const WindiCSSWebpackPlugin = require('windicss-webpack-plugin').default

module.exports = {
	webpack: config => {
		config.plugins.push(new WindiCSSWebpackPlugin({
			scan: {
				dirs:    ['./', './pages'],
				exclude: ['node_modules', '.git', '.next/**/*'],
			},
		}))

		return config
	},

	basePath:      process.env.TUBER_PREFIX || '/tuber',
	trailingSlash: true,
}
