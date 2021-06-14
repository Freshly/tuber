const WindiCSSWebpackPlugin = require('windicss-webpack-plugin').default

const TUBER_PREFIX = process.env.TUBER_PREFIX || '/tuber'

module.exports = {
	webpack: config => {
		config.plugins.push(new WindiCSSWebpackPlugin())
		return config
	},

	trailingSlash: true,
	basePath:      TUBER_PREFIX,
	assetPrefix:   TUBER_PREFIX,
	env:           { TUBER_PREFIX },
	publicRuntimeConfig: {
		basePath: TUBER_PREFIX,
	},
}
